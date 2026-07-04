package models

import (
	"net"
	"strings"
	"testing"

	"github.com/asiffer/wg-easy-vpn/crypto"
	"github.com/asiffer/wg-easy-vpn/utils"
)

func TestNewWGVPN(t *testing.T) {
	serverNet := []net.IPNet{}
	server := NewWGServer(serverNet, false, 51820)
	networks := []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}}
	dns := []net.IP{net.ParseIP("1.1.1.1")}
	routes := []net.IPNet{{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)}}

	vpn, err := NewWGVPN("test-vpn", server, "vpn.example.com:51820", networks, dns, routes)
	if err != nil {
		t.Fatalf("failed to create VPN: %v", err)
	}

	if vpn.name != "test-vpn" {
		t.Errorf("expected name 'test-vpn', got '%s'", vpn.name)
	}
	if vpn.endpoint != "vpn.example.com:51820" {
		t.Errorf("expected endpoint 'vpn.example.com:51820', got '%s'", vpn.endpoint)
	}
	if len(vpn.dns) != 1 {
		t.Errorf("expected 1 DNS server, got %d", len(vpn.dns))
	}
	if len(vpn.networks) != 1 {
		t.Errorf("expected 1 network, got %d", len(vpn.networks))
	}
	if len(vpn.routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(vpn.routes))
	}

	// Server should have been assigned first IP from network
	if len(vpn.server.address) == 0 {
		t.Fatal("expected server to have an address")
	}
	// First available IP in 10.0.0.0/24 is 10.0.0.1 (skips network address)
	serverIP := vpn.server.address[0].IP.String()
	if serverIP != "10.0.0.1" {
		t.Errorf("expected server IP '10.0.0.1', got '%s'", serverIP)
	}
}

func TestWGVPNNumberOfPeers(t *testing.T) {
	vpn := &WGVPN{
		peers: make([]*WGClientAsPeer, 0),
	}

	if vpn.NumberOfPeers() != 0 {
		t.Errorf("expected 0 peers, got %d", vpn.NumberOfPeers())
	}

	vpn.peers = append(vpn.peers, &WGClientAsPeer{})
	if vpn.NumberOfPeers() != 1 {
		t.Errorf("expected 1 peer, got %d", vpn.NumberOfPeers())
	}

	vpn.peers = append(vpn.peers, &WGClientAsPeer{}, &WGClientAsPeer{})
	if vpn.NumberOfPeers() != 3 {
		t.Errorf("expected 3 peers, got %d", vpn.NumberOfPeers())
	}
}

func TestWGVPNRemovePeer(t *testing.T) {
	key1 := crypto.NewRandomKey()
	key2 := crypto.NewRandomKey()

	vpn := &WGVPN{
		peers: []*WGClientAsPeer{
			{WGPeer: WGPeer{public: key1}},
			{WGPeer: WGPeer{public: key2}},
		},
	}

	t.Run("remove existing peer", func(t *testing.T) {
		err := vpn.RemovePeer(key1)
		if err != nil {
			t.Fatalf("failed to remove peer: %v", err)
		}
		if vpn.NumberOfPeers() != 1 {
			t.Errorf("expected 1 peer after removal, got %d", vpn.NumberOfPeers())
		}
		if vpn.peers[0].public.Base64() != key2.Base64() {
			t.Error("wrong peer was removed")
		}
	})

	t.Run("remove non-existing peer", func(t *testing.T) {
		nonExistingKey := crypto.NewRandomKey()
		err := vpn.RemovePeer(nonExistingKey)
		if err == nil {
			t.Error("expected error when removing non-existing peer")
		}
	})
}

func TestWGVPNReservedIPs(t *testing.T) {
	serverNet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	server := NewWGServer(serverNet, true, 51820)

	vpn := &WGVPN{
		server: server,
		peers: []*WGClientAsPeer{
			{WGPeer: WGPeer{
				allowedIPs: []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(32, 32)}},
			}},
			{WGPeer: WGPeer{
				allowedIPs: []net.IPNet{{IP: net.ParseIP("10.0.0.3"), Mask: net.CIDRMask(32, 32)}},
			}},
		},
	}

	reserved := vpn.ReservedIPs()

	if len(reserved) != 3 {
		t.Fatalf("expected 3 reserved IPs, got %d", len(reserved))
	}

	// Check all IPs are present
	ips := make(map[string]bool)
	for _, ip := range reserved {
		ips[ip.String()] = true
	}

	if !ips["10.0.0.1"] {
		t.Error("expected 10.0.0.1 in reserved IPs")
	}
	if !ips["10.0.0.2"] {
		t.Error("expected 10.0.0.2 in reserved IPs")
	}
	if !ips["10.0.0.3"] {
		t.Error("expected 10.0.0.3 in reserved IPs")
	}
}

