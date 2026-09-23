#!/usr/bin/env bash
# One-shot installer for the home manager on Ubuntu.
# Usage: sudo ./deploy/install.sh   (build the binary first with `make build`)
set -euo pipefail

BIN="home"
BIN_DIR="/opt/home"
DATA_DIR="/var/lib/home"
USER="home"
SERVICE="home.service"

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ $EUID -ne 0 ]]; then
  echo "error: run as root (sudo ./deploy/install.sh)" >&2
  exit 1
fi

if [[ ! -f "$REPO_DIR/$BIN" ]]; then
  echo "error: '$BIN' not found in $REPO_DIR — run 'make build' first." >&2
  exit 1
fi

echo "==> Creating service user"
if ! id -u "$USER" >/dev/null 2>&1; then
  useradd --system --home-dir "$BIN_DIR" --shell /usr/sbin/nologin "$USER"
fi

echo "==> Installing binary + data dir"
install -d -m 0755 "$BIN_DIR"
install -d -m 0750 -o "$USER" -g "$USER" "$DATA_DIR"
install -m 0755 "$REPO_DIR/$BIN" "$BIN_DIR/$BIN"

echo "==> Installing systemd unit"
install -m 0644 "$REPO_DIR/deploy/$SERVICE" "/etc/systemd/system/$SERVICE"
systemctl daemon-reload

echo "==> Starting service"
systemctl enable --now "$SERVICE"

sleep 1
if curl -fsS http://127.0.0.1:8080/api/health >/dev/null 2>&1; then
  echo "OK — the app is up at http://127.0.0.1:8080"
else
  echo "Started, but the health check failed. Check:"
  echo "  systemctl status $SERVICE --no-pager"
  echo "  journalctl -u $SERVICE -e --no-pager"
fi

echo
echo "Next: expose it over Tailscale (see README 'Deployment')."
echo "  sudo tailscale up"
echo "  sudo tailscale serve --bg 127.0.0.1:8080"
