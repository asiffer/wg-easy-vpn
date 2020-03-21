//
//
package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"testing"
)

const (
	dns      = "1.1.1.1"
	network  = "10.0.10.1/24"
	port     = 12000
	endpoint = "wg.example.org"
	route    = "0.0.0.0/0"
)

func _Reset() {
	// runtime
	initRuntime()
	// init App
	initApp()
}

func createServer(name string) (string, error) {
	_Reset()
	// TempDir
	dir, err := ioutil.TempDir(os.TempDir(), "wg_")
	if err != nil {
		return dir, err
	}

	cmdl := []string{
		app.Name,
		"create",
		"--server-dir", dir,
		"--net", network,
		"--port", fmt.Sprintf("%d", port),
		"--endpoint", endpoint,
		"--dns", dns, name,
	}

	return dir, app.Run(cmdl)
}

func createServerWithClients(name string, nbClients int) (string, error) {
	// TempDir
	dir, err := createServer(name)
	if err != nil {
		return dir, err
	}

	clientDir := path.Join(dir, "clients")

	_Reset()
	// fake commands
	cmdl := []string{
		app.Name,
		"add",
		"--server-dir", dir,
		"--route", route,
		"--client-dir", clientDir,
	}
	for i := 0; i < nbClients; i++ {
		cmdl = append(cmdl, "--client", fmt.Sprintf("client%d", i))
	}

	// add connection name
	cmdl = append(cmdl, name)

	// parse fake commands
	// f.Parse(cmdl)

	// create context and launch command
	// c := cli.NewContext(app, f, nil)
	return dir, app.Run(cmdl)
}

func TestCreate(t *testing.T) {
	title("Testing creating WG server")

	dir, err := createServer("wgtest")
	if err != nil {
		t.Fatalf("Erorr while creating server (%v)", err)
	}
	// check metadata
	metadata, err := LoadMetadata("wgtest", filepath.Join(dir, ".wg-easy-vpn.conf"))
	if err != nil {
		t.Errorf("Error after metadata loading (%v)", err)
	}
	if !net.ParseIP(dns).Equal(metadata.dns[0]) {
		t.Errorf("Error in dns IP, expected %s, got %v", dns, metadata.dns)
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("%v", err)
	}
}

func TestCreateWithError(t *testing.T) {
	title("Testing creating WG server twice")
	name := "wgerror"
	dir, _ := createServer(name)

	_Reset()
	// fake commands
	cmdl := []string{
		app.Name,
		"create",
		"--server-dir", dir,
		"--port", fmt.Sprintf("%d", port),
		"--endpoint", endpoint,
		"--dns", dns,
		name,
	}

	// the server config file already exists
	if err := app.Run(cmdl); err == nil {
		t.Errorf("An error should occur (same connection name)")
	}

	// removing the server config file
	if err := os.RemoveAll(path.Join(dir, name+DefaultConfigSuffix)); err != nil {
		t.Errorf("%v", err)
	}

	_Reset()
	// try again (now metadata fails)
	if err := app.Run(cmdl); err == nil {
		t.Errorf("An error should occur (same connection name)")
	}

	_Reset()
	// now force
	cmdl = []string{
		app.Name,
		"create",
		"--server-dir", dir,
		"--net", network,
		"--port", fmt.Sprintf("%d", port),
		"--endpoint", endpoint,
		"--dns", dns,
		"--force",
		name,
	}

	if err := app.Run(cmdl); err != nil {
		t.Errorf("Error while forcing server creation (%v)", err)
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("%v", err)
	}
}

func TestAdd(t *testing.T) {
	title("Testing adding clients to WG server")

	connName := "wg0"

	dir, err := createServer(connName)
	if err != nil {
		t.Fatalf("Erorr while creating server (%v)", err)
	}

	clientDir := path.Join(dir, "clients/")

	_Reset()
	// fake commands
	cmdl := []string{
		app.Name,
		"add",
		"--server-dir", dir,
		"--route", route,
		"--client-dir", clientDir,
		"--client", "client0",
		"--client", "client1",
		"--client", "client2",
		"--export",
		connName,
	}

	if err := app.Run(cmdl); err != nil {
		t.Errorf("%v", err)
	}

	// check connection
	vpn, err := ReadVPN(path.Join(dir, connName+DefaultConfigSuffix))
	if err != nil {
		t.Fatalf("Error while reading VPN config %s (%v)", path.Join(dir, connName), err)
	}
	if p := vpn.NumberOfPeers(); p != 3 {
		t.Errorf("Expected 3 peers, got %d", p)
	}

	base := *vpn.server.address.Copy()
	for _, peer := range vpn.peers {
		base.Increment()
		for i, a := range *peer.allowedIPs {
			if !a.IP.Equal(base[i].IP) {
				t.Errorf("Expected %s, got %s", base.String(), peer.allowedIPs.String())
			}
		}

	}

	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("%v", err)
	}

}

func TestShow(t *testing.T) {
	title("Testing showing clients of WG server")

	connName := "wg1"

	dir, err := createServerWithClients(connName, 5)
	if err != nil {
		t.Fatalf("Error while creating server with clients (%v)", err)
	}
	clientDir := path.Join(dir, "clients")

	_Reset()
	cmdl := []string{
		app.Name, "show",
		"--server-dir", dir,
		"--client-dir", clientDir,
		connName,
	}

	if err := app.Run(cmdl); err != nil {
		t.Errorf("%v", err)
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("%v", err)
	}

}

func TestRm(t *testing.T) {
	title("Testing removing clients")

	connName := "wgRM"

	dir, err := createServerWithClients(connName, 3)
	if err != nil {
		t.Fatalf("Erorr while creating server (%v)", err)
	}
	clientDir := path.Join(dir, "clients")

	_Reset()
	cmdl := []string{
		app.Name, "rm",
		"--server-dir", dir,
		"--client-dir", clientDir,
		"-c", "client0", "-c", "client1",
		connName,
	}

	if err := app.Run(cmdl); err != nil {
		t.Errorf("%v", err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("%v", err)
	}
}
