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
BUILD_TIME := $(shell date +'%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_FLAGS := -ldflags="-X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"
GO_TEST_FLAGS ?= -race -coverprofile=coverage.out -covermode=atomic

# ======================================
# 🧹 PHONY 声明
# ======================================
.PHONY: all run dev build clean docker-up docker-down swag generate test lint fmt help check-deps deps

# ======================================
# 🚀 默认目标
# ======================================
all: build

# ======================================
# 🚀 开发与运行
# ======================================
run: check-deps
	@echo "🚀 Running $(APP_NAME)"
	go run $(APP_CMD)

dev: check-deps
	@echo "🔥 Starting dev server with hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "⚠️  air not found. Install with 'go install github.com/air-verse/air@latest'"; \
		exit 1; \
	fi

# ======================================
# 🏗️ 构建
# ======================================
build:
	@mkdir -p $(BIN_PATH)
	@echo "🔨 Building $(APP_NAME)..."
	@go build $(GO_FLAGS) -o $(BIN_PATH)/$(APP_NAME) $(APP_CMD)
	@echo "✅ Build complete: $(BIN_PATH)/$(APP_NAME)"

build-linux:
	@mkdir -p $(BIN_PATH)
	@echo "🔨 Building $(APP_NAME) for linux/amd64..."
	@GOOS=linux GOARCH=amd64 go build $(GO_FLAGS) -o $(BIN_PATH)/$(APP_NAME)-linux-amd64 $(APP_CMD)
	@echo "✅ Build complete: $(BIN_PATH)/$(APP_NAME)-linux-amd64"

build-darwin:
	@mkdir -p $(BIN_PATH)
	@echo "🔨 Building $(APP_NAME) for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 go build $(GO_FLAGS) -o $(BIN_PATH)/$(APP_NAME)-darwin-amd64 $(APP_CMD)
	@echo "✅ Build complete: $(BIN_PATH)/$(APP_NAME)-darwin-amd64"

# ======================================
# 🧠 代码生成
# ======================================
generate:
	@echo "⚙️ Generating sqlc code..."
	@sqlc generate
	@echo "✅ sqlc done"

swag:
	@echo "📘 Generating Swagger docs..."
	@swag init -g $(APP_CMD) -o docs --parseDependency --exclude pkg/retry 2>/dev/null || \
	swag init -g $(APP_CMD) -o docs --parseDependency
	@echo "✅ Swagger generated at ./docs"

# ======================================
# 🧪 测试与代码质量
# ======================================
test:
	@echo "🧪 Running unit tests..."
	@go test ./... $(GO_TEST_FLAGS)
	@echo "✅ Tests complete. Coverage report: coverage.out"

test-unit:
	@echo "🧪 Running unit tests without race detector..."
	@go test ./... -cover

test-race:
	@echo "🧪 Running tests with race detector..."
	@go test ./... -race

test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@echo "📊 Coverage report:"
	@go tool cover -func=coverage.out | tail -5

lint:
	@echo "🔍 Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not found. Install with 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"; \
	fi

lint-fix:
	@echo "🔧 Running golangci-lint with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./... --fix; \
	else \
		echo "⚠️  golangci-lint not found. Install with 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"; \
	fi

fmt:
	@echo "🧹 Formatting Go code..."
	@go fmt ./...
	@go vet ./...

fmt-check:
	@echo "🔍 Checking code format..."
	@gofmt -d ./...

# ======================================
# 📦 依赖管理
# ======================================
deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@echo "✅ Dependencies downloaded"

deps-tidy:
	@echo "📦 Tidying dependencies..."
	@go mod tidy
	@echo "✅ Dependencies tidied"

deps-verify:
	@echo "🔍 Verifying dependencies..."
	@go mod verify
	@echo "✅ Dependencies verified"

# ======================================
# 🐳 Docker 管理
# ======================================
docker-up:
	@echo "🐳 Starting docker-compose..."
	docker compose up --build -d

docker-down:
	@echo "🧹 Stopping and removing docker containers..."
	docker compose down -v

docker-logs:
	@echo "📋 Showing docker logs..."
	docker compose logs -f

# ======================================
# 🧼 清理
# ======================================
clean:
	@echo "🧹 Cleaning build outputs..."
	@rm -rf $(BIN_PATH)
	@rm -rf ./docs
	@rm -f coverage.out
	@rm -f *.db
	@echo "✅ Clean complete"

clean-all: clean
	@echo "🧹 Cleaning go cache..."
	@go clean -cache -testcache -modcache
	@echo "✅ All clean complete"

# ======================================
# 🔍 检查
# ======================================
check-deps:
	@echo "🔍 Checking dependencies..."
	@which go >/dev/null || (echo "❌ go not found" && exit 1)
	@which sqlc >/dev/null || echo "⚠️  sqlc not found (run: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest)"
	@which swag >/dev/null || echo "⚠️  swag not found (run: go install github.com/swaggo/swag/cmd/swag@latest)"

check-config:
	@echo "🔍 Checking configuration..."
	@if [ -f .env ]; then echo "✅ .env found"; else echo "⚠️  .env not found (copy .env.example to .env)"; fi

# ======================================
# 📜 帮助信息
# ======================================
help:
	@echo "📖 Makefile Commands:"
	@echo ""
	@echo "🚀 Run & Dev:"
	@echo "  make run           - 运行程序"
	@echo "  make dev           - 开发模式运行 (air热加载)"
	@echo ""
	@echo "🏗️ Build:"
	@echo "  make build         - 构建可执行文件"
	@echo "  make build-linux   - 构建 Linux 版本"
	@echo "  make build-darwin  - 构建 macOS 版本"
	@echo ""
	@echo "🧠 Code Generation:"
	@echo "  make generate      - 运行 sqlc 代码生成"
	@echo "  make swag          - 生成 Swagger 文档"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test          - 运行单元测试 (带 race 检测)"
	@echo "  make test-unit     - 快速单元测试"
	@echo "  make test-race     - 仅 race 检测"
	@echo "  make test-coverage - 生成覆盖率报告"
	@echo ""
	@echo "📦 Dependencies:"
	@echo "  make deps          - 下载依赖"
	@echo "  make deps-tidy     - 整理依赖"
	@echo "  make deps-verify   - 验证依赖"
	@echo ""
	@echo "🧹 Code Quality:"
	@echo "  make fmt           - 格式化与静态检查"
	@echo "  make fmt-check     - 检查格式问题"
	@echo "  make lint          - 运行代码检查"
	@echo "  make lint-fix      - 自动修复代码问题"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  make docker-up     - 启动 docker compose"
	@echo "  make docker-down   - 停止 docker compose"
	@echo "  make docker-logs   - 查看 docker 日志"
	@echo ""
	@echo "🧼 Cleanup:"
	@echo "  make clean         - 清理构建产物"
	@echo "  make clean-all     - 清理所有缓存"
	@echo ""
	@echo "🔍 Check:"
	@echo "  make check-deps    - 检查依赖工具"
	@echo "  make check-config  - 检查配置文件"
	@echo "  make help          - 显示此帮助信息"
