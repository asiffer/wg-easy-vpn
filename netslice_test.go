//
//

package main

import (
	"net"
	"strings"
	"testing"
)

// NewNetSliceFromString inits a new slice based on a
// string looking like "addr/mask, addr/mask ..."
func TestNewNetSliceFromString(t *testing.T) {
	slice := NewNetSlice()
	if len(slice) != 0 {
		t.Errorf("Expected an empty slice, got %d", len(slice))
	}

	_, n, err := net.ParseCIDR("127.0.0.1/32")
	if err != nil {
		t.Errorf("%v", err)
	}
	slice.Append(n)
	if len(slice) != 1 {
		t.Errorf("Error when appending, expected a slice with size 1, got %d", len(slice))
	}

	str := " 192.168.0.1/24, fe80::/64"
	slice, err = NewNetSliceFromString(str)
	if err != nil {
		t.Errorf("An error occured (%v)", err)
	}
	if len(slice) != 2 {
		t.Errorf("Expected a slice of size 2, got %d", len(slice))
	}

	if s := slice.String(); s != strings.TrimSpace(str) {
		t.Errorf("Bad string, expected %s, got %s", strings.TrimSpace(str), s)
	}
}

func TestIterate(t *testing.T) {
	network := &net.IPNet{
		IP:   net.IPv4(192, 168, 3, 10),
		Mask: net.CIDRMask(17, 32),
	}

	flow, _ := Iterate(network)
	for {
		ip, ok := <-flow
		if !ok {
			break
		}
		if !network.Contains(ip) {
			t.Errorf("The network %s does not contain %s", network.String(), ip.String())
		}
	}

	// early stop
	flow, stop := Iterate(network)
	for i := 0; i < 10; i++ {
		<-flow
	}
	stop <- true
}

// Append adds a new element in the slice
// func (ns NetSlice) Append(n *net.IPNet) {
// 	ns = append(ns, n)
// }

// Increment change every IP in the nets of the slice
// func (ns NetSlice) Increment() error {
// 	for _, n := range ns {
// 		newIP := incrementIP(n.IP)
// 		if newIP == nil {
// 			return errors.New("Cannot increment (limit case, last byte is 255)")
// 		}
// 		n.IP = newIP
// 	}
// 	return nil
// }

// Copy creates a copy of the slice
// func (ns NetSlice) Copy() NetSlice {
// 	cp := NewNetSlice()
// 	for _, n := range ns {
// 		cp.Append(
// 			&net.IPNet{
// 				IP:   n.IP,
// 				Mask: n.Mask,
// 			},
// 		)
// 	}
// 	return cp
// }
