package cmd

import (
	"context"
	"fmt"
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
		&wanFlag,
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
	wan      string // WAN interface for NAT masquerading (empty = disabled, non-empty = interface name)
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
		wan:      c.String("wan"),
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

	// configure WAN masquerading if requested
	if config.wan != "" {
		wanIface := config.wan
		if wanIface == "auto" {
			detected, err := utils.GetDefaultInterface()
			if err != nil {
				return fmt.Errorf("failed to auto-detect WAN interface: %w", err)
			}
			wanIface = detected
			log.Debug().Str("interface", wanIface).Msg("Auto-detected WAN interface")
		}
		preUp, postDown := generateMasqueradeHooks(config.networks, wanIface)
		server.SetHooks(preUp, postDown)
		log.Debug().Strs("preUp", preUp).Strs("postDown", postDown).Msg("WAN masquerading hooks configured")
	}

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

// generateMasqueradeHooks generates PreUp and PostDown commands for NAT masquerading.
// It configures:
//   - IP forwarding (sysctl) only if not already enabled on the system
//   - iptables/ip6tables MASQUERADE rules for each network
//
// If IP forwarding was disabled before, it will be restored to 0 in PostDown.
func generateMasqueradeHooks(networks []net.IPNet, wanIface string) (preUp, postDown []string) {
	needsIPv4Forwarding := false
	needsIPv6Forwarding := false

	for _, network := range networks {
		if network.IP.To4() != nil {
			needsIPv4Forwarding = true
		} else {
			needsIPv6Forwarding = true
		}
	}

	// Enable IP forwarding only if not already enabled
	if needsIPv4Forwarding && !utils.IsIPv4ForwardingEnabled() {
		preUp = append(preUp, "sysctl -q -w net.ipv4.ip_forward=1")
		postDown = append(postDown, "sysctl -q -w net.ipv4.ip_forward=0")
	}
	if needsIPv6Forwarding && !utils.IsIPv6ForwardingEnabled() {
		preUp = append(preUp, "sysctl -q -w net.ipv6.conf.all.forwarding=1")
		postDown = append(postDown, "sysctl -q -w net.ipv6.conf.all.forwarding=0")
	}

	// Add masquerade rules
	for _, network := range networks {
		if network.IP.To4() != nil {
			preUp = append(preUp, fmt.Sprintf("iptables -t nat -A POSTROUTING -s %s -o %s -j MASQUERADE", network.String(), wanIface))
			postDown = append(postDown, fmt.Sprintf("iptables -t nat -D POSTROUTING -s %s -o %s -j MASQUERADE", network.String(), wanIface))
		} else {
			preUp = append(preUp, fmt.Sprintf("ip6tables -t nat -A POSTROUTING -s %s -o %s -j MASQUERADE", network.String(), wanIface))
			postDown = append(postDown, fmt.Sprintf("ip6tables -t nat -D POSTROUTING -s %s -o %s -j MASQUERADE", network.String(), wanIface))
		}
	}
	return
}
