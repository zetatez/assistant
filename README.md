# Assistant

Local HTTP API service exposing system commands and utilities for a personal
desktop automation hub. Runs on **Arch Linux (X11/dwm)** and **macOS**.

Author: **zetatez** - [github.com/zetatez/suckless-dwm](https://github.com/zetatez/suckless-dwm)

## Features

- **System control**: volume, brightness, display layout, power menu, WiFi/Bluetooth/SSH
- **Clipboard**: smart detection (path/URL), translate, format code
- **AI**: LeetCode solving (screenshot), translation, reporting
- **News notify**: periodic RSS fetch pushed to the desktop notifier
- **Background**: daemon auto-restart, wallpaper slideshow

All OS-specific plumbing lives in `internal/platform` and is selected at build
time: `platform_linux.go` keeps the original Arch/X11 behaviour, while
`platform_darwin.go` uses system tools (`osascript`, `screencapture`,
`networksetup`, `pmset`, `open`). Features with no built-in macOS equivalent
return a clear error, and every canned action can be overridden in config.

## Quick Start

### Arch Linux

```bash
curl -sL https://github.com/zetatez/suckless-dwm/raw/master/assistant/install.sh | sh
# or: make install    # systemd --user unit
```

Install runtime dependencies used by the Linux implementation (Arch):

```bash
sudo pacman -S rofi xdotool xclip xrandr feh flameshot light \
  networkmanager bluez bluez-utils alsa-utils slock ffmpeg \
  fd fzf ripgrep bat neovim kitty
```

`install.sh` and `make install` keep an existing `~/.config/assistant/config.yaml`
untouched and only generate one from `config.default.yaml` when it is missing.

### macOS

```bash
curl -sL https://github.com/zetatez/suckless-dwm/raw/master/assistant/install.sh | sh
# or: make install    # launchd user agent (com.assistant)
```

The macOS build only needs the system toolchain; optional features degrade to
a clear error when a helper is missing:

| Feature | macOS implementation | Optional helper |
|---|---|---|
| Notifications | `osascript display notification` | `terminal-notifier` |
| Prompts / menus | `osascript` dialogs | — |
| Volume / mic | `osascript set volume` | — |
| Brightness | brightness media keys via AppleScript | `brightness` (config override) |
| Screenshot | `screencapture` | — |
| WiFi | `networksetup` | — |
| Bluetooth | `blueutil` | `brew install blueutil` |
| Display layouts | not scriptable with system tools | `displayplacer` (config override) |
| Keyboard backlight | not scriptable | config override |
| Wallpaper | System Events | — |
| Recording | `ffmpeg -f avfoundation` | `brew install ffmpeg` |
| Popup vim editing | Terminal + AppleScript paste | Accessibility permission |

Grant **Accessibility** (popup editing / media keys / synthetic keystrokes),
**Screen Recording** (screenshots) and **Automation** (AppleScript) permissions
to the terminal or to `assistant` in
`System Settings → Privacy & Security`.

## API

`http://<host>:4321/api/`.

| Prefix             | Description                                    |
|--------------------|------------------------------------------------|
| `/api/svr`         | ~50+ system/network/file/AI endpoints          |
| `/api/health`      | health check                                   |

`scripts/` has one curl script per endpoint.

## Structure

```
cmd/assistant/           # entrypoint
internal/
├── app/modules/         # gin modules: svc, health
├── platform/            # OS abstraction: linux + darwin implementations
├── bootstrap/psl/       # config, logger, llm client, background tasks
└── news/                # news notify service
pkg/
├── llm/                 # OpenAI-compatible LLM client (single provider)
├── aiapi/               # structured AI APIs (translator, reporter, ...)
├── news_collector/      # RSS fetcher
├── dwmblocknotify/      # status notifications (dwmblocks on Linux, toast on macOS)
└── utils/, xlog/, ...   # utilities
scripts/                 # curl scripts
config.default.yaml      # config template (works on both OSes)
assistant.service        # systemd --user unit (Linux)
assistant.plist          # launchd agent (macOS)
```

## Configuration

See `config.default.yaml`. Empty values fall back to per-OS defaults
(interface, monitor, terminal, keyboard backlight path, background daemons).
Key sections: `app`, `settings`, `llm`, `news`, `background`.

Per-intent command overrides let you plug in any external tool without code
changes:

```yaml
settings:
  commands:
    brightness_up: "brightness 0.05"
    keyboard_light_toggle: "kbdlight toggle"
```

## Compatibility

The refactor does not change existing Arch Linux behaviour: all OS-specific
commands still run the same tools, and every new `settings` field is optional
with a default that matches the previous hard-coded value. An existing
`config.yaml` keeps working unmodified (e.g. `interface`, `default_monitor`,
`default_terminal`, `filepath_keyboard_brightness` and `background.procs` are
all honoured as-is; the former AI output path is expressed as
`dir_ai_output`, defaulting to `~/git/test`).

Small, non-breaking differences on Linux:

- Wallpaper slideshow also picks up `.jpg` / `.jpeg` / `.png` (previously only `*.JPG`).
- Wi-Fi password input is masked; cancelling it no longer attempts an
  empty-password connect.
- Power menu lists `lock` instead of `slock` (same `slock` action).
- Bluetooth disconnect reports an error when no device is connected and uses
  `bluetoothctl devices Connected` (the old argument-less `bluetoothctl info`
  was unreliable).
- `SysReset` runs the same commands (Master unmute/80%, Speaker unmute/64%,
  Capture cap/64%, Headphone 64%, `light -S 48`, `xset r rate 158 128`); only
  the notification text changed.
- Recording windows keep the `scratchpad` class, so existing dwm rules apply.

## Build

```bash
make build                 # host OS
make build-linux           # linux/amd64
make build-darwin          # darwin/arm64 + darwin/amd64
```

## Running

```bash
make dev                   # development
make build                 # production build
systemctl --user enable --now assistant    # Linux
launchctl load -w ~/Library/LaunchAgents/com.assistant.plist   # macOS
```
