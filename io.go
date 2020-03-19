//
//
//
package main

import (
	"fmt"
	"io"
)

const (
	// InterfaceHeader is the header which starts the interface section
	InterfaceHeader = "[Interface]"
	// PeerHeader is the header which starts a peer section
	PeerHeader = "[Peer]"
)

func (node *WGNode) Write(w io.Writer) (int, error) {
	section := InterfaceHeader + "\n"
	section += fmt.Sprintf("Address = %s\n", node.Address())
	section += fmt.Sprintf("PrivateKey = %s\n", node.Private())
	return w.Write([]byte(section))
}

func (server *WGServer) Write(w io.Writer) (int, error) {
	n, err := server.WGNode.Write(w)
	if err != nil {
		return n, err
	}
	add := fmt.Sprintf("ListenPort = %d\n", server.port)
	m, err := w.Write([]byte(add))
	return n + m, err
}

func (client *WGClient) Write(w io.Writer) (int, error) {
	n, err := client.WGNode.Write(w)
	if err != nil {
		return n, err
	}
	if client.dns != nil {
		add := fmt.Sprintf("DNS = %s\n", client.DNS())
		m, err := w.Write([]byte(add))
		return n + m, err
	}
	return n, err
}

func (peer *WGPeer) Write(w io.Writer) (int, error) {
	section := PeerHeader + "\n"
	section += fmt.Sprintf("PublicKey = %s\n", peer.Public())
	if peer.psk != nil {
		section += fmt.Sprintf("PresharedKey = %s\n", peer.PSK())
	}
	section += fmt.Sprintf("AllowedIPs = %s\n", peer.AllowedIPs())
	return w.Write([]byte(section))
}

func (server *WGServerAsPeer) Write(w io.Writer) (int, error) {
	n, err := server.WGPeer.Write(w)
	if err != nil {
		return n, err
	}
	add := fmt.Sprintf("Endpoint = %s\n", server.Endpoint())
	m, err := w.Write([]byte(add))
	return m + n, err
}

func (client *WGClientAsPeer) Write(w io.Writer) (int, error) {
	return client.WGPeer.Write(w)
}
