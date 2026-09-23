// caddy-build compiles a Caddy binary that bundles the Cloudflare DNS plugin,
// so HTTPS certificates for your own domain can be issued via DNS-01 without
// needing xcaddy or Go on the server. Built by `make caddy-dist`.
package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	// standard modules + plugins
	_ "github.com/caddy-dns/cloudflare"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

func main() {
	caddycmd.Main()
}
