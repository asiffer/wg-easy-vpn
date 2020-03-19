// +build linux
//
//

// # wg-easy-vpn
//
// wg-easy-vpn is a tool designed to ease the set-up of a
// WireGuard VPN. In particular you can easily create a server
// and then add clients. You can also export the clients
// configurations through QR codes.
// When your vpn is set up, you just have to invoke `wg-quick`
// for instance.
//
// ## Installation
//
// ### From sources
//
// Basically you download the binary from this repo:
// <pre><code lang="console">go install github.com/asiffer/wg-easy-vpn</code></pre>
//
// The advantage is that the tool is build according to your architecture. The drawback is
// the need to have `Go` installed on your host.
//
// ### Debian package
//
// I also made a debian package to make it available on various debian-like platforms.
//
// ## Usage
//
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
