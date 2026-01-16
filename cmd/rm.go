package cmd

import (
	"context"
	"fmt"

	"github.com/asiffer/wg-easy-vpn/crypto"
	"github.com/asiffer/wg-easy-vpn/models"
	"github.com/asiffer/wg-easy-vpn/utils"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

var rmCmd = cli.Command{
	Name:                  "rm",
	Usage:                 "Remove a client from an existing Wireguard VPN",
	EnableShellCompletion: true,
	Suggest:               true,
	Flags: []cli.Flag{
		&peerFlag,
	},
	Arguments: []cli.Argument{
		&connArg,
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		config, err := buildRmCmdConfig(c)
		if err != nil {
			return err
		}
		return rmAction(ctx, config)
	},
}

type rmConfig struct {
	name    string
	peerKey string
}

func buildRmCmdConfig(c *cli.Command) (*rmConfig, error) {
	cfg := &rmConfig{
		name:    c.StringArg(CONNECTION_ARG),
		peerKey: c.String("peer"),
	}
	log.Debug().
		Str("name", cfg.name).
		Str("peer", cfg.peerKey).
		Msg("rm command configuration")

	return cfg, nil
}

func rmAction(_ context.Context, config *rmConfig) error {
	// Get connection name and path
	_, path, err := ConfigurationInfo(config.name)
	if err != nil {
		return err
	}
	log.Debug().Str("path", path).Msg("Loading VPN configuration")

	// Parse existing VPN configuration
	file, err := utils.ParseFile(path)
	if err != nil {
		return err
	}

	// Load VPN from file
	vpn, err := models.VPNFromFile(config.name, file)
	if err != nil {
		return err
	}

	// Parse the public key
	key := crypto.NewKey()
	err = key.UpdateFromBase64(config.peerKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	// Remove the peer
	err = vpn.RemovePeer(key)
	if err != nil {
		return err
	}
	log.Info().Str("peer", config.peerKey).Msg("Peer removed from VPN")

	// Update server configuration file
	newServerFile := utils.NewFile()
	vpn.Populate(newServerFile)
	err = newServerFile.Save(path)
	if err != nil {
		return err
	}
	log.Info().Str("path", path).Msg("Wireguard VPN configuration updated")

	return nil
}
