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

# --- Cloudflare API helpers ------------------------------------------------
cf_api() {  # cf_api <METHOD> <PATH> [JSON BODY]
  local method="$1" path="$2" body="${3:-}"
  if [[ -n "$body" ]]; then
    curl -sS -X "$method" "https://api.cloudflare.com/client/v4$path" \
      -H "Authorization: Bearer $CF_TOKEN" -H "Content-Type: application/json" --data "$body"
  else
    curl -sS -X "$method" "https://api.cloudflare.com/client/v4$path" \
      -H "Authorization: Bearer $CF_TOKEN"
  fi
}

json_first_id() {      # first 32-hex "id" in a Cloudflare JSON response
  grep -o "\"id\":\"[0-9a-f]\{32\}\"" | head -n1 | cut -d'"' -f4
}
json_first_content() { # first "content" value in a Cloudflare JSON response
  grep -o '"content":"[^"]*"' | head -n1 | cut -d'"' -f4
}

ensure_dns_a_record() {  # ensure_dns_a_record <tailnet ip>
  local ip="$1"
  local zone="${CF_ZONE_NAME:-${DOMAIN#*.}}"  # home.andrinoff.com -> andrinoff.com
  local zone_id="${CF_ZONE_ID:-}"

  if [[ -z "$ip" ]]; then
    echo "     warning: no Tailscale IPv4 found on this machine; cannot auto-create the A record." >&2
    echo "     Add it manually: Type A | Name ${DOMAIN%%.*} | <tailscale ip> | Proxy OFF" >&2
    return 0
  fi

  if [[ -z "$zone_id" ]]; then
    zone_id="$(cf_api GET "/zones?name=$zone&per_page=1" | json_first_id || true)"
  fi
  if [[ -z "$zone_id" ]]; then
    echo "     warning: could not look up zone '$zone' (token may lack Zone:Read)." >&2
    echo "     Re-run with CF_ZONE_ID=<id> (on the zone overview page), or add the A record manually." >&2
    return 0
  fi

  local payload="{\"type\":\"A\",\"name\":\"$DOMAIN\",\"content\":\"$ip\",\"proxied\":false,\"ttl\":1}"
  local existing rec_id cur_content
  existing="$(cf_api GET "/zones/$zone_id/dns_records?type=A&name=$DOMAIN" || true)"
  rec_id="$(printf '%s' "$existing" | json_first_id || true)"
  cur_content="$(printf '%s' "$existing" | json_first_content || true)"

  if [[ -z "$rec_id" ]]; then
    cf_api POST "/zones/$zone_id/dns_records" "$payload" >/dev/null
    echo "     created A record: $DOMAIN -> $ip (proxied off)"
  elif [[ "$cur_content" != "$ip" ]]; then
    cf_api PUT "/zones/$zone_id/dns_records/$rec_id" "$payload" >/dev/null
    echo "     updated A record: $DOMAIN -> $ip"
  else
    echo "     A record already points at $ip"
  fi
}

echo "==> Step 0: DNS A record"
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

if [[ $USE_LOCAL_CA -eq 0 ]]; then
  echo "==> Step 2.5: Create/update the DNS A record via the Cloudflare API"
  ensure_dns_a_record "$TS_IP"
fi

echo "==> Step 3: Render Caddyfile (token stored in the file, never in env)"
mkdir -p "$CONF_DIR"
if [[ $USE_LOCAL_CA -eq 0 ]]; then
  # IMPORTANT: the token goes into the Caddyfile (kept at mode 0640, root:caddy),
  # NOT into the process environment — Caddy prints its environment at startup,
  # which would write the token into the systemd journal.
  TLS_BLOCK=$'tls {\n\t\tdns cloudflare '"$CF_TOKEN"$'\n\t}'
else
  TLS_BLOCK=$'tls internal'
fi

# Clean up any token left there by earlier versions of this script.
rm -f "$CONF_DIR/env" "/etc/systemd/system/caddy.service.d/env.conf"
rmdir "/etc/systemd/system/caddy.service.d" 2>/dev/null || true
systemctl daemon-reload

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
chown root:caddy "$CONF_DIR/Caddyfile"
chmod 0640 "$CONF_DIR/Caddyfile"
unset CF_TOKEN

echo "==> Step 4: Start Caddy"
systemctl daemon-reload
systemctl enable caddy >/dev/null 2>&1 || true

# Fail fast instead of letting Caddy die on "address already in use".
if ss -lnt 2>/dev/null | grep -q ':443 ' && ! systemctl is-active --quiet caddy; then
  echo "error: something is already listening on port 443." >&2
  if command -v tailscale >/dev/null 2>&1; then
    echo "  If you used 'tailscale serve' earlier, disable it (Caddy takes over HTTPS):" >&2
    echo "    sudo tailscale serve off" >&2
  fi
  echo "  Find the culprit with:  ss -lntp 'sport = :443'" >&2
  exit 1
fi
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
