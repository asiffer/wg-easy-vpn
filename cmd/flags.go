package cmd

import "github.com/urfave/cli/v3"

var connArg = cli.StringArg{
	Name:      CONNECTION_ARG,
	UsageText: "Wireguard connection name (ex: wg0)",
	Config:    cli.StringConfig{TrimSpace: true},
}

var noPSKFlag = cli.BoolFlag{
	Name:  "no-psk",
	Usage: "Do not generate preshared keys",
	Value: false,
}

var endpointFlag = cli.StringFlag{
	Name:     "endpoint",
	Usage:    "Public endpoint (IP or domain) of the Wireguard server (ex: mydomain.com:52820)",
	Required: true,
}

var portFlag = cli.Uint16Flag{
	Name:  "port",
	Usage: "UDP port the Wireguard server will listen on",
	Value: 52820,
}

var networksFlag = cli.StringSliceFlag{
	Name:    "networks",
	Aliases: []string{"n"},
	Usage:   "VPN networks",
	Value:   []string{"10.8.0.0/24"},
}

var routesFlag = cli.StringSliceFlag{
	Name:    "routes",
	Aliases: []string{"r"},
	Usage:   "Routes tunneled through the VPN",
	Value:   []string{"0.0.0.0/0", "::/0"},
}

var dnsFlag = cli.StringSliceFlag{
	Name:  "dns",
	Usage: "DNS servers for the VPN clients",
	Value: nil,
}

var qrcodeFlag = cli.BoolFlag{
	Name:  "qrcode",
	Usage: "export the config to a qrcode",
	Value: false,
}

var peerFlag = cli.StringFlag{
	Name:     "peer",
	Aliases:  []string{"p"},
	Usage:    "Peer to add/remove from the VPN",
	Required: true,
}

var clientFlag = cli.StringFlag{
	Name:     "client",
	Aliases:  []string{"c"},
	Usage:    "New client to add to the VPN",
	Required: true,
}

var wanFlag = cli.StringFlag{
	Name:  "wan",
	Usage: "WAN interface for NAT masquerading (auto = auto-detect, empty = disabled)",
	Value: "",
}
