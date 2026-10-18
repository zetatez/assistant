// Package platform abstracts the OS-specific desktop integrations used by
// assistant. Arch Linux (X11/dwm) and macOS are supported. Every other OS gets
// an unsupported stub so the project still compiles everywhere.
//
// Callers must never shell out to OS tools directly; they go through the
// Platform returned by Get so a single build works on both targets.
package platform

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"assistant/pkg/utils"
)

// ErrUnsupported is returned when a feature has no implementation on the
// current OS. The wrapped message is meant to be shown to the user.
var ErrUnsupported = errors.New("not supported on this OS")

// GOOS values used by the platform selection.
const (
	Linux  = "linux"
	Darwin = "darwin"
)

// Intent names a canned system action whose concrete command differs per OS.
// Users may override any intent's command through config (see SetOverrides).
type Intent string

const (
	VolumeUp     Intent = "volume_up"
	VolumeDown   Intent = "volume_down"
	VolumeToggle Intent = "volume_toggle"
	VolumeSet    Intent = "volume_set"
	// AudioReset restores the default output/mic levels (used by SysReset).
	AudioReset Intent = "audio_reset"

	MicUp     Intent = "mic_up"
	MicDown   Intent = "mic_down"
	MicToggle Intent = "mic_toggle"
	MicSet    Intent = "mic_set"

	BrightnessUp   Intent = "brightness_up"
	BrightnessDown Intent = "brightness_down"
	BrightnessSet  Intent = "brightness_set"

	KeyboardLightToggle Intent = "keyboard_light_toggle"

	ScreenOff Intent = "screen_off"
	Suspend   Intent = "suspend"
	Reboot    Intent = "reboot"
	Shutdown  Intent = "shutdown"
	Lock      Intent = "lock"

	KeyRepeatSet Intent = "key_repeat_set"
)

// AllIntents lists every intent with a built-in command, useful for docs.
var AllIntents = []Intent{
	VolumeUp, VolumeDown, VolumeToggle, VolumeSet, AudioReset,
	MicUp, MicDown, MicToggle, MicSet,
	BrightnessUp, BrightnessDown, BrightnessSet,
	KeyboardLightToggle,
	ScreenOff, Suspend, Reboot, Shutdown, Lock,
	KeyRepeatSet,
}

// CaptureConfig carries the device names used to build ffmpeg command lines.
// Empty fields fall back to sensible per-OS defaults.
type CaptureConfig struct {
	ScreenDevice string
	AudioDevice  string
	VideoDevice  string
}

