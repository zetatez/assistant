# Assistant

Personal AI assistant with long-term memory and local wiki search.

## Features

- **Tars**: AI assistant with multi-tier memory
- **Local Wiki**: Grep + LLM rerank over markdown files
- **Knowledge Graph**: Entity extraction from conversations
- **Multi-Channel**: Feishu (extensible)

## Quick Start

```bash
# Clone and setup
make generate && make swag

# Configure (edit config.yaml)
# - Set LLM provider and API key
# - Set wiki directory: tars.wiki.dir: ~/wiki
# - Set Feishu app credentials

# Run
make dev       # dev mode
make run       # production

# Docker
make docker-up
```

## Configuration

```yaml
# --- App ---
app:
  name: assistant
  port: 9876
  interface: wlan0

# --- Auth ---
auth:
  root:
    username: admin
    password: AAaa00__
    email: admin@example.com
  jwt:
    secret: your-256-bit-secret-key-here
    expiry_hours: 24

# --- Database ---
db:
  driver: mysql
  dsn: assistant:AAaa00__@tcp(127.0.0.1:3306)/assistant?charset=utf8mb4&parseTime=True&loc=Local
  pool:
    max_open_conns: 20
    max_idle_conns: 10
    conn_max_lifetime: 1h
    conn_max_idle_time: 30m

# --- LLM ---
llm:
  provider: minimax
  api_key: "${LLM_API_KEY}"
  base_url: https://api.minimaxi.com
  model: MiniMax-M2.5
  timeout: 60
  max_tokens: 4096
  temperature: 0.2

# --- Channel ---
channel:
  provider: feishu
  feishu:
    app_id: ""
    app_secret: ""

# --- Tars AI ---
tars:
  enabled: true
  llm_temperature: 0.7
  persona:
    humor_level: 50
    honesty_level: 80
  memory:
    max_history: 64
    ttl_minutes: 60
  wiki:
    enabled: true
    dir: ~/share/github/obsidian/

# --- Monitor ---
monitor:
  tracing:
    enabled: true
    sample_rate: 1.0
  metrics:
    enabled: true
    path: /metrics

# --- Log ---
log:
  level: debug
  filename: ./logs/assistant.log
  max_size_mb: 16
  max_backups: 7
  max_age_days: 14
  compress: true
  format: text
  console: true
```

## Tars Memory Architecture

```
System Prompt → Memory Doc → Session State → Wiki → Knowledge → Conversation
```

- Wiki search: Grep + LLM rerank, only results with relevance >= 5.0
- Token limit: 12,000 per request
- Background: entity extraction, session refresh every 5 messages

## Project Structure

```
internal/app/modules/
├── tars/              # AI assistant
│   ├── handler.go     # Message processing
│   ├── memory/        # Short-term + long-term memory
│   └── knowledge/     # Entity extraction
pkg/wiki/              # Local wiki search
pkg/channel/feishu/    # Feishu integration
```

## Make Commands

```
make run        # build and run
make dev        # hot reload
make test       # run tests
make generate   # generate sqlc code
make swag       # generate swagger
make docker-up  # start docker
```
