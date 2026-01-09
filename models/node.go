package models

import (
	"fmt"
	"net"
	"strings"

	"github.com/asiffer/wg-easy-vpn/crypto"
	"github.com/asiffer/wg-easy-vpn/utils"
)

// WGNode defines a Wireguard node (client or server)
type WGNode struct {
	address []net.IPNet
	private crypto.Key
	psk     crypto.PresharedKey
}

// NewWGNode creates a new Node (generates random key and psk)
func NewWGNode(ipnet []net.IPNet, noPSK bool) *WGNode {
	node := &WGNode{
		address: ipnet,
		private: crypto.NewRandomKey(),
		psk:     crypto.NewRandomPresharedKey(),
	}
	if noPSK {
		node.psk = nil
	}
	return node

}

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
	return strings.Join(utils.StringifyNetworks(node.address), ", ")
}

func (node *WGNode) String() string {
	s := ""
	s += fmt.Sprintf("Address = %s\n", node.Address())
	s += fmt.Sprintf("PrivateKey = %s\n", node.Private())
	if node.psk != nil {
		s += fmt.Sprintf("PresharedKey = %s\n", node.PSK())
	}
	return s
}

// Section fills a section with node attributes
func (node *WGNode) Populate(section *utils.Section) {
	section.Set("Address", node.Address())
	section.Set("PrivateKey", node.Private())
}

// ToPeer turns a Node into a Peer
func (node *WGNode) ToPeer() *WGPeer {
	allowedIPs := make([]net.IPNet, len(node.address))
	// fmt.Println("AllowedIPs", node.address)
	// allowedIPs := NewNetSlice()
	// by default set the allowed IPs to the node networks with full mask
	// this is the default behavior when a client becomes a peer (server side so)
	// this configuration is later overriden when the server becomes a peer (client side)
	for i, ipnet := range node.address {
		_, size := ipnet.IP.DefaultMask().Size()
		allowedIPs[i] = net.IPNet{
			IP:   ipnet.IP,
			Mask: net.CIDRMask(size, size),
		}
	}

	return &WGPeer{
		allowedIPs: allowedIPs,
		public:     node.private.Public(),
		psk:        node.psk,
	}
}
