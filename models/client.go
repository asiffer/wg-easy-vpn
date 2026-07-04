package models

import (
	"fmt"
	"net"
	"strings"

	"github.com/asiffer/wg-easy-vpn/crypto"
	"github.com/asiffer/wg-easy-vpn/utils"
)

// WGClient is a particular node which tries to reach a server
type WGClient struct {
	WGNode
	dns    []net.IP
	routes []net.IPNet
}

// NewWGClient creates a new client
func NewWGClient(ipnet []net.IPNet, noPSK bool, dns []net.IP, routes []net.IPNet) *WGClient {
	// fmt.Println("CLIENT ROUTES:", routes)
	return &WGClient{
		WGNode: *NewWGNode(ipnet, noPSK),
		dns:    dns,
		routes: routes,
	}
}

// DNS returns the DNS address that client should use
func (client *WGClient) DNS() string {
	nDNS := len(client.dns)
	strDNS := make([]string, nDNS)
	for i := 0; i < nDNS; i++ {
		strDNS[i] = client.dns[i].String()
	}
	return strings.Join(strDNS, ", ")
}

// ToPeer turns a WGClient into a Peer
func (client *WGClient) ToPeer() *WGClientAsPeer {
	return &WGClientAsPeer{
		WGPeer: *client.WGNode.ToPeer(),
	}
}

func (client *WGClient) String() string {
	s := client.WGNode.String()
	if len(client.dns) > 0 {
		s += fmt.Sprintf("DNS = %s\n", client.DNS())
	}
	return s
}

// Populate enriches a section with client attributes
func (client *WGClient) Populate(section *utils.Section) {
	client.WGNode.Populate(section)
	if len(client.dns) > 0 {
		section.Set("DNS", client.DNS())
	}
}

// PopulateClient writes the client config into a file
func (client *WGClient) PopulateClient(file *utils.File, vpn *WGVPN) {
	// client section ([Interface])
	sec := file.AddSection("Interface")
	client.Populate(sec)

	// use client routes if provided
	routes := vpn.routes
	if len(client.routes) > 0 {
		routes = client.routes
	}
	// server as peer
	peer := vpn.server.ToPeer(routes, vpn.endpoint)
	// add client PSK to server as peer
	peer.psk = crypto.PresharedKey(client.psk)

	// Peer section
	sec = file.AddSection("Peer")
	peer.Populate(sec)
}
