// node_test.go

package main

import (
	"fmt"
	"testing"
)

// NewWGNode creates a new Node (generates random key and psk)
func TestNewWGNode(t *testing.T) {
	ns := "192.168.0.1/24, fe80::/64"
	slice, err := NewNetSliceFromString(ns)
	if err != nil {
		t.Errorf("Error while parsing netslice %s (%v)", ns, err)
	}
	node := NewWGNode(&slice, false)
	if node.psk != nil {
		t.Errorf("The PSK should not be set")
	}
}

func TestNodeString(t *testing.T) {
	ns := "192.168.0.1/24, fe80::/64"
	slice, err := NewNetSliceFromString(ns)
	if err != nil {
		t.Errorf("Error while parsing netslice %s (%v)", ns, err)
	}
	node := NewWGNode(&slice, true)
	truth := fmt.Sprintf("Address = %s\nPrivateKey = %s\nPSK = %s\n",
		node.Address(), node.Private(), node.PSK())
	if node.String() != truth {
		t.Errorf("Expected %s, got %s", truth, node.String())
	}
}

func TestNewWGServerToPeer(t *testing.T) {
	ns := "192.168.0.1/24, fe80::/64"
	slice, err := NewNetSliceFromString(ns)
	if err != nil {
		t.Errorf("Error while parsing netslice %s (%v)", ns, err)
	}
	server := NewWGServer(&slice, true, 15000)
	peer := server.ToPeer(nil, "wg.example.net")

	truth := NewNetSlice()
	truth.Append(&IPv4ZeroNet)
	truth.Append(&IPv6ZeroNet)
	if peer.allowedIPs.String() != truth.String() {
		t.Errorf("Expecting %s, got %s", truth.String(), peer.allowedIPs.String())
	}
	// (ipnet *NetSlice, withPSK bool, port int) *WGServer {
	// 	return &WGServer{
	// 		WGNode: *NewWGNode(ipnet, withPSK),
	// 		port:   port,
	// 	}
}
