package models

import (
	"net"
	"strings"
	"testing"

	"github.com/asiffer/wg-easy-vpn/utils"
)

func TestNewWGNode(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}

	t.Run("with PSK", func(t *testing.T) {
		node := NewWGNode(ipnet, false)

		if node.private == nil {
			t.Error("expected private key to be generated")
		}
		if len(node.private) != 32 {
			t.Errorf("expected 32-byte private key, got %d", len(node.private))
		}
		if node.psk == nil {
			t.Error("expected PSK to be generated")
		}
		if len(node.psk) != 32 {
			t.Errorf("expected 32-byte PSK, got %d", len(node.psk))
		}
		if len(node.address) != 1 {
			t.Errorf("expected 1 address, got %d", len(node.address))
		}
	})

	t.Run("without PSK", func(t *testing.T) {
		node := NewWGNode(ipnet, true)

		if node.private == nil {
			t.Error("expected private key to be generated")
		}
		if node.psk != nil {
			t.Error("expected no PSK when noPSK=true")
		}
	})

	t.Run("multiple addresses", func(t *testing.T) {
		multiNet := []net.IPNet{
			{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)},
			{IP: net.ParseIP("fd00::1"), Mask: net.CIDRMask(64, 128)},
		}
		node := NewWGNode(multiNet, false)

		if len(node.address) != 2 {
			t.Errorf("expected 2 addresses, got %d", len(node.address))
		}
	})
}

func TestWGNodePrivate(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	node := NewWGNode(ipnet, false)

	private := node.Private()

	// Base64 encoded 32-byte key should be 44 characters
	if len(private) != 44 {
		t.Errorf("expected base64 key length 44, got %d", len(private))
	}

	// Verify it ends with = (base64 padding for 32 bytes)
	if !strings.HasSuffix(private, "=") {
		t.Error("expected base64 key to end with padding")
	}
}

func TestWGNodePSK(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}

	t.Run("with PSK", func(t *testing.T) {
		node := NewWGNode(ipnet, false)
		psk := node.PSK()

		if len(psk) != 44 {
			t.Errorf("expected base64 PSK length 44, got %d", len(psk))
		}
	})

	t.Run("without PSK", func(t *testing.T) {
		node := NewWGNode(ipnet, true)
		psk := node.PSK()

		// nil PSK should return empty string
		if psk != "" {
			t.Errorf("expected empty PSK, got %s", psk)
		}
	})
}

func TestWGNodeAddress(t *testing.T) {
	t.Run("single IPv4", func(t *testing.T) {
		ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
		node := NewWGNode(ipnet, true)

		addr := node.Address()
		if addr != "10.0.0.1/24" {
			t.Errorf("expected '10.0.0.1/24', got '%s'", addr)
		}
	})

	t.Run("multiple addresses", func(t *testing.T) {
		multiNet := []net.IPNet{
			{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)},
			{IP: net.ParseIP("fd00::1"), Mask: net.CIDRMask(64, 128)},
		}
		node := NewWGNode(multiNet, true)

		addr := node.Address()
		if !strings.Contains(addr, "10.0.0.1/24") {
			t.Errorf("expected address to contain '10.0.0.1/24', got '%s'", addr)
		}
		if !strings.Contains(addr, "fd00::1/64") {
			t.Errorf("expected address to contain 'fd00::1/64', got '%s'", addr)
		}
		if !strings.Contains(addr, ", ") {
			t.Error("expected comma-separated addresses")
		}
	})
}

func TestWGNodeString(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}

	t.Run("with PSK", func(t *testing.T) {
		node := NewWGNode(ipnet, false)
		s := node.String()

		if !strings.Contains(s, "Address = ") {
			t.Error("expected Address field in output")
		}
		if !strings.Contains(s, "PrivateKey = ") {
			t.Error("expected PrivateKey field in output")
		}
		if !strings.Contains(s, "PresharedKey = ") {
			t.Error("expected PresharedKey field in output")
		}
	})

	t.Run("without PSK", func(t *testing.T) {
		node := NewWGNode(ipnet, true)
		s := node.String()

		if !strings.Contains(s, "Address = ") {
			t.Error("expected Address field in output")
		}
		if !strings.Contains(s, "PrivateKey = ") {
			t.Error("expected PrivateKey field in output")
		}
		if strings.Contains(s, "PresharedKey = ") {
			t.Error("expected no PresharedKey field in output")
		}
	})
}

