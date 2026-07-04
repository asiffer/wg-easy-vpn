package cmd

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/asiffer/wg-easy-vpn/utils"
)

// setupVPNWithClient creates a VPN with one client and returns the config path and peer's public key
func setupVPNWithClient(t *testing.T, dir string) (string, string) {
	t.Helper()
	configPath := testConfigPath(t, dir, "wg0")

	// Init VPN
	initCfg := &initConfig{
		noPSK:    false,
		endpoint: "vpn.example.com:51820",
		networks: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
		port:     51820,
		conn:     configPath,
	}
	initAction(context.Background(), initCfg)

	// Add client and capture its public key
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	addCfg := &addConfig{
		name:   configPath,
		client: "client1",
	}
	addAction(context.Background(), addCfg)

	w.Close()
	os.Stdout = oldStdout

	// Get the peer's public key from server config
	file, _ := utils.ParseFile(configPath)
	var peerKey string
	for _, sec := range file.Sections() {
		if sec.Name() == "Peer" {
			peerKey, _ = sec.Get("PublicKey")
			break
		}
	}

	return configPath, peerKey
}

func TestRmAction(t *testing.T) {
	t.Run("removes peer from VPN", func(t *testing.T) {
		dir := testDir(t)
		configPath, peerKey := setupVPNWithClient(t, dir)

		// Verify peer exists
		file, _ := utils.ParseFile(configPath)
		peerCountBefore := 0
		for _, sec := range file.Sections() {
			if sec.Name() == "Peer" {
				peerCountBefore++
			}
		}
		if peerCountBefore != 1 {
			t.Fatalf("expected 1 peer before removal, got %d", peerCountBefore)
		}

		// Remove the peer
		rmCfg := &rmConfig{
			name:    configPath,
			peerKey: peerKey,
		}

		err := rmAction(context.Background(), rmCfg)
		if err != nil {
			t.Fatalf("rmAction failed: %v", err)
		}

		// Verify peer is gone
		file, _ = utils.ParseFile(configPath)
		peerCountAfter := 0
		for _, sec := range file.Sections() {
			if sec.Name() == "Peer" {
				peerCountAfter++
			}
		}
		if peerCountAfter != 0 {
			t.Errorf("expected 0 peers after removal, got %d", peerCountAfter)
		}
	})

	t.Run("fails with invalid public key", func(t *testing.T) {
		dir := testDir(t)
		configPath, _ := setupVPNWithClient(t, dir)

		rmCfg := &rmConfig{
			name:    configPath,
			peerKey: "invalid-key",
		}

		err := rmAction(context.Background(), rmCfg)
		if err == nil {
			t.Error("expected error for invalid public key")
		}
	})

	t.Run("fails with non-existent peer", func(t *testing.T) {
		dir := testDir(t)
		configPath, _ := setupVPNWithClient(t, dir)

		// Use a valid but non-existent key
		rmCfg := &rmConfig{
			name:    configPath,
			peerKey: "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=",
		}

		err := rmAction(context.Background(), rmCfg)
		if err == nil {
			t.Error("expected error for non-existent peer")
		}
	})

	t.Run("fails on non-existent config", func(t *testing.T) {
		dir := testDir(t)
		nonExistentPath := filepath.Join(dir, "does-not-exist.conf")

		rmCfg := &rmConfig{
			name:    nonExistentPath,
			peerKey: "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=",
		}

		err := rmAction(context.Background(), rmCfg)
		if err == nil {
			t.Error("expected error for non-existent config")
		}
	})
}
