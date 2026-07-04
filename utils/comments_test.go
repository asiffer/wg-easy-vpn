package utils

import "testing"

func TestRemoveComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "hash comment",
			input:    "# This is a comment",
			expected: "",
		},
		{
			name:     "semicolon comment",
			input:    "; This is a comment",
			expected: "",
		},
		{
			name:     "double slash comment",
			input:    "// This is a comment",
			expected: "",
		},
		{
			name:     "no comment",
			input:    "ListenPort = 52820",
			expected: "ListenPort = 52820",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "hash in middle (not a comment)",
			input:    "key = value # not stripped",
			expected: "key = value # not stripped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComment(tt.input)
			if result != tt.expected {
				t.Errorf("removeComment(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
