package cmd

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/asiffer/wg-easy-vpn/utils"
)

func TestInitAction(t *testing.T) {
	t.Run("creates valid server config", func(t *testing.T) {
		dir := testDir(t)
		configPath := testConfigPath(t, dir, "wg0")

		config := &initConfig{
			noPSK:    false,
			endpoint: "vpn.example.com:51820",
			networks: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
			dns:      []net.IP{net.ParseIP("1.1.1.1")},
			routes:   []net.IPNet{{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)}},
			port:     51820,
			conn:     configPath,
		}

		err := initAction(context.Background(), config)
		if err != nil {
			t.Fatalf("initAction failed: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatal("config file was not created")
		}

		// Parse and verify content
		file, err := utils.ParseFile(configPath)
		if err != nil {
			t.Fatalf("failed to parse created config: %v", err)
		}

		// Check Interface section exists
		sections := file.Sections()
		hasInterface := false
		for _, sec := range sections {
			if sec.Name() == "Interface" {
				hasInterface = true

				// Verify essential fields
				if !sec.HasKey("Address") {
					t.Error("Interface section missing Address")
				}
				if !sec.HasKey("PrivateKey") {
					t.Error("Interface section missing PrivateKey")
				}
				if !sec.HasKey("ListenPort") {
					t.Error("Interface section missing ListenPort")
				}

				port, _ := sec.Get("ListenPort")
				if port != "51820" {
					t.Errorf("expected ListenPort 51820, got %s", port)
				}
			}
		}
		if !hasInterface {
			t.Error("config missing Interface section")
		}

		// Check metadata in default section
		defaultSec := file.GetorCreateSection(utils.DEFAULT_SECTION)
		endpoint, err := defaultSec.Get("Endpoint")
		if err != nil {
			t.Error("missing Endpoint in metadata")
		} else if endpoint != "vpn.example.com:51820" {
			t.Errorf("expected endpoint 'vpn.example.com:51820', got '%s'", endpoint)
		}
	})

	t.Run("creates config without PSK", func(t *testing.T) {
		dir := testDir(t)
		configPath := testConfigPath(t, dir, "wg-nopsk")

		config := &initConfig{
			noPSK:    true,
			endpoint: "vpn.example.com:51820",
			networks: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
			port:     51820,
			conn:     configPath,
		}

		err := initAction(context.Background(), config)
		if err != nil {
			t.Fatalf("initAction failed: %v", err)
		}

		// File should exist
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatal("config file was not created")
		}
	})

	t.Run("creates config with custom port", func(t *testing.T) {
		dir := testDir(t)
		configPath := testConfigPath(t, dir, "wg-custom-port")

		config := &initConfig{
			noPSK:    false,
			endpoint: "vpn.example.com:12345",
			networks: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
			port:     12345,
			conn:     configPath,
		}

		err := initAction(context.Background(), config)
		if err != nil {
			t.Fatalf("initAction failed: %v", err)
		}

		file, _ := utils.ParseFile(configPath)
		for _, sec := range file.Sections() {
			if sec.Name() == "Interface" {
				port, _ := sec.Get("ListenPort")
				if port != "12345" {
					t.Errorf("expected ListenPort 12345, got %s", port)
				}
			}
		}
	})

	t.Run("creates config with multiple networks", func(t *testing.T) {
		dir := testDir(t)
		configPath := testConfigPath(t, dir, "wg-multi-net")

		config := &initConfig{
			noPSK:    false,
			endpoint: "vpn.example.com:51820",
			networks: []net.IPNet{
				{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)},
				{IP: net.ParseIP("192.168.100.0"), Mask: net.CIDRMask(24, 32)},
			},
			port: 51820,
			conn: configPath,
		}

		err := initAction(context.Background(), config)
		if err != nil {
			t.Fatalf("initAction failed: %v", err)
		}

		file, _ := utils.ParseFile(configPath)
		defaultSec := file.GetorCreateSection(utils.DEFAULT_SECTION)
		network, _ := defaultSec.Get("Network")

		if !strings.Contains(network, "10.0.0.0/24") {
			t.Errorf("expected network to contain 10.0.0.0/24, got %s", network)
		}
		if !strings.Contains(network, "192.168.100.0/24") {
			t.Errorf("expected network to contain 192.168.100.0/24, got %s", network)
		}
	})

	t.Run("creates config with DNS servers", func(t *testing.T) {
		dir := testDir(t)
		configPath := testConfigPath(t, dir, "wg-dns")

		config := &initConfig{
			noPSK:    false,
			endpoint: "vpn.example.com:51820",
			networks: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
			dns:      []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("8.8.8.8")},
			port:     51820,
			conn:     configPath,
		}

		err := initAction(context.Background(), config)
		if err != nil {
			t.Fatalf("initAction failed: %v", err)
		}

		file, _ := utils.ParseFile(configPath)
		defaultSec := file.GetorCreateSection(utils.DEFAULT_SECTION)
		dns, err := defaultSec.Get("DNS")
		if err != nil {
			t.Error("DNS not found in metadata")
		}
		if !strings.Contains(dns, "1.1.1.1") || !strings.Contains(dns, "8.8.8.8") {
			t.Errorf("expected DNS servers, got %s", dns)
		}
	})
}
