# Assistant

Local HTTP API service that exposes system commands and utilities via REST API. Designed to be triggered from **dwm** (suckless window manager) keybindings or any HTTP client. Acts as a personal automation hub for a Linux desktop environment.

Author: **zetatez** — [github.com/zetatez/suckless-dwm](https://github.com/zetatez/suckless-dwm) (the `assistant/` subdirectory).

## Features

- **System control**: volume, brightness, microphone, display layout, power menu, keyboard backlight
- **Clipboard**: smart detection (paths/URLs/text), translate, format code
- **URL handling**: open in Chrome / qutebrowser / app mode
- **File management**: search (name/content/type), browse, open, upload (with embedded web UI)
- **Process management**: toggle/kill applications, music control, screen recording
- **AI/LLM integration**: LeetCode solving via screenshot, code formatting, translation, reporting
- **Network**: WiFi (rofi menu), Bluetooth, SSH connection, VPN status, IP detection
- **TARS agent**: AI conversational agent (Interstellar-inspired) running over Feishu, with memory, wiki search, and web search
- **Background daemons**: auto-restart dwmblocks, picom, dunst, fcitx5 on crash; wallpaper slideshow
- **Feishu (Lark)**: send messages, integrate with team communication
- **Kindle**: reading statistics dashboard with embedded web UI

## Tech Stack

 | Component   | Technology                                            |
 | ----------- | ------------                                          |
 | Language    | Go 1.25.3                                             |
 | HTTP        | Gin v1.11.0                                           |
 | Config      | Viper (YAML + env var expansion)                      |
 | Logging     | Logrus + Lumberjack (rotation)                        |
 | API Docs    | swaggo/swag (Swagger)                                 |
 | AI          | OpenAI-compatible API (Gemini, Ark, DeepSeek, NVIDIA) |
 | Cron        | robfig/cron/v3                                        |
 | Feishu      | larksuite/oapi-sdk-go/v3                              |

## Quick Start

```bash
# One-line install
curl -sL https://github.com/zetatez/suckless-dwm/raw/master/assistant/install.sh | sh

# Or manually:
# 1. make install
# 2. systemctl --user enable --now assistant
```

## API

All endpoints under `http://<host>:4321/api/`.

Auth: HTTP Basic (`auth.username` / `auth.password` from config).

Modules:

 | Prefix          | Module      | Description                                       |
 | --------        | --------    | -------------                                     |
 | `/api/health`   | health      | Health check                                      |
 | `/api/svr`      | svc         | ~50+ system/network/file/AI endpoints             |
 | `/api/files`    | filebrowser | File listing, raw read, download, upload + web UI |
 | `/api/llm`      | llmproxy    | Anthropic-compatible LLM proxy                    |
 | `/api/kindle`   | kindle      | Kindle reading statistics + web UI                |
 | `/swagger/*any` | —           | Swagger UI                                        |

See `scripts/` for all available endpoints. Each file is a single curl command:

```bash
./scripts/sys-shortcut               # rofi power menu
./scripts/sys-display                # rofi display layout
./scripts/toggle.sh flameshot
./scripts/launch.sh inkscape
./scripts/open-url-chrome.sh "https://github.com"
./scripts/solve-leetcode             # AI solve via screenshot
./scripts/translate-clipboard
./scripts/send-to-feishu
```

## Project Structure

```
cmd/assistant/            # entrypoint (signal handling, bootstrap)
internal/
├── app/                  # Gin server + module registration
│   └── modules/
│       ├── health/       # health check
│       ├── svc/          # all service endpoints (handler + service)
│       ├── filebrowser/  # file management + embedded web UI
│       ├── llmproxy/     # LLM proxy endpoint
│       └── kindle/       # Kindle dashboard + web UI
├── bootstrap/            # init config, log, LLM client, background tasks, TARS
│   └── psl/              # package-level singletons (config, logger, llmclient, background, shutdown)
└── tars/                 # TARS AI agent (handler, memory, parser, prompt, react, tools)
pkg/                      # shared libraries
├── aiapi/                # AI-powered APIs (diagnoser, translator, reporter, etc.)
├── cache/                # in-memory LRU cache
├── channel/feishu/       # Feishu messaging
├── consts/               # constants
├── cron/                 # cron scheduling helpers
├── dwmblocknotify/       # dwm status bar notifications
├── encrypt/              # AES-GCM encryption
├── llm/                  # LLM client abstraction + proxy (multi-provider failover)
├── network/              # IP detection
├── news/                 # RSS/news feed fetcher
├── response/             # standard JSON response helpers
├── retry/                # retry with backoff (sync + async)
├── utils/                # utilities (fs, exec, git, glob, grep, encoding, etc.)
└── xlog/                 # logrus logger factory + rotation
scripts/                  # 73 curl scripts (one per endpoint)
docs/                     # Swagger generated docs
config.default.yaml       # configuration template
config.yaml               # active configuration
assistant.service         # systemd user unit
tars.sys.prompt.md        # TARS system prompt
Makefile                  # build, install, dev, clean, swag
install.sh                # one-line installer
```

## Configuration

See `config.default.yaml`. Key sections:

- `auth` — HTTP Basic credentials
- `server` — host, port
- `log` — level, file, rotation
- `background` — daemon watch list, wallpaper, news
- `llm` — provider(s), API keys, model, proxy
- `feishu` — app credentials for TARS agent
- `svc` — custom scripts, paths, network interfaces

## Running

```bash
# Development (with go run)
make dev

# Build and run binary
make build && ./bin/assistant

# As a systemd user service
systemctl --user enable --now assistant
```

## Architecture

```
cmd/assistant/main.go
  └─ bootstrap.Run()
       ├─ InitConfig → InitLog → InitLLMClient
       ├─ StartBackgroundTasks (daemon watch, wallpaper, news)
       ├─ InitTars (Feishu agent)
       └─ app.Run() → Gin server
            ├─ /api/health/
            ├─ /api/svr/*       (svc module)
            ├─ /api/files/*     (filebrowser module)
            ├─ /api/llm/*       (llmproxy module)
            ├─ /api/kindle/*    (kindle module)
            └─ /swagger/*
```
