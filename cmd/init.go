package cmd

import (
	"context"
	"net"

	"github.com/asiffer/wg-easy-vpn/models"
	"github.com/asiffer/wg-easy-vpn/utils"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

var initCmd = cli.Command{
	Name:                  "init",
	Usage:                 "Init a new Wireguard VPN (create a server)",
	EnableShellCompletion: true,
	Suggest:               true,
	Flags: []cli.Flag{
		&noPSKFlag,
		&endpointFlag,
		&networksFlag,
		&dnsFlag,
		&routesFlag,
		&portFlag,
	},
	Arguments: []cli.Argument{
		&connArg,
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		config, err := buildInitCmdConfig(c)
		if err != nil {
			return err
		}
		return initAction(ctx, config)
	},
}

type initConfig struct {
	noPSK    bool
	endpoint string
	networks []net.IPNet
	dns      []net.IP
	routes   []net.IPNet
	port     uint16
	conn     string
}

func buildInitCmdConfig(c *cli.Command) (*initConfig, error) {
	networks, err := utils.ParseIPNetList(c.StringSlice("networks"))
	if err != nil {
		return nil, err
	}
	routes, err := utils.ParseIPNetList(c.StringSlice("routes"))
	if err != nil {
		return nil, err
	}
	dns, err := utils.ParseIPList(c.StringSlice("dns"))
	if err != nil {
		return nil, err
	}
	cfg := &initConfig{
		noPSK:    c.Bool("no-psk"),
		endpoint: c.String("endpoint"),
		networks: networks,
		dns:      dns,
		port:     c.Uint16("port"),
		conn:     c.StringArg(CONNECTION_ARG),
		routes:   routes,
	}
	log.Debug().
		Bool("no-psk", cfg.noPSK).
		Str("endpoint", cfg.endpoint).
		Strs("networks", utils.StringifyNetworks(cfg.networks)).
		Strs("routes", utils.StringifyNetworks(cfg.routes)).
		Strs("dns", utils.StringifyIPs(cfg.dns)).
		Uint16("port", cfg.port).
		Str("conn", cfg.conn).
		Msg("Init command configuration")

	return cfg, nil
}

func initAction(_ context.Context, config *initConfig) error {
	// name of the connection and path to the config file
	name, path, err := ConfigurationInfo(config.conn)
	if err != nil {
		return err
	}
	log.Debug().Str("name", name).Str("path", path).Msg("Parsing connection location")

	// create the server first (without networks)
	server := models.NewWGServer(nil, config.noPSK, config.port)

	// create the VPN. It provides the networks to the server
	vpn, err := models.NewWGVPN(
		name,
		server,
		config.endpoint,
		config.networks,
		config.dns,
		config.routes,
	)
	if err != nil {
		return err
	}
	vpn.Log(log.Debug()).Msg("Creating new vpn")

	file := utils.NewFile()
	vpn.Populate(file)
	file.Log(log.Debug()).Msg("Populating config file in memory")
	if err := file.Save(path); err != nil {
		return err
	}
	log.Info().Str("path", path).Msg("Wireguard VPN configuration file created")
	return nil
}
