//
//
//

package main

import (
	"bytes"
	"fmt"
	"net"
	"testing"
)

var (
	key        = NewRandomKey()
	psk        = NewPresharedKey()
	address, _ = NewNetSliceFromString("10.10.10.10/24")
	writer     = bytes.NewBuffer(make([]byte, 0))
	nodeDNS    = []net.IP{net.ParseIP("1.1.1.1")}
)

func TestWriteNode(t *testing.T) {
	title("Test writing Node structure into buffers")
	writer.Reset()

	node := NewWGNode(&address, true)
	psk = node.psk
	_, err := node.Write(writer)
	if err != nil {
		t.Errorf("Error while writing WGNode to buffer (%v)", err)
	}

	truth := fmt.Sprintf("[Interface]\nAddress = %s\nPrivateKey = %s\n", node.address.String(), node.private.Base64())
	got := writer.String()
	if got != truth {
		t.Errorf("Strings mismatch")
		fmt.Println(truth, len(truth))
		fmt.Println(got, len(got))
	}
}

func TestWriteClient(t *testing.T) {
	title("Test writing Client structure into buffers")
	writer.Reset()

	client := NewWGClient(&address, true, nodeDNS)
	psk = client.psk
	_, err := client.Write(writer)
	if err != nil {
		t.Errorf("Error while writing WGClient to buffer (%v)", err)
	}

	truth := fmt.Sprintf("[Interface]\nAddress = %s\nPrivateKey = %s\nDNS = %s\n",
		client.address.String(), client.private.Base64(), client.DNS())
	got := writer.String()
	if got != truth {
		t.Errorf("Strings mismatch")
		fmt.Println(truth, len(truth))
		fmt.Println(got, len(got))
	}

	peer := client.ToPeer()
	writer.Reset()
	_, err = peer.Write(writer)
	if err != nil {
		t.Errorf("Error while writing WGClientAsPeer to buffer (%v)", err)
	}
	truth = fmt.Sprintf("[Peer]\nPublicKey = %s\nPresharedKey = %s\nAllowedIPs = %s\n",
		peer.Public(), peer.PSK(), peer.AllowedIPs())
	got = writer.String()
	if got != truth {
		t.Errorf("Strings mismatch")
		fmt.Println(truth, len(truth))
		fmt.Println(got, len(got))
	}

}

func TestWriteServer(t *testing.T) {
	title("Test writing Server structure into buffers")
	writer.Reset()

	server := NewWGServer(&address, true, port)
	psk = server.psk
	_, err := server.Write(writer)
	if err != nil {
		t.Errorf("Error while writing WGServer to buffer (%v)", err)
	}

	truth := fmt.Sprintf("[Interface]\nAddress = %s\nPrivateKey = %s\nListenPort = %d\n",
		server.address.String(), server.private.Base64(), server.port)
	got := writer.String()
	if got != truth {
		t.Errorf("Strings mismatch")
		fmt.Println(truth, len(truth))
		fmt.Println(got, len(got))
	}

	r, err := NewNetSliceFromString("0.0.0.0/0")
	if err != nil {
		t.Fatalf("Error while parsing route (%v)", err)
	}
	peer := server.ToPeer(&r, endpoint)
	writer.Reset()
	_, err = peer.Write(writer)
	if err != nil {
		t.Errorf("Error while writing WGServerAsPeer to buffer (%v)", err)
	}
	truth = fmt.Sprintf("[Peer]\nPublicKey = %s\nPresharedKey = %s\nAllowedIPs = %s\nEndpoint = %s\n",
		peer.Public(), peer.PSK(), peer.AllowedIPs(), fmt.Sprintf("%s:%d", endpoint, port))
	got = writer.String()
	if got != truth {
		t.Errorf("Strings mismatch")
		fmt.Println(truth, len(truth))
		fmt.Println(got, len(got))
	}
}
