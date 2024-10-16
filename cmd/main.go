//go:build exclude

package main

import (
	_ "github.com/Mirror0/caddy-langneg"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

func main() {
	caddycmd.Main()
}
