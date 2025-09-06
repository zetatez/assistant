# Assistant

A Go-based web application built with Gin framework featuring distributed locking, user management, AI chatbot with long-term memory, and RESTful API capabilities.

## Features

- **RESTful API**: Built with Gin framework with automatic Swagger documentation
- **Distributed Locking**: Database-based distributed lock implementation for multi-instance coordination
- **User Management**: Built-in admin user system with JWT authentication
- **AI Chatbot (Tars)**: Channel-abstracted AI assistant with short-term and long-term memory
- **Channel Abstraction**: Pluggable message channel system (Feishu supported)
- **Health Monitoring**: Built-in health check endpoints
- **LLM Integration**: Multi-provider LLM support (DeepSeek/OpenAI/Ollama/Qwen/MiniMax/GLM/Gemini/Doubao)
- **Database Integration**: MySQL with SQLc for type-safe database operations
- **Docker Support**: Production-ready Docker configuration

## Tech Stack

- **Language**: Go 1.22
- **Web Framework**: Gin
- **Database**: MySQL 8.0
- **Code Generation**: SQLc
- **API Documentation**: Swagger
- **Container**: Docker

## Quick Start

### Prerequisites

- Go 1.22+
- MySQL 8.0+
- Docker & Docker Compose (optional)
- Make

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd assistant
   ```

2. **Generate database code**
   ```bash
   make generate
   ```

3. **Generate API documentation**
   ```bash
   make swag
   ```

### Configuration

Modify `config.yaml` according to your environment:

```yaml
app:
  name: assistant
  port: 9876
  root:
    username: admin
    password: your_secure_password
    email: admin@example.com
  feishu:
    app_id: ""
    app_secret: ""

channel:
  provider: feishu

tars:
  enabled: false
  max_history: 30
  memory_ttl_minutes: 60
  temperature: 0.7

llm:
  provider: minimax
  api_key: "${MINIMAX_API_KEY}"
  base_url: https://api.minimaxi.com
  model: MiniMax-M2.5
  timeout: 60

db:
  driver_name: mysql
  dsn: assistant:AAaa00__@tcp(127.0.0.1:3306)/assistant?charset=utf8mb4&parseTime=True&loc=Local
  max_open_conns: 20
  max_idle_conns: 10

dislock:
  default_ttl: 30
  max_ttl: 300

log:
  level: debug
  filename: ./logs/assistant.log
```

### Database Setup

**Using Docker Compose**
```bash
docker compose up -d mysql
```

**Local MySQL**
Ensure MySQL is running and create the database:
```sql
CREATE DATABASE assistant;
```

### Running the Application

**Development Mode (with hot reload)**
```bash
make dev
```

**Production Mode**
```bash
make run
```

**Build Binary**
```bash
make build
./bin/assistant
```

### Docker Deployment

**Build and Run**
```bash
make docker-up
```

**Stop Containers**
```bash
make docker-down
```

## API Documentation

Once the application is running, access Swagger documentation at:
```
http://localhost:9876/swagger/index.html
```

## Available Make Commands

| Command | Description |
|---------|-------------|
| `make run` | Build and run the application |
| `make dev` | Run with hot reload (requires air) |
| `make build` | Build executable binary |
| `make test` | Run unit tests with race detection |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code and run vet |
| `make generate` | Generate SQLc database code |
| `make swag` | Generate Swagger documentation |
| `make docker-up` | Start Docker containers |
| `make docker-down` | Stop Docker containers |
| `make clean` | Clean build artifacts |

## Project Structure

```
assistant/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── app/
│   │   ├── module/           # Module interface definition
│   │   ├── modules/          # Feature modules
│   │   │   ├── health/       # Health check module
│   │   │   ├── sys_distributed_lock/  # Distributed locking
│   │   │   ├── sys_server/   # Server management
│   │   │   ├── sys_user/     # User management
│   │   │   └── tars/         # AI chatbot with memory
│   │   ├── repo/             # Database repositories (SQLc)
│   │   └── server.go         # Server initialization
│   └── bootstrap/
│       └── psl/              # Bootstrap utilities (config/db/log)
├── pkg/
│   ├── channel/              # Channel abstraction
│   │   ├── channel.go        # Channel interface
│   │   └── feishu/          # Feishu channel implementation
│   ├── middleware/           # HTTP middleware (auth/ratelimit)
│   ├── llm/                 # LLM provider integrations
│   │   └── providers/        # DeepSeek/OpenAI/Ollama/Qwen/MiniMax/GLM/Gemini/Doubao
│   ├── cache/               # In-memory cache with TTL
│   └── response/            # Unified API response
├── db/
│   ├── schema/               # Database table schemas
│   └── queries/              # SQL queries (SQLc)
├── docs/                     # Swagger documentation
├── docker-compose.yml        # Docker Compose configuration
├── Dockerfile               # Docker build configuration
├── Makefile                 # Build automation
└── config.yaml              # Application configuration
```

## Module System

The application uses a modular architecture where each feature is implemented as a separate module implementing the `Module` interface:

```go
type Module interface {
    Name() string
    Register(r *gin.RouterGroup)
    Middleware() []gin.HandlerFunc
}
```

## Tars AI Chatbot

Tars is an AI chatbot with dual-memory system:

- **Short-term Memory**: In-memory cache (30 messages default)
- **Long-term Memory**: Database storage with LLM-powered keyword extraction and summarization
- **Channel Abstraction**: Works with any messaging platform implementing the `Channel` interface

Supported channels:
- Feishu (Lark)

Configure Tars in `config.yaml`:

```yaml
tars:
  enabled: true
  max_history: 30
  memory_ttl_minutes: 60
  temperature: 0.7
```

## Distributed Lock

The system includes a database-based distributed lock implementation for multi-instance coordination:

```yaml
dislock:
  default_ttl: 30    # Default lock TTL in seconds
  max_ttl: 300       # Maximum lock TTL in seconds
```

## Troubleshooting

**Port already in use**
```bash
lsof -i :9876
kill <PID>
```

**Database connection failed**
- Ensure MySQL is running
- Verify database credentials in `config.yaml`
- Check network connectivity

**Docker build fails**
- Ensure Docker daemon is running
- Check available disk space

**Swagger not working**
- Run `make swag` to regenerate documentation

## License

This project is proprietary software.
