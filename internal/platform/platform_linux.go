//go:build linux

package platform

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"assistant/pkg/utils"
)

const linuxNotifyURL = "http://127.0.0.1:8765/notify"

type linuxPlatform struct{}

func newPlatform() Platform { return linuxPlatform{} }

func (linuxPlatform) GOOS() string { return "linux" }

// ── notifications ────────────────────────────────────────────────

type notifyBody struct {
	Msg        string `json:"msg"`
	TTLSeconds int    `json:"ttl_seconds"`
}

func (linuxPlatform) Notify(msg string, ttl time.Duration) error {
	return notifyRequest(http.MethodPut, msg, ttl)
}

func (linuxPlatform) NotifyQueue(msg string, ttl time.Duration) error {
	return notifyRequest(http.MethodPost, msg, ttl)
}

func (linuxPlatform) NotifyClear() error {
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodDelete, linuxNotifyURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func notifyRequest(method, msg string, ttl time.Duration) error {
	buf, err := json.Marshal(notifyBody{Msg: msg, TTLSeconds: normalizeTTL(ttl)})
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(method, linuxNotifyURL, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// ── interactive UI ───────────────────────────────────────────────

func (linuxPlatform) Select(prompt string, options []string) (string, error) {
	if _, err := exec.LookPath("rofi"); err != nil {
		return "", fmt.Errorf("rofi not found: %w", err)
	}
	listFile, err := writeTempList("assistant-select-*", options)
	if err != nil {
		return "", fmt.Errorf("write option list: %w", err)
	}
	defer os.Remove(listFile)
	script := fmt.Sprintf("rofi -dmenu -p %s < %s", shellQuote(prompt), shellQuote(listFile))
	out, _, err := utils.RunScript("bash", script)
	if err != nil {
		if strings.TrimSpace(out) == "" {
			return "", nil // cancelled
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (linuxPlatform) Input(prompt string, secret bool) (string, error) {
	if _, err := exec.LookPath("rofi"); err != nil {
		return "", fmt.Errorf("rofi not found: %w", err)
	}
	flag := ""
	if secret {
		flag = " -password"
	}
	script := fmt.Sprintf("rofi -dmenu%s -p %s < /dev/null", flag, shellQuote(prompt))
	out, _, err := utils.RunScript("bash", script)
	if err != nil {
		if strings.TrimSpace(out) == "" {
			return "", nil // cancelled
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ── canned actions ───────────────────────────────────────────────

var linuxIntents = map[Intent]string{
	VolumeUp:     "amixer set Master unmute && amixer set Master 5%+",
	VolumeDown:   "amixer set Master unmute && amixer set Master 5%-",
	VolumeToggle: "amixer set Master toggle",
	VolumeSet:    "amixer set Master {0}%",
	AudioReset: `has() { amixer scontrols | grep -q "'$1'"; }
has Master && amixer set Master unmute && amixer set Master 80%
has Speaker && amixer set Speaker unmute && amixer set Speaker 64%
has Capture && amixer set Capture cap && amixer set Capture 64%
has Headphone && amixer set Headphone 64%
true`,
	MicUp:          "amixer set Capture 5%+",
	MicDown:        "amixer set Capture 5%-",
	MicToggle:      "amixer set Capture toggle",
	MicSet:         "amixer set Capture {0}%",
	BrightnessUp:   "sudo light -A 1",
	BrightnessDown: "sudo light -N 1 && sudo light -U 1",
	BrightnessSet:  "sudo light -S {0}",
	ScreenOff:      "sleep .5; xset dpms force off",
	Suspend:        "systemctl suspend",
	Reboot:         "systemctl reboot",
	Shutdown:       "systemctl poweroff",
	Lock:           "slock",
	KeyRepeatSet:   "xset r rate {0} {1}",
}

func (linuxPlatform) Run(i Intent, args ...string) error {
	return runIntent(linuxIntents, i, args)
}

func (linuxPlatform) KeyboardLightToggle(path string) (string, error) {
	if cmd, ok := overrideCommand(KeyboardLightToggle, nil); ok {
		if err := runShell(cmd); err != nil {
			return "", err
		}
		return "toggled", nil
	}
	if path == "" {
		return "", fmt.Errorf("%w: keyboard backlight path not configured", ErrUnsupported)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read keyboard brightness: %w", err)
	}
	current := strings.TrimSpace(string(data))
	newVal := "1"
	if current == "1" {
		newVal = "0"
	}
	if _, _, err := utils.RunScript("bash", fmt.Sprintf("sudo sh -c 'echo %s > %s'", newVal, path)); err != nil {
		return "", fmt.Errorf("set keyboard brightness: %w", err)
	}
	return newVal, nil
}

// ── terminal ─────────────────────────────────────────────────────

func (linuxPlatform) TerminalCommand(term, script string) (string, error) {
	if strings.TrimSpace(term) == "" {
		term = "st"
	}
	return fmt.Sprintf("%s -e bash -c %s", term, shellQuote(script)), nil
}

func (p linuxPlatform) FloatingTerminalCommand(term, script string) (string, error) {
	if strings.TrimSpace(term) == "" {
		term = "st"
	}
	sw, sh := p.ScreenSize()
	w, h := 53, 8
	x := int(float64(sw) * 0.72)
	y := int(float64(sh) * 0.04)
	return fmt.Sprintf("%s -g %dx%d+%d+%d -t float -c float -e bash -c %s",
		term, w, h, x, y, shellQuote(script)), nil
}

func (linuxPlatform) ScratchpadTerminalCommand(term, script string) (string, error) {
	if strings.TrimSpace(term) == "" {
		term = "st"
	}
	return fmt.Sprintf("%s -t scratchpad -c scratchpad -e bash -c %s",
		term, shellQuote(script)), nil
}

// ── capture ──────────────────────────────────────────────────────

func (linuxPlatform) ScreenSize() (int, int) {
	out, _, err := utils.RunScript("bash", "xdpyinfo|awk '/dimensions/{split($2,a,\"x\");print a[1],a[2]}'")
	if err != nil {
		return 1920, 1080
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) < 2 {
		return 1920, 1080
	}
	w, _ := strconv.Atoi(parts[0])
	h, _ := strconv.Atoi(parts[1])
	if w <= 0 || h <= 0 {
		return 1920, 1080
	}
	return w, h
}

func (linuxPlatform) Screenshot() ([]byte, error) {
	if _, err := exec.LookPath("flameshot"); err != nil {
		return nil, fmt.Errorf("screenshot tool not found: flameshot")
	}
	cmd := exec.Command("flameshot", "full", "--raw")
	cmd.Env = append(os.Environ(), "DISPLAY=:0")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}
	return out, nil
}

func (linuxPlatform) CaptureScreenArgs(out string, w, h int, cfg CaptureConfig) []string {
	display := cfg.ScreenDevice
	if display == "" {
		display = os.Getenv("DISPLAY")
	}
	audio := cfg.AudioDevice
	if audio == "" {
		audio = "default"
	}
	return []string{ "ffmpeg", "-y", "-s", fmt.Sprintf("%dx%d", w, h), "-r", "60", "-f", "x11grab", "-i", display, "-f", "alsa", "-i", audio, "-c:v", "libx264rgb", "-crf", "0", "-preset", "ultrafast", "-color_range", "2", "-c:a", "aac", out }
}

func (linuxPlatform) CaptureAudioArgs(out string, cfg CaptureConfig) []string {
	audio := cfg.AudioDevice
	if audio == "" {
		audio = "default"
	}
	return []string{"ffmpeg", "-y", "-r", "60", "-f", "alsa", "-i", audio, "-c:a", "flac", out}
}

func (linuxPlatform) CaptureWebcamArgs(out string, cfg CaptureConfig) []string {
	audio := cfg.AudioDevice
	if audio == "" {
		audio = "default"
	}
	video := cfg.VideoDevice
	if video == "" {
		video = "/dev/video0"
	}
	return []string{ "ffmpeg", "-f", "pulse", "-ac", "2", "-i", audio, "-f", "v4l2", "-i", video, "-t", "00:00:20", "-vcodec", "libx264", out }
}

func (linuxPlatform) PreviewWebcamCmd(cfg CaptureConfig) []string {
	video := cfg.VideoDevice
	if video == "" {
		video = "/dev/video0"
	}
	return []string{"ffplay", "-loglevel", "quiet", "-framedrop", "-fast", "-alwaysontop", "-i", video}
}

// ── desktop ──────────────────────────────────────────────────────

func (linuxPlatform) SetWallpaper(path string) error {
	if _, err := exec.LookPath("feh"); err != nil {
		return fmt.Errorf("feh not found: %w", err)
	}
	return runShell(fmt.Sprintf("feh --bg-fill %s", shellQuote(path)))
}

func (linuxPlatform) DefaultInterfaceName() string {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	bestIface := ""
	bestMetric := uint32(^uint32(0))
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[1] != "00000000" || len(fields[1]) != 8 {
			continue
		}
		var metric uint32
		if len(fields) >= 7 {
			fmt.Sscanf(fields[6], "%d", &metric)
		}
		if metric < bestMetric {
			bestMetric = metric
			bestIface = fields[0]
		}
	}
	return bestIface
}

func (linuxPlatform) OpenURL(url string) error {
	if _, err := exec.LookPath("xdg-open"); err != nil {
		return fmt.Errorf("xdg-open not found: %w", err)
	}
	return utils.StartScript("bash", fmt.Sprintf("xdg-open %s", shellQuote(url)))
}

func (linuxPlatform) OpenImages(dir string) error {
	if dir == "" {
		dir = "."
	}
	script := fmt.Sprintf(`fd --type f --extension jpg --extension jpeg --extension png --extension gif --extension bmp --extension tiff --exclude repos . %s | sxiv -ftio`, shellQuote(dir))
	return utils.StartScript("bash", script)
}

func (linuxPlatform) OpenBrowser(browser, url, proxy string, asApp bool) error {
	switch browser {
	case "chrome":
		if asApp {
			return utils.StartScript("bash", fmt.Sprintf("chrome --proxy-server=%s --app=%s", shellQuote(proxy), shellQuote(url)))
		}
		return utils.StartScript("bash", fmt.Sprintf("chrome --proxy-server=%s %s", shellQuote(proxy), shellQuote(url)))
	case "qutebrowser":
		if asApp {
			return utils.StartScript("bash", fmt.Sprintf("qutebrowser --set content.proxy %s --target window %s", shellQuote(proxy), shellQuote(url)))
		}
		return utils.StartScript("bash", fmt.Sprintf("qutebrowser --set content.proxy %s %s", shellQuote(proxy), shellQuote(url)))
	default:
		return linuxPlatform{}.OpenURL(url)
	}
}

// ── network ──────────────────────────────────────────────────────

func (linuxPlatform) WifiNetworks() ([]string, error) {
	if _, err := exec.LookPath("nmcli"); err != nil {
		return nil, fmt.Errorf("nmcli not found: %w", err)
	}
	out, stderr, err := utils.RunScript("bash", "nmcli -t -f SSID device wifi list")
	if err != nil {
		return nil, fmt.Errorf("list wifi: %s", strings.TrimSpace(stderr))
	}
	seen := make(map[string]struct{})
	var networks []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(strings.ToLower(line), "ssid") {
			continue
		}
		ssid := line
		if strings.Contains(line, "\\:") {
			ssid = strings.ReplaceAll(line, "\\:", ":")
		}
		ssid = strings.TrimSpace(ssid)
		if ssid == "" {
			continue
		}
		if _, ok := seen[ssid]; ok {
			continue
		}
		seen[ssid] = struct{}{}
		networks = append(networks, ssid)
	}
	return networks, nil
}

func (linuxPlatform) WifiConnect(ssid, password string) error {
	if _, err := exec.LookPath("nmcli"); err != nil {
		return fmt.Errorf("nmcli not found: %w", err)
	}
	cmd := fmt.Sprintf("nmcli device wifi connect %s password %s", shellQuote(ssid), shellQuote(password))
	_, stderr, err := utils.RunScript("bash", cmd)
	if err != nil {
		return fmt.Errorf("connect wifi: %s", strings.TrimSpace(stderr))
	}
	return nil
}

// ── bluetooth ────────────────────────────────────────────────────

func (linuxPlatform) BluetoothPaired() ([]string, error) {
	out, _, err := utils.RunScript("bash", "bluetoothctl devices")
	if err != nil {
		return nil, fmt.Errorf("list bluetooth devices: %w", err)
	}
	return parseBluetoothctlDevices(out), nil
}

func (linuxPlatform) BluetoothConnected() ([]string, error) {
	out, _, err := utils.RunScript("bash", "bluetoothctl devices Connected")
	if err == nil && strings.TrimSpace(out) != "" {
		return parseBluetoothctlDevices(out), nil
	}
	out, _, err = utils.RunScript("bash", "bluetoothctl info")
	if err != nil {
		return nil, fmt.Errorf("list connected bluetooth devices: %w", err)
	}
	return parseBluetoothctlDevices(out), nil
}

func parseBluetoothctlDevices(out string) []string {
	var devices []string
	re := regexp.MustCompile(`Device\s+([0-9A-Fa-f:]{17})\s+(.+)$`)
	for _, line := range strings.Split(out, "\n") {
		if m := re.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			devices = append(devices, fmt.Sprintf("%s %s", m[1], strings.TrimSpace(m[2])))
		}
	}
	return devices
}

func (linuxPlatform) BluetoothScan() ([]string, error) {
	if _, err := exec.LookPath("bluetoothctl"); err != nil {
		return nil, fmt.Errorf("bluetoothctl not found: %w", err)
	}
	cmd := exec.Command("bluetoothctl")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("bluetooth stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("bluetooth stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start bluetoothctl: %w", err)
	}
	if _, err := stdin.Write([]byte("scan on\n")); err != nil {
		return nil, fmt.Errorf("write scan on: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	found := make(map[string]string)
	re := regexp.MustCompile(`Device\s+([0-9A-Fa-f:]{17})\s+(.+)$`)

	timer := time.After(6 * time.Second)
	scanDone := make(chan struct{})
	go func() {
		for scanner.Scan() {
			if m := re.FindStringSubmatch(strings.TrimSpace(scanner.Text())); m != nil {
				found[m[1]] = strings.TrimSpace(m[2])
			}
		}
		close(scanDone)
	}()

	select {
	case <-timer:
	case <-scanDone:
	}

	_, _ = stdin.Write([]byte("scan off\nexit\n"))
	_ = cmd.Wait()

	var list []string
	for mac, name := range found {
		list = append(list, fmt.Sprintf("%s %s", mac, name))
	}
	return list, nil
}

func (linuxPlatform) BluetoothConnect(mac string) error {
	for _, subcmd := range []string{
		fmt.Sprintf("bluetoothctl pair %s", mac),
		fmt.Sprintf("bluetoothctl trust %s", mac),
		fmt.Sprintf("bluetoothctl connect %s", mac),
	} {
		_, stderr, err := utils.RunScript("bash", subcmd)
		if err != nil {
			return fmt.Errorf("%s failed: %s", subcmd, strings.TrimSpace(stderr))
		}
	}
	return nil
}

func (linuxPlatform) BluetoothDisconnect(mac string) error {
	_, stderr, err := utils.RunScript("bash", fmt.Sprintf("bluetoothctl disconnect %s", mac))
	if err != nil {
		return fmt.Errorf("disconnect failed: %s", strings.TrimSpace(stderr))
	}
	return nil
}

// ── display ──────────────────────────────────────────────────────

func (linuxPlatform) DisplayLayouts(primary string) ([]string, error) {
	stdout, _, err := utils.RunScript("bash", fmt.Sprintf("xrandr|grep ' connected'|grep -v %s|awk '{print $1}'", shellQuote(primary)))
	if err != nil {
		return nil, fmt.Errorf("detect monitors: %w", err)
	}
	if strings.TrimSpace(stdout) == "" {
		return nil, fmt.Errorf("only one monitor detected")
	}
	return []string{
		"default",
		"clone",
		"primary_only",
		"second_only",
		"left_of",
		"right_of",
		"above",
		"below",
		"rotate_left_left_of",
		"rotate_right_left_of",
		"rotate_left_right_of",
		"rotate_right_right_of",
	}, nil
}

func (linuxPlatform) ApplyDisplayLayout(layout, primary, wallpaperDir string) error {
	stdout, _, err := utils.RunScript("bash", fmt.Sprintf("xrandr|grep ' connected'|grep -v %s|awk '{print $1}'", shellQuote(primary)))
	if err != nil {
		return fmt.Errorf("detect monitors: %w", err)
	}
	second := strings.TrimSpace(stdout)
	if second == "" {
		return fmt.Errorf("only one monitor detected")
	}
	layouts := map[string]string{
		"default":               fmt.Sprintf("xrandr --output %s --auto --output %s --off", primary, second),
		"clone":                 fmt.Sprintf("xrandr --output %s --mode 1920x1080", second),
		"primary_only":          fmt.Sprintf("xrandr --output %s --auto --output %s --off", primary, second),
		"second_only":           fmt.Sprintf("xrandr --output %s --auto --output %s --off", second, primary),
		"left_of":               fmt.Sprintf("xrandr --output %s --auto --left-of %s --auto", second, primary),
		"right_of":              fmt.Sprintf("xrandr --output %s --auto --right-of %s --auto", second, primary),
		"above":                 fmt.Sprintf("xrandr --output %s --auto --above %s --auto", second, primary),
		"below":                 fmt.Sprintf("xrandr --output %s --auto --below %s --auto", second, primary),
		"rotate_left_left_of":   fmt.Sprintf("xrandr --output %s --auto --rotate left --left-of %s --auto", second, primary),
		"rotate_right_left_of":  fmt.Sprintf("xrandr --output %s --auto --rotate right --left-of %s --auto", second, primary),
		"rotate_left_right_of":  fmt.Sprintf("xrandr --output %s --auto --rotate left --right-of %s --auto", second, primary),
		"rotate_right_right_of": fmt.Sprintf("xrandr --output %s --auto --rotate right --right-of %s --auto", second, primary),
	}
	cmd, ok := layouts[layout]
	if !ok {
		return fmt.Errorf("unknown layout: %s", layout)
	}
	if _, stderr, err := utils.RunScript("bash", cmd); err != nil {
		return fmt.Errorf("set display: %s", strings.TrimSpace(stderr))
	}
	return linuxPlatform{}.SetWallpaper(wallpaperDir)
}

// ── ssh ──────────────────────────────────────────────────────────

func (linuxPlatform) SSHConnect(term, user, host, password string, port int) error {
	script := fmt.Sprintf("sshpass -p %s ssh -o 'StrictHostKeyChecking no' -p %d %s@%s",
		shellQuote(password), port, user, host)
	cmd, err := linuxPlatform{}.TerminalCommand(term, script)
	if err != nil {
		return err
	}
	return utils.StartScript("bash", cmd)
}

// ── popup text editing ──────────────────────────────────────────

func (linuxPlatform) VimEditInput(term string) error {
	if wid, ok := runningVimPopup(); ok {
		_, _, _ = utils.RunScript("bash", "xdotool windowkill "+wid+" 2>/dev/null")
		_, _, _ = utils.RunScript("bash", "pkill -f '[v]im-edit-input' 2>/dev/null || true")
		return nil
	}

	winID, err := activeWindowID()
	if err != nil {
		return fmt.Errorf("get active window: %w", err)
	}
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

	popup := fmt.Sprintf(`
%s -c float -t vim-edit-input -e nvim +startinsert %s &
STPID=$!
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
  xdotool mousemove --window "$WID" $((WW/2)) $((WH/2)) 2>/dev/null
  sleep 0.1
  xdotool windowfocus --sync "$WID" 2>/dev/null
fi
wait "$STPID"`, term, shellQuote(tmpPath))
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

	_, _, _ = utils.RunScript("bash", fmt.Sprintf(
		"xdotool mousemove --window %s 30 30 2>/dev/null; sleep 0.1; xdotool windowfocus --sync %s 2>/dev/null", winID, winID))
	_, _, _ = utils.RunScript("bash", "sleep 0.15")

	return typeIntoWindow(winID, text)
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

func typeIntoWindow(winID, text string) error {
	pasteKey := "ctrl+v"
	if cls := windowClassName(winID); isTerminalClass(cls) {
		pasteKey = "ctrl+shift+v"
	}
	cmd := fmt.Sprintf("printf '%%s' %s | xclip -selection clipboard && xdotool key --clearmodifiers %s", shellQuote(text), pasteKey)
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
