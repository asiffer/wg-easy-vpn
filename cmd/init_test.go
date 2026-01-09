package cmd

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestInitCmd(t *testing.T) {
	// resetConfig()
	ctx := context.Background()
	args := []string{
		initCmd.Name,
		"--endpoint", "example.org:52820",
		"/tmp/wg0.conf",
	}
	if err := initCmd.Run(ctx, args); err != nil {
		t.Fatalf("initCmd.Run() failed: %v", err)
	}
}

func TestAddCmd(t *testing.T) {
	// resetConfig()
	// First, initialize a VPN
	ctx := context.Background()
	initArgs := []string{
		initCmd.Name,
		"--endpoint", "example.org:52820",
		"/tmp/wg0.conf",
	}
	if err := initCmd.Run(ctx, initArgs); err != nil {
		t.Fatalf("initCmd.Run() failed: %v", err)
	}

	bytes, err := os.ReadFile("/tmp/wg0.conf")
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	fmt.Println(string(bytes))
	fmt.Println("--------------------------------------------------------")

	// resetConfig()
	// Now, add a client
	addArgs := []string{
		addCmd.Name,
		"--client", "client1",
		// "--routes", "0.0.0.0/0",
		// "--qrcode",
		"/tmp/wg0.conf",
	}
	if err := addCmd.Run(ctx, addArgs); err != nil {
		t.Fatalf("addCmd.Run() failed: %v", err)
	}
	fmt.Println("--------------------------------------------------------")

	// resetConfig()
	// Now, add a client
	addArgs2 := []string{
		addCmd.Name,
		"--client", "client2",
		"--routes", "10.8.0.0/16",
		// "--qrcode",
		"/tmp/wg0.conf",
	}
	if err := addCmd.Run(ctx, addArgs2); err != nil {
		t.Fatalf("addCmd.Run() failed: %v", err)
	}

	fmt.Println("--------------------------------------------------------")

	bytes, err = os.ReadFile("/tmp/wg0.conf")
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	fmt.Println(string(bytes))

}
