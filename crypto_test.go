//
//
//
package main

import (
	"testing"
)

const (
	privateKeyTest = "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE="
	publicKeyTest  = "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024="
	pskTest        = "qCJhKwR0uMEx8LbqvJbBx9LetPHA3zZp61M6TXcTaJ8="
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
}
