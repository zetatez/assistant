#!/bin/bash
set -e

REPO="zetatez/suckless-dwm"
BRANCH="master"
SRC_DIR="assistant"

CONFIG_DIR="$HOME/.config/assistant"
LOG_DIR="$CONFIG_DIR/logs"

OS="$(uname -s)"

if command -v curl >/dev/null 2>&1; then
	DL="curl -sL"
else
	DL="wget -qO-"
fi

echo "==> Downloading $REPO..."
TMP=$(mktemp -d)
$DL "https://github.com/$REPO/archive/refs/heads/$BRANCH.tar.gz" | tar xz -C "$TMP"
cd "$TMP/suckless-dwm-$BRANCH/$SRC_DIR"

echo "==> Building..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o assistant ./cmd/assistant/

# Portable in-place sed (BSD sed on macOS requires an explicit backup suffix).
sed_inplace() {
	local expr="$1" file="$2"
	sed "$expr" "$file" >"$file.tmp" && mv "$file.tmp" "$file"
}

if [ "$OS" = "Darwin" ]; then
	PLIST="$HOME/Library/LaunchAgents/com.assistant.plist"
	BIN="$HOME/.local/bin/assistant"

	echo "==> Stopping launchd agent..."
	launchctl unload "$PLIST" 2>/dev/null || true

	echo "==> Installing binary..."
	mkdir -p "$HOME/.local/bin"
	cp assistant "$BIN"
	chmod +x "$BIN"

	echo "==> Installing scripts..."
	rm -rf "$CONFIG_DIR/scripts"
	mkdir -p "$CONFIG_DIR/scripts"
	cp -r scripts/* "$CONFIG_DIR/scripts/"
	chmod +x "$CONFIG_DIR/scripts/"*

	echo "==> Configuring..."
	mkdir -p "$CONFIG_DIR" "$LOG_DIR"
	if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
		cp config.default.yaml "$CONFIG_DIR/config.yaml"
		echo "    created $CONFIG_DIR/config.yaml"
		echo "    >> edit it to set your auth password and LLM API key <<"
	else
		echo "    keeping existing $CONFIG_DIR/config.yaml"
	fi
	sed_inplace "s|filename:.*|filename: $LOG_DIR/assistant.log|" "$CONFIG_DIR/config.yaml"

	echo "==> Installing launchd agent..."
	mkdir -p "$(dirname "$PLIST")"
	sed "s|%h|$HOME|g" assistant.plist >"$PLIST"
	launchctl load -w "$PLIST"

	echo ""
	echo "================================"
	echo "  Install complete!"
	echo "================================"
	echo ""
	echo "  Binary:     $BIN"
	echo "  Config:     $CONFIG_DIR/config.yaml"
	echo "  Logs:       $LOG_DIR"
	echo ""
	echo "  Status:     launchctl list | grep assistant"
	echo "  Logs:       tail -f $LOG_DIR/assistant.out.log"
	echo "  API:        curl -u user:pass http://127.0.0.1:4321/api/svr/get-datetime"
	echo ""
	echo "  macOS permissions: allow Terminal/assistant in"
	echo "  System Settings > Privacy & Security > Accessibility / Screen Recording"
	echo "  for popup editing, screenshots and notifications."
	echo ""
else
	INSTALL_DIR="/opt/assistant"
	SERVICE_FILE="$HOME/.config/systemd/user/assistant.service"

	echo "==> Stopping service..."
	systemctl --user disable --now assistant 2>/dev/null || true

	echo "==> Installing binary..."
	sudo mkdir -p "$INSTALL_DIR"
	sudo cp assistant "$INSTALL_DIR/"

	echo "==> Installing scripts..."
	rm -rf "$CONFIG_DIR/scripts"
	mkdir -p "$CONFIG_DIR/scripts"
	cp -r scripts/* "$CONFIG_DIR/scripts/"
	chmod +x "$CONFIG_DIR/scripts/"*

	echo "==> Configuring..."
	mkdir -p "$CONFIG_DIR" "$LOG_DIR"
	if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
		cp config.default.yaml "$CONFIG_DIR/config.yaml"
		echo "    created $CONFIG_DIR/config.yaml"
		echo "    >> edit it to set your auth password and LLM API key <<"
	else
		echo "    keeping existing $CONFIG_DIR/config.yaml"
	fi
	sed_inplace "s|filename:.*|filename: $LOG_DIR/assistant.log|" "$CONFIG_DIR/config.yaml"

	echo "==> Installing systemd service..."
	mkdir -p "$(dirname "$SERVICE_FILE")"
	sed "s|%h|$HOME|g; s|/opt/assistant/assistant|$INSTALL_DIR/assistant|" assistant.service >"$SERVICE_FILE"
	systemctl --user daemon-reload

	echo "==> Starting service..."
	systemctl --user enable --now assistant 2>/dev/null || true

	echo ""
	echo "================================"
	echo "  Install complete!"
	echo "================================"
	echo ""
	echo "  Binary:     $INSTALL_DIR/assistant"
	echo "  Config:     $CONFIG_DIR/config.yaml"
	echo "  Logs:       $LOG_DIR"
	echo ""
	echo "  Status:     systemctl --user status assistant"
	echo "  Logs:       journalctl --user -fu assistant"
	echo "  API:        curl -u user:pass http://127.0.0.1:4321/api/svr/get-datetime"
	echo ""
fi

rm -rf "$TMP"
