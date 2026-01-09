//go:build linux
// +build linux

//
//

// wg-easy-vpn is a tool designed to ease the set-up of a
// WireGuard VPN. In particular you can easily create a server
// and then add clients. You can also export the clients
// configurations through QR codes.
// When your vpn is set up, you just have to invoke `wg-quick`
// for instance.
package main

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/asiffer/wg-easy-vpn/cmd"
)

func main() {
	ctx := context.Background()
	// app is defined in cli.go
	if err := cmd.App.Run(ctx, os.Args); err != nil {
		log.Err(err).Send()
		os.Exit(1)
	}
}
