//
//
//
package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

const (
	// KeyLen is the curve25519 key size (in bytes)
	KeyLen = 32
	// PSKLen is the length (number of bytes) of pre-shared kay
	PSKLen = 32
)

// Key is either a Public or a Private key (32 bytes)
type Key []byte

// PresharedKey is a 32 bytes random array (Wireguard option)
// WireGuard rests upon peers exchanging static public keys with each othera priori,
// as their static identities. The secrecy of all data sent relies on the
// security of the Curve25519 ECDH function. In order to mitigate any
// futureadvances in quantum computing, WireGuard also supports a mode in
// which any pair of peers might additionally pre-share a single 256-bit
// symmetric encryption key between themselves, in order to add an additional
// layer of symmetric encryption. The attack model here is that adversaries
// may be recording encrypted traffic on a longterm basis, in hopes of someday
// being able to break Curve25519 and decrypt past traffic. While pre-sharing
// symmetric encryption keys is usually troublesome from a key management
// perspective and might be more likely stolen, the idea is that by the time
// quantum computing advances to break Curve25519, this pre-shared symmetric
// key has been long forgotten. And, more importantly, in the shorter term,
// if the pre-shared symmetric key is compromised, the Curve25519 keys still
// provide more than sufficient protection. In lieu of using a completely
// post-quantum crypto system, which as of writing are not practical for use
// here, this optional hybrid approach ofa pre-shared symmetric key to
// complement the elliptic curve cryptography provides a sound and acceptable
// trade-off for the extremely paranoid. Furthermore, it allows for building
// on top of WireGuard sophisticated key-rotation schemes, in order to achieve
// varying types of post-compromise security.
type PresharedKey []byte

// NewKey creates a new key filled with zeros
func NewKey() Key {
	return make([]byte, KeyLen)
}

// NewRandomKey generates a new random key
func NewRandomKey() Key {
	key := NewKey()
	rand.Read(key)
	return key
}

// NewPresharedKey creates a new PSK filled with zeros
func NewPresharedKey() PresharedKey {
	return make([]byte, PSKLen)
}

// NewRandomPresharedKey generates a new random PSK
func NewRandomPresharedKey() PresharedKey {
	psk := NewPresharedKey()
	rand.Read(psk)
	return psk
}

// UpdateFromBytes update a key from a given slice
func (key Key) UpdateFromBytes(slice []byte) error {
	if len(slice) < KeyLen {
		return fmt.Errorf("The input slice is not big enough (expected %d, got %d)", KeyLen, len(slice))
	}
	for i, v := range slice {
		key[i] = v
	}
	return nil
}

// UpdateFromBase64 update a key from a base64 encoded string
func (key Key) UpdateFromBase64(s string) error {
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return key.UpdateFromBytes(raw)
}

func curve25519KeyPair() (Key, Key) {
	private := NewRandomKey()
	return private.Public(), private
}

// Base64 encodes a key
func (key Key) Base64() string {
	return base64.StdEncoding.EncodeToString(key)
}

// Public returns the corresponding public key
func (key Key) Public() Key {
	public, err := curve25519.X25519(key, curve25519.Basepoint)
	if err != nil {
		panic(err)
	}
	return public
}

// Base64 encodes a preshared key
func (psk PresharedKey) Base64() string {
	return base64.StdEncoding.EncodeToString(psk)
}

// IsNull check if the PSK is full of zeros
func (psk PresharedKey) IsNull() bool {
	for i := 0; i < PSKLen; i++ {
		if psk[i] != 0 {
			return false
		}
	}
	return true
}

// UpdateFromBytes update a PSK from a given slice
func (psk PresharedKey) UpdateFromBytes(slice []byte) error {
	if len(slice) < PSKLen {
		return fmt.Errorf("The input slice is not big enough (expected %d, got %d)", KeyLen, len(slice))
	}
	for i, v := range slice {
		psk[i] = v
	}
	return nil
}

// UpdateFromBase64 update a PSK from a base64 encoded string
func (psk PresharedKey) UpdateFromBase64(s string) error {
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return psk.UpdateFromBytes(raw)
}
