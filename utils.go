//
//
//
package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
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

func getPublicKeyFromFile(p string) (string, error) {
	// vpn, err := ReadVPN(file)
	file, err := ParseFile(p)
	if err != nil {
		return "", fmt.Errorf("Error while parsing %s (%v)", file, err)
	}

	sec, err := file.GetSection("Interface")
	if err != nil {
		return "", fmt.Errorf("Error while getting Interface section from %s (%v)", file, err)
	}

	key, err := sec.GetKeyFromBase64("PrivateKey")
	if err != nil {
		return "", fmt.Errorf("Error while retrieving PrivateKey from %s (%v)", file, err)
	}
	return key.Public().Base64(), nil
}

func extractPairsFromFolder(folder string) map[string]string {
	pairs := make(map[string]string)
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil
	}
	for _, f := range files {
		// fmt.Println(f.Name(), f)
		name := strings.TrimSuffix(f.Name(), DefaultConfigSuffix)
		if !f.IsDir() {
			// fmt.Println(path.Join(folder, f.Name()))
			// fmt.Println(getPublicKeyFromFile(path.Join(folder, f.Name())))
			if pk, err := getPublicKeyFromFile(path.Join(folder, f.Name())); err == nil {
				pairs[pk] = name
			}
		}
	}
	return pairs
}

// func fatal(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }

func isIPv4(ip net.IP) bool {
	return ip.To4() != nil
}

// fullMask returns an IP address with a mask with all bits set to 1 (/32 or /128)
// func fullMask(ip net.IP) net.IP {
// 	var mask net.IPMask
// 	if isIPv4(ip) {
// 		mask = net.CIDRMask(IPv4Len, IPv4Len)
// 	} else {
// 		mask = net.CIDRMask(IPv6Len, IPv6Len)
// 	}
// 	return ip.Mask(mask)
// }

// func nullMask(ip net.IP) net.IP {
// 	var mask net.IPMask
// 	if isIPv4(ip) {
// 		mask = net.CIDRMask(0, IPv4Len)
// 	} else {
// 		mask = net.CIDRMask(0, IPv6Len)
// 	}
// 	return ip.Mask(mask)
// }

// Return a 32 bytes (256 bits) random
// pre-shared key (PSK)
// func genPSK() []byte {
// 	// init PSK
// 	psk := make([]byte, PSKLen)
// 	// fill it with random
// 	rand.Read(psk)
// 	return psk
// }

func printIPNet(ipnet *net.IPNet) string {
	return ipnet.String()
}

func parseAddressAndMask(ipMask string) (*net.IPNet, error) {
	ip, ipnet, err := net.ParseCIDR(ipMask)
	if err != nil {
		return nil, err
	}
	return &net.IPNet{IP: ip, Mask: ipnet.Mask}, nil
}

// func fullMaskIPNet(ip *net.IPNet) *net.IPNet {
// 	_, total := ip.Mask.Size()
// 	return &net.IPNet{IP: ip.IP, Mask: net.CIDRMask(total, total)}
// }

// func maxLenKey(m map[string]string) int {
// 	mlk := 0
// 	for k := range m {
// 		if len(k) > mlk {
// 			mlk = len(k)
// 		}
// 	}
// 	return mlk
// }

func incrementIP(ip net.IP) net.IP {
	bytesIP := []byte(ip)
	lastByte := int(bytesIP[len(bytesIP)-1])
	if lastByte == 255 {
		return nil
	}
	bytesIP[len(bytesIP)-1] = byte(lastByte + 1)
	return net.IP(bytesIP)
}

// it increments an array an create a copy
// func incrementIPSlice(array []*net.IPNet) []*net.IPNet {
// 	out := make([]*net.IPNet, len(array))
// 	for i, n := range array {
// 		out[i] = &net.IPNet{
// 			IP:   incrementIP(n.IP),
// 			Mask: n.Mask,
// 		}
// 	}
// 	return out
// }

// create a copy
// func copyIPSlice(array []*net.IPNet) []*net.IPNet {
// 	out := make([]*net.IPNet, len(array))
// 	for i, n := range array {
// 		out[i] = &net.IPNet{
// 			IP:   n.IP,
// 			Mask: n.Mask,
// 		}
// 	}
// 	return out
// }

// cleanString removes characters which are not allowed
func cleanString(input string) string {
	output := ""
	input = strings.TrimSpace(input)
	for _, c := range input {
		if strings.ContainsRune(AllowedChars, c) {
			output += string(c)
		}
	}
	return output
}

func mapIPStrList(l []string) ([]net.IP, error) {
	length := len(l)
	out := make([]net.IP, length)
	for i := 0; i < length; i++ {
		out[i] = net.ParseIP(l[i])
		if out[i] == nil {
			return nil, fmt.Errorf("Error while parsing IP: %s", l[i])
		}
	}
	return out, nil
}

func mapIPList(l []net.IP) []string {
	length := len(l)
	out := make([]string, length)
	for i := 0; i < length; i++ {
		out[i] = l[i].String()
	}
	return out
}

// func mapIPNetStrList(l []string) ([]*net.IPNet, error) {
// 	var ip net.IP
// 	var err error
// 	length := len(l)
// 	out := make([]*net.IPNet, length)
// 	for i := 0; i < length; i++ {
// 		ip, out[i], err = net.ParseCIDR(l[i])
// 		if err != nil {
// 			return nil, fmt.Errorf("Error while parsing network %s (%w)", l[i], err)
// 		}
// 		out[i].IP = ip
// 	}
// 	return out, nil
// }

// func mapIPNetList(l []*net.IPNet) []string {
// 	length := len(l)
// 	out := make([]string, length)
// 	for i := 0; i < length; i++ {
// 		out[i] = l[i].String()
// 	}
// 	return out
// }

func fileExist(file string) bool {
	_, err := os.Stat(file)
	if err == nil {
		return true
	}
	return false
}
