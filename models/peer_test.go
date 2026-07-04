package models

import (
	"net"
	"strings"
	"testing"

	"github.com/asiffer/wg-easy-vpn/crypto"
	"github.com/asiffer/wg-easy-vpn/utils"
)

func TestWGPeerAllowedIPs(t *testing.T) {
	t.Run("single network", func(t *testing.T) {
		peer := &WGPeer{
			allowedIPs: []net.IPNet{
				{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(32, 32)},
			},
			public: crypto.NewRandomKey(),
		}

		allowed := peer.AllowedIPs()
		if allowed != "10.0.0.1/32" {
			t.Errorf("expected '10.0.0.1/32', got '%s'", allowed)
		}
	})

	t.Run("multiple networks", func(t *testing.T) {
		peer := &WGPeer{
			allowedIPs: []net.IPNet{
				{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(32, 32)},
				{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(24, 32)},
			},
			public: crypto.NewRandomKey(),
		}

		allowed := peer.AllowedIPs()
		if allowed != "10.0.0.1/32, 192.168.0.0/24" {
			t.Errorf("expected '10.0.0.1/32, 192.168.0.0/24', got '%s'", allowed)
		}
	})

	t.Run("dual stack", func(t *testing.T) {
		peer := &WGPeer{
			allowedIPs: []net.IPNet{
				{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)},
				{IP: net.ParseIP("::"), Mask: net.CIDRMask(0, 128)},
			},
			public: crypto.NewRandomKey(),
		}

		allowed := peer.AllowedIPs()
		if !strings.Contains(allowed, "0.0.0.0/0") {
			t.Errorf("expected '0.0.0.0/0' in allowed IPs, got '%s'", allowed)
		}
		if !strings.Contains(allowed, "::/0") {
			t.Errorf("expected '::/0' in allowed IPs, got '%s'", allowed)
		}
	})
}

func TestWGPeerPublic(t *testing.T) {
	key := crypto.NewRandomKey()
	peer := &WGPeer{
		allowedIPs: []net.IPNet{},
		public:     key,
	}

	pub := peer.Public()

	// Base64 encoded 32-byte key should be 44 characters
	if len(pub) != 44 {
		t.Errorf("expected base64 key length 44, got %d", len(pub))
	}

	if pub != key.Base64() {
		t.Error("public key mismatch")
	}
}

func TestWGPeerPSK(t *testing.T) {
	t.Run("with PSK", func(t *testing.T) {
		psk := crypto.NewRandomPresharedKey()
		peer := &WGPeer{
			allowedIPs: []net.IPNet{},
			public:     crypto.NewRandomKey(),
			psk:        psk,
		}

		pskStr := peer.PSK()
		if len(pskStr) != 44 {
			t.Errorf("expected base64 PSK length 44, got %d", len(pskStr))
		}
		if pskStr != psk.Base64() {
			t.Error("PSK mismatch")
		}
	})

	t.Run("without PSK", func(t *testing.T) {
		peer := &WGPeer{
			allowedIPs: []net.IPNet{},
			public:     crypto.NewRandomKey(),
			psk:        nil,
		}

		pskStr := peer.PSK()
		if pskStr != "" {
			t.Errorf("expected empty PSK, got '%s'", pskStr)
		}
	})
}

func TestWGPeerString(t *testing.T) {
	t.Run("with PSK", func(t *testing.T) {
		peer := &WGPeer{
			allowedIPs: []net.IPNet{
				{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(32, 32)},
			},
			public: crypto.NewRandomKey(),
			psk:    crypto.NewRandomPresharedKey(),
		}

		s := peer.String()

		if !strings.Contains(s, "PublicKey = ") {
			t.Error("expected PublicKey field in output")
		}
		if !strings.Contains(s, "AllowedIPs = ") {
			t.Error("expected AllowedIPs field in output")
		}
		if !strings.Contains(s, "PresharedKey = ") {
			t.Error("expected PresharedKey field in output")
		}
	})

	t.Run("without PSK", func(t *testing.T) {
		peer := &WGPeer{
			allowedIPs: []net.IPNet{
				{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(32, 32)},
			},
			public: crypto.NewRandomKey(),
			psk:    nil,
		}

		s := peer.String()

		if !strings.Contains(s, "PublicKey = ") {
			t.Error("expected PublicKey field in output")
		}
		if strings.Contains(s, "PresharedKey = ") {
			t.Error("expected no PresharedKey field when PSK is nil")
		}
	})
}

