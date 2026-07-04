package models

import (
	"fmt"
	"net"
	"strings"

	"github.com/asiffer/wg-easy-vpn/crypto"
	"github.com/asiffer/wg-easy-vpn/utils"
)

// WGPeer defines a Wireguard peer (client or server)
type WGPeer struct {
	allowedIPs []net.IPNet
	public     crypto.Key
	psk        crypto.PresharedKey
}

// WGServerAsPeer is a server seen from a client (peer of a client)
type WGServerAsPeer struct {
	WGPeer
	endpoint string
}

// WGClientAsPeer is a server seen from a client (peer of a client)
type WGClientAsPeer struct {
	WGPeer
}

func PeerFromSection(sec *utils.Section) (*WGClientAsPeer, error) {
	// Public key
	pubkey, err := sec.GetKeyFromBase64("PublicKey")
	if err != nil {
		return nil, fmt.Errorf("error while retrieving peer public key (%w)", err)
	}

	// AllowedIPs (array of string)
	ips, err := sec.GetNetworks("AllowedIPs")
	if err != nil {
		return nil, fmt.Errorf("error while retrieving peer allowedIPs (%w)", err)
	}

	// PSK
	var psk crypto.PresharedKey
	if sec.HasKey("PresharedKey") {
		pskKey, err := sec.GetKeyFromBase64("PresharedKey")
		// pskKey, err := sec.GetKey("PresharedKey")
		if err != nil {
			return nil, fmt.Errorf("error while retrieving peer psk (%w)", err)
		}
		psk = crypto.PresharedKey(pskKey)
	} else {
		psk = nil
	}

	// return peer
	return &WGClientAsPeer{
		WGPeer: WGPeer{
			allowedIPs: ips,
			public:     pubkey,
			psk:        psk,
		},
	}, nil

}

// AllowedIPs return the list of teh allowed addresses
func (peer *WGPeer) AllowedIPs() string {
	str := make([]string, len(peer.allowedIPs))
	for i, ipnet := range peer.allowedIPs {
		str[i] = ipnet.String()
	}
	return strings.Join(str, ", ")
}

// Public returns the peer public key base64 encoded
func (peer *WGPeer) Public() string {
	return peer.public.Base64()
}

// PSK returns the pre shared key as a base64 encoded string
func (peer *WGPeer) PSK() string {
	return peer.psk.Base64()
}

func (peer *WGPeer) String() string {
	s := ""
	s += fmt.Sprintf("PublicKey = %s\n", peer.Public())
	s += fmt.Sprintf("AllowedIPs = %s\n", peer.AllowedIPs())
	if peer.psk != nil {
		s += fmt.Sprintf("PresharedKey = %s\n", peer.PSK())
	}
	return s
}

// Populate enriches a section with server attributes
func (peer *WGPeer) Populate(section *utils.Section) {
	section.Set("AllowedIPs", peer.AllowedIPs())
	section.Set("PublicKey", peer.Public())
	if peer.psk != nil {
		section.Set("PresharedKey", peer.PSK())
	}
}

// Endpoint returns the server endpoint addr:port
func (server *WGServerAsPeer) Endpoint() string {
	return server.endpoint
}

// Populate enrich a section with server attributes
func (server *WGServerAsPeer) Populate(section *utils.Section) {
	server.WGPeer.Populate(section)
	section.Set("Endpoint", server.endpoint)
}
