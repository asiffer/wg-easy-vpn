package utils

import "testing"

func TestCheckKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		expectErr bool
	}{
		{
			name:      "valid key",
			key:       "ValidKey",
			expectErr: false,
		},
		{
			name:      "key with underscore",
			key:       "Valid_Key",
			expectErr: false,
		},
		{
			name:      "key with dash - not allowed",
			key:       "Valid-Key",
			expectErr: true,
		},
		{
			name:      "key with numbers",
			key:       "ValidKey123",
			expectErr: false,
		},
		{
			name:      "empty key - no validation",
			key:       "",
			expectErr: false,
		},
		{
			name:      "key with spaces",
			key:       "Invalid Key",
			expectErr: true,
		},
		{
			name:      "key with special chars",
			key:       "Invalid@Key",
			expectErr: true,
		},
		{
			name:      "key starting with number - allowed",
			key:       "1InvalidKey",
			expectErr: false,
		},
		{
			name:      "standard WireGuard keys",
			key:       "PrivateKey",
			expectErr: false,
		},
		{
			name:      "PublicKey",
			key:       "PublicKey",
			expectErr: false,
		},
		{
			name:      "PresharedKey",
			key:       "PresharedKey",
			expectErr: false,
		},
		{
			name:      "AllowedIPs",
			key:       "AllowedIPs",
			expectErr: false,
		},
		{
			name:      "Endpoint",
			key:       "Endpoint",
			expectErr: false,
		},
		{
			name:      "ListenPort",
			key:       "ListenPort",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkKey(tt.key)
			if tt.expectErr {
				if err == nil {
					t.Errorf("checkKey(%q) expected error, got nil", tt.key)
				}
			} else {
				if err != nil {
					t.Errorf("checkKey(%q) unexpected error: %v", tt.key, err)
				}
			}
		})
	}
}
