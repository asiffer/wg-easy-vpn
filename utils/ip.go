package utils

import (
	"context"
	"fmt"
	"net"
	"strings"
)

const (
	// AllowedChars is the list of chars you can use to define
	// the server or the clients
	AllowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-."
	// IPv4Len is the lenght (in bits) of an IPv4 address
	IPv4Len = 8 * net.IPv4len
	// IPv6Len is the lenght (in bits) of an IPv6 address
	IPv6Len = 8 * net.IPv6len
)

var (
	// IPv4ZeroNet is the null IPv4 network 0.0.0.0/0
	IPv4ZeroNet = net.IPNet{
		IP:   net.IPv4zero,
		Mask: []byte{0, 0, 0, 0},
	}
	// IPv6ZeroNet is the null IPv6 network ::/0
	IPv6ZeroNet = net.IPNet{
		IP:   net.IPv6zero,
		Mask: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
)

// cleanString removes characters which are not allowed
func CleanString(input string) string {
	output := ""
	input = strings.TrimSpace(input)
	for _, c := range input {
		if strings.ContainsRune(AllowedChars, c) {
			output += string(c)
		}
	}
	return output
}

func ParseIPList(l []string) ([]net.IP, error) {
	length := len(l)
	out := make([]net.IP, length)
	for i := 0; i < length; i++ {
		out[i] = net.ParseIP(l[i])
		if out[i] == nil {
			return nil, fmt.Errorf("error while parsing IP: %s", l[i])
		}
	}
	return out, nil
}

func ParseIPNetList(l []string) ([]net.IPNet, error) {
	length := len(l)
	out := make([]net.IPNet, length)
	for i := 0; i < length; i++ {
		_, n, err := net.ParseCIDR(l[i])
		if err != nil {
			return nil, fmt.Errorf("error while parsing IPNet: %s", l[i])
		}
		out[i] = *n
	}
	return out, nil
}

func StringifyIPs(l []net.IP) []string {
	length := len(l)
	out := make([]string, length)
	for i := 0; i < length; i++ {
		out[i] = l[i].String()
	}
	return out
}

func CopyIP(ip net.IP) net.IP {
	tmp := make([]byte, len(ip))
	copy(tmp, ip)
	return tmp
}

func FindIP(ip net.IP, slice []net.IP) int {
	for i, addr := range slice {
		if ip.Equal(addr) {
			return i
		}
	}
	return -1
}

func StringifyNetworks(nets []net.IPNet) []string {
	strs := make([]string, len(nets))
	for i, n := range nets {
		strs[i] = n.String()
	}
	return strs
}

// Iterate returns a channel yielding IP addresses
// included in IP network. It accepts a context for cancellation.
// When the context is cancelled, the iteration stops and the channel is closed.
func Iterate(ctx context.Context, n *net.IPNet) chan net.IP {
	// create a copy of the IP address setting (.Mask() creates a new slice)
	// free bits to zero
	base := n.IP.Mask(n.Mask)
	// init channels
	c := make(chan net.IP)
	//
	size := len(base)
	// get the mask
	frozen, total := n.Mask.Size()
	// number of IP (2^(n-k))
	nIP := 1 << uint(total-frozen)
	// run
	go func() {
		defer close(c)

		for i := 0; i < nIP-1; i++ {
			// byte index (starting from the end)
			k := 1
			// increment last byte
			base[size-k]++
			for base[size-k] == 0 {
				// increment last byte if 255 is reached
				k++
				base[size-k]++
			}
			// send buffer
			select {
			case c <- CopyIP(base):
			case <-ctx.Done():
				return
			}
		}
	}()

	return c
}
