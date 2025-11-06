# ======================================
# 🧠 项目信息
# ======================================
APP_NAME := assistant
APP_CMD := ./cmd/server/main.go
BIN_PATH := ./bin
SQLC_PATH := sqlc.yaml

# ======================================
# 🧩 通用变量
# ======================================
APP_ADDR ?= :8080
BUILD_TIME := $(shell date +'%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GO_FLAGS := -ldflags="-X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

# ======================================
# 🧹 PHONY 声明
# ======================================
.PHONY: all run dev build clean docker-up docker-down swag generate test lint fmt help

# ======================================
# 🚀 开发与运行
# ======================================
run:
	@echo "🚀 Running $(APP_NAME) at $(APP_ADDR)"
	go run $(APP_CMD)

# ======================================
# 🏗️ 构建
# ======================================
build:
	@mkdir -p $(BIN_PATH)
	@echo "🔨 Building $(APP_NAME)..."
	@go build $(GO_FLAGS) -o $(BIN_PATH)/$(APP_NAME) $(APP_CMD)
	@echo "✅ Build complete: $(BIN_PATH)/$(APP_NAME)"

# ======================================
# 🧠 代码生成
# ======================================
generate:
	@echo "⚙️ Generating sqlc code..."
	sqlc generate
	@echo "✅ sqlc done"

swag:
	@echo "📘 Generating Swagger docs..."
	swag init -g $(APP_CMD) -o docs --parseDependency --exclude pkg/retry
	@echo "✅ Swagger generated at ./docs"

# ======================================
# 🧪 测试与代码质量
# ======================================
test:
	@echo "🧪 Running unit tests..."
	@go test ./... -cover

lint:
	@echo "🔍 Running golangci-lint (if installed)..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found. Install with 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"; \
	fi

fmt:
	@echo "🧹 Formatting Go code..."
	@go fmt ./...
	@go vet ./...

# ======================================
# 🐳 Docker 管理
# ======================================
docker-up:
	@echo "🐳 Starting docker-compose..."
	docker compose up --build -d

docker-down:
	@echo "🧹 Stopping and removing docker containers..."
	docker compose down -v

# ======================================
# 🧼 清理
# ======================================
clean:
	@echo "🧹 Cleaning build outputs..."
	rm -rf $(BIN_PATH)
	rm -rf ./docs
	@echo "✅ Clean complete"

# ======================================
# 📜 帮助信息
# ======================================
help:
	@echo "📖 Makefile Commands:"
	@echo "  make run           - 运行程序"
	@echo "  make dev           - 开发模式运行 (air热加载)"
	@echo "  make build         - 构建可执行文件"
	@echo "  make generate      - 运行 sqlc 代码生成"
	@echo "  make swag          - 生成 Swagger 文档"
	@echo "  make test          - 运行单元测试"
	@echo "  make fmt           - 格式化与静态检查"
	@echo "  make lint          - 运行代码检查 (golangci-lint)"
	@echo "  make docker-up     - 启动 docker compose"
	@echo "  make docker-down   - 停止 docker compose"
	@echo "  make clean         - 清理构建产物"
