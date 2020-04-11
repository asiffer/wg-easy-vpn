// vpn.go
//
//

package main

import (
	"fmt"
	"net"
	"path"
	"strings"
)

// WGVPN denotes a Wireguard VPN (one server and several clients)
type WGVPN struct {
	name     string            // name of the connection
	server   *WGServer         // server
	peers    []*WGClientAsPeer // clients
	metadata *Metadata         // vpn metadata
}

func parseName(p string) string {
	base := path.Base(p)
	ext := path.Ext(base)
	if ext != "" {
		return strings.TrimSuffix(base, ext)
	}
	return base
}

func serverFromSection(sec *Section) (*WGServer, error) {
	networks, err := sec.GetNetSlice("Address")
	if err != nil {
		return nil, err
	}

	// PrivateKey
	private, err := sec.GetKeyFromBase64("PrivateKey")
	if err != nil {
		return nil, err
	}

	// ListenPort
	port, err := sec.GetInt("ListenPort")
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

func peerFromSection(sec *Section) (*WGClientAsPeer, error) {
	// Public key
	pubkey, err := sec.GetKeyFromBase64("PublicKey")
	if err != nil {
		return nil, fmt.Errorf("Error while retrieving peer public key (%w)", err)
	}

	// AllowedIPs (array of string)
	ips, err := sec.GetNetSlice("AllowedIPs")
	if err != nil {
		return nil, fmt.Errorf("Error while retrieving peer allowedIPs (%w)", err)
	}

	// PSK
	psk := NewPresharedKey()
	if sec.HasKey("PresharedKey") {
		pskKey, err := sec.GetKeyFromBase64("PresharedKey")
		// pskKey, err := sec.GetKey("PresharedKey")
		if err != nil {
			return nil, fmt.Errorf("Error while retrieving peer psk (%w)", err)
		}
		psk = PresharedKey(pskKey)
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

// ReadVPN reads a config file and load information in a minimal Wireguard VPN
func ReadVPN(p string) (*WGVPN, error) {
	cfg, err := ParseFile(p)
	if err != nil {
		return nil, err
	}

	// Read server/node
	var server *WGServer
	peers := make([]*WGClientAsPeer, 0)

	for _, sec := range cfg.Sections() {
		switch sec.Name() {
		case "Interface":
			server, err = serverFromSection(sec)
			if err != nil {
				return nil, err
			}
		case "Peer":
			peer, err := peerFromSection(sec)
			if err != nil {
				return nil, err
			}
			if peer != nil {
				peers = append(peers, peer)
			}
		default:
			// non-blocking
		}
		// fmt.Println(sec.String())
	}

	return &WGVPN{
		name:   parseName(p),
		server: server,
		peers:  peers,
	}, nil
}

// AddMetadata completes the VPN configuration by reading a metadata file
func (vpn *WGVPN) AddMetadata(path string) error {
	meta, err := LoadMetadata(vpn.name, path)
	if err != nil {
		return err
	}
	vpn.metadata = meta
	return nil
}

// NumberOfPeers returns the number of clients in the vpn
func (vpn *WGVPN) NumberOfPeers() int {
	return len(vpn.peers)
}

// RemovePeerFromPublicKey does what it says
func (vpn *WGVPN) RemovePeerFromPublicKey(k Key) error {
	for i, p := range vpn.peers {
		if k.Base64() == p.Public() {
			vpn.peers = append(vpn.peers[:i], vpn.peers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("This peer (%s) is not in the VPN", k.Base64())
}

// Save write the vpn config into a file (server conf only)
func (vpn *WGVPN) Save(file string) error {
	f := NewFile()
	section := f.AddSection("Interface")
	vpn.server.Section(section)

	for _, peer := range vpn.peers {
		section = f.AddSection("Peer")
		peer.Section(section)
	}

	return f.Save(file)
}

// PeerPublicKeys returns a list of the public keys of the clients
func (vpn *WGVPN) PeerPublicKeys() []string {
	keys := make([]string, vpn.NumberOfPeers())
	for i, p := range vpn.peers {
		keys[i] = p.Public()
	}
	return keys
}

// ReservedIPs return a list of all the IP already reserved in
// the VPN
func (vpn *WGVPN) ReservedIPs() []net.IP {
	reserved := make([]net.IP, 0)
	// loop over the server addresses
	for _, n := range *vpn.server.address {
		reserved = append(reserved, copyIP(n.IP))
	}
	// loop over the clients addresses
	for _, client := range vpn.peers {
		for _, n := range *client.allowedIPs {
			reserved = append(reserved, copyIP(n.IP))
		}
	}
	return reserved
}

// ProvideNetSlice generates a new netslice (for a new client
// for example). It ensures that there is no overlap. An error
// is raised whan no ip are available
func (vpn *WGVPN) ProvideNetSlice() (*NetSlice, error) {
	reserved := vpn.ReservedIPs()
	// new netslice
	out := NewNetSlice()
	// loop over the networks
	for _, n := range *vpn.metadata.networks {
		// loop over the addresses within each network
		list, stop := Iterate(n)
		loop := true
		for loop {
			// generate an IP
			ip, ok := <-list
			// return error if no IP remains
			if !ok {
				return nil, fmt.Errorf("No available IP")
			}
			// check if the ip is reserved
			// and if it is a special address
			special := ip.IsMulticast() || ip.IsUnspecified()
			if !special && findIP(ip, reserved) < 0 {
				out.Append(&net.IPNet{
					IP:   ip,
					Mask: n.Mask,
				})

				// close the generator
				stop <- true
				// stop the loop
				loop = false
			}
		}
	}
	return &out, nil
}
