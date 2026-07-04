package cmd

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.port != DefaultListeningPort {
		t.Errorf("expected default port %d, got %d", DefaultListeningPort, cfg.port)
	}
	if cfg.noPSK != false {
		t.Error("expected noPSK to default to false")
	}
	if cfg.qrcode != false {
		t.Error("expected qrcode to default to false")
	}
	if len(cfg.routes) != 2 {
		t.Errorf("expected 2 default routes, got %d", len(cfg.routes))
	}
	if len(cfg.networks) != 1 {
		t.Errorf("expected 1 default network, got %d", len(cfg.networks))
	}
}

func TestConfigNetworks(t *testing.T) {
	cfg := defaultConfig()
	networks := cfg.Networks()

	if len(networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(networks))
	}

	// Default network from config.go: DefaultNetwork = "10.8.0.1/24"
	// But when parsed, the IP is masked to network address: 10.8.0.0/24
	expected := "10.8.0.0/24"
	if networks[0].String() != expected {
		t.Errorf("expected network %s, got %s", expected, networks[0].String())
	}
}

func TestConfigRoutes(t *testing.T) {
	cfg := defaultConfig()
	routes := cfg.Routes()

	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}

	// Should have IPv4 and IPv6 catch-all routes
	hasIPv4 := false
	hasIPv6 := false
	for _, r := range routes {
		if r.String() == "0.0.0.0/0" {
			hasIPv4 = true
		}
		if r.String() == "::/0" {
			hasIPv6 = true
		}
	}
	if !hasIPv4 {
		t.Error("missing IPv4 default route")
	}
	if !hasIPv6 {
		t.Error("missing IPv6 default route")
	}
}

func TestConfigDNS(t *testing.T) {
	cfg := defaultConfig()
	dns := cfg.DNS()

	// Default has no DNS
	if len(dns) != 0 {
		t.Errorf("expected 0 DNS servers by default, got %d", len(dns))
	}
}
