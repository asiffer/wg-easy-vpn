package utils

import (
	"runtime"
	"testing"
)

func TestGetDefaultInterface(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping test on non-Linux platform")
	}

	iface, err := GetDefaultInterface()
	if err != nil {
		t.Fatalf("GetDefaultInterface() error = %v", err)
	}
	if iface == "" {
		t.Error("GetDefaultInterface() returned empty interface name")
	}
	t.Logf("default interface: %s", iface)
}
