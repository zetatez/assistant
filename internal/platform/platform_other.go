//go:build !linux && !darwin

package platform

import (
	"fmt"
	"time"
)

type unsupportedPlatform struct{}

func newPlatform() Platform { return unsupportedPlatform{} }

func (unsupportedPlatform) GOOS() string { return "other" }

func (unsupportedPlatform) Notify(msg string, ttl time.Duration) error {
	return fmt.Errorf("%w: notifications", ErrUnsupported)
}

func (unsupportedPlatform) NotifyQueue(msg string, ttl time.Duration) error {
	return fmt.Errorf("%w: notifications", ErrUnsupported)
}

func (unsupportedPlatform) NotifyClear() error { return nil }

func (unsupportedPlatform) Select(prompt string, options []string) (string, error) {
	return "", fmt.Errorf("%w: interactive selection", ErrUnsupported)
}

func (unsupportedPlatform) Input(prompt string, secret bool) (string, error) {
	return "", fmt.Errorf("%w: interactive input", ErrUnsupported)
}

func (unsupportedPlatform) Run(i Intent, args ...string) error {
	return fmt.Errorf("%w: %s", ErrUnsupported, i)
}

func (unsupportedPlatform) KeyboardLightToggle(path string) (string, error) {
	return "", fmt.Errorf("%w: keyboard backlight", ErrUnsupported)
}

func (unsupportedPlatform) TerminalCommand(term, script string) (string, error) {
	return "", fmt.Errorf("%w: terminal integration", ErrUnsupported)
}

func (unsupportedPlatform) FloatingTerminalCommand(term, script string) (string, error) {
	return "", fmt.Errorf("%w: terminal integration", ErrUnsupported)
}

func (unsupportedPlatform) ScratchpadTerminalCommand(term, script string) (string, error) {
	return "", fmt.Errorf("%w: terminal integration", ErrUnsupported)
}

func (unsupportedPlatform) ScreenSize() (int, int) { return 1920, 1080 }

func (unsupportedPlatform) Screenshot() ([]byte, error) {
	return nil, fmt.Errorf("%w: screenshots", ErrUnsupported)
}

func (unsupportedPlatform) CaptureScreenArgs(out string, w, h int, cfg CaptureConfig) []string {
	return nil
}

func (unsupportedPlatform) CaptureAudioArgs(out string, cfg CaptureConfig) []string {
	return nil
}

func (unsupportedPlatform) CaptureWebcamArgs(out string, cfg CaptureConfig) []string {
	return nil
}

func (unsupportedPlatform) PreviewWebcamCmd(cfg CaptureConfig) []string { return nil }

func (unsupportedPlatform) SetWallpaper(path string) error {
	return fmt.Errorf("%w: wallpaper", ErrUnsupported)
}

func (unsupportedPlatform) DefaultInterfaceName() string { return "" }

func (unsupportedPlatform) OpenURL(url string) error {
	return fmt.Errorf("%w: opening URLs", ErrUnsupported)
}

func (unsupportedPlatform) OpenBrowser(browser, url, proxy string, asApp bool) error {
	return fmt.Errorf("%w: opening browsers", ErrUnsupported)
}

func (unsupportedPlatform) OpenImages(dir string) error {
	return fmt.Errorf("%w: image viewer", ErrUnsupported)
}

func (unsupportedPlatform) WifiNetworks() ([]string, error) {
	return nil, fmt.Errorf("%w: wifi", ErrUnsupported)
}

func (unsupportedPlatform) WifiConnect(ssid, password string) error {
	return fmt.Errorf("%w: wifi", ErrUnsupported)
}

func (unsupportedPlatform) BluetoothPaired() ([]string, error) {
	return nil, fmt.Errorf("%w: bluetooth", ErrUnsupported)
}

func (unsupportedPlatform) BluetoothConnected() ([]string, error) {
	return nil, fmt.Errorf("%w: bluetooth", ErrUnsupported)
}

func (unsupportedPlatform) BluetoothScan() ([]string, error) {
	return nil, fmt.Errorf("%w: bluetooth", ErrUnsupported)
}

func (unsupportedPlatform) BluetoothConnect(mac string) error {
	return fmt.Errorf("%w: bluetooth", ErrUnsupported)
}

func (unsupportedPlatform) BluetoothDisconnect(mac string) error {
	return fmt.Errorf("%w: bluetooth", ErrUnsupported)
}

func (unsupportedPlatform) DisplayLayouts(primary string) ([]string, error) {
	return nil, fmt.Errorf("%w: display layouts", ErrUnsupported)
}

func (unsupportedPlatform) ApplyDisplayLayout(layout, primary, wallpaperDir string) error {
	return fmt.Errorf("%w: display layouts", ErrUnsupported)
}

func (unsupportedPlatform) SSHConnect(term, user, host, password string, port int) error {
	return fmt.Errorf("%w: ssh", ErrUnsupported)
}

func (unsupportedPlatform) VimEditInput(term string) error {
	return fmt.Errorf("%w: vim edit input", ErrUnsupported)
}
