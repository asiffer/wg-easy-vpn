package models

import (
	"net"
	"strings"
	"testing"

	"github.com/asiffer/wg-easy-vpn/utils"
)

func TestNewWGClient(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}}
	dns := []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("8.8.8.8")}
	routes := []net.IPNet{{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)}}

	t.Run("with all options", func(t *testing.T) {
		client := NewWGClient(ipnet, false, dns, routes)

		if client.private == nil {
			t.Error("expected private key to be generated")
		}
		if client.psk == nil {
			t.Error("expected PSK to be generated")
		}
		if len(client.dns) != 2 {
			t.Errorf("expected 2 DNS servers, got %d", len(client.dns))
		}
		if len(client.routes) != 1 {
			t.Errorf("expected 1 route, got %d", len(client.routes))
		}
	})

	t.Run("without PSK", func(t *testing.T) {
		client := NewWGClient(ipnet, true, dns, routes)

		if client.psk != nil {
			t.Error("expected no PSK when noPSK=true")
		}
	})

	t.Run("without DNS", func(t *testing.T) {
		client := NewWGClient(ipnet, false, nil, routes)

		if client.dns != nil {
			t.Errorf("expected nil DNS, got %v", client.dns)
		}
	})

	t.Run("without routes", func(t *testing.T) {
		client := NewWGClient(ipnet, false, dns, nil)

		if client.routes != nil {
			t.Errorf("expected nil routes, got %v", client.routes)
		}
	})
}

func TestWGClientDNS(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}}

	t.Run("single DNS", func(t *testing.T) {
		dns := []net.IP{net.ParseIP("1.1.1.1")}
		client := NewWGClient(ipnet, true, dns, nil)

		dnsStr := client.DNS()
		if dnsStr != "1.1.1.1" {
			t.Errorf("expected '1.1.1.1', got '%s'", dnsStr)
		}
	})

	t.Run("multiple DNS", func(t *testing.T) {
		dns := []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("8.8.8.8")}
		client := NewWGClient(ipnet, true, dns, nil)

		dnsStr := client.DNS()
		if dnsStr != "1.1.1.1, 8.8.8.8" {
			t.Errorf("expected '1.1.1.1, 8.8.8.8', got '%s'", dnsStr)
		}
	})

	t.Run("empty DNS", func(t *testing.T) {
		client := NewWGClient(ipnet, true, nil, nil)

		dnsStr := client.DNS()
		if dnsStr != "" {
			t.Errorf("expected empty string, got '%s'", dnsStr)
		}
	})

	t.Run("IPv6 DNS", func(t *testing.T) {
		dns := []net.IP{net.ParseIP("2606:4700:4700::1111")}
		client := NewWGClient(ipnet, true, dns, nil)

		dnsStr := client.DNS()
		if dnsStr != "2606:4700:4700::1111" {
			t.Errorf("expected '2606:4700:4700::1111', got '%s'", dnsStr)
		}
	})
}

func TestWGClientToPeer(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}}
	dns := []net.IP{net.ParseIP("1.1.1.1")}
	routes := []net.IPNet{{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)}}
	client := NewWGClient(ipnet, false, dns, routes)

	peer := client.ToPeer()

	// Verify it's a WGClientAsPeer
	if peer == nil {
		t.Fatal("expected non-nil peer")
	}

	// Verify public key
	if peer.public.Base64() != client.private.Public().Base64() {
		t.Error("peer public key doesn't match client's derived public key")
	}

	// Verify PSK
	if peer.psk.Base64() != client.psk.Base64() {
		t.Error("peer PSK doesn't match client PSK")
	}

	// Verify allowedIPs has full mask
	ones, bits := peer.allowedIPs[0].Mask.Size()
	if ones != bits {
		t.Errorf("expected full mask (/%d), got /%d", bits, ones)
	}
}

func TestWGClientString(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}}

	t.Run("with DNS", func(t *testing.T) {
		dns := []net.IP{net.ParseIP("1.1.1.1")}
		client := NewWGClient(ipnet, true, dns, nil)

		s := client.String()

		if !strings.Contains(s, "Address = ") {
			t.Error("expected Address field in output")
		}
		if !strings.Contains(s, "PrivateKey = ") {
			t.Error("expected PrivateKey field in output")
		}
		if !strings.Contains(s, "DNS = 1.1.1.1") {
			t.Error("expected DNS field in output")
		}
	})

	t.Run("without DNS", func(t *testing.T) {
		client := NewWGClient(ipnet, true, nil, nil)

		s := client.String()

		if strings.Contains(s, "DNS = ") {
			t.Error("expected no DNS field when DNS is empty")
		}
	})
}

func TestWGClientPopulate(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}}

	t.Run("with DNS", func(t *testing.T) {
		dns := []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("8.8.8.8")}
		client := NewWGClient(ipnet, false, dns, nil)

		section := utils.NewSection("Interface")
		client.Populate(section)

		addr, err := section.Get("Address")
		if err != nil {
			t.Fatalf("failed to get Address: %v", err)
		}
		if addr != "10.0.0.2/24" {
			t.Errorf("expected Address '10.0.0.2/24', got '%s'", addr)
		}

		dnsVal, err := section.Get("DNS")
		if err != nil {
			t.Fatalf("failed to get DNS: %v", err)
		}
		if dnsVal != "1.1.1.1, 8.8.8.8" {
			t.Errorf("expected DNS '1.1.1.1, 8.8.8.8', got '%s'", dnsVal)
		}
	})

	t.Run("without DNS", func(t *testing.T) {
		client := NewWGClient(ipnet, false, nil, nil)

		section := utils.NewSection("Interface")
		client.Populate(section)

		if section.HasKey("DNS") {
			t.Error("expected no DNS key when DNS is empty")
		}
	})
}

