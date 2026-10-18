package svc

import (
	"fmt"
	"strings"
	"time"

	"assistant/internal/bootstrap/psl"
	"assistant/internal/platform"
	"assistant/pkg/dwmblocknotify"
	"assistant/pkg/utils"
)

func (s *Service) SysShortcut() error {
	actions := []string{"suspend", "poweroff", "reboot", "off-display", "lock"}
	intents := map[string]platform.Intent{
		"suspend":     platform.Suspend,
		"poweroff":    platform.Shutdown,
		"reboot":      platform.Reboot,
		"off-display": platform.ScreenOff,
		"lock":        platform.Lock,
	}

	action, err := platform.Get().Select("power", actions)
	if err != nil {
		return err
	}
	if action == "" {
		return nil
	}
	intent, ok := intents[action]
	if !ok {
		return fmt.Errorf("unknown action: %s", action)
	}
	return platform.Get().Run(intent)
}

func (s *Service) SysDisplay() error {
	cfg := psl.GetConfig().Settings
	p := platform.Get()

	layouts, err := p.DisplayLayouts(cfg.DefaultMonitor)
	if err != nil {
		return err
	}
	layout, err := p.Select("screen strategy", layouts)
	if err != nil {
		dwmblocknotify.PUT(fmt.Sprintf("display: prompt failed: %v", err), 3*time.Second)
		layout = "default"
	}
	if layout == "" {
		layout = "default"
	}
	if err := p.ApplyDisplayLayout(layout, cfg.DefaultMonitor, cfg.DirWallpaper); err != nil {
		return fmt.Errorf("set display: %w", err)
	}
	return nil
}

func (s *Service) SysKeyboardLight() (string, error) {
	path := psl.GetConfig().Settings.PathKeyboardBrightness
	return platform.Get().KeyboardLightToggle(path)
}

func (s *Service) SysVolumeUp() error {
	err := platform.Get().Run(platform.VolumeUp)
	if err == nil {
		dwmblocknotify.PUT("vol +5%", 1*time.Second)
	}
	return err
}

func (s *Service) SysVolumeDown() error {
	err := platform.Get().Run(platform.VolumeDown)
	if err == nil {
		dwmblocknotify.PUT("vol -5%", 1*time.Second)
	}
	return err
}

func (s *Service) SysVolumeToggle() error {
	err := platform.Get().Run(platform.VolumeToggle)
	if err == nil {
		dwmblocknotify.PUT("vol toggle", 1*time.Second)
	}
	return err
}

func (s *Service) SysMicroUp() error {
	err := platform.Get().Run(platform.MicUp)
	if err == nil {
		dwmblocknotify.PUT("micro +5%", 1*time.Second)
	}
	return err
}

func (s *Service) SysMicroDown() error {
	err := platform.Get().Run(platform.MicDown)
	if err == nil {
		dwmblocknotify.PUT("micro -5%", 1*time.Second)
	}
	return err
}

func (s *Service) SysMicroToggle() error {
	err := platform.Get().Run(platform.MicToggle)
	if err == nil {
		dwmblocknotify.PUT("micro toggle", 1*time.Second)
	}
	return err
}

func (s *Service) SysDisplayLightUp() error {
	err := platform.Get().Run(platform.BrightnessUp)
	if err == nil {
		dwmblocknotify.PUT("light +1%", 1*time.Second)
	}
	return err
}

func (s *Service) SysDisplayLightDown() error {
	err := platform.Get().Run(platform.BrightnessDown)
	if err == nil {
		dwmblocknotify.PUT("light -1%", 1*time.Second)
	}
	return err
}

func (s *Service) SysReset() error {
	dwmblocknotify.PUT("reset sys default", 1*time.Second)
	p := platform.Get()

	type step struct {
		intent platform.Intent
		args   []string
		msg    string
	}
	steps := []step{
		{platform.BrightnessSet, []string{"48"}, "set light to 48%"},
		{platform.AudioReset, nil, "reset audio defaults"},
		{platform.KeyRepeatSet, []string{"158", "128"}, "set keyboard rate to 158 128"},
	}
	for _, st := range steps {
		if err := p.Run(st.intent, st.args...); err != nil {
			// macOS has no built-in equivalent for some of these; keep going so
			// the remaining steps still apply.
			if !strings.Contains(err.Error(), platform.ErrUnsupported.Error()) {
				dwmblocknotify.PUT(fmt.Sprintf("%s failed: %v", st.msg, err), 3*time.Second)
			}
			continue
		}
		dwmblocknotify.PUT(st.msg, 1*time.Second)
	}
	return nil
}

func (s *Service) SysKill() error {
	term := psl.GetConfig().Settings.DefaultTerminal
	script := `pids=$(ps -axo pid,command | tail -n +2 | fzf --prompt="kill -9 >" --select-1 --exit-0 | awk '{print $1}')
[ -n "$pids" ] && kill -9 $pids`
	cmd, err := platform.Get().TerminalCommand(term, script)
	if err != nil {
		return err
	}
	return utils.StartScript("bash", cmd)
}

func (s *Service) SysWifiConnect() error {
	p := platform.Get()
	networks, err := p.WifiNetworks()
	if err != nil {
		return err
	}
	ssid, err := p.Select("connect to wifi", networks)
	if err != nil {
		return err
	}
	if ssid == "" {
		ssid, err = p.Input("wifi ssid", false)
		if err != nil || strings.TrimSpace(ssid) == "" {
			return err
		}
	}
	password, err := p.Input("password", true)
	if err != nil {
		return err
	}
	if password == "" {
		return nil // cancelled
	}
	return p.WifiConnect(strings.TrimSpace(ssid), password)
}

func (s *Service) SysBluetoothConnect() error {
	p := platform.Get()
	devices, err := p.BluetoothPaired()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return fmt.Errorf("no paired bluetooth devices found")
	}
	choice, err := p.Select("connect to", devices)
	if err != nil || choice == "" {
		return err
	}
	mac := platform.ExtractMAC(choice)
	if mac == "" {
		return fmt.Errorf("invalid bluetooth device: %s", choice)
	}
	return p.BluetoothConnect(mac)
}

func (s *Service) SysBluetoothDisconnect() error {
	p := platform.Get()
	devices, err := p.BluetoothConnected()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return fmt.Errorf("no connected bluetooth devices found")
	}
	choice, err := p.Select("disconnect from", devices)
	if err != nil || choice == "" {
		return err
	}
	mac := platform.ExtractMAC(choice)
	if mac == "" {
		return fmt.Errorf("invalid bluetooth device: %s", choice)
	}
	return p.BluetoothDisconnect(mac)
}

func (s *Service) SysBluetoothScanConnect() error {
	p := platform.Get()
	devices, err := p.BluetoothScan()
	if err != nil {
		return fmt.Errorf("scan bluetooth: %w", err)
	}
	if len(devices) == 0 {
		return fmt.Errorf("no bluetooth devices found")
	}
	choice, err := p.Select("connect bluetooth", devices)
	if err != nil || choice == "" {
		return err
	}
	mac := platform.ExtractMAC(choice)
	if mac == "" {
		return fmt.Errorf("invalid bluetooth device: %s", choice)
	}
	return p.BluetoothConnect(mac)
}
