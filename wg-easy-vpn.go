// +build linux
//
//

//
// wg-easy-vpn is a tool designed to ease the set-up of a
// WireGuard VPN. In particular you can easily create a server
// and then add clients. You can also export the clients
// configurations through QR codes.
// When your vpn is set up, you just have to invoke `wg-quick`
// for instance.
//
package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func main() {
	fmt.Println(os.Args)
	// app is defined in cli.go
	if err := app.Run(os.Args); err != nil {
		color.Red("%v", err)
	}

	if doc, err := app.ToMarkdown(); err == nil {
		fmt.Println(doc)
	}

}
