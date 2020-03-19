//
//
//
package main

import (
	"fmt"
	"testing"
)

func TestCheckKey(t *testing.T) {
	title("Test checking key")
	keysOK := []string{"alert", "y3s", "Why_n0t", "0x_AZZ9"}
	keysNO := []string{"no-key", "why not", "Y€S", "#", "'", `'`}
	for _, key := range keysNO {
		if err := checkKey(key); err == nil {
			t.Errorf("Expected error, got nil")
		}
	}
	for _, key := range keysOK {
		if err := checkKey(key); err != nil {
			t.Errorf("Expected error, got %v", err)
		}
	}
}

func TestSectionName(t *testing.T) {
	title("Test getting section name")
	name := "testnamesection"
	s := NewSection(name)
	if s.Name() != name {
		t.Errorf("Bad name, expect %s, got %s", name, s.Name())
	}
}

func TestSectionKey(t *testing.T) {
	title("Testing get/set section keys")
	name := "test"
	key := "kkey"
	value := "###"
	unknownKey := key + "-unknown"
	s := NewSection(name)
	s.Set(key, value)
	if s.HasKey(key) != true {
		t.Errorf("Key %s not found", key)
	}
	if v, err := s.Get(key); err != nil || v != value {
		t.Errorf("Expected %s, got %s", value, v)
	}
	if s.HasKey(unknownKey) {
		t.Errorf("Unkown key %s found", unknownKey)
	}
}

func TestSectionGetInt(t *testing.T) {
	title("Testing GetInt")
	name := "integerSection"
	key := "integer"
	value := "17"
	badValue := "-z"
	s := NewSection(name)

	s.Set(key, value)
	if i, err := s.GetInt(key); err != nil {
		t.Errorf("Expected %s, got %d", value, i)
	}
	s.Set(key, badValue)
	if i, err := s.GetInt(key); err == nil {
		t.Errorf("Error expected, got %d", i)
	}
}

func TestBase64(t *testing.T) {
	title("Testing GetBytesFromBase64")
	name := "bytesTest"
	key := "bytes"
	value := "YmFzZTY0"
	badValue := "YmFzZTY01"

	s := NewSection(name)

	s.Set(key, value)
	if b, err := s.GetBytesFromBase64(key); err != nil || string(b) != "base64" {
		t.Errorf("Expected 'base64' encoded in base64")
	}

	s.Set(key, badValue)
	if _, err := s.GetBytesFromBase64(key); err == nil {
		t.Errorf("Expected an error")
	}

}

func TestParseFile(t *testing.T) {
	title("Test parsing file")
	file, err := ParseFile(conf)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(file.String())
	}
}
