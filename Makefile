.PHONY: run dev build docker-up docker-down

run:
	go run ./cmd/server

dev:
	APP_ADDR=:8080 air || go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v
