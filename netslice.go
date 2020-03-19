// netslice.go
//

package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// NetSlice represents a slice of IPNet
type NetSlice []*net.IPNet

// NewNetSlice inits a new slice
func NewNetSlice() NetSlice {
	return make([]*net.IPNet, 0)
}

// NewNetSliceFromString inits a new slice based on a
// string looking like "addr/mask, addr/mask ..."
func NewNetSliceFromString(s string) (NetSlice, error) {
	slice := NewNetSlice()
	// fmt.Println(s, strings.Split(s, ","))
	for _, sn := range strings.Split(s, ",") {
		// fmt.Println(sn)
		n, err := parseAddressAndMask(strings.TrimSpace(sn))
		if err != nil {
			return nil, err
		}
		slice.Append(n)
	}
	return slice, nil
}

// NewNetSliceFromStringSlice inits a new slice based on a
// slice of strings looking like ["addr/mask", "addr/mask" ...]
func NewNetSliceFromStringSlice(slice []string) (NetSlice, error) {
	if len(slice) == 0 {
		return nil, fmt.Errorf("Empty input slice")
	}
	return NewNetSliceFromString(strings.Join(slice, ","))
}

// Append adds a new element in the slice
func (ns *NetSlice) Append(n *net.IPNet) {
	*ns = append(*ns, n)
	// fmt.Printf("%v\n", ns)
}

// Increment change every IP in the nets of the slice
func (ns *NetSlice) Increment() error {
	for _, n := range *ns {
		newIP := incrementIP(n.IP)
		if newIP == nil {
			return errors.New("Cannot increment (limit case, last byte is 255)")
		}
		n.IP = newIP
	}
	return nil
}

// Copy creates a copy of the slice
func (ns *NetSlice) Copy() *NetSlice {
	cp := NewNetSlice()
	for _, n := range *ns {
		// empty structure
		ipn := net.IPNet{
			IP:   make([]byte, len(n.IP)),
			Mask: make([]byte, len(n.Mask)),
		}
		// copy slices
		copy(ipn.IP, n.IP)
		copy(ipn.Mask, n.Mask)
		// append network
		cp.Append(&ipn)
	}
	return &cp
}

// Len return the length of the netslice
func (ns *NetSlice) Len() int {
	return len(*ns)
}

func (ns *NetSlice) String() string {
	s := make([]string, ns.Len())
	for i, n := range *ns {
		s[i] = n.String()
	}
	return strings.Join(s, ", ")
}
