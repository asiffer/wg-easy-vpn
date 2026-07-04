package models

import (
	"fmt"
	"net"

	"github.com/asiffer/wg-easy-vpn/utils"
)

// WGServer is a particular node which listens on a new UDP port
type WGServer struct {
	WGNode
	port uint16
}

// NewWGServer creates a new server
func NewWGServer(ipnet []net.IPNet, noPSK bool, port uint16) *WGServer {
	return &WGServer{
		WGNode: *NewWGNode(ipnet, noPSK),
		port:   port,
	}
}

// ToPeer turns a WGServer into a Peer (so it is used client-side).
// routes is the destinations which will pass through the vpn
// dnsname is the public address of the server
func (server *WGServer) ToPeer(routes []net.IPNet, endpoint string) *WGServerAsPeer {
	peer := server.WGNode.ToPeer()
	if len(routes) > 0 {
		peer.allowedIPs = routes
	} else {
		peer.allowedIPs = []net.IPNet{
			utils.IPv4ZeroNet,
			utils.IPv6ZeroNet,
		}
	}
	return &WGServerAsPeer{
		WGPeer:   *peer,
		endpoint: endpoint,
	}
}

func (server *WGServer) String() string {
	s := server.WGNode.String()
	s += fmt.Sprintf("ListenPort = %d\n", server.port)
	return s
}

// Section enrich a section with server attributes
func (server *WGServer) Populate(section *utils.Section) {
	server.WGNode.Populate(section)
	section.Set("ListenPort", fmt.Sprintf("%d", server.port))
}

func ServerFromSection(sec *utils.Section) (*WGServer, error) {
	networks, err := sec.GetNetworks("Address")
	if err != nil {
		return nil, err
	}
	fmt.Println("networks:", networks)

	// PrivateKey
	private, err := sec.GetKeyFromBase64("PrivateKey")
	if err != nil {
		return nil, err
	}

	// ListenPort
	port, err := sec.GetUint16("ListenPort")
	if err != nil {
		return nil, err
	}

	return &WGServer{
		WGNode: WGNode{
			address: networks,
			private: private,
			psk:     nil,
		},
		port: port,
	}, nil
}
