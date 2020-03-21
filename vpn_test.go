//
//
//
package main

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"
)

var (
	conf string
	meta string
)

func init() {
	conf = path.Join(testpath, "testfile.conf")
	meta = path.Join(testpath, "testfile.meta")
}

func TestReadVPN(t *testing.T) {
	title("Reading VPN config")
	vpn, err := ReadVPN(conf)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Server:")
	fmt.Print(vpn.server.String())

	for _, peer := range vpn.peers {
		fmt.Println("Peer:")
		fmt.Print(peer.String())
	}

	c := filepath.Join(testpath, "bad_keys.conf")
	if _, err := ReadVPN(c); err == nil {
		t.Error("Expected an error while reading VPN config")
	} else {
		fmt.Println(err)
	}

	c = filepath.Join(testpath, "bad_peer.conf")
	if _, err := ReadVPN(c); err == nil {
		t.Error("Expected an error while reading VPN config")
	} else {
		fmt.Println(err)
	}
}

func TestAddMetadata(t *testing.T) {
	title("Testing adding metadata")
	vpn, err := ReadVPN(conf)
	if err != nil {
		t.Fatal(err)
	}
	err = vpn.AddMetadata(meta)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", vpn.metadata)

}

func TestServerFromSection(t *testing.T) {
	s := NewSection("noname")
	s.Set("Address", "192.168.0.-1")
	if _, err := serverFromSection(s); err == nil {
		t.Errorf("Expected an error")
	} else {
		fmt.Println(err)
	}
	s.Set("Address", "192.168.0.1/24")

	s.Set("PrivateKey", badCharKeyTest)
	if _, err := serverFromSection(s); err == nil {
		t.Errorf("Expected an error")
	} else {
		fmt.Println(err)
	}
	s.Set("PrivateKey", privateKeyTest)

	s.Set("ListenPort", "3.4")
	if _, err := serverFromSection(s); err == nil {
		t.Errorf("Expected an error")
	} else {
		fmt.Println(err)
	}
}

func TestPeerFromSection(t *testing.T) {
	s := NewSection("noname")
	s.Set("PublicKey", badLengthKeyTest)
	if _, err := peerFromSection(s); err == nil {
		t.Errorf("Expected an error")
	} else {
		fmt.Println(err)
	}
	s.Set("PublicKey", publicKeyTest)

	s.Set("AllowedIPs", "10.10.10.10/34, fa50:cafe::/32")
	if _, err := peerFromSection(s); err == nil {
		t.Errorf("Expected an error")
	} else {
		fmt.Println(err)
	}
	s.Set("AllowedIPs", "10.10.10.10/32, fa50:cafe::/32")

	s.Set("PresharedKey", badLengthKeyTest)
	if _, err := peerFromSection(s); err == nil {
		t.Errorf("Expected an error")
	} else {
		fmt.Println(err)
	}
}