func TestWGVPNProvideNetworks(t *testing.T) {
	t.Run("allocates next available IP", func(t *testing.T) {
		serverNet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
		server := NewWGServer(serverNet, true, 51820)

		vpn := &WGVPN{
			server:   server,
			networks: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
			peers:    make([]*WGClientAsPeer, 0),
		}

		nets, err := vpn.ProvideNetworks()
		if err != nil {
			t.Fatalf("failed to provide networks: %v", err)
		}

		if len(nets) != 1 {
			t.Fatalf("expected 1 network, got %d", len(nets))
		}

		// Server has 10.0.0.1, so next should be 10.0.0.2
		if nets[0].IP.String() != "10.0.0.2" {
			t.Errorf("expected IP '10.0.0.2', got '%s'", nets[0].IP.String())
		}
	})

	t.Run("skips reserved IPs", func(t *testing.T) {
		serverNet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
		server := NewWGServer(serverNet, true, 51820)

		vpn := &WGVPN{
			server:   server,
			networks: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
			peers: []*WGClientAsPeer{
				{WGPeer: WGPeer{
					allowedIPs: []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(32, 32)}},
				}},
			},
		}

		nets, err := vpn.ProvideNetworks()
		if err != nil {
			t.Fatalf("failed to provide networks: %v", err)
		}

		// Server has 10.0.0.1, peer has 10.0.0.2, so next should be 10.0.0.3
		if nets[0].IP.String() != "10.0.0.3" {
			t.Errorf("expected IP '10.0.0.3', got '%s'", nets[0].IP.String())
		}
	})

	t.Run("dual stack allocation", func(t *testing.T) {
		// Note: Using /120 for IPv6 instead of /64 because the Iterate function
		// uses bit shifting (1 << (total-frozen)) which overflows for large subnets.
		// A /64 would require iterating 2^64 addresses which causes overflow.
		serverNet := []net.IPNet{
			{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)},
			{IP: net.ParseIP("fd00::1"), Mask: net.CIDRMask(120, 128)},
		}
		server := NewWGServer(serverNet, true, 51820)

		vpn := &WGVPN{
			server: server,
			networks: []net.IPNet{
				{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)},
				{IP: net.ParseIP("fd00::"), Mask: net.CIDRMask(120, 128)},
			},
			peers: make([]*WGClientAsPeer, 0),
		}

		nets, err := vpn.ProvideNetworks()
		if err != nil {
			t.Fatalf("failed to provide networks: %v", err)
		}

		if len(nets) != 2 {
			t.Fatalf("expected 2 networks (dual stack), got %d", len(nets))
		}
	})
}

func TestWGVPNAddClient(t *testing.T) {
	serverNet := []net.IPNet{}
	server := NewWGServer(serverNet, true, 51820)
	networks := []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}}

	vpn, _ := NewWGVPN("test", server, "vpn.example.com:51820", networks, nil, nil)

	t.Run("add first client", func(t *testing.T) {
		client := NewWGClient(nil, true, nil, nil)
		err := vpn.AddClient(client)
		if err != nil {
			t.Fatalf("failed to add client: %v", err)
		}

		if vpn.NumberOfPeers() != 1 {
			t.Errorf("expected 1 peer, got %d", vpn.NumberOfPeers())
		}

		// Client should have been assigned 10.0.0.2 (10.0.0.1 is server)
		if len(client.address) == 0 {
			t.Fatal("expected client to have an address")
		}
		if client.address[0].IP.String() != "10.0.0.2" {
			t.Errorf("expected client IP '10.0.0.2', got '%s'", client.address[0].IP.String())
		}
	})

	t.Run("add second client", func(t *testing.T) {
		client := NewWGClient(nil, true, nil, nil)
		err := vpn.AddClient(client)
		if err != nil {
			t.Fatalf("failed to add client: %v", err)
		}

		if vpn.NumberOfPeers() != 2 {
			t.Errorf("expected 2 peers, got %d", vpn.NumberOfPeers())
		}

		// Second client should have 10.0.0.3
		if client.address[0].IP.String() != "10.0.0.3" {
			t.Errorf("expected client IP '10.0.0.3', got '%s'", client.address[0].IP.String())
		}
	})
}

