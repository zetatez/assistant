//go:build darwin

package platform

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"assistant/pkg/utils"
)

type darwinPlatform struct{}

func newPlatform() Platform { return darwinPlatform{} }

func (darwinPlatform) GOOS() string { return "darwin" }

// runAppleScript executes a script and returns its trimmed stdout. stderr is
// folded into the error so callers can recognise AppleScript codes such as
// -128 ("User canceled").
func runAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	var out, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errBuf
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(errBuf.String()); msg != "" {
			return "", fmt.Errorf("%w: %s", err, msg)
		}
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// ── notifications ────────────────────────────────────────────────

func (darwinPlatform) Notify(msg string, ttl time.Duration) error {
	return darwinNotify(msg)
}

func (darwinPlatform) NotifyQueue(msg string, ttl time.Duration) error {
	return darwinNotify(msg)
}

// NotifyClear is a no-op: macOS notifications are one-shot, there is no queue.
func (darwinPlatform) NotifyClear() error { return nil }

func darwinNotify(msg string) error {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return nil
	}
	if path, err := exec.LookPath("terminal-notifier"); err == nil {
		cmd := exec.Command(path, "-title", "Assistant", "-message", msg)
		return cmd.Run()
	}
	script := fmt.Sprintf(`display notification "%s" with title "Assistant"`, appleQuote(msg))
	_, err := runAppleScript(script)
	return err
}

// ── interactive UI ───────────────────────────────────────────────

