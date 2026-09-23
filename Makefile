BINARY := home
DATA_DIR := ./data

.PHONY: dev-api dev-web build test install clean fmt

## Run the Go API server (frontend served from web/dist if built).
dev-api:
	go run ./cmd/home -addr 127.0.0.1:8080 -data $(DATA_DIR)

## Run the Vite dev server with /api proxied to the Go server on :8080.
dev-web:
	cd web && npm run dev

## Build the embedded frontend, then the single Go binary (no CGO dependency).
build:
	cd web && npm ci --no-audit --no-fund && npm run build
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/home

## Vet and test the Go code.
test:
	go vet ./...
	go test ./...

## Install on a Ubuntu server (run as root, or with sudo).
install: clean
	sudo ./deploy/install.sh

clean:
	rm -f $(BINARY)

## Build caddy+cloudflare binaries for serving a custom domain over Tailscale.
caddy-dist:
	cd deploy/caddy-build && go mod tidy
	cd deploy/caddy-build && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ../dist/caddy-linux-amd64 .
	cd deploy/caddy-build && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o ../dist/caddy-linux-arm64 .
	@echo "built deploy/dist/ — scp these to the server next to deploy/ and run setup-domain.sh"

fmt:
	cd web && npx tsc --noEmit
	gofmt -w ./cmd ./internal ./web/embed.go
