package utils

import (
	"os"
	"strings"
)

// IsIPv4ForwardingEnabled checks if IPv4 forwarding is enabled
func IsIPv4ForwardingEnabled() bool {
	return readSysctl("/proc/sys/net/ipv4/ip_forward") == "1"
}

// IsIPv6ForwardingEnabled checks if IPv6 forwarding is enabled
func IsIPv6ForwardingEnabled() bool {
	return readSysctl("/proc/sys/net/ipv6/conf/all/forwarding") == "1"
}

func readSysctl(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
