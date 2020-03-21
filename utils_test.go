//
//
//
package main

import (
	"fmt"
	"net"
	"path"
	"testing"
)

var (
	ipList = []net.IP{
		net.IPv4(127, 0, 0, 100),
		net.IPv4(10, 10, 10, 10),
		net.IPv6loopback,
	}
	ipStrList = []string{
		"127.0.0.100",
		"10.10.10.10",
		"::1",
	}
)

func TestMapIPStrList(t *testing.T) {
	title("Testing mapping ip string list")
	list, err := mapIPStrList(ipStrList)
	if err != nil {
		t.Errorf("%w", err)
	}
	for i, ip := range list {
		if !ip.Equal(ipList[i]) {
			t.Errorf("Expecting %v, got %v", ipList[i], ip)
		}
	}
}

func TestMapIPList(t *testing.T) {
	title("Testing mapping ip list")
	for i, ip := range mapIPList(ipList) {
		if ip != ipStrList[i] {
			t.Errorf("Expecting %s, got %s", ipStrList[i], ip)
		}
	}
}

func TestGetPublicKeyFromFile(t *testing.T) {
	title("Testing getting public key from file")
	test := path.Join(testpath, "malformed_key.conf")
	if _, err := getPublicKeyFromFile(test); err == nil {
		t.Errorf("An error must occured while parsing %s", test)
	} else {
		fmt.Println(err)
	}

	test = path.Join(testpath, "no_interface_section.conf")
	if _, err := getPublicKeyFromFile(test); err == nil {
		t.Errorf("An error must occured while parsing %s", test)
	} else {
		fmt.Println(err)
	}

	test = path.Join(testpath, "bad_keys.conf")
	if _, err := getPublicKeyFromFile(test); err == nil {
		t.Errorf("An error must occured while parsing %s", test)
	} else {
		fmt.Println(err)
	}
}

func TestParseAddressAndMask(t *testing.T) {
	ips := []string{
		"192.168.256.0/24",
		"192.168.1.0/33",
		"192.168.256.0.1",
	}

	for _, ip := range ips {
		if _, err := parseAddressAndMask(ip); err == nil {
			t.Errorf("An error must occured while parsing %s", ip)
		}
	}
}
