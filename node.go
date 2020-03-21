//
//
//
package main

import (
	"fmt"
	"net"
)

// WGNode defines a Wireguard node (client or server)
type WGNode struct {
	address *NetSlice
	private Key
	psk     PresharedKey
}

// WGServer is a particular node which listens on a new UDP port
type WGServer struct {
	WGNode
	port int
}

// WGClient is a particular node which tries to reach a server
type WGClient struct {
	WGNode
	dns []net.IP
}

// NewWGNode creates a new Node (generates random key and psk)
func NewWGNode(ipnet *NetSlice, withPSK bool) *WGNode {
	if withPSK {
		return &WGNode{
			address: ipnet.Copy(),
			private: NewRandomKey(),
			psk:     NewRandomPresharedKey(),
		}
	}
	return &WGNode{address: ipnet.Copy(), private: NewRandomKey(), psk: nil}
}

// NewWGServer creates a new server
func NewWGServer(ipnet *NetSlice, withPSK bool, port int) *WGServer {
	return &WGServer{
		WGNode: *NewWGNode(ipnet, withPSK),
		port:   port,
	}
}

// NewWGClient creates a new client
func NewWGClient(ipnet *NetSlice, withPSK bool, dns []net.IP) *WGClient {
	return &WGClient{
		WGNode: *NewWGNode(ipnet, withPSK),
		dns:    dns,
	}
}

// Public returns the public key of the node (base64 encoded string)
// func (node *WGNode) Public() string {
// 	return node.private.Public().Base64()
// }

// Private returns the private key of the node (base64 encoded string)
func (node *WGNode) Private() string {
	return node.private.Base64()
}

// PSK returns the pre shared key as a base64 encoded string
func (node *WGNode) PSK() string {
	return node.psk.Base64()
}

// Address returns the node address
func (node *WGNode) Address() string {
	// return strings.Join(mapIPNetList(node.address), ", ")
	return node.address.String()
}

func (node *WGNode) String() string {
	s := ""
	s += fmt.Sprintf("Address = %s\n", node.Address())
	s += fmt.Sprintf("PrivateKey = %s\n", node.Private())
	if node.psk != nil {
		s += fmt.Sprintf("PSK = %s\n", node.PSK())
	}
	return s
}

// Section fills a section with node attributes
func (node *WGNode) Section(section *Section) {
	section.Set("Address", node.Address())
	section.Set("PrivateKey", node.Private())
	if node.psk != nil {
		section.Set("PSK", node.PSK())
	}
}

// DNS returns the DNS address that client should use
// func (client *WGClient) DNS() string {
// 	nDNS := len(client.dns)
// 	strDNS := make([]string, nDNS)
// 	for i := 0; i < nDNS; i++ {
// 		strDNS[i] = client.dns[i].String()
// 	}
// 	return strings.Join(strDNS, ", ")
// }

// ToPeer turns a Node into a Peer
func (node *WGNode) ToPeer() *WGPeer {
	// allowedIPs := make([]*net.IPNet, node.address.Len())
	allowedIPs := NewNetSlice()
	for _, ipnet := range *node.address {
		_, size := ipnet.IP.DefaultMask().Size()
		allowedIPs.Append(&net.IPNet{
			IP:   ipnet.IP,
			Mask: net.CIDRMask(size, size),
		})
	}

	return &WGPeer{
		allowedIPs: &allowedIPs,
		public:     node.private.Public(),
		psk:        node.psk,
	}
}

// ToPeer turns a WGServer into a Peer.
// routes is the destinations which will pass through the vpn
// dnsname is the public address of teh server
func (server *WGServer) ToPeer(routes *NetSlice, endpoint string) *WGServerAsPeer {
	peer := server.WGNode.ToPeer()
	if routes != nil {
		peer.allowedIPs = routes
	} else {
		n := NewNetSlice()
		n.Append(&IPv4ZeroNet)
		n.Append(&IPv6ZeroNet)
		// peer.allowedIPs = []*net.IPNet{&IPv4ZeroNet, &IPv6ZeroNet}
		peer.allowedIPs = &n
	}
	return &WGServerAsPeer{
		WGPeer:   *peer,
		endpoint: fmt.Sprintf("%s:%d", endpoint, server.port),
	}
}

func (server *WGServer) String() string {
	s := server.WGNode.String()
	s += fmt.Sprintf("ListenPort = %d\n", server.port)
	return s
}

// Section enrich a section with server attributes
func (server *WGServer) Section(section *Section) {
	server.WGNode.Section(section)
	section.Set("ListenPort", fmt.Sprintf("%d", server.port))
}

// ToPeer turns a WGClient into a Peer
func (client *WGClient) ToPeer() *WGClientAsPeer {
	return &WGClientAsPeer{
		WGPeer: *client.WGNode.ToPeer(),
	}
}