func TestWGPeerPopulate(t *testing.T) {
	t.Run("with PSK", func(t *testing.T) {
		peer := &WGPeer{
			allowedIPs: []net.IPNet{
				{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(32, 32)},
			},
			public: crypto.NewRandomKey(),
			psk:    crypto.NewRandomPresharedKey(),
		}

		section := utils.NewSection("Peer")
		peer.Populate(section)

		pubKey, err := section.Get("PublicKey")
		if err != nil {
			t.Fatalf("failed to get PublicKey: %v", err)
		}
		if pubKey != peer.Public() {
			t.Error("PublicKey mismatch")
		}

		allowed, err := section.Get("AllowedIPs")
		if err != nil {
			t.Fatalf("failed to get AllowedIPs: %v", err)
		}
		if allowed != "10.0.0.1/32" {
			t.Errorf("expected AllowedIPs '10.0.0.1/32', got '%s'", allowed)
		}

		psk, err := section.Get("PresharedKey")
		if err != nil {
			t.Fatalf("failed to get PresharedKey: %v", err)
		}
		if psk != peer.PSK() {
			t.Error("PresharedKey mismatch")
		}
	})

	t.Run("without PSK", func(t *testing.T) {
		peer := &WGPeer{
			allowedIPs: []net.IPNet{
				{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(32, 32)},
			},
			public: crypto.NewRandomKey(),
			psk:    nil,
		}

		section := utils.NewSection("Peer")
		peer.Populate(section)

		if section.HasKey("PresharedKey") {
			t.Error("expected no PresharedKey when PSK is nil")
		}
	})
}

func TestPeerFromSection(t *testing.T) {
	t.Run("valid section with PSK", func(t *testing.T) {
		section := utils.NewSection("Peer")
		section.Set("PublicKey", "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=")
		section.Set("AllowedIPs", "10.0.0.2/32")
		section.Set("PresharedKey", "qCJhKwR0uMEx8LbqvJbBx9LetPHA3zZp61M6TXcTaJ8=")

		peer, err := PeerFromSection(section)
		if err != nil {
			t.Fatalf("failed to parse peer: %v", err)
		}

		if peer.Public() != "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=" {
			t.Error("public key mismatch")
		}
		if peer.PSK() != "qCJhKwR0uMEx8LbqvJbBx9LetPHA3zZp61M6TXcTaJ8=" {
			t.Error("PSK mismatch")
		}
		if len(peer.allowedIPs) != 1 {
			t.Errorf("expected 1 allowedIP, got %d", len(peer.allowedIPs))
		}
	})

	t.Run("valid section without PSK", func(t *testing.T) {
		section := utils.NewSection("Peer")
		section.Set("PublicKey", "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=")
		section.Set("AllowedIPs", "10.0.0.2/32")

		peer, err := PeerFromSection(section)
		if err != nil {
			t.Fatalf("failed to parse peer: %v", err)
		}

		if peer.psk != nil {
			t.Error("expected nil PSK")
		}
	})

	t.Run("multiple AllowedIPs", func(t *testing.T) {
		section := utils.NewSection("Peer")
		section.Set("PublicKey", "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=")
		section.Set("AllowedIPs", "10.0.0.2/32, 192.168.1.0/24")

		peer, err := PeerFromSection(section)
		if err != nil {
			t.Fatalf("failed to parse peer: %v", err)
		}

		if len(peer.allowedIPs) != 2 {
			t.Errorf("expected 2 allowedIPs, got %d", len(peer.allowedIPs))
		}
	})

	t.Run("missing PublicKey", func(t *testing.T) {
		section := utils.NewSection("Peer")
		section.Set("AllowedIPs", "10.0.0.2/32")

		_, err := PeerFromSection(section)
		if err == nil {
			t.Error("expected error for missing PublicKey")
		}
	})

	t.Run("missing AllowedIPs", func(t *testing.T) {
		section := utils.NewSection("Peer")
		section.Set("PublicKey", "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=")

		_, err := PeerFromSection(section)
		if err == nil {
			t.Error("expected error for missing AllowedIPs")
		}
	})

	t.Run("invalid PublicKey", func(t *testing.T) {
		section := utils.NewSection("Peer")
		section.Set("PublicKey", "invalid-key")
		section.Set("AllowedIPs", "10.0.0.2/32")

		_, err := PeerFromSection(section)
		if err == nil {
			t.Error("expected error for invalid PublicKey")
		}
	})

	t.Run("invalid PresharedKey", func(t *testing.T) {
		section := utils.NewSection("Peer")
		section.Set("PublicKey", "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=")
		section.Set("AllowedIPs", "10.0.0.2/32")
		section.Set("PresharedKey", "invalid-psk")

		_, err := PeerFromSection(section)
		if err == nil {
			t.Error("expected error for invalid PresharedKey")
		}
	})
}

