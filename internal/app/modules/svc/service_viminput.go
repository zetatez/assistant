package svc

import (
	"fmt"
	"os"
	"strings"

	"assistant/internal/bootstrap/psl"
	"assistant/pkg/utils"
)

// VimEditInput toggles the vim compose popup for the currently focused
// input box: if a popup is already running it is killed (abandoning the
// current input), otherwise a floating centered vim window is opened
// (insert mode). On save and exit, the text is typed into the previously
// focused window.
func (s *Service) VimEditInput() error {
	// Toggle off: a popup is already up -> kill the popup window/process to
	// abandon the current input. No synthetic keyboard events are injected
	// (typing :q! through xdotool can leak into the focused window/desktop).
	if wid, ok := runningVimPopup(); ok {
		_, _, _ = utils.RunScript("bash", "xdotool windowkill "+wid+" 2>/dev/null")
		// Belt and suspenders: also reap the terminal/vim processes behind the
		// popup. The [v] bracket trick avoids matching this shell itself.
		_, _, _ = utils.RunScript("bash", "pkill -f '[v]im-edit-input' 2>/dev/null || true")
		return nil
	}

	winID, err := activeWindowID()
	if err != nil {
		return fmt.Errorf("get active window: %w", err)
	}

	term := psl.GetConfig().Settings.DefaultTerminal
	if term == "" {
		term = "st"
	}

	f, err := os.CreateTemp("", "vim-input-*.txt")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := f.Name()
	f.Close()
	defer os.Remove(tmpPath)

	// Block until the user finishes editing (vim exits, then st closes).
	// The popup is launched floating and centered: st uses the shared
	// WM_CLASS "float" (same as other floating popups, e.g. tty-clock), and
	// xdotool recenters/sizes it once mapped.
	popup := fmt.Sprintf(`
%s -c float -t vim-edit-input -e nvim +startinsert %s &
STPID=$!
# Fast-relocate: poll tightly, then size/move the window as soon as it is
# actually mapped so the popup appears promptly (and centered).
for i in $(seq 1 150); do
  WID=$(xdotool search --name vim-edit-input 2>/dev/null | head -1)
  if [ -n "$WID" ] && xwininfo -id "$WID" -stats 2>/dev/null | grep -q "Map State: IsViewable"; then
    break
  fi
  sleep 0.02
done
if [ -n "$WID" ]; then
  eval "$(xdotool getdisplaygeometry | awk '{print "SW="$1" SH="$2}')"
  WW=$((SW*50/100)); WH=$((SH*50/100))
  xdotool windowsize "$WID" "$WW" "$WH"
  xdotool windowmove "$WID" $(((SW-WW)/2)) $(((SH-WH)/2))
  # dwm focuses by the pointer: move the mouse onto the popup first, then set
  # X focus. No synthetic click (clicking activates nvim's mouse mode and
  # paints a crosshair cursor).
  xdotool mousemove --window "$WID" $((WW/2)) $((WH/2)) 2>/dev/null
  sleep 0.1
  xdotool windowfocus --sync "$WID" 2>/dev/null
fi
wait "$STPID"`, term, shquote(tmpPath))
	if _, _, err := utils.RunScript("bash", popup); err != nil {
		return fmt.Errorf("vim popup: %w", err)
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("read composed text: %w", err)
	}
	text := strings.TrimRight(string(content), "\n")
	if text == "" {
		return nil
	}

	// Refocus the original window before pasting: move the pointer onto it,
	// then set X focus (no click).
	_, _, _ = utils.RunScript("bash", fmt.Sprintf(
		"xdotool mousemove --window %s 30 30 2>/dev/null; sleep 0.1; xdotool windowfocus --sync %s 2>/dev/null", winID, winID))
	_, _, _ = utils.RunScript("bash", "sleep 0.15")

	return s.typeIntoWindow(winID, text)
}

func activeWindowID() (string, error) {
	out, _, err := utils.RunScript("bash", "xdotool getactivewindow 2>/dev/null")
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(out)
	if id == "" {
		return "", fmt.Errorf("no active window")
	}
	return id, nil
}

// runningVimPopup returns the window id of an already-open vim popup.
func runningVimPopup() (string, bool) {
	out, _, err := utils.RunScript("bash", "xdotool search --title vim-edit-input 2>/dev/null | head -1")
	if err != nil {
		return "", false
	}
	wid := strings.TrimSpace(out)
	if wid == "" {
		return "", false
	}
	return wid, true
}

func (s *Service) typeIntoWindow(winID, text string) error {
	// Paste via clipboard + Ctrl+V. Terminals treat Ctrl+V as literal input,
	// so use Ctrl+Shift+V there; everything else gets the standard Ctrl+V.
	pasteKey := "ctrl+v"
	if cls := windowClassName(winID); isTerminalClass(cls) {
		pasteKey = "ctrl+shift+v"
	}
	cmd := fmt.Sprintf("printf '%%s' %s | xclip -selection clipboard && xdotool key --clearmodifiers %s", shquote(text), pasteKey)
	if _, _, err := utils.RunScript("bash", cmd); err != nil {
		return fmt.Errorf("paste text: %w", err)
	}
	return nil
}

func windowClassName(winID string) string {
	out, _, err := utils.RunScript("bash", "xdotool getwindowclassname "+winID+" 2>/dev/null")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func isTerminalClass(cls string) bool {
	switch strings.ToLower(strings.TrimSpace(cls)) {
	case "st", "st-256color", "kitty", "alacritty", "urxvt", "urxvt-unicode",
		"xterm", "xterm-256color", "gnome-terminal", "konsole", "xfce4-terminal",
		"terminator", "wezterm", "foot", "tilix", "ghostty":
		return true
	}
	return false
}

func shquote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
