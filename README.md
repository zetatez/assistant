# Assistant

Local HTTP API service exposing system commands and utilities, designed for **dwm** keybindings. A personal automation hub for Linux desktop.

Author: **zetatez** - [github.com/zetatez/suckless-dwm](https://github.com/zetatez/suckless-dwm)

## Features

- **System control**: volume, brightness, display layout, power menu, WiFi/Bluetooth/SSH
- **Clipboard**: smart detection (path/URL), translate, format code
- **AI**: LeetCode solving (screenshot), translation, reporting
- **News notify**: periodic RSS fetch pushed to dwm status bar
- **Background**: daemon auto-restart, wallpaper slideshow

## Quick Start

```bash
curl -sL https://github.com/zetatez/suckless-dwm/raw/master/assistant/install.sh | sh
# or: make install && systemctl --user enable --now assistant
```

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
├── bootstrap/psl/       # config, logger, llm client, background tasks
└── news/                # news notify service
pkg/
├── llm/                 # OpenAI-compatible LLM client (single provider)
├── aiapi/               # structured AI APIs (translator, reporter, ...)
├── news_collector/      # RSS fetcher
├── dwmblocknotify/      # dwm status bar notifications
└── utils/, xlog/, ...   # utilities
scripts/                 # curl scripts
config.default.yaml      # config template
```

## Configuration

See `config.default.yaml`. Key sections: `app`, `llm` (single provider base_url/api_key + text/vision models), `news`, `background`.

## Running

```bash
make dev              # development
make build            # production build
systemctl --user enable --now assistant
```
