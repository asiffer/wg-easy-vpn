package models

import (
	"net"
	"strings"
	"testing"

	"github.com/asiffer/wg-easy-vpn/utils"
)

func TestNewWGServer(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}

	t.Run("basic creation", func(t *testing.T) {
		server := NewWGServer(ipnet, false, 51820)

		if server.port != 51820 {
			t.Errorf("expected port 51820, got %d", server.port)
		}
		if server.private == nil {
			t.Error("expected private key to be generated")
		}
		if server.psk == nil {
			t.Error("expected PSK to be generated")
		}
	})

	t.Run("without PSK", func(t *testing.T) {
		server := NewWGServer(ipnet, true, 51820)

		if server.psk != nil {
			t.Error("expected no PSK when noPSK=true")
		}
	})

	t.Run("custom port", func(t *testing.T) {
		server := NewWGServer(ipnet, false, 12345)

		if server.port != 12345 {
			t.Errorf("expected port 12345, got %d", server.port)
		}
	})
}

func TestWGServerToPeer(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	server := NewWGServer(ipnet, false, 51820)

	t.Run("with custom routes", func(t *testing.T) {
		routes := []net.IPNet{
			{IP: net.ParseIP("192.168.1.0"), Mask: net.CIDRMask(24, 32)},
			{IP: net.ParseIP("10.10.0.0"), Mask: net.CIDRMask(16, 32)},
		}
		peer := server.ToPeer(routes, "vpn.example.com:51820")

		if peer.endpoint != "vpn.example.com:51820" {
			t.Errorf("expected endpoint 'vpn.example.com:51820', got '%s'", peer.endpoint)
		}

		if len(peer.allowedIPs) != 2 {
			t.Fatalf("expected 2 allowedIPs, got %d", len(peer.allowedIPs))
		}

		// Verify routes are used as allowedIPs
		if peer.allowedIPs[0].String() != "192.168.1.0/24" {
			t.Errorf("expected first route '192.168.1.0/24', got '%s'", peer.allowedIPs[0].String())
		}
	})

	t.Run("with empty routes defaults to all traffic", func(t *testing.T) {
		peer := server.ToPeer(nil, "vpn.example.com:51820")

		if len(peer.allowedIPs) != 2 {
			t.Fatalf("expected 2 allowedIPs (0.0.0.0/0 and ::/0), got %d", len(peer.allowedIPs))
		}

		// Should default to 0.0.0.0/0 and ::/0
		allowedStr := peer.AllowedIPs()
		if !strings.Contains(allowedStr, "0.0.0.0/0") {
			t.Errorf("expected 0.0.0.0/0 in allowedIPs, got '%s'", allowedStr)
		}
		if !strings.Contains(allowedStr, "::/0") {
			t.Errorf("expected ::/0 in allowedIPs, got '%s'", allowedStr)
		}
	})

	t.Run("public key derivation", func(t *testing.T) {
		peer := server.ToPeer(nil, "vpn.example.com:51820")

		if peer.public.Base64() != server.private.Public().Base64() {
			t.Error("peer public key doesn't match server's derived public key")
		}
	})
}

func TestWGServerString(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	server := NewWGServer(ipnet, true, 51820)

	s := server.String()

	if !strings.Contains(s, "Address = ") {
		t.Error("expected Address field in output")
	}
	if !strings.Contains(s, "PrivateKey = ") {
		t.Error("expected PrivateKey field in output")
	}
	if !strings.Contains(s, "ListenPort = 51820") {
		t.Error("expected ListenPort = 51820 in output")
	}
}