func TestWGVPNPeerPublicKeys(t *testing.T) {
	key1 := crypto.NewRandomKey()
	key2 := crypto.NewRandomKey()

	vpn := &WGVPN{
		peers: []*WGClientAsPeer{
			{WGPeer: WGPeer{public: key1}},
			{WGPeer: WGPeer{public: key2}},
		},
	}

	keys := vpn.PeerPublicKeys()

	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != key1.Base64() {
		t.Error("first key mismatch")
	}
	if keys[1] != key2.Base64() {
		t.Error("second key mismatch")
	}
}

func TestWGVPNPopulateServer(t *testing.T) {
	serverNet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	server := NewWGServer(serverNet, true, 51820)

	key1 := crypto.NewRandomKey()
	vpn := &WGVPN{
		server: server,
		peers: []*WGClientAsPeer{
			{WGPeer: WGPeer{
				public:     key1,
				allowedIPs: []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(32, 32)}},
			}},
		},
	}

	file := utils.NewFile()
	vpn.PopulateServer(file)

	sections := file.Sections()
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections (Interface + 1 Peer), got %d", len(sections))
	}

	// Verify Interface section
	if sections[0].Name() != "Interface" {
		t.Errorf("expected first section to be Interface, got %s", sections[0].Name())
	}

	// Verify Peer section
	if sections[1].Name() != "Peer" {
		t.Errorf("expected second section to be Peer, got %s", sections[1].Name())
	}
}

func TestWGVPNPopulate(t *testing.T) {
	serverNet := []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}}
	server := NewWGServer(serverNet, true, 51820)

	vpn := &WGVPN{
		name:     "test-vpn",
		server:   server,
		endpoint: "vpn.example.com:51820",
		dns:      []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("8.8.8.8")},
		networks: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
		routes:   []net.IPNet{{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)}},
		peers:    make([]*WGClientAsPeer, 0),
	}

	file := utils.NewFile()
	vpn.Populate(file)

	// Should have default section with metadata
	defaultSec := file.GetorCreateSection(utils.DEFAULT_SECTION)

	endpoint, err := defaultSec.Get("Endpoint")
	if err != nil {
		t.Fatalf("failed to get Endpoint: %v", err)
	}
	if endpoint != "vpn.example.com:51820" {
		t.Errorf("expected Endpoint 'vpn.example.com:51820', got '%s'", endpoint)
	}

	dns, err := defaultSec.Get("DNS")
	if err != nil {
		t.Fatalf("failed to get DNS: %v", err)
	}
	if !strings.Contains(dns, "1.1.1.1") || !strings.Contains(dns, "8.8.8.8") {
		t.Errorf("expected DNS to contain both servers, got '%s'", dns)
	}

	network, err := defaultSec.Get("Network")
	if err != nil {
		t.Fatalf("failed to get Network: %v", err)
	}
	if network != "10.0.0.0/24" {
		t.Errorf("expected Network '10.0.0.0/24', got '%s'", network)
	}

	routes, err := defaultSec.Get("Routes")
	if err != nil {
		t.Fatalf("failed to get Routes: %v", err)
	}
	if routes != "0.0.0.0/0" {
		t.Errorf("expected Routes '0.0.0.0/0', got '%s'", routes)
	}
}

