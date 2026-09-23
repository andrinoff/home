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

## 3. Optional: use your own domain (`home.andrinoff.com`)

Tailscale's built-in serving only supports its `.ts.net` names. To use your
own domain **while staying tailnet-only**, run a reverse proxy on the server and
point the domain at the server's Tailscale IP. Caddy is included in `deploy/`.

1. **DNS**: in your DNS provider, add
   - Type `A`, name `home` (→ `home.andrinoff.com`), value = **your server's
     Tailscale IP** (the `100.x.y.z` from `tailscale ip -4`), proxy **off**
     (grey cloud).
   - Tailscale IPs are only routable from inside your tailnet, so this does
     **not** publish the site to the internet.

2. **Install Caddy** (official repo):

   ```bash
   sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
   sudo apt update && sudo apt install caddy
   ```

3. **Configure**: copy `deploy/Caddyfile` to `/etc/caddy/Caddyfile` (edit the
   domain) and reload:

   ```bash
   sudo cp deploy/Caddyfile /etc/caddy/Caddyfile
   sudo systemctl reload caddy
   ```

4. **Trust the local CA once per device** so the browser shows a green lock
   (Caddy uses `tls internal`, i.e. its own private CA):

   ```bash
   # on the server, print the root cert
   sudo cat /var/lib/caddy/.local/share/caddy/pki/authorities/local/root.crt
   ```

   Add that certificate as a **trusted root** on each device that will browse
   the site. (Alternative for a real public cert without this step: build
   Caddy with the `caddy-dns/cloudflare` plugin and use `tls { dns cloudflare
   <token> }` — see comments in the Caddyfile.)

5. Visit `https://home.andrinoff.com` from a device connected to your tailnet.

### Which to pick?

- **`.ts.net` + `tailscale serve`**: zero configuration, real cert, recommend
  for most people.
- **Custom domain + Caddy**: nicer name, but requires the one-time CA-trust
  step on each device (or a plugin build for a public cert).

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