func TestWGServerPopulate(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	server := NewWGServer(ipnet, false, 51820)

	section := utils.NewSection("Interface")
	server.Populate(section)

	addr, err := section.Get("Address")
	if err != nil {
		t.Fatalf("failed to get Address: %v", err)
	}
	if addr != "10.0.0.1/24" {
		t.Errorf("expected Address '10.0.0.1/24', got '%s'", addr)
	}

	port, err := section.Get("ListenPort")
	if err != nil {
		t.Fatalf("failed to get ListenPort: %v", err)
	}
	if port != "51820" {
		t.Errorf("expected ListenPort '51820', got '%s'", port)
	}

	privKey, err := section.Get("PrivateKey")
	if err != nil {
		t.Fatalf("failed to get PrivateKey: %v", err)
	}
	if privKey != server.Private() {
		t.Error("PrivateKey mismatch")
	}
}

func TestServerFromSection(t *testing.T) {
	t.Run("valid section", func(t *testing.T) {
		section := utils.NewSection("Interface")
		section.Set("Address", "10.0.0.1/24")
		section.Set("PrivateKey", "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=")
		section.Set("ListenPort", "51820")

		server, err := ServerFromSection(section)
		if err != nil {
			t.Fatalf("failed to parse server from section: %v", err)
		}

		if server.port != 51820 {
			t.Errorf("expected port 51820, got %d", server.port)
		}
		if len(server.address) != 1 {
			t.Errorf("expected 1 address, got %d", len(server.address))
		}
		if server.address[0].IP.String() != "10.0.0.1" {
			t.Errorf("expected IP 10.0.0.1, got %s", server.address[0].IP.String())
		}
	})

	t.Run("missing Address", func(t *testing.T) {
		section := utils.NewSection("Interface")
		section.Set("PrivateKey", "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=")
		section.Set("ListenPort", "51820")

		_, err := ServerFromSection(section)
		if err == nil {
			t.Error("expected error for missing Address")
		}
	})

	t.Run("missing PrivateKey", func(t *testing.T) {
		section := utils.NewSection("Interface")
		section.Set("Address", "10.0.0.1/24")
		section.Set("ListenPort", "51820")

		_, err := ServerFromSection(section)
		if err == nil {
			t.Error("expected error for missing PrivateKey")
		}
	})

	t.Run("missing ListenPort", func(t *testing.T) {
		section := utils.NewSection("Interface")
		section.Set("Address", "10.0.0.1/24")
		section.Set("PrivateKey", "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=")

		_, err := ServerFromSection(section)
		if err == nil {
			t.Error("expected error for missing ListenPort")
		}
	})

	t.Run("invalid PrivateKey", func(t *testing.T) {
		section := utils.NewSection("Interface")
		section.Set("Address", "10.0.0.1/24")
		section.Set("PrivateKey", "invalid-key")
		section.Set("ListenPort", "51820")

		_, err := ServerFromSection(section)
		if err == nil {
			t.Error("expected error for invalid PrivateKey")
		}
	})

	t.Run("dual stack addresses", func(t *testing.T) {
		section := utils.NewSection("Interface")
		section.Set("Address", "10.0.0.1/24, fd00::1/64")
		section.Set("PrivateKey", "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=")
		section.Set("ListenPort", "51820")

		server, err := ServerFromSection(section)
		if err != nil {
			t.Fatalf("failed to parse server: %v", err)
		}

		if len(server.address) != 2 {
			t.Errorf("expected 2 addresses, got %d", len(server.address))
		}
	})
}

func TestServerRoundTrip(t *testing.T) {
	// Create a server, populate a section, parse it back
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	original := NewWGServer(ipnet, true, 51820)

	section := utils.NewSection("Interface")
	original.Populate(section)

	parsed, err := ServerFromSection(section)
	if err != nil {
		t.Fatalf("failed to parse server: %v", err)
	}

	if parsed.port != original.port {
		t.Errorf("port mismatch: expected %d, got %d", original.port, parsed.port)
	}
	if parsed.Private() != original.Private() {
		t.Error("private key mismatch")
	}
	if parsed.Address() != original.Address() {
		t.Errorf("address mismatch: expected %s, got %s", original.Address(), parsed.Address())
	}
}
