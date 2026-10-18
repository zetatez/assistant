.PHONY: build build-linux build-darwin run dev install install-linux install-darwin clean fmt swag

OS := $(shell uname -s)

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o assistant ./cmd/assistant/

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o assistant-linux-amd64 ./cmd/assistant/

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o assistant-darwin-arm64 ./cmd/assistant/
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o assistant-darwin-amd64 ./cmd/assistant/

run: build
	@./assistant

dev:
	go run ./cmd/assistant --config config.yaml

# Install for the host OS: systemd --user on Linux, launchd on macOS.
install:
	$(MAKE) swag
	$(MAKE) build
ifeq ($(OS),Darwin)
	$(MAKE) install-darwin
else
	$(MAKE) install-linux
endif

install-linux:
	systemctl --user disable --now assistant 2>/dev/null || true
	sudo mkdir -p /opt/assistant
	sudo cp assistant /opt/assistant/assistant
	rm -f assistant
	rm -rf ~/.config/assistant/scripts
	mkdir -p ~/.config/assistant/scripts
	cp -r scripts/* ~/.config/assistant/scripts/
	chmod +x ~/.config/assistant/scripts/*
	mkdir -p ~/.config/assistant/logs
	cp -f config.default.yaml ~/.config/assistant/config.yaml
	sed -i "s|filename:.*|filename: $$HOME/.config/assistant/logs/assistant.log|" ~/.config/assistant/config.yaml
	mkdir -p ~/.config/systemd/user
	sed "s|%h|$$HOME|g" assistant.service > ~/.config/systemd/user/assistant.service
	systemctl --user daemon-reload 2>/dev/null || true
	systemctl --user enable --now assistant 2>/dev/null || true
	@echo "installed (systemd --user)"

install-darwin:
	launchctl unload ~/Library/LaunchAgents/com.assistant.plist 2>/dev/null || true
	mkdir -p ~/.local/bin ~/.config/assistant/scripts ~/.config/assistant/logs
	cp -f assistant ~/.local/bin/assistant
	chmod +x ~/.local/bin/assistant
	rm -f assistant
	rm -rf ~/.config/assistant/scripts
	mkdir -p ~/.config/assistant/scripts
	cp -r scripts/* ~/.config/assistant/scripts/
	chmod +x ~/.config/assistant/scripts/*
	cp -f config.default.yaml ~/.config/assistant/config.yaml
	mkdir -p ~/Library/LaunchAgents
	sed "s|%h|$$HOME|g" assistant.plist > ~/Library/LaunchAgents/com.assistant.plist
	launchctl load -w ~/Library/LaunchAgents/com.assistant.plist
	@echo "installed (launchd)"

clean:
	@rm -f assistant assistant-linux-amd64 assistant-darwin-arm64 assistant-darwin-amd64 coverage.out

fmt:
	gofmt -w ./cmd ./internal ./pkg && go vet ./...

swag:
	swag init -g ./cmd/assistant/main.go --output docs/
	@echo "Swagger Addr: http://127.0.0.1:4321/swagger/index.html"