func TestWGClientPopulateClient(t *testing.T) {
	// Setup VPN with server
	serverNet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	server := NewWGServer(serverNet, false, 51820)

	networks := []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}}
	routes := []net.IPNet{
		{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)},
	}
	dns := []net.IP{net.ParseIP("1.1.1.1")}

	vpn := &WGVPN{
		name:     "test",
		server:   server,
		peers:    make([]*WGClientAsPeer, 0),
		dns:      dns,
		endpoint: "vpn.example.com:51820",
		networks: networks,
		routes:   routes,
	}

	t.Run("client uses VPN routes", func(t *testing.T) {
		clientNet := []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}}
		client := NewWGClient(clientNet, false, dns, nil)

		file := utils.NewFile()
		client.PopulateClient(file, vpn)

		sections := file.Sections()
		if len(sections) != 2 {
			t.Fatalf("expected 2 sections (Interface, Peer), got %d", len(sections))
		}

		// Verify Interface section
		interfaceSec := sections[0]
		if interfaceSec.Name() != "Interface" {
			t.Errorf("expected first section to be Interface, got %s", interfaceSec.Name())
		}

		// Verify Peer section (server as peer)
		peerSec := sections[1]
		if peerSec.Name() != "Peer" {
			t.Errorf("expected second section to be Peer, got %s", peerSec.Name())
		}

		endpoint, err := peerSec.Get("Endpoint")
		if err != nil {
			t.Fatalf("failed to get Endpoint: %v", err)
		}
		if endpoint != "vpn.example.com:51820" {
			t.Errorf("expected Endpoint 'vpn.example.com:51820', got '%s'", endpoint)
		}

		allowedIPs, err := peerSec.Get("AllowedIPs")
		if err != nil {
			t.Fatalf("failed to get AllowedIPs: %v", err)
		}
		if !strings.Contains(allowedIPs, "0.0.0.0/0") {
			t.Errorf("expected AllowedIPs to contain '0.0.0.0/0', got '%s'", allowedIPs)
		}
	})

	t.Run("client uses own routes when specified", func(t *testing.T) {
		clientNet := []net.IPNet{{IP: net.ParseIP("10.0.0.3"), Mask: net.CIDRMask(24, 32)}}
		clientRoutes := []net.IPNet{{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)}}
		client := NewWGClient(clientNet, false, dns, clientRoutes)

		file := utils.NewFile()
		client.PopulateClient(file, vpn)

		peerSec := file.Sections()[1]
		allowedIPs, _ := peerSec.Get("AllowedIPs")

		if !strings.Contains(allowedIPs, "192.168.0.0/16") {
			t.Errorf("expected AllowedIPs to contain client's route '192.168.0.0/16', got '%s'", allowedIPs)
		}
		if strings.Contains(allowedIPs, "0.0.0.0/0") {
			t.Error("expected AllowedIPs to NOT contain VPN routes when client has own routes")
		}
	})

	t.Run("PSK is included in peer section", func(t *testing.T) {
		clientNet := []net.IPNet{{IP: net.ParseIP("10.0.0.4"), Mask: net.CIDRMask(24, 32)}}
		client := NewWGClient(clientNet, false, dns, nil)

		file := utils.NewFile()
		client.PopulateClient(file, vpn)

		peerSec := file.Sections()[1]
		if !peerSec.HasKey("PresharedKey") {
			t.Error("expected PresharedKey in peer section")
		}

		psk, _ := peerSec.Get("PresharedKey")
		if psk != client.PSK() {
			t.Error("PresharedKey mismatch")
		}
	})
}

func TestWGClientDualStack(t *testing.T) {
	multiNet := []net.IPNet{
		{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)},
		{IP: net.ParseIP("fd00::2"), Mask: net.CIDRMask(64, 128)},
	}
	dns := []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("2606:4700:4700::1111")}
	client := NewWGClient(multiNet, true, dns, nil)

	t.Run("address contains both", func(t *testing.T) {
		addr := client.Address()
		if !strings.Contains(addr, "10.0.0.2/24") {
			t.Errorf("expected IPv4 in address, got '%s'", addr)
		}
		if !strings.Contains(addr, "fd00::2/64") {
			t.Errorf("expected IPv6 in address, got '%s'", addr)
		}
	})

	t.Run("DNS contains both", func(t *testing.T) {
		dnsStr := client.DNS()
		if !strings.Contains(dnsStr, "1.1.1.1") {
			t.Errorf("expected IPv4 DNS, got '%s'", dnsStr)
		}
		if !strings.Contains(dnsStr, "2606:4700:4700::1111") {
			t.Errorf("expected IPv6 DNS, got '%s'", dnsStr)
		}
	})

	t.Run("peer has both allowedIPs", func(t *testing.T) {
		peer := client.ToPeer()
		if len(peer.allowedIPs) != 2 {
			t.Fatalf("expected 2 allowedIPs, got %d", len(peer.allowedIPs))
		}
	})
}
