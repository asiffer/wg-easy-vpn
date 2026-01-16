package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// testDir creates a temporary directory for test files
func testDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "wg-easy-vpn-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// testConfigPath returns a path for a test config file
func testConfigPath(t *testing.T, dir, name string) string {
	t.Helper()
	return filepath.Join(dir, name+".conf")
}
