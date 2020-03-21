//
//
//
package main

import (
	"testing"
)

const (
	privateKeyTest   = "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE="
	publicKeyTest    = "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024="
	pskTest          = "qCJhKwR0uMEx8LbqvJbBx9LetPHA3zZp61M6TXcTaJ8="
	badLengthKeyTest = "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024"
	badCharKeyTest   = "IYIgnBITiOdCJUyg/c0jpP!!0+OWVhcWw/CS5FIpG024"
)

func TestKeyPair(t *testing.T) {
	title("Testing key pair")
	private := NewRandomKey()
	private.UpdateFromBase64(privateKeyTest)
	public := private.Public()
	if public.Base64() != publicKeyTest {
		t.Errorf("Bad key (expected %s, got %s)", publicKeyTest, public.Base64())
	}
}

func TestUpdatePSK(t *testing.T) {
	title("Testing PSK update")
	psk := NewPresharedKey()
	psk.UpdateFromBase64(pskTest)
	if psk.Base64() != pskTest {
		t.Errorf("Bad PSK, expected %s, got %s", pskTest, psk.Base64())
	}

	if err := psk.UpdateFromBase64(badLengthKeyTest); err == nil {
		t.Errorf("Expected a bad length key")
	}

	if err := psk.UpdateFromBytes([]byte{0, 0, 1, 1}); err == nil {
		t.Errorf("Expected a bad length key")
	}
}

func TestUpdateKy(t *testing.T) {
	title("Testing Key update")
	k := NewKey()
	k.UpdateFromBase64(publicKeyTest)

	if err := k.UpdateFromBase64(badLengthKeyTest); err == nil {
		t.Errorf("Expected a bad length key")
	}

	if err := k.UpdateFromBytes([]byte{0, 0, 1, 1}); err == nil {
		t.Errorf("Expected a bad length key")
	}
}
