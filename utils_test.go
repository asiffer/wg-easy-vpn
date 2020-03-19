//
//
//
package main

import (
	"net"
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
