//
//
//
package main

import (
	"fmt"
	"strings"
)

// WGPeer defines a Wireguard peer (client or server)
type WGPeer struct {
	allowedIPs *NetSlice
	public     Key
	psk        PresharedKey
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

// AllowedIPs return the list of teh allowed addresses
func (peer *WGPeer) AllowedIPs() string {
	str := make([]string, peer.allowedIPs.Len())
	for i, ipnet := range *peer.allowedIPs {
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

// Section enrich a section with server attributes
func (peer *WGPeer) Section(section *Section) {
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

// Section enrich a section with server attributes
func (server *WGServerAsPeer) Section(section *Section) {
	server.WGPeer.Section(section)
	section.Set("Endpoint", server.endpoint)
}
