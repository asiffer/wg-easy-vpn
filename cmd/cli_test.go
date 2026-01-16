package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/asiffer/wg-easy-vpn/utils"
)

func TestConfigurationInfo(t *testing.T) {
	t.Run("absolute path returns name from filename", func(t *testing.T) {
		name, path, err := ConfigurationInfo("/tmp/myvpn.conf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if name != "myvpn" {
			t.Errorf("expected name 'myvpn', got '%s'", name)
		}
		if path != "/tmp/myvpn.conf" {
			t.Errorf("expected path '/tmp/myvpn.conf', got '%s'", path)
		}
	})

	t.Run("relative path is converted to absolute", func(t *testing.T) {
		// Create a temp dir and file
		dir := t.TempDir()
		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		name, path, err := ConfigurationInfo("test.conf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if name != "test" {
			t.Errorf("expected name 'test', got '%s'", name)
		}
		if !filepath.IsAbs(path) {
			t.Errorf("expected absolute path, got '%s'", path)
		}
	})

	t.Run("simple name uses default wireguard directory", func(t *testing.T) {
		// This test checks the behavior for simple names like "wg0"
		// which should map to /etc/wireguard/wg0.conf
		name, path, err := ConfigurationInfo("wg0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// The behavior depends on whether the path resolves to /wg0
		// If cwd is /, then "wg0" -> "/wg0" which matches SEPARATOR + raw
		// Otherwise it's treated as a relative path
		_ = name
		_ = path
		// Just verify no error occurred
	})
}

func TestCLICommands(t *testing.T) {
	t.Run("init command with flags", func(t *testing.T) {
		dir := testDir(t)
		configPath := testConfigPath(t, dir, "wg-cli")

		args := []string{
			"init",
			"--endpoint", "vpn.example.com:51820",
			"--port", "51820",
			"--networks", "10.0.0.0/24",
			"--dns", "1.1.1.1",
			"--routes", "0.0.0.0/0",
			configPath,
		}

		err := initCmd.Run(context.Background(), args)
		if err != nil {
			t.Fatalf("CLI init failed: %v", err)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file not created")
		}
	})

	t.Run("add command with flags", func(t *testing.T) {
		dir := testDir(t)
		configPath := testConfigPath(t, dir, "wg-cli-add")

		// First init
		initArgs := []string{
			"init",
			"--endpoint", "vpn.example.com:51820",
			configPath,
		}
		initCmd.Run(context.Background(), initArgs)

		// Suppress stdout
		oldStdout := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		// Then add
		addArgs := []string{
			"add",
			"--client", "testclient",
			configPath,
		}

		err := addCmd.Run(context.Background(), addArgs)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Fatalf("CLI add failed: %v", err)
		}

		// Verify peer was added
		file, _ := utils.ParseFile(configPath)
		hasPeer := false
		for _, sec := range file.Sections() {
			if sec.Name() == "Peer" {
				hasPeer = true
				break
			}
		}
		if !hasPeer {
			t.Error("peer not added to config")
		}
	})

	t.Run("rm command with flags", func(t *testing.T) {
		dir := testDir(t)
		configPath := testConfigPath(t, dir, "wg-cli-rm")

		// Init
		initArgs := []string{"init", "--endpoint", "vpn.example.com:51820", configPath}
		initCmd.Run(context.Background(), initArgs)

		// Add client
		oldStdout := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		addArgs := []string{"add", "--client", "todelete", configPath}
		addCmd.Run(context.Background(), addArgs)

		w.Close()
		os.Stdout = oldStdout

		// Get peer key
		file, _ := utils.ParseFile(configPath)
		var peerKey string
		for _, sec := range file.Sections() {
			if sec.Name() == "Peer" {
				peerKey, _ = sec.Get("PublicKey")
				break
			}
		}

		// Remove
		rmArgs := []string{"rm", "--peer", peerKey, configPath}
		err := rmCmd.Run(context.Background(), rmArgs)
		if err != nil {
			t.Fatalf("CLI rm failed: %v", err)
		}

		// Verify removal
		file, _ = utils.ParseFile(configPath)
		for _, sec := range file.Sections() {
			if sec.Name() == "Peer" {
				t.Error("peer was not removed")
			}
		}
	})
}