func TestPeerRoundTrip(t *testing.T) {
	// Create peer, populate section, parse back
	original := &WGPeer{
		allowedIPs: []net.IPNet{
			{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(32, 32)},
		},
		public: crypto.NewRandomKey(),
		psk:    crypto.NewRandomPresharedKey(),
	}

	section := utils.NewSection("Peer")
	original.Populate(section)

	parsed, err := PeerFromSection(section)
	if err != nil {
		t.Fatalf("failed to parse peer: %v", err)
	}

	if parsed.Public() != original.Public() {
		t.Error("public key mismatch after round-trip")
	}
	if parsed.PSK() != original.PSK() {
		t.Error("PSK mismatch after round-trip")
	}
	if parsed.AllowedIPs() != original.AllowedIPs() {
		t.Errorf("AllowedIPs mismatch: expected '%s', got '%s'",
			original.AllowedIPs(), parsed.AllowedIPs())
	}
}

func TestWGServerAsPeerEndpoint(t *testing.T) {
	peer := &WGServerAsPeer{
		WGPeer: WGPeer{
			allowedIPs: []net.IPNet{
				{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)},
			},
			public: crypto.NewRandomKey(),
		},
		endpoint: "vpn.example.com:51820",
	}

	if peer.Endpoint() != "vpn.example.com:51820" {
		t.Errorf("expected 'vpn.example.com:51820', got '%s'", peer.Endpoint())
	}
}

func TestWGServerAsPeerPopulate(t *testing.T) {
	peer := &WGServerAsPeer{
		WGPeer: WGPeer{
			allowedIPs: []net.IPNet{
				{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)},
			},
			public: crypto.NewRandomKey(),
			psk:    crypto.NewRandomPresharedKey(),
		},
		endpoint: "vpn.example.com:51820",
	}

	section := utils.NewSection("Peer")
	peer.Populate(section)

	// Verify base peer fields
	if !section.HasKey("PublicKey") {
		t.Error("expected PublicKey in section")
	}
	if !section.HasKey("AllowedIPs") {
		t.Error("expected AllowedIPs in section")
	}
	if !section.HasKey("PresharedKey") {
		t.Error("expected PresharedKey in section")
	}

	// Verify server-specific field
	endpoint, err := section.Get("Endpoint")
	if err != nil {
		t.Fatalf("failed to get Endpoint: %v", err)
	}
	if endpoint != "vpn.example.com:51820" {
		t.Errorf("expected Endpoint 'vpn.example.com:51820', got '%s'", endpoint)
	}
}

func TestWGClientAsPeer(t *testing.T) {
	// WGClientAsPeer is just a wrapper around WGPeer
	// It's mainly a type distinction, verify it works correctly
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}}
	client := NewWGClient(ipnet, false, nil, nil)

	clientAsPeer := client.ToPeer()

	// Verify it's a valid peer
	if clientAsPeer.Public() != client.private.Public().Base64() {
		t.Error("public key mismatch")
	}

	// Can be populated like any peer
	section := utils.NewSection("Peer")
	clientAsPeer.Populate(section)

	if !section.HasKey("PublicKey") {
		t.Error("expected PublicKey in section")
	}
	if !section.HasKey("AllowedIPs") {
		t.Error("expected AllowedIPs in section")
	}
}
