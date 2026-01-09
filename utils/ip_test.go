package utils

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestCleanString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid characters",
			input:    "client-name_123",
			expected: "client-name_123",
		},
		{
			name:     "with spaces",
			input:    "  client name  ",
			expected: "clientname",
		},
		{
			name:     "special characters",
			input:    "client@#$%name",
			expected: "clientname",
		},
		{
			name:     "dots and dashes",
			input:    "client.name-123",
			expected: "client.name-123",
		},
		{
			name:     "mixed case",
			input:    "ClientName",
			expected: "ClientName",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special chars",
			input:    "@#$%^&*()",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanString(tt.input)
			if result != tt.expected {
				t.Errorf("CleanString(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseIPList(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		expectErr bool
	}{
		{
			name:      "valid IPv4 list",
			input:     []string{"192.168.1.1", "10.0.0.1"},
			expectErr: false,
		},
		{
			name:      "valid IPv6 list",
			input:     []string{"::1", "2001:db8::1"},
			expectErr: false,
		},
		{
			name:      "mixed IP versions",
			input:     []string{"192.168.1.1", "::1"},
			expectErr: false,
		},
		{
			name:      "invalid IP",
			input:     []string{"192.168.1.256"},
			expectErr: true,
		},
		{
			name:      "empty list",
			input:     []string{},
			expectErr: false,
		},
		{
			name:      "invalid format",
			input:     []string{"not-an-ip"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseIPList(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseIPList(%v) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseIPList(%v) unexpected error: %v", tt.input, err)
				}
				if len(result) != len(tt.input) {
					t.Errorf("ParseIPList(%v) returned %d IPs, expected %d", tt.input, len(result), len(tt.input))
				}
			}
		})
	}
}

func TestParseIPNetList(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		expectErr bool
	}{
		{
			name:      "valid CIDR list",
			input:     []string{"192.168.1.0/24", "10.0.0.0/8"},
			expectErr: false,
		},
		{
			name:      "valid IPv6 CIDR",
			input:     []string{"2001:db8::/32"},
			expectErr: false,
		},
		{
			name:      "invalid CIDR",
			input:     []string{"192.168.1.0/33"},
			expectErr: true,
		},
		{
			name:      "missing mask",
			input:     []string{"192.168.1.0"},
			expectErr: true,
		},
		{
			name:      "empty list",
			input:     []string{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseIPNetList(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseIPNetList(%v) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseIPNetList(%v) unexpected error: %v", tt.input, err)
				}
				if len(result) != len(tt.input) {
					t.Errorf("ParseIPNetList(%v) returned %d networks, expected %d", tt.input, len(result), len(tt.input))
				}
			}
		})
	}
}

func TestStringifyIPs(t *testing.T) {
	tests := []struct {
		name     string
		input    []net.IP
		expected int // expected number of strings
	}{
		{
			name:     "IPv4 addresses",
			input:    []net.IP{net.ParseIP("192.168.1.1"), net.ParseIP("10.0.0.1")},
			expected: 2,
		},
		{
			name:     "IPv6 addresses",
			input:    []net.IP{net.ParseIP("::1"), net.ParseIP("2001:db8::1")},
			expected: 2,
		},
		{
			name:     "empty list",
			input:    []net.IP{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringifyIPs(tt.input)
			if len(result) != tt.expected {
				t.Errorf("StringifyIPs() returned %d strings, expected %d", len(result), tt.expected)
			}
		})
	}
}

func TestCopyIP(t *testing.T) {
	original := net.ParseIP("192.168.1.1")
	copy := CopyIP(original)

	if !original.Equal(copy) {
		t.Errorf("CopyIP() did not create equal IP")
	}

	// Modify the copy
	copy[len(copy)-1] = 99

	// Original should be unchanged
	if original.Equal(copy) {
		t.Errorf("CopyIP() did not create independent copy")
	}
}

func TestFindIP(t *testing.T) {
	slice := []net.IP{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.2"),
		net.ParseIP("::1"),
	}

	tests := []struct {
		name     string
		ip       net.IP
		expected int
	}{
		{
			name:     "found at beginning",
			ip:       net.ParseIP("192.168.1.1"),
			expected: 0,
		},
		{
			name:     "found in middle",
			ip:       net.ParseIP("192.168.1.2"),
			expected: 1,
		},
		{
			name:     "found at end",
			ip:       net.ParseIP("::1"),
			expected: 2,
		},
		{
			name:     "not found",
			ip:       net.ParseIP("10.0.0.1"),
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindIP(tt.ip, slice)
			if result != tt.expected {
				t.Errorf("FindIP() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestStringifyNetworks(t *testing.T) {
	_, net1, _ := net.ParseCIDR("192.168.1.0/24")
	_, net2, _ := net.ParseCIDR("10.0.0.0/8")

	networks := []net.IPNet{*net1, *net2}
	result := StringifyNetworks(networks)

	if len(result) != 2 {
		t.Errorf("StringifyNetworks() returned %d strings, expected 2", len(result))
	}

	if result[0] != "192.168.1.0/24" {
		t.Errorf("StringifyNetworks()[0] = %q, expected %q", result[0], "192.168.1.0/24")
	}
}

func TestIterate(t *testing.T) {
	_, network, err := net.ParseCIDR("192.168.1.0/30")
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}

	ctx := context.Background()
	ipChan := Iterate(ctx, network)

	ips := make([]net.IP, 0)
	for ip := range ipChan {
		ips = append(ips, ip)
	}

	// /30 network should yield 3 IPs (excluding network address)
	// Network: 192.168.1.0 (not yielded by the iterator starting from base+1)
	// Usable: 192.168.1.1, 192.168.1.2
	// Broadcast: 192.168.1.3 (last one)
	expectedCount := 3 // Based on the implementation: (2^(32-30))-1 = 3
	if len(ips) != expectedCount {
		t.Errorf("Iterate() yielded %d IPs, expected %d", len(ips), expectedCount)
	}
}

func TestIterateWithCancel(t *testing.T) {
	_, network, err := net.ParseCIDR("192.168.0.0/24")
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ipChan := Iterate(ctx, network)

	// Read a few IPs then cancel
	count := 0
	maxCount := 5
	for ip := range ipChan {
		if ip != nil {
			count++
		}
		if count >= maxCount {
			cancel()
		}
	}

	if count != maxCount {
		t.Errorf("Iterate() with cancel: got %d IPs, expected %d", count, maxCount)
	}
}

func TestIterateIPv6(t *testing.T) {
	_, network, err := net.ParseCIDR("2001:db8::/126")
	if err != nil {
		t.Fatalf("Failed to parse IPv6 CIDR: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ipChan := Iterate(ctx, network)

	ips := make([]net.IP, 0)
	for ip := range ipChan {
		ips = append(ips, ip)
	}

	// /126 network should yield 3 IPs (2^(128-126) - 1)
	expectedCount := 3
	if len(ips) != expectedCount {
		t.Errorf("Iterate() IPv6 yielded %d IPs, expected %d", len(ips), expectedCount)
	}
}
