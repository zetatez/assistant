APP_NAME := assistant
APP_CMD := ./cmd/assistant/main.go
BIN_PATH := ./bin

BUILD_TIME := $(shell date +'%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_LDFLAGS := -s -w -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'

.PHONY: all build run dev test lint fmt generate swag docker-up docker-down clean help

all: build

build:
	@mkdir -p $(BIN_PATH)
	@echo "Building $(APP_NAME)..."
	@CGO_ENABLED=0 go build -ldflags="$(GO_LDFLAGS)" -o $(BIN_PATH)/$(APP_NAME) $(APP_CMD)

run: build
	@$(BIN_PATH)/$(APP_NAME)

dev:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found. Install: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi

test:
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...

lint:
	@golangci-lint run ./... || echo "golangci-lint not found"

fmt:
	@go fmt ./... && go vet ./...

generate:
	@sqlc generate

swag:
	@swag init -g $(APP_CMD) -o docs --parseDependency 2>/dev/null || \
   swag init -g $(APP_CMD) -o docs --parseDependency

docker-up:
	@docker compose up --build -d

docker-down:
	@docker compose down -v

clean:
	@rm -rf $(BIN_PATH) ./docs coverage.out *.db

help:
	@grep -E '^[a-zA-Z_-]+:' $(MAKEFILE_LIST) | sort | while read line; do \
		printf "%-20s\n" $$line; \
	done