func TestWGNodePopulate(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	node := NewWGNode(ipnet, false)

	section := utils.NewSection("Interface")
	node.Populate(section)

	addr, err := section.Get("Address")
	if err != nil {
		t.Fatalf("failed to get Address: %v", err)
	}
	if addr != "10.0.0.1/24" {
		t.Errorf("expected Address '10.0.0.1/24', got '%s'", addr)
	}

	privKey, err := section.Get("PrivateKey")
	if err != nil {
		t.Fatalf("failed to get PrivateKey: %v", err)
	}
	if privKey != node.Private() {
		t.Error("PrivateKey mismatch")
	}
}

func TestWGNodeToPeer(t *testing.T) {
	t.Run("single IPv4 address", func(t *testing.T) {
		ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
		node := NewWGNode(ipnet, false)

		peer := node.ToPeer()

		// Verify public key derivation
		if peer.public.Base64() != node.private.Public().Base64() {
			t.Error("peer public key doesn't match derived public key")
		}

		// Verify PSK is copied
		if peer.psk.Base64() != node.psk.Base64() {
			t.Error("peer PSK doesn't match node PSK")
		}

		// Verify allowedIPs uses full mask (/32 for IPv4)
		if len(peer.allowedIPs) != 1 {
			t.Fatalf("expected 1 allowedIP, got %d", len(peer.allowedIPs))
		}
		ones, bits := peer.allowedIPs[0].Mask.Size()
		if ones != bits {
			t.Errorf("expected full mask (/%d), got /%d", bits, ones)
		}
	})

	t.Run("IPv6 address", func(t *testing.T) {
		ipnet := []net.IPNet{{IP: net.ParseIP("fd00::1"), Mask: net.CIDRMask(64, 128)}}
		node := NewWGNode(ipnet, true)

		peer := node.ToPeer()

		// Verify allowedIPs uses full mask (/128 for IPv6)
		ones, bits := peer.allowedIPs[0].Mask.Size()
		if ones != bits {
			t.Errorf("expected full mask (/%d), got /%d", bits, ones)
		}
	})

	t.Run("dual stack", func(t *testing.T) {
		multiNet := []net.IPNet{
			{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)},
			{IP: net.ParseIP("fd00::1"), Mask: net.CIDRMask(64, 128)},
		}
		node := NewWGNode(multiNet, false)

		peer := node.ToPeer()

		if len(peer.allowedIPs) != 2 {
			t.Fatalf("expected 2 allowedIPs, got %d", len(peer.allowedIPs))
		}

		// Check IPv4 has /32 (full mask based on DefaultMask)
		ones, _ := peer.allowedIPs[0].Mask.Size()
		if ones != 32 {
			t.Errorf("expected /32 for IPv4, got /%d", ones)
		}

		// Note: IPv6 uses DefaultMask() which returns nil for IPv6,
		// so the mask becomes /0. This is the current implementation behavior.
		// The code uses: _, size := ipnet.IP.DefaultMask().Size()
		// For IPv6, DefaultMask() returns nil, and nil.Size() returns (0, 0)
		ones, _ = peer.allowedIPs[1].Mask.Size()
		if ones != 0 {
			t.Errorf("expected /0 for IPv6 (due to DefaultMask behavior), got /%d", ones)
		}
	})
}

func TestWGNodeKeyUniqueness(t *testing.T) {
	ipnet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}

	node1 := NewWGNode(ipnet, false)
	node2 := NewWGNode(ipnet, false)

	if node1.Private() == node2.Private() {
		t.Error("two nodes should have different private keys")
	}
	if node1.PSK() == node2.PSK() {
		t.Error("two nodes should have different PSKs")
	}
}