func (darwinPlatform) Select(prompt string, options []string) (string, error) {
	if len(options) == 0 {
		return "", nil
	}
	script := fmt.Sprintf(`set picked to choose from list %s with title "Assistant" with prompt "%s"
if picked is false then
	return ""
else
	return item 1 of picked
end if`, appleList(options), appleQuote(prompt))
	out, err := runAppleScript(script)
	if err != nil {
		// User cancellation surfaces as AppleScript error -128.
		if strings.Contains(err.Error(), "-128") {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (darwinPlatform) Input(prompt string, secret bool) (string, error) {
	hidden := ""
	if secret {
		hidden = " with hidden answer"
	}
	script := fmt.Sprintf(`set res to display dialog "%s" default answer ""%s with title "Assistant"
return text returned of res`, appleQuote(prompt), hidden)
	out, err := runAppleScript(script)
	if err != nil {
		if strings.Contains(err.Error(), "-128") {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ── canned actions ───────────────────────────────────────────────

var darwinIntents = map[Intent]string{
	VolumeUp:     `osascript -e 'set volume output volume ((output volume of (get volume settings)) + 5)'`,
	VolumeDown:   `osascript -e 'set volume output volume ((output volume of (get volume settings)) - 5)'`,
	VolumeToggle: `osascript -e 'set volume output muted not (output muted of (get volume settings))'`,
	VolumeSet:    `osascript -e 'set volume output volume {0}'`,
	AudioReset:   `osascript -e 'set volume output muted false' -e 'set volume output volume 80' -e 'try' -e 'set volume input volume 64' -e 'end try'`,
	MicUp:          `osascript -e 'set volume input volume ((input volume of (get volume settings)) + 5)'`,
	MicDown:        `osascript -e 'set volume input volume ((input volume of (get volume settings)) - 5)'`,
	MicToggle:      `osascript -e 'if input volume of (get volume settings) is 0 then' -e 'set volume input volume 64' -e 'else' -e 'set volume input volume 0' -e 'end if'`,
	MicSet:         `osascript -e 'set volume input volume {0}'`,
	BrightnessUp:   `osascript -e 'tell application "System Events" to key code 144'`,
	BrightnessDown: `osascript -e 'tell application "System Events" to key code 145'`,
	ScreenOff:      `pmset displaysleepnow`,
	Suspend:        `pmset sleepnow`,
	Reboot:         `osascript -e 'tell application "System Events" to restart'`,
	Shutdown:       `osascript -e 'tell application "System Events" to shut down'`,
	Lock:           `osascript -e 'tell application "System Events" to keystroke "q" using {control down, command down}'`,
}

func (darwinPlatform) Run(i Intent, args ...string) error {
	return runIntent(darwinIntents, i, args)
}

func (darwinPlatform) KeyboardLightToggle(path string) (string, error) {
	if cmd, ok := overrideCommand(KeyboardLightToggle, nil); ok {
		if err := runShell(cmd); err != nil {
			return "", err
		}
		return "toggled", nil
	}
	return "", fmt.Errorf("%w: keyboard backlight control", ErrUnsupported)
}

// ── terminal ─────────────────────────────────────────────────────

func (darwinPlatform) TerminalCommand(term, script string) (string, error) {
	switch strings.TrimSpace(term) {
	case "", "Terminal", "terminal", "Apple_Terminal":
		// Block until the command finishes when the caller uses RunScript,
		// while still working as fire-and-forget via StartScript. This keeps
		// the "read the result afterwards" endpoints (e.g. snip-fzf) correct.
		as := fmt.Sprintf(`tell application "Terminal"
	do script "%s"
	activate
end tell
delay 0.3
tell application "Terminal"
	repeat while busy of front window
		delay 0.2
	end repeat
end tell`, appleQuote(script))
		return fmt.Sprintf("osascript -e %s", shellQuote(as)), nil
	case "iTerm", "iterm", "iTerm2":
		as := fmt.Sprintf(`tell application "iTerm" to create window with default profile command "%s"`, appleQuote(script))
		return fmt.Sprintf("osascript -e %s", shellQuote(as)), nil
	default:
		// kitty/alacritty/wezterm and friends accept `-e <command>`.
		return fmt.Sprintf("%s -e bash -c %s", term, shellQuote(script)), nil
	}
}

func (p darwinPlatform) FloatingTerminalCommand(term, script string) (string, error) {
	return p.TerminalCommand(term, script)
}

func (p darwinPlatform) ScratchpadTerminalCommand(term, script string) (string, error) {
	return p.TerminalCommand(term, script)
}

// ── capture ──────────────────────────────────────────────────────

func (darwinPlatform) ScreenSize() (int, int) {
	out, err := runAppleScript(`tell application "Finder" to get bounds of window of desktop`)
	if err == nil {
		re := regexp.MustCompile(`-?\d+`)
		nums := re.FindAllString(out, -1)
		if len(nums) >= 4 {
			w, _ := strconv.Atoi(nums[2])
			h, _ := strconv.Atoi(nums[3])
			if w > 0 && h > 0 {
				return w, h
			}
		}
	}
	return 1920, 1080
}

func (darwinPlatform) Screenshot() ([]byte, error) {
	out, err := exec.Command("screencapture", "-x", "-t", "jpeg", "-").Output()
	if err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}
	return out, nil
}

func (darwinPlatform) CaptureScreenArgs(out string, w, h int, cfg CaptureConfig) []string {
	device := cfg.ScreenDevice
	if device == "" {
		device = "1:none"
	}
	return []string{
		"ffmpeg", "-y", "-f", "avfoundation", "-capture_cursor", "1",
		"-i", device, "-c:v", "libx264", "-crf", "0", "-preset", "ultrafast",
		"-pix_fmt", "yuv420p", out,
	}
}

func (darwinPlatform) CaptureAudioArgs(out string, cfg CaptureConfig) []string {
	device := cfg.AudioDevice
	if device == "" {
		device = "0"
	}
	return []string{"ffmpeg", "-y", "-f", "avfoundation", "-i", ":" + device, "-c:a", "flac", out}
}

func (darwinPlatform) CaptureWebcamArgs(out string, cfg CaptureConfig) []string {
	device := cfg.VideoDevice
	if device == "" {
		device = "0:0"
	}
	return []string{
		"ffmpeg", "-y", "-f", "avfoundation", "-i", device,
		"-t", "00:00:20", "-vcodec", "libx264", "-pix_fmt", "yuv420p", out,
	}
}

func (darwinPlatform) PreviewWebcamCmd(cfg CaptureConfig) []string {
	device := cfg.VideoDevice
	if device == "" {
		device = "0"
	}
	return []string{"ffplay", "-loglevel", "quiet", "-framedrop", "-fast", "-f", "avfoundation", "-i", device}
}

// ── desktop ──────────────────────────────────────────────────────

func (darwinPlatform) SetWallpaper(path string) error {
	script := fmt.Sprintf(`tell application "System Events" to set picture of every desktop to "%s"`, appleQuote(path))
	_, err := runAppleScript(script)
	if err != nil {
		return fmt.Errorf("set wallpaper: %w", err)
	}
	return nil
}

func (darwinPlatform) DefaultInterfaceName() string {
	out, err := exec.Command("route", "-n", "get", "default").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) >= 2 && fields[0] == "interface:" {
				return fields[1]
			}
		}
	}
	return "en0"
}

func (darwinPlatform) OpenImages(dir string) error {
	if strings.TrimSpace(dir) == "" {
		dir = "."
	}
	script := fmt.Sprintf(`find %s -type f \( -iname '*.jpg' -o -iname '*.jpeg' -o -iname '*.png' -o -iname '*.gif' -o -iname '*.bmp' -o -iname '*.tiff' \) -print0 | xargs -0 open -a Preview`, shellQuote(dir))
	return utils.StartScript("bash", script)
}

func (darwinPlatform) OpenURL(url string) error {
	return utils.StartScript("bash", fmt.Sprintf("open %s", shellQuote(url)))
}

func (darwinPlatform) OpenBrowser(browser, url, proxy string, asApp bool) error {
	switch browser {
	case "chrome":
		if !dirExists("/Applications/Google Chrome.app") {
			return darwinPlatform{}.OpenURL(url)
		}
		args := []string{"-na", "Google Chrome", "--args"}
		if proxy != "" {
			args = append(args, "--proxy-server="+proxy)
		}
		if asApp {
			args = append(args, "--app="+url)
		} else {
			args = append(args, url)
		}
		return utils.StartScript("bash", "open "+CommandLine(args))
	case "qutebrowser":
		if _, err := exec.LookPath("qutebrowser"); err != nil {
			return darwinPlatform{}.OpenURL(url)
		}
		if asApp {
			return utils.StartScript("bash", fmt.Sprintf("qutebrowser --set content.proxy %s --target window %s", shellQuote(proxy), shellQuote(url)))
		}
		return utils.StartScript("bash", fmt.Sprintf("qutebrowser --set content.proxy %s %s", shellQuote(proxy), shellQuote(url)))
	default:
		return darwinPlatform{}.OpenURL(url)
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ── network ──────────────────────────────────────────────────────

func (darwinPlatform) WifiNetworks() ([]string, error) {
	iface := darwinPlatform{}.wifiInterface()
	out, stderr, err := utils.RunScript("bash", fmt.Sprintf("networksetup -listpreferredwirelessnetworks %s", shellQuote(iface)))
	if err != nil {
		return nil, fmt.Errorf("list wifi: %s", strings.TrimSpace(stderr))
	}
	var networks []string
	for _, line := range strings.Split(out, "\n") {
		ssid := strings.TrimSpace(line)
		if ssid == "" || strings.HasPrefix(ssid, "Preferred networks") {
			continue
		}
		networks = append(networks, ssid)
	}
	return networks, nil
}

func (darwinPlatform) WifiConnect(ssid, password string) error {
	iface := darwinPlatform{}.wifiInterface()
	cmd := fmt.Sprintf("networksetup -setairportnetwork %s %s", shellQuote(iface), shellQuote(ssid))
	if password != "" {
		cmd += " " + shellQuote(password)
	}
	if _, stderr, err := utils.RunScript("bash", cmd); err != nil {
		return fmt.Errorf("connect wifi: %s", strings.TrimSpace(stderr))
	}
	return nil
}

func (darwinPlatform) wifiInterface() string {
	out, _, err := utils.RunScript("bash", "networksetup -listallhardwareports")
	if err != nil {
		return "en0"
	}
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "Hardware Port: Wi-Fi" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			l := strings.TrimSpace(lines[j])
			if strings.HasPrefix(l, "Device: ") {
				return strings.TrimSpace(strings.TrimPrefix(l, "Device: "))
			}
			if strings.HasPrefix(l, "Hardware Port:") {
				break
			}
		}
	}
	return "en0"
}

// ── bluetooth ────────────────────────────────────────────────────

func blueutilPath() (string, error) {
	path, err := exec.LookPath("blueutil")
	if err != nil {
		return "", fmt.Errorf("%w: bluetooth control (brew install blueutil)", ErrUnsupported)
	}
	return path, nil
}

func (darwinPlatform) BluetoothPaired() ([]string, error) {
	path, err := blueutilPath()
	if err != nil {
		return nil, err
	}
	out, stderr, err := utils.RunScript("bash", fmt.Sprintf("%s --paired", shellQuote(path)))
	if err != nil {
		return nil, fmt.Errorf("list bluetooth devices: %s", strings.TrimSpace(stderr))
	}
	return parseBlueutil(out), nil
}

func (darwinPlatform) BluetoothConnected() ([]string, error) {
	path, err := blueutilPath()
	if err != nil {
		return nil, err
	}
	out, stderr, err := utils.RunScript("bash", fmt.Sprintf("%s --connected", shellQuote(path)))
	if err != nil {
		return nil, fmt.Errorf("list connected bluetooth devices: %s", strings.TrimSpace(stderr))
	}
	return parseBlueutil(out), nil
}

func (darwinPlatform) BluetoothScan() ([]string, error) {
	path, err := blueutilPath()
	if err != nil {
		return nil, err
	}
	out, stderr, err := utils.RunScript("bash", fmt.Sprintf("%s --inquiry", shellQuote(path)))
	if err != nil {
		return nil, fmt.Errorf("scan bluetooth: %s", strings.TrimSpace(stderr))
	}
	return parseBlueutil(out), nil
}

var blueutilMACRe = regexp.MustCompile(`(?i)[0-9a-f]{2}([:-][0-9a-f]{2}){5}`)

func parseBlueutil(out string) []string {
	var devices []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		raw := blueutilMACRe.FindString(line)
		if raw == "" {
			continue
		}
		mac := strings.ReplaceAll(raw, "-", ":")
		name := strings.TrimSpace(strings.Replace(line, raw, "", 1))
		name = strings.Trim(name, ",; \t\"'")
		if name == "" {
			devices = append(devices, mac)
			continue
		}
		devices = append(devices, mac+" "+name)
	}
	return devices
}

func (darwinPlatform) BluetoothConnect(mac string) error {
	path, err := blueutilPath()
	if err != nil {
		return err
	}
	if _, stderr, err := utils.RunScript("bash", fmt.Sprintf("%s --connect %s", shellQuote(path), shellQuote(mac))); err != nil {
		return fmt.Errorf("bluetooth connect: %s", strings.TrimSpace(stderr))
	}
	return nil
}

func (darwinPlatform) BluetoothDisconnect(mac string) error {
	path, err := blueutilPath()
	if err != nil {
		return err
	}
	if _, stderr, err := utils.RunScript("bash", fmt.Sprintf("%s --disconnect %s", shellQuote(path), shellQuote(mac))); err != nil {
		return fmt.Errorf("bluetooth disconnect: %s", strings.TrimSpace(stderr))
	}
	return nil
}

// ── display ──────────────────────────────────────────────────────

func (darwinPlatform) DisplayLayouts(primary string) ([]string, error) {
	return nil, fmt.Errorf("%w: display layout switching (install displayplacer or use System Settings)", ErrUnsupported)
}

func (darwinPlatform) ApplyDisplayLayout(layout, primary, wallpaperDir string) error {
	return fmt.Errorf("%w: display layout switching (install displayplacer or use System Settings)", ErrUnsupported)
}

// ── ssh ──────────────────────────────────────────────────────────

func (darwinPlatform) SSHConnect(term, user, host, password string, port int) error {
	script := fmt.Sprintf("ssh -o 'StrictHostKeyChecking no' -p %d %s@%s", port, user, host)
	if _, err := exec.LookPath("sshpass"); err == nil {
		script = fmt.Sprintf("sshpass -p %s %s", shellQuote(password), script)
	}
	cmd, err := darwinPlatform{}.TerminalCommand(term, script)
	if err != nil {
		return err
	}
	return utils.StartScript("bash", cmd)
}

// ── popup text editing ──────────────────────────────────────────

func (darwinPlatform) VimEditInput(term string) error {
	// Toggle off: an editor is already waiting on a vim-input temp file.
	if utils.ProcessRunning("nvim") {
		_, _, _ = utils.RunScript("bash", "pkill -f 'nvim.*vim-input-' 2>/dev/null || true")
		return nil
	}

	frontApp, _ := runAppleScript(`tell application "System Events" to get name of first application process whose frontmost is true`)

	f, err := os.CreateTemp("", "vim-input-*.txt")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := f.Name()
	f.Close()
	defer os.Remove(tmpPath)

	sentinel := tmpPath + ".done"
	defer os.Remove(sentinel)

	script := fmt.Sprintf("nvim +startinsert %s; touch %s", shellQuote(tmpPath), shellQuote(sentinel))
	cmd, err := darwinPlatform{}.TerminalCommand(term, script)
	if err != nil {
		return err
	}
	if err := utils.StartScript("bash", cmd); err != nil {
		return fmt.Errorf("vim popup: %w", err)
	}

	deadline := time.Now().Add(30 * time.Minute)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(sentinel); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if _, err := os.Stat(sentinel); err != nil {
		return fmt.Errorf("vim popup: timed out waiting for editor")
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("read composed text: %w", err)
	}
	text := strings.TrimRight(string(content), "\n")
	if text == "" {
		return nil
	}

	// Put the text on the clipboard and paste it into the app that was
	// focused before the popup opened (requires Accessibility permission).
	_, _, _ = utils.RunScript("bash", fmt.Sprintf("printf '%%s' %s | pbcopy", shellQuote(text)))
	if frontApp != "" {
		_, _ = runAppleScript(fmt.Sprintf(`tell application "%s" to activate`, appleQuote(frontApp)))
		time.Sleep(150 * time.Millisecond)
	}
	_, err = runAppleScript(`tell application "System Events" to keystroke "v" using command down`)
	if err != nil {
		return fmt.Errorf("paste text: %w", err)
	}
	return nil
}
