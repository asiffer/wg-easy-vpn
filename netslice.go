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
	for _, sn := range strings.Split(s, ",") {
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

// Iterate returns a channel yielding IP addresses
// included in IP network
func Iterate(n *net.IPNet) (chan net.IP, chan bool) {
	// create a copy of the IP address setting
	// free bits to zero
	base := n.IP.Mask(n.Mask)
	// init channels
	c := make(chan net.IP, 0)
	stop := make(chan bool, 1)
	//
	size := len(base)
	// get the mask
	frozen, total := n.Mask.Size()
	// number of IP (2^(n-k))
	nIP := 1 << uint(total-frozen)
	// run
	go func() {
		c <- base
		for i := 0; i < nIP-1; i++ {
			select {
			case <-stop:
				// early close
				close(stop)
				close(c)
				// stop
				return
			default:
			}
			// create a new buffer
			// tmp := make([]byte, len(base))
			// byte index (starting from the end)
			k := 1
			// increment last byte
			base[size-k]++
			for base[size-k] == 0 {
				// increment last byte if 255 is reached
				k++
				base[size-k]++
			}
			// copy base into tmp
			// copy(tmp, base)
			// send buffer
			c <- copyIP(base)
		}
		// close
		close(c)
	}()

	return c, stop
}
