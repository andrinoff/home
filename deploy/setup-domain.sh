#!/usr/bin/env bash
# Serve the home manager on your own domain over Tailscale, with a REAL,
# publicly-trusted HTTPS cert — no public IP, no open ports.
#
#   sudo ./deploy/setup-domain.sh [home.example.com]
#
# How:
#   * An A record points the domain at your server's Tailscale IP (grey-cloud),
#     so it is only reachable inside your tailnet.
#   * Caddy fronting 127.0.0.1:8080 obtains a Let's Encrypt cert via the
#     Cloudflare DNS-01 challenge (a TXT record); needs no inbound ports.
#   * The Cloudflare-capable Caddy comes from deploy/dist (see `make caddy-dist`)
#     or is built here with xcaddy. Without a token, Caddy falls back to its own
#     local CA (trust once per device).
set -euo pipefail

DOMAIN="${1:-home.andrinoff.com}"
APP_TARGET="127.0.0.1:8080"
CONF_DIR="/etc/caddy"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64 | amd64) ARCH_SHORT="amd64" ;;
  aarch64 | arm64) ARCH_SHORT="arm64" ;;
  *) echo "error: unsupported architecture: $ARCH" >&2; exit 1 ;;
esac
PREBUILT="$SCRIPT_DIR/dist/caddy-linux-$ARCH_SHORT"

if [[ $EUID -ne 0 ]]; then
  echo "error: run as root (sudo $0 [$DOMAIN])" >&2
  exit 1
fi

echo "==> Step 0: DNS A record (do this once in your DNS provider)"
TS_IP="$(tailscale ip -4 2>/dev/null | head -n1 || true)"
echo "     Type: A | Name: ${DOMAIN%%.*} | Content: ${TS_IP:-<Tailscale IP of this server>} | Proxy: OFF (grey cloud)"
echo "     (A Tailscale IP also routes only inside your tailnet, so the site stays private.)"

echo "==> Step 1: Install Caddy (official repo) for the systemd unit + config dir"
if ! command -v caddy >/dev/null 2>&1; then
  apt-get install -y --no-install-recommends curl apt-transport-https >/dev/null
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list >/dev/null
  apt-get update
  apt-get install -y caddy
fi

CF_TOKEN="${CF_DNS_API_TOKEN:-}"
if [[ -z "$CF_TOKEN" ]]; then
  while true; do
    read -r -p "Cloudflare API token (Zone:DNS:Edit) for a real cert [Enter to use a local CA]: " CF_TOKEN
    [[ -z "$CF_TOKEN" ]] && break || break
  done
fi

USE_LOCAL_CA=0
if [[ -n "$CF_TOKEN" ]]; then
  echo "==> Step 2: Install a Caddy that can do Cloudflare DNS-01"
  if [[ -f "$PREBUILT" ]]; then
    echo "     using prebuilt binary: $PREBUILT"
    install -m 0755 "$PREBUILT" /usr/bin/caddy
  else
    echo "     no $PREBUILT found — building on the server (needs Go)"
    if command -v go >/dev/null 2>&1; then
      [[ -x "$(go env GOPATH)/bin/xcaddy" ]] || go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
      export PATH="$(go env GOPATH)/bin:$PATH"
      VER="$(caddy version | awk '{print $1}')"
      # note: xcaddy uses --output (no -o shorthand)
      xcaddy build "${VER}" --with github.com/caddy-dns/cloudflare --output /usr/bin/caddy
    else
      echo "error: Go not installed, and no prebuilt Caddy here." >&2
      echo "  Either build one on your Mac:  make caddy-dist && scp <that> deploy/dist/caddy-linux-$ARCH_SHORT  [to the server]" >&2
      exit 1
    fi
  fi
else
  USE_LOCAL_CA=1
  echo "==> Step 2: no token — the stock Caddy is fine (local CA)"
fi

echo "==> Step 3: Write token env + render Caddyfile"
mkdir -p "$CONF_DIR"
if [[ $USE_LOCAL_CA -eq 0 ]]; then
  # token must reach the caddy process even on service restarts
  mkdir -p "/etc/systemd/system/caddy.service.d"
  (umask 177; printf 'CF_DNS_API_TOKEN=%s\n' "$CF_TOKEN" > "$CONF_DIR/env")
  printf '[Service]\nEnvironmentFile=%s/env\n' "$CONF_DIR" > "/etc/systemd/system/caddy.service.d/env.conf"
  systemctl daemon-reload
  TLS_BLOCK='tls {
        dns cloudflare {env.CF_DNS_API_TOKEN}
    }'
else
  TLS_BLOCK='tls internal'
fi

cat > "$CONF_DIR/Caddyfile" <<EOF
$DOMAIN {
    $TLS_BLOCK

    encode zstd gzip

    handle_errors {
        @404 {
            path /api/*
        }
        respond @404 \`{"error":"not found"}\` 404
    }

    reverse_proxy $APP_TARGET
}
EOF
unset CF_TOKEN

echo "==> Step 4: Start Caddy"
systemctl daemon-reload
systemctl enable caddy >/dev/null 2>&1 || true
systemctl restart caddy

sleep 1
if caddy validate --config "$CONF_DIR/Caddyfile" >/dev/null 2>&1; then
  echo "     Caddyfile valid."
else
  echo "     warning: Caddyfile validation reported issues; see journal logs." >&2
fi

echo
echo "==> Done"
if [[ $USE_LOCAL_CA -eq 1 ]]; then
  echo "  No token was supplied, so TLS uses a local CA. Trust it once per device:"
  echo "    sudo cat /var/lib/caddy/.local/share/caddy/pki/authorities/local/root.crt"
fi
echo "  The certificate for $DOMAIN is fetched on the first HTTPS request."
echo "  After a minute, test from any tailnet device:"
echo "    https://$DOMAIN/api/health"
echo "  Debugging:"
echo "    journalctl -u caddy -e --no-pager"
echo "    caddy cert list"
