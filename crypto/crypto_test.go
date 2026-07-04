package crypto

import (
	"testing"
)

const (
	privateKeyTest   = "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE="
	publicKeyTest    = "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024="
	pskTest          = "qCJhKwR0uMEx8LbqvJbBx9LetPHA3zZp61M6TXcTaJ8="
	badLengthKeyTest = "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024"
	badCharKeyTest   = "IYIgnBITiOdCJUyg/c0jpP!!0+OWVhcWw/CS5FIpG024="
)

func TestKeyPair(t *testing.T) {
	private := NewRandomKey()
	if err := private.UpdateFromBase64(privateKeyTest); err != nil {
		t.Fatalf("Failed to update private key: %v", err)
	}
	public := private.Public()
	if public.Base64() != publicKeyTest {
		t.Errorf("Bad key (expected %s, got %s)", publicKeyTest, public.Base64())
	}
}

func TestNewKey(t *testing.T) {
	k := NewKey()
	if len(k) != KeyLen {
		t.Errorf("Expected key length %d, got %d", KeyLen, len(k))
	}
	for _, b := range k {
		if b != 0 {
			t.Errorf("Expected zero-filled key, got non-zero byte")
			break
		}
	}
}

func TestNewRandomKey(t *testing.T) {
	k1 := NewRandomKey()
	k2 := NewRandomKey()

	if len(k1) != KeyLen {
		t.Errorf("Expected key length %d, got %d", KeyLen, len(k1))
	}

	if k1.Base64() == k2.Base64() {
		t.Error("Two random keys should not be identical")
	}
}

func TestKeyUpdateFromBytes(t *testing.T) {
	k := NewKey()
	testBytes := make([]byte, KeyLen)
	for i := range testBytes {
		testBytes[i] = byte(i)
	}

	if err := k.UpdateFromBytes(testBytes); err != nil {
		t.Fatalf("Failed to update key from bytes: %v", err)
	}

	for i := range testBytes {
		if k[i] != testBytes[i] {
			t.Errorf("Byte mismatch at index %d: expected %d, got %d", i, testBytes[i], k[i])
		}
	}
}

func TestKeyUpdateFromBytesShortSlice(t *testing.T) {
	k := NewKey()
	if err := k.UpdateFromBytes([]byte{0, 0, 1, 1}); err == nil {
		t.Error("Expected error for short byte slice")
	}
}

func TestKeyUpdateFromBase64(t *testing.T) {
	k := NewKey()
	if err := k.UpdateFromBase64(publicKeyTest); err != nil {
		t.Fatalf("Failed to update key from base64: %v", err)
	}
	if k.Base64() != publicKeyTest {
		t.Errorf("Expected %s, got %s", publicKeyTest, k.Base64())
	}
}

func TestKeyUpdateFromBase64BadLength(t *testing.T) {
	k := NewKey()
	if err := k.UpdateFromBase64(badLengthKeyTest); err == nil {
		t.Error("Expected error for bad length key")
	}
}

func TestKeyUpdateFromBase64BadChars(t *testing.T) {
	k := NewKey()
	if err := k.UpdateFromBase64(badCharKeyTest); err == nil {
		t.Error("Expected error for bad character key")
	}
}

func TestNewPresharedKey(t *testing.T) {
	psk := NewPresharedKey()
	if len(psk) != PSKLen {
		t.Errorf("Expected PSK length %d, got %d", PSKLen, len(psk))
	}
	for _, b := range psk {
		if b != 0 {
			t.Errorf("Expected zero-filled PSK, got non-zero byte")
			break
		}
	}
}

func TestNewRandomPresharedKey(t *testing.T) {
	psk1 := NewRandomPresharedKey()
	psk2 := NewRandomPresharedKey()

	if len(psk1) != PSKLen {
		t.Errorf("Expected PSK length %d, got %d", PSKLen, len(psk1))
	}

	if psk1.Base64() == psk2.Base64() {
		t.Error("Two random PSKs should not be identical")
	}
}

func TestPSKUpdateFromBase64(t *testing.T) {
	psk := NewPresharedKey()
	if err := psk.UpdateFromBase64(pskTest); err != nil {
		t.Fatalf("Failed to update PSK: %v", err)
	}
	if psk.Base64() != pskTest {
		t.Errorf("Bad PSK, expected %s, got %s", pskTest, psk.Base64())
	}
}

func TestPSKUpdateFromBase64BadLength(t *testing.T) {
	psk := NewPresharedKey()
	if err := psk.UpdateFromBase64(badLengthKeyTest); err == nil {
		t.Error("Expected error for bad length key")
	}
}

func TestPSKUpdateFromBytes(t *testing.T) {
	psk := NewPresharedKey()
	testBytes := make([]byte, PSKLen)
	for i := range testBytes {
		testBytes[i] = byte(i)
	}

	if err := psk.UpdateFromBytes(testBytes); err != nil {
		t.Fatalf("Failed to update PSK from bytes: %v", err)
	}

	for i := range testBytes {
		if psk[i] != testBytes[i] {
			t.Errorf("Byte mismatch at index %d: expected %d, got %d", i, testBytes[i], psk[i])
		}
	}
}

func TestPSKUpdateFromBytesShortSlice(t *testing.T) {
	psk := NewPresharedKey()
	if err := psk.UpdateFromBytes([]byte{0, 0, 1, 1}); err == nil {
		t.Error("Expected error for short byte slice")
	}
}

func TestKeyBase64RoundTrip(t *testing.T) {
	k1 := NewRandomKey()
	encoded := k1.Base64()

	k2 := NewKey()
	if err := k2.UpdateFromBase64(encoded); err != nil {
		t.Fatalf("Failed to decode key: %v", err)
	}

	if k1.Base64() != k2.Base64() {
		t.Error("Round-trip encoding/decoding failed")
	}
}

func TestPSKBase64RoundTrip(t *testing.T) {
	psk1 := NewRandomPresharedKey()
	encoded := psk1.Base64()

	psk2 := NewPresharedKey()
	if err := psk2.UpdateFromBase64(encoded); err != nil {
		t.Fatalf("Failed to decode PSK: %v", err)
	}

	if psk1.Base64() != psk2.Base64() {
		t.Error("Round-trip encoding/decoding failed")
	}
}

func TestPublicKeyDerivation(t *testing.T) {
	private := NewRandomKey()
	public1 := private.Public()
	public2 := private.Public()

	if len(public1) != KeyLen {
		t.Errorf("Expected public key length %d, got %d", KeyLen, len(public1))
	}

	if public1.Base64() != public2.Base64() {
		t.Error("Multiple calls to Public() should return same result")
	}
}