func TestVPNFromFile(t *testing.T) {
	t.Run("parse complete config", func(t *testing.T) {
		file := utils.NewFile()

		// Add default section with metadata
		def := file.GetorCreateSection(utils.DEFAULT_SECTION)
		def.Set("Endpoint", "vpn.example.com:51820")
		def.Set("DNS", "1.1.1.1,8.8.8.8")
		def.Set("Network", "10.0.0.0/24")
		def.Set("Routes", "0.0.0.0/0")

		// Add Interface section
		iface := file.AddSection("Interface")
		iface.Set("Address", "10.0.0.1/24")
		iface.Set("PrivateKey", "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=")
		iface.Set("ListenPort", "51820")

		// Add Peer section
		peer := file.AddSection("Peer")
		peer.Set("PublicKey", "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=")
		peer.Set("AllowedIPs", "10.0.0.2/32")

		vpn, err := VPNFromFile("test", file)
		if err != nil {
			t.Fatalf("failed to parse VPN from file: %v", err)
		}

		if vpn.name != "test" {
			t.Errorf("expected name 'test', got '%s'", vpn.name)
		}
		if vpn.endpoint != "vpn.example.com:51820" {
			t.Errorf("expected endpoint 'vpn.example.com:51820', got '%s'", vpn.endpoint)
		}
		if len(vpn.dns) != 2 {
			t.Errorf("expected 2 DNS servers, got %d", len(vpn.dns))
		}
		if len(vpn.networks) != 1 {
			t.Errorf("expected 1 network, got %d", len(vpn.networks))
		}
		if len(vpn.routes) != 1 {
			t.Errorf("expected 1 route, got %d", len(vpn.routes))
		}
		if vpn.server == nil {
			t.Fatal("expected server to be parsed")
		}
		if vpn.server.port != 51820 {
			t.Errorf("expected server port 51820, got %d", vpn.server.port)
		}
		if vpn.NumberOfPeers() != 1 {
			t.Errorf("expected 1 peer, got %d", vpn.NumberOfPeers())
		}
	})

	t.Run("parse config without optional fields", func(t *testing.T) {
		file := utils.NewFile()

		// Add Interface section only
		iface := file.AddSection("Interface")
		iface.Set("Address", "10.0.0.1/24")
		iface.Set("PrivateKey", "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=")
		iface.Set("ListenPort", "51820")

		vpn, err := VPNFromFile("minimal", file)
		if err != nil {
			t.Fatalf("failed to parse VPN: %v", err)
		}

		if vpn.dns != nil {
			t.Error("expected nil DNS")
		}
		if vpn.networks != nil {
			t.Error("expected nil networks")
		}
		if vpn.routes != nil {
			t.Error("expected nil routes")
		}
		if vpn.endpoint != "" {
			t.Errorf("expected empty endpoint, got '%s'", vpn.endpoint)
		}
	})

	t.Run("parse config with multiple peers", func(t *testing.T) {
		file := utils.NewFile()

		iface := file.AddSection("Interface")
		iface.Set("Address", "10.0.0.1/24")
		iface.Set("PrivateKey", "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=")
		iface.Set("ListenPort", "51820")

		peer1 := file.AddSection("Peer")
		peer1.Set("PublicKey", "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=")
		peer1.Set("AllowedIPs", "10.0.0.2/32")

		peer2 := file.AddSection("Peer")
		peer2.Set("PublicKey", "qCJhKwR0uMEx8LbqvJbBx9LetPHA3zZp61M6TXcTaJ8=")
		peer2.Set("AllowedIPs", "10.0.0.3/32")

		vpn, err := VPNFromFile("multi-peer", file)
		if err != nil {
			t.Fatalf("failed to parse VPN: %v", err)
		}

		if vpn.NumberOfPeers() != 2 {
			t.Errorf("expected 2 peers, got %d", vpn.NumberOfPeers())
		}
	})
}

func TestVPNRoundTrip(t *testing.T) {
	// Create VPN, populate file, parse back
	serverNet := []net.IPNet{}
	server := NewWGServer(serverNet, true, 51820)
	networks := []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}}
	dns := []net.IP{net.ParseIP("1.1.1.1")}
	routes := []net.IPNet{{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(0, 32)}}

	original, err := NewWGVPN("roundtrip", server, "vpn.example.com:51820", networks, dns, routes)
	if err != nil {
		t.Fatalf("failed to create VPN: %v", err)
	}

	// Add a client
	client := NewWGClient(nil, true, nil, nil)
	original.AddClient(client)

	// Populate file
	file := utils.NewFile()
	original.Populate(file)

	// Parse back
	parsed, err := VPNFromFile("roundtrip", file)
	if err != nil {
		t.Fatalf("failed to parse VPN: %v", err)
	}

	// Compare
	if parsed.endpoint != original.endpoint {
		t.Errorf("endpoint mismatch: expected '%s', got '%s'", original.endpoint, parsed.endpoint)
	}
	if len(parsed.dns) != len(original.dns) {
		t.Errorf("DNS count mismatch: expected %d, got %d", len(original.dns), len(parsed.dns))
	}
	if len(parsed.networks) != len(original.networks) {
		t.Errorf("networks count mismatch: expected %d, got %d", len(original.networks), len(parsed.networks))
	}
	if len(parsed.routes) != len(original.routes) {
		t.Errorf("routes count mismatch: expected %d, got %d", len(original.routes), len(parsed.routes))
	}
	if parsed.NumberOfPeers() != original.NumberOfPeers() {
		t.Errorf("peer count mismatch: expected %d, got %d", original.NumberOfPeers(), parsed.NumberOfPeers())
	}
	if parsed.server.port != original.server.port {
		t.Errorf("server port mismatch: expected %d, got %d", original.server.port, parsed.server.port)
	}
}
