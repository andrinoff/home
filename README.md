# Home

A small, zero-login home manager: grocery lists, schedule/calendar, events,
tasks, and notes. Designed to run on your own server and be reached only over
[Tailscale](https://tailscale.com) — no public IP, no exposed ports, no account
system.

- **Backend**: Go, single static binary, embeds the frontend.
- **Data**: SQLite (pure-Go driver, so the server needs no gcc/CGO toolchain).
- **Frontend**: React + Vite + TypeScript, compiled and embedded into the binary.
- **Deploy**: one binary + one SQLite file, run as a systemd service.

## Features

| Module | Notes |
| --- | --- |
| Dashboard | Today's tasks, upcoming events, unchecked groceries at a glance |
| Groceries | Multiple lists, quantity, tap-to-check, clear-checked, reorder |
| Calendar | Month grid of events; click a day to add/edit/delete |
| Events | Upcoming list grouped by day, plus a collapsed past section |
| Tasks | Priority, due dates, overdue/today highlighting |
| Notes | Two-pane editor, unsaved-changes guard, Cmd/Ctrl+S to save |

There is **no authentication** — the app binds to `127.0.0.1` only and is meant
to stay inside your tailnet (which already requires authenticated devices).

## Project layout

```
cmd/home/           entry point (flags, server wiring)
internal/store/     SQLite schema + queries per module
internal/api/       REST routes + JSON handlers
web/                React frontend (Vite build -> embedded)
deploy/             systemd unit, optional Caddyfile, install.sh
Makefile            build / dev / test / install
```

## Develop locally

You need Go ≥ 1.22 and Node ≥ 18. Three terminals:

```bash
# terminal 1: Go API on :8080
make dev-api

# terminal 2: Vite dev server on :5173 (proxies /api to :8080)
make dev-web
```

Open http://localhost:5173. HMR is on; data lives in `./data/home.db`.

## Build

```bash
make build
```

Builds `web/dist`, then produces a single self-contained `./home` binary with
`CGO_ENABLED=0`. Run it against the embedded frontend:

```bash
./home -addr 127.0.0.1:8080 -data ./data
```

Flags: `-addr` (default `127.0.0.1:8080`, always loopback) and `-data`
(directory for the `home.db` SQLite file).

## Test

```bash
make test   # go vet + go test (integration tests against a temp DB)
```

## Deploy on your Ubuntu server

### 1. Build and install

```bash
make build
sudo ./deploy/install.sh
```

The installer creates a `home` system user, copies the binary to `/opt/home`,
writes the SQLite DB under `/var/lib/home`, installs and starts a systemd unit.

Verify locally on the server:

```bash
systemctl status home --no-pager     # active (running)
journalctl -u home -e --no-pager     # logs
curl -s 127.0.0.1:8080/api/health    # {"status":"ok"}
```

Upgrading later is just `make build && sudo ./deploy/install.sh` again — it
replaces the binary and restarts the service; your data is untouched.

**Backups**: copy `/var/lib/home/home.db` somewhere safe. That single file is
everything.

---

## 2. Reach it over Tailscale

This is the recommended, zero-config path. Install Tailscale on the server
once, then serve the app to your tailnet:

```bash
# on the server
sudo tailscale up

# enable HTTPS on your tailnet (one-time, in the admin console):
#   https://login.tailscale.com/admin/dns  ->  DNS -> HTTPS Certificates -> Enable HTTPS

# serve 127.0.0.1:8080 at https://<server>.<tailnet>.ts.net/ (detached, survives reboot)
sudo tailscale serve --bg 127.0.0.1:8080
```

Then open `https://<server>.<tailnet>.ts.net` from **any device in your
tailnet** (phone, laptop, whatever) — a real, trusted Let's Encrypt cert, no
browser warnings, nothing exposed to the public internet.

Useful commands:

```bash
tailscale serve status   # what's being served
tailscale serve off      # stop serving
tailscale serve reset    # clear all serve config
```

> Tailscale frees you from managing ports, firewalls and certificates. If you
> don't need a branded domain, stop here.

---

## 3. Use your own domain (`home.andrinoff.com`)

Tailscale's built-in serving only supports `.ts.net` names. To use your own
domain **while staying tailnet-only**, Caddy (run on the server) fronts the app
and issues a real Let's Encrypt certificate via Cloudflare's **DNS-01**
challenge — so no public IP and no open ports are needed.

### Prepare a Cloudflare-capable Caddy (do this on your Mac, once)

```bash
make caddy-dist    # builds deploy/dist/caddy-linux-{amd64,arm64} locally
```

These are the Caddy builtin + `caddy-dns/cloudflare` compiled for Ubuntu, so
the server **does not need Go or xcaddy**. Copy them to the server next to
`deploy/`:

```bash
scp -r deploy/dist user@server:/path/to/home/deploy/
```

### Run the setup on the server

```bash
# on the server, from the repo
sudo ./deploy/setup-domain.sh home.andrinoff.com
```

The script picks up `CF_DNS_API_TOKEN` from the environment if set (or prompts
for it). What it does underneath:

1. **DNS** — add in Cloudflare:
   - Type `A`, name `home`, value = **your server's Tailscale IP**
     (`tailscale ip -4`), proxy **OFF** (grey cloud).
   - A Tailscale IP is only routable inside your tailnet, so this does **not**
     publish the site to the internet.
2. **Install Caddy** from the official repo (`apt install caddy`, for the
   systemd unit + config dir).
3. **Swap in the Cloudflare-capable Caddy** from the `deploy/dist` binaries
   above. (If they are missing and Go is present, it builds one with `xcaddy`
   instead.)
4. **API token** — create a Cloudflare token scoped to `Zone:DNS:Edit` and pass
   it in. Caddy writes it to `/etc/caddy/env` (guarded by a systemd drop-in) and
   requests the certificate for `home.andrinoff.com`.
5. **Visit** `https://home.andrinoff.com` from any tailnet device — trusted
   cert, green lock, no per-device setup.

### Fallback: no token → local CA
If you skip the token, the script uses `tls internal` (Caddy's own private CA).
The site works the same way, but each device must trust Caddy's root cert once:

```bash
sudo cat /var/lib/caddy/.local/share/caddy/pki/authorities/local/root.crt
```

Add that certificate as a trusted root on each phone/laptop.

### Troubleshooting
```bash
journalctl -u caddy -e --no-pager     # caddy errors
caddy cert list                       # obtained certificates
# make sure the A record value equals the Tailscale IP AND proxy is OFF
```

Notes:
- `apt upgrade` can overwrite `/usr/bin/caddy` with a stock build and drop the
  Cloudflare module, which breaks cert renewal. Either `sudo apt-mark hold
  caddy`, or just re-run `sudo ./deploy/setup-domain.sh home.andrinoff.com`
  after upgrading.
- Prefer the prebuilt binaries (`make caddy-dist` → copy `deploy/dist` to the
  server): then the server never needs Go and the setup is fast. The on-server
  `xcaddy` build is only a fallback.

## API summary

```
GET  /api/health
GET  /api/dashboard
GET|POST            /api/grocery/lists
PUT|DELETE          /api/grocery/lists/{id}
GET|POST            /api/grocery/lists/{id}/items
PUT          /api/grocery/items/{id}
PATCH|...    /api/grocery/items/{id}/toggle
DELETE       /api/grocery/items/{id}
POST         /api/grocery/lists/{id}/clear-checked

GET  /api/events?from=ISO&to=ISO
POST /api/events
PUT|DELETE /api/events/{id}

GET  /api/tasks?status=open|done|all
POST /api/tasks
PUT /api/tasks/{id}
PATCH /api/tasks/{id}/toggle
DELETE /api/tasks/{id}

GET|POST /api/notes
GET|PUT|DELETE /api/notes/{id}
```

## Notes & limitations

- SQLite is suited to a single writer — only this binary touches the DB
  (`SetMaxOpenConns(1)` + WAL are set for you).
- Timestamps are stored as RFC3339 (UTC); dates as `YYYY-MM-DD`. The browser
  renders them in the viewer's timezone.
- No auth by design. Keep the app on loopback and only reach it through your
  tailnet. If you ever need it, Caddy can add `basic_auth` with two lines.
