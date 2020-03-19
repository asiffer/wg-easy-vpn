//
//
//
package main

import (
	"fmt"
	"path"
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
}

func TestAddMetadata(t *testing.T) {
	title("Testing adding metadata")
	vpn, err := ReadVPN(conf)
	if err != nil {
		t.Error(err)
	}
	err = vpn.AddMetadata(meta)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", vpn.metadata)
}
