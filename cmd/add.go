package cmd

import (
	"context"
	"net"
	"os"

	"github.com/asiffer/wg-easy-vpn/models"
	"github.com/asiffer/wg-easy-vpn/utils"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

var addCmd = cli.Command{
	Name:                  "add",
	Usage:                 "Add a new client to an existing Wireguard VPN",
	EnableShellCompletion: true,
	Suggest:               true,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "no-psk",
			Usage: "Do not generate preshared keys",
			Value: false,
		},
		&cli.StringFlag{
			Name:     "client",
			Aliases:  []string{"c"},
			Usage:    "New client to add to the VPN",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:    "routes",
			Aliases: []string{"r"},
			Usage:   "Routes tunneled through the VPN",
			Value:   []string{"0.0.0.0/0", "::/0"},
		},
		&cli.StringSliceFlag{
			Name:        "dns",
			Usage:       "DNS servers for the VPN clients",
			DefaultText: "DNS servers",
			Value:       nil,
		},
		&cli.BoolFlag{
			Name:  "qrcode",
			Usage: "export the config to a qrcode",
			Value: false,
		},
	},
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:   CONNECTION_ARG,
			Config: cli.StringConfig{TrimSpace: true},
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		config, err := buildAddCmdConfig(c)
		if err != nil {
			return err
		}
		return addAction(ctx, config)
	},
}

type addConfig struct {
	name   string
	noPSK  bool
	client string
	routes []net.IPNet
	dns    []net.IP
	qrcode bool
}

func buildAddCmdConfig(c *cli.Command) (*addConfig, error) {
	routes, err := utils.ParseIPNetList(c.StringSlice("routes"))
	if err != nil {
		return nil, err
	}
	dns, err := utils.ParseIPList(c.StringSlice("dns"))
	if err != nil {
		return nil, err
	}
	cfg := &addConfig{
		name:   c.StringArg(CONNECTION_ARG),
		noPSK:  c.Bool("no-psk"),
		client: c.String("client"),
		routes: routes,
		dns:    dns,
		qrcode: c.Bool("qrcode"),
	}
	log.Debug().
		Bool("no-psk", cfg.noPSK).
		Str("client", cfg.client).
		Strs("routes", utils.StringifyNetworks(cfg.routes)).
		Strs("dns", utils.StringifyIPs(cfg.dns)).
		Str("name", cfg.name).
		Bool("qrcode", cfg.qrcode).
		Msg("Add command configuration")
	return cfg, nil
}

func addAction(ctx context.Context, config *addConfig) error {
	// Get connection name and path
	name, path, err := ConfigurationInfo(config.name)
	if err != nil {
		return err
	}
	log.Debug().Str("name", name).Str("path", path).Msg("Parsing connection location")

	// Parse existing VPN configuration
	file, err := utils.ParseFile(path)
	if err != nil {
		return err
	}
	log.Debug().Str("path", path).Msg("Loaded existing VPN configuration")

	// Load VPN from file
	vpn, err := models.VPNFromFile(config.name, file)
	if err != nil {
		return err
	}

	clientName := config.client

	// ips are provided by the vpn when adding the client
	client := models.NewWGClient(nil, config.noPSK, config.dns, config.routes)
	log.Debug().
		Str("client", clientName).
		Bool("no-psk", config.noPSK).
		Strs("dns", utils.StringifyIPs(config.dns)).
		Strs("routes", utils.StringifyNetworks(config.routes)).
		Msg("Creating new client")

	err = vpn.AddClient(client)
	if err != nil {
		return err
	}

	// Get the public key from the peer representation
	peer := client.ToPeer()
	log.Info().
		Str("client", clientName).
		Str("public_key", peer.Public()).
		Msg("Client added to VPN")

	// Prepare client configuration file
	clientFile := utils.NewFile()
	// add the client name inside the file
	sec := clientFile.GetorCreateSection(utils.DEFAULT_SECTION)
	sec.AddComment(clientName)
	// fill the file with client config
	client.PopulateClient(clientFile, vpn)
	clientFile.Log(log.Debug()).Msg("Populating client config file in memory")

	// write to stdout
	if config.qrcode {
		_, err = clientFile.WriteQRCodeTo(os.Stdout)
	} else {
		_, err = clientFile.WriteTo(os.Stdout)
	}
	if err != nil {
		return err
	}

	// update server file
	newServerFile := utils.NewFile()
	vpn.Populate(newServerFile)
	newServerFile.Log(log.Debug()).Msg("Populating server config file in memory")

	err = newServerFile.Save(path)
	if err != nil {
		return err
	}
	log.Info().Str("path", path).Msg("Wireguard VPN configuration file updated")

	return nil
}
