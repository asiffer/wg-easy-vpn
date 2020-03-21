// metadata_test.go
//
//

package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestLoadMetadata(t *testing.T) {
	title("Testing metadata loading")
	conn := "wg0"
	if _, err := LoadMetadata(conn, "./?"); err == nil {
		t.Errorf("An error should occur while loading metadata")
	} else {
		fmt.Println(err)
	}

	f := filepath.Join(testpath, "bad_keys.conf")
	if _, err := LoadMetadata(conn, f); err == nil {
		t.Errorf("An error should occur while loading metadata")
	} else {
		fmt.Println(err)
	}

	f = filepath.Join(testpath, "bad_dns.conf")
	if _, err := LoadMetadata(conn, f); err == nil {
		t.Errorf("An error should occur while loading metadata")
	} else {
		fmt.Println(err)
	}

	f = filepath.Join(testpath, "no_endpoint.conf")
	if _, err := LoadMetadata(conn, f); err == nil {
		t.Errorf("An error should occur while loading metadata")
	} else {
		fmt.Println(err)
	}

	f = filepath.Join(testpath, "no_network.conf")
	if _, err := LoadMetadata(conn, f); err == nil {
		t.Errorf("An error should occur while loading metadata")
	} else {
		fmt.Println(err)
	}

	f = filepath.Join(testpath, "bad_network.conf")
	if _, err := LoadMetadata(conn, f); err == nil {
		t.Errorf("An error should occur while loading metadata")
	} else {
		fmt.Println(err)
	}
}
