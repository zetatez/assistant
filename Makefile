.PHONY: run dev build docker-up docker-down


generate:
	sqlc generate

run:
	go run ./cmd/server/main.go

dev:
	APP_ADDR=:8080 air || go run ./cmd/server

swag:
	swag init -g cmd/server/main.go -o docs --parseDependency --exclude pkg/retry

build:
	go build -o bin/server ./cmd/server/main.go

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v