// Platform is the OS abstraction. Implementations live in platform_linux.go
// and platform_darwin.go and are selected at build time.
type Platform interface {
	// GOOS returns the runtime OS ("linux" or "darwin").
	GOOS() string

	// ── notifications ────────────────────────────────────────────
	// Notify preempts the current message; NotifyQueue enqueues it.
	Notify(msg string, ttl time.Duration) error
	NotifyQueue(msg string, ttl time.Duration) error
	// NotifyClear drops the current message and the queue.
	NotifyClear() error

	// ── interactive UI ───────────────────────────────────────────
	// Select asks the user to pick one of options; returns "" when cancelled.
	Select(prompt string, options []string) (string, error)
	// Input asks the user for free text; secret masks the input.
	Input(prompt string, secret bool) (string, error)

	// ── canned actions ───────────────────────────────────────────
	// Run executes the intent, substituting args into {0}, {1}, ...
	Run(i Intent, args ...string) error
	// KeyboardLightToggle toggles the keyboard backlight and returns the new
	// value ("0"/"1"). path is the sysfs file on Linux.
	KeyboardLightToggle(path string) (string, error)

	// ── terminal ─────────────────────────────────────────────────
	// TerminalCommand wraps script so it runs inside the given terminal.
	TerminalCommand(term, script string) (string, error)
	// FloatingTerminalCommand is like TerminalCommand but asks the window
	// manager for a small floating window when supported (dwm/st on Linux).
	FloatingTerminalCommand(term, script string) (string, error)
	// ScratchpadTerminalCommand opens script in the WM scratchpad window
	// (class "scratchpad" under dwm/st on Linux). Used by long-running jobs
	// such as recordings so they do not hijack a normal floating window.
	ScratchpadTerminalCommand(term, script string) (string, error)

	// ── capture ──────────────────────────────────────────────────
	ScreenSize() (int, int)
	Screenshot() ([]byte, error)
	CaptureScreenArgs(out string, w, h int, cfg CaptureConfig) []string
	CaptureAudioArgs(out string, cfg CaptureConfig) []string
	CaptureWebcamArgs(out string, cfg CaptureConfig) []string
	PreviewWebcamCmd(cfg CaptureConfig) []string

	// ── desktop ──────────────────────────────────────────────────
	SetWallpaper(path string) error
	DefaultInterfaceName() string
	OpenURL(url string) error
	OpenBrowser(browser, url, proxy string, asApp bool) error
	// OpenImages opens every image under dir in the platform image viewer.
	OpenImages(dir string) error

	// ── network ──────────────────────────────────────────────────
	// WifiNetworks returns "SSID" lines.
	WifiNetworks() ([]string, error)
	WifiConnect(ssid, password string) error

	// ── bluetooth ────────────────────────────────────────────────
	// BluetoothPaired/Connected/Scan return "MAC NAME" lines.
	BluetoothPaired() ([]string, error)
	BluetoothConnected() ([]string, error)
	BluetoothScan() ([]string, error)
	BluetoothConnect(mac string) error
	BluetoothDisconnect(mac string) error

	// ── display ──────────────────────────────────────────────────
	DisplayLayouts(primary string) ([]string, error)
	ApplyDisplayLayout(layout, primary, wallpaperDir string) error

	// ── ssh ──────────────────────────────────────────────────────
	SSHConnect(term, user, host, password string, port int) error

	// ── popup text editing ───────────────────────────────────────
	// VimEditInput composes text in a vim popup and types it back into the
	// previously focused input box.
	VimEditInput(term string) error
}

var (
	onceOnce sync.Once
	instance Platform
)

// Get returns the process-wide platform implementation.
func Get() Platform {
	onceOnce.Do(func() { instance = newPlatform() })
	return instance
}

var (
	overrideMu sync.RWMutex
	overrides  map[Intent]string
)

// SetOverrides installs per-OS command overrides (from config). Templates may
// reference positional args as {0}, {1}, ... An empty command removes the
// override, restoring the built-in default.
func SetOverrides(m map[string]string) {
	overrideMu.Lock()
	defer overrideMu.Unlock()
	overrides = make(map[Intent]string, len(m))
	for k, v := range m {
		if strings.TrimSpace(v) == "" {
			continue
		}
		overrides[Intent(k)] = v
	}
}

// overrideCommand returns the configured override for i, if any.
func overrideCommand(i Intent, args []string) (string, bool) {
	overrideMu.RLock()
	tmpl, ok := overrides[i]
	overrideMu.RUnlock()
	if !ok {
		return "", false
	}
	return expand(tmpl, args), true
}

// expand substitutes {0}, {1}, ... with args.
func expand(tmpl string, args []string) string {
	out := tmpl
	for idx, a := range args {
		out = strings.ReplaceAll(out, fmt.Sprintf("{%d}", idx), a)
	}
	return out
}

// runShell executes a command line through bash and folds stderr into the
// returned error so callers can surface a useful message.
func runShell(script string) error {
	if strings.TrimSpace(script) == "" {
		return fmt.Errorf("%w: empty command", ErrUnsupported)
	}
	_, stderr, err := utils.RunScript("bash", script)
	if err != nil {
		stderr = strings.TrimSpace(stderr)
		if stderr == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, stderr)
	}
	return nil
}

// runIntent is the shared implementation of Platform.Run: config override
// first, then the per-OS command table.
func runIntent(intents map[Intent]string, i Intent, args []string) error {
	if cmd, ok := overrideCommand(i, args); ok {
		return runShell(cmd)
	}
	tmpl, ok := intents[i]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnsupported, i)
	}
	return runShell(expand(tmpl, args))
}

// unsupported builds a consistent "not supported" error with a hint.
func unsupported(feature, hint string) error {
	if hint == "" {
		return fmt.Errorf("%w: %s", ErrUnsupported, feature)
	}
	return fmt.Errorf("%w: %s (%s)", ErrUnsupported, feature, hint)
}
