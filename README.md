# Assistant

Local HTTP API service exposing system commands and utilities, designed for **dwm** keybindings. A personal automation hub for Linux desktop.

Author: **zetatez** - [github.com/zetatez/suckless-dwm](https://github.com/zetatez/suckless-dwm)

## Features

- **System control**: volume, brightness, display layout, power menu, WiFi/Bluetooth/SSH
- **Clipboard**: smart detection (path/URL), translate, format code
- **File management**: search, browse, upload (web UI)
- **AI**: LeetCode solving (screenshot), translation, reporting
- **LLM proxy**: OpenAI/Responses/Anthropic-compatible, multi-provider failover, vision-model auto routing
- **TARS agent**: Feishu AI bot with memory and ReAct tool loop (`grep_wiki`/`read_wiki`/`web_search`)
- **News notify**: periodic RSS fetch pushed to dwm status bar
- **Background**: daemon auto-restart, wallpaper slideshow

## Quick Start

```bash
curl -sL https://github.com/zetatez/suckless-dwm/raw/master/assistant/install.sh | sh
# or: make install && systemctl --user enable --now assistant
```

## API

`http://<host>:4321/api/` - Basic Auth for svr/filebrowser, Bearer token for llmproxy.

| Prefix             | Description                                    |
|--------------------|------------------------------------------------|
| `/api/svr`         | ~50+ system/network/file/AI endpoints          |
| `/api/filebrowser` | file management + web UI                       |
| `/api/llmproxy`    | LLM proxy (OpenAI/Responses/Anthropic)         |
| `/api/health`      | health check                                   |

`scripts/` has one curl script per endpoint.

## Structure

```
cmd/assistant/           # entrypoint
internal/
├── app/modules/         # gin modules: svc, filebrowser, llm, health
├── bootstrap/psl/       # config, logger, llm client, background tasks
├── news/                # news notify service
└── tars/                # Feishu AI agent (react, memory, tools)
pkg/
├── llmproxy/            # LLM client + multi-provider proxy
├── aiapi/               # structured AI APIs (translator, reporter, ...)
├── news_collector/      # RSS fetcher
├── channel/feishu/      # Feishu messaging
├── dwmblocknotify/      # dwm status bar notifications
└── utils/, xlog/, ...   # utilities
scripts/                 # curl scripts
config.default.yaml      # config template
```

## Configuration

See `config.default.yaml`. Key sections: `app`, `llm_proxy` (providers + vision_models), `tars` (wiki_search, web_search), `news`, `background`, `filebrowser`.

## Running

```bash
make dev              # development
make build            # production build
systemctl --user enable --now assistant
```
