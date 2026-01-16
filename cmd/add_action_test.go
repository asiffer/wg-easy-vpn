package cmd

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asiffer/wg-easy-vpn/utils"
)

// setupVPN creates an initialized VPN for add tests
func setupVPN(t *testing.T, dir string) string {
	t.Helper()
	configPath := testConfigPath(t, dir, "wg0")
	initCfg := &initConfig{
		noPSK:    false,
		endpoint: "vpn.example.com:51820",
		networks: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
		dns:      []net.IP{net.ParseIP("1.1.1.1")},
		routes:   []net.IPNet{{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)}},
		port:     51820,
		conn:     configPath,
	}
	if err := initAction(context.Background(), initCfg); err != nil {
		t.Fatalf("setup VPN failed: %v", err)
	}
	return configPath
}

func TestAddAction(t *testing.T) {
	t.Run("adds client to VPN", func(t *testing.T) {
		dir := testDir(t)
		configPath := setupVPN(t, dir)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		addCfg := &addConfig{
			name:   configPath,
			noPSK:  false,
			client: "client1",
			qrcode: false,
		}

		err := addAction(context.Background(), addCfg)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Fatalf("addAction failed: %v", err)
		}

		// Read captured output
		buf := make([]byte, 4096)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		// Output should contain client config
		if !strings.Contains(output, "[Interface]") {
			t.Error("expected client output to contain [Interface]")
		}
		if !strings.Contains(output, "[Peer]") {
			t.Error("expected client output to contain [Peer]")
		}
		if !strings.Contains(output, "PrivateKey") {
			t.Error("expected client output to contain PrivateKey")
		}

		// Server config should now have a peer
		file, _ := utils.ParseFile(configPath)
		peerCount := 0
		for _, sec := range file.Sections() {
			if sec.Name() == "Peer" {
				peerCount++
			}
		}
		if peerCount != 1 {
			t.Errorf("expected 1 peer in server config, got %d", peerCount)
		}
	})

	t.Run("adds multiple clients", func(t *testing.T) {
		dir := testDir(t)
		configPath := setupVPN(t, dir)

		// Suppress stdout
		oldStdout := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		// Add first client
		addCfg1 := &addConfig{
			name:   configPath,
			client: "client1",
		}
		if err := addAction(context.Background(), addCfg1); err != nil {
			t.Fatalf("addAction for client1 failed: %v", err)
		}

		// Add second client
		addCfg2 := &addConfig{
			name:   configPath,
			client: "client2",
		}
		if err := addAction(context.Background(), addCfg2); err != nil {
			t.Fatalf("addAction for client2 failed: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		// Server should have 2 peers
		file, _ := utils.ParseFile(configPath)
		peerCount := 0
		for _, sec := range file.Sections() {
			if sec.Name() == "Peer" {
				peerCount++
			}
		}
		if peerCount != 2 {
			t.Errorf("expected 2 peers, got %d", peerCount)
		}
	})

	t.Run("client gets sequential IP", func(t *testing.T) {
		dir := testDir(t)
		configPath := setupVPN(t, dir)

		// Suppress stdout
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

		// Check server config for peer's AllowedIPs
		file, _ := utils.ParseFile(configPath)
		for _, sec := range file.Sections() {
			if sec.Name() == "Peer" {
				allowed, _ := sec.Get("AllowedIPs")
				// Server has 10.0.0.1, client should get 10.0.0.2
				if !strings.Contains(allowed, "10.0.0.2") {
					t.Errorf("expected client to get 10.0.0.2, got %s", allowed)
				}
			}
		}
	})

	t.Run("adds client with custom routes", func(t *testing.T) {
		dir := testDir(t)
		configPath := setupVPN(t, dir)

		// Capture stdout to check client config
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		addCfg := &addConfig{
			name:   configPath,
			client: "client-custom-routes",
			routes: []net.IPNet{{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)}},
		}

		err := addAction(context.Background(), addCfg)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Fatalf("addAction failed: %v", err)
		}

		buf := make([]byte, 4096)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		// Client config should have custom AllowedIPs for server peer
		if !strings.Contains(output, "192.168.0.0/16") {
			t.Errorf("expected custom route in client config, got: %s", output)
		}
	})

	t.Run("adds client with custom DNS", func(t *testing.T) {
		dir := testDir(t)
		configPath := setupVPN(t, dir)

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		addCfg := &addConfig{
			name:   configPath,
			client: "client-dns",
			dns:    []net.IP{net.ParseIP("9.9.9.9")},
		}

		addAction(context.Background(), addCfg)

		w.Close()
		os.Stdout = oldStdout

		buf := make([]byte, 4096)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		if !strings.Contains(output, "DNS") || !strings.Contains(output, "9.9.9.9") {
			t.Errorf("expected DNS 9.9.9.9 in client config, got: %s", output)
		}
	})

	t.Run("fails on non-existent config", func(t *testing.T) {
		dir := testDir(t)
		nonExistentPath := filepath.Join(dir, "does-not-exist.conf")

		addCfg := &addConfig{
			name:   nonExistentPath,
			client: "client1",
		}

		err := addAction(context.Background(), addCfg)
		if err == nil {
			t.Error("expected error for non-existent config")
		}
	})
}
