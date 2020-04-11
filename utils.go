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
		return "", fmt.Errorf("Error while getting Interface section from %s (%v)", p, err)
	}

	key, err := sec.GetKeyFromBase64("PrivateKey")
	if err != nil {
		return "", fmt.Errorf("Error while retrieving PrivateKey from %s (%v)", file, err)
	}
	return key.Public().Base64(), nil
}

// map clientName->Key
func extractPairsFromFolder(folder string, clientAsKey bool) map[string]string {
	pairs := make(map[string]string)
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil
	}
	for _, f := range files {
		// fmt.Println(f.Name(), f)
		name := strings.TrimSuffix(f.Name(), DefaultConfigSuffix)
		if !f.IsDir() {
			if pk, err := getPublicKeyFromFile(path.Join(folder, f.Name())); err == nil {
				if clientAsKey {
					pairs[name] = pk
				} else {
					pairs[pk] = name
				}
			}
		}
	}
	return pairs
}

// func getClientsFromFolder(folder string) []string {
// 	c := make([]string, 0)
// 	files, err := ioutil.ReadDir(folder)
// 	if err != nil {
// 		return nil
// 	}
// 	for _, f := range files {
// 		name := strings.TrimSuffix(f.Name(), DefaultConfigSuffix)
// 		if !f.IsDir() {
// 			c = append(c, name)
// 		}
// 	}
// 	return c
// }

// func isIPv4(ip net.IP) bool {
// 	return ip.To4() != nil
// }

// func printIPNet(ipnet *net.IPNet) string {
// 	return ipnet.String()
// }

func parseAddressAndMask(ipMask string) (*net.IPNet, error) {
	ip, ipnet, err := net.ParseCIDR(ipMask)
	if err != nil {
		return nil, err
	}
	return &net.IPNet{IP: ip, Mask: ipnet.Mask}, nil
}

func incrementIP(ip net.IP) net.IP {
	bytesIP := []byte(ip)
	lastByte := int(bytesIP[len(bytesIP)-1])
	if lastByte == 255 {
		return nil
	}
	bytesIP[len(bytesIP)-1] = byte(lastByte + 1)
	return net.IP(bytesIP)
}

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

func fileExist(file string) bool {
	_, err := os.Stat(file)
	if err == nil {
		return true
	}
	return false
}

func copyIP(ip net.IP) net.IP {
	tmp := make([]byte, len(ip))
	copy(tmp, ip)
	return tmp
}

func findIP(ip net.IP, slice []net.IP) int {
	for i, addr := range slice {
		if ip.Equal(addr) {
			return i
		}
	}
	return -1
}
