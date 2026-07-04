package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFile(t *testing.T) {
	f := NewFile()
	if f == nil {
		t.Fatal("NewFile() returned nil")
	}
	if len(f.sections) != 0 {
		t.Errorf("NewFile() sections not empty, got %d sections", len(f.sections))
	}
}

func TestFileAddSection(t *testing.T) {
	f := NewFile()

	sec1 := f.AddSection("Interface")
	if sec1 == nil {
		t.Fatal("AddSection() returned nil")
	}
	if sec1.Name() != "Interface" {
		t.Errorf("AddSection().Name() = %q, expected %q", sec1.Name(), "Interface")
	}

	sec2 := f.AddSection("Peer")
	if len(f.sections) != 2 {
		t.Errorf("AddSection() - file has %d sections, expected 2", len(f.sections))
	}

	if sec2.Name() != "Peer" {
		t.Errorf("AddSection().Name() = %q, expected %q", sec2.Name(), "Peer")
	}
}

func TestFileHasSection(t *testing.T) {
	f := NewFile()
	f.AddSection("Interface")
	f.AddSection("Peer")

	if !f.HasSection("Interface") {
		t.Error("HasSection(\"Interface\") returned false")
	}
	if !f.HasSection("Peer") {
		t.Error("HasSection(\"Peer\") returned false")
	}
	if f.HasSection("NonExistent") {
		t.Error("HasSection(\"NonExistent\") returned true")
	}
}

func TestFileGetSection(t *testing.T) {
	f := NewFile()
	f.AddSection("Interface")
	f.AddSection("Peer")

	sec, err := f.GetSection("Interface")
	if err != nil {
		t.Fatalf("GetSection(\"Interface\") failed: %v", err)
	}
	if sec.Name() != "Interface" {
		t.Errorf("GetSection().Name() = %q, expected %q", sec.Name(), "Interface")
	}

	_, err = f.GetSection("NonExistent")
	if err == nil {
		t.Error("GetSection(\"NonExistent\") should return error")
	}
}

func TestFileGetOrCreateSection(t *testing.T) {
	f := NewFile()
	f.AddSection("Interface")

	// Get existing section
	sec1 := f.GetorCreateSection("Interface")
	if sec1.Name() != "Interface" {
		t.Errorf("GetorCreateSection() returned wrong section")
	}

	// Create new section
	sec2 := f.GetorCreateSection("Peer")
	if sec2.Name() != "Peer" {
		t.Errorf("GetorCreateSection() did not create new section")
	}

	if !f.HasSection("Peer") {
		t.Error("GetorCreateSection() did not add section to file")
	}
}

func TestFileSections(t *testing.T) {
	f := NewFile()
	f.AddSection("Interface")
	f.AddSection("Peer")

	sections := f.Sections()
	if len(sections) != 2 {
		t.Errorf("Sections() returned %d sections, expected 2", len(sections))
	}
}

func TestFileString(t *testing.T) {
	f := NewFile()

	// Add Interface section
	iface := f.AddSection("Interface")
	iface.Set("PrivateKey", "testkey123")
	iface.Set("Address", "192.168.1.1/24")
	iface.Set("ListenPort", "52820")

	// Add Peer section
	peer := f.AddSection("Peer")
	peer.Set("PublicKey", "peerpublickey")
	peer.Set("AllowedIPs", "192.168.1.2/32")

	result := f.String()

	// Check Interface section
	if !strings.Contains(result, "[Interface]") {
		t.Error("String() does not contain [Interface] section")
	}
	if !strings.Contains(result, "PrivateKey = testkey123") {
		t.Error("String() does not contain PrivateKey")
	}

	// Check Peer section
	if !strings.Contains(result, "[Peer]") {
		t.Error("String() does not contain [Peer] section")
	}
	if !strings.Contains(result, "PublicKey = peerpublickey") {
		t.Error("String() does not contain PublicKey")
	}
}

func TestFileStringWithDefaultSection(t *testing.T) {
	f := NewFile()

	// Add DEFAULT section (no header)
	def := f.AddSection(DEFAULT_SECTION)
	def.Set("Endpoint", "vpn.example.com:52820")
	def.Set("Network", "192.168.0.0/24")

	// Add Interface section
	iface := f.AddSection("Interface")
	iface.Set("PrivateKey", "testkey123")

	result := f.String()

	// DEFAULT section should not have a header
	if strings.Contains(result, "[DEFAULT]") {
		t.Error("String() contains [DEFAULT] header (should not)")
	}

	// But should contain the values
	if !strings.Contains(result, "Endpoint = vpn.example.com:52820") {
		t.Error("String() does not contain DEFAULT section content")
	}

	// Interface section should have header
	if !strings.Contains(result, "[Interface]") {
		t.Error("String() does not contain [Interface] section")
	}
}

func TestFileSave(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.conf")

	f := NewFile()
	iface := f.AddSection("Interface")
	iface.Set("PrivateKey", "testkey123")
	iface.Set("Address", "192.168.1.1/24")

	err := f.Save(filePath)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	if !FileExists(filePath) {
		t.Error("Save() did not create file")
	}

	// Read and verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "[Interface]") {
		t.Error("Saved file does not contain [Interface] section")
	}
	if !strings.Contains(contentStr, "PrivateKey = testkey123") {
		t.Error("Saved file does not contain PrivateKey")
	}
}

func TestParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.conf")

	// Create a test config file
	configContent := `# This is a comment
[Interface]
PrivateKey = wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=
Address = 192.168.1.1/24
ListenPort = 52820

[Peer]
PublicKey = IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=
AllowedIPs = 192.168.1.2/32
`

	err := os.WriteFile(filePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	f, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	// Check Interface section
	iface, err := f.GetSection("Interface")
	if err != nil {
		t.Fatalf("GetSection(\"Interface\") failed: %v", err)
	}

	privateKey, err := iface.Get("PrivateKey")
	if err != nil {
		t.Fatalf("Get(\"PrivateKey\") failed: %v", err)
	}
	if privateKey != "wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=" {
		t.Errorf("PrivateKey = %q, expected correct value", privateKey)
	}

	port, err := iface.GetUint16("ListenPort")
	if err != nil {
		t.Fatalf("GetUint16(\"ListenPort\") failed: %v", err)
	}
	if port != 52820 {
		t.Errorf("ListenPort = %d, expected 52820", port)
	}

	// Check Peer section
	peer, err := f.GetSection("Peer")
	if err != nil {
		t.Fatalf("GetSection(\"Peer\") failed: %v", err)
	}

	publicKey, err := peer.Get("PublicKey")
	if err != nil {
		t.Fatalf("Get(\"PublicKey\") failed: %v", err)
	}
	if publicKey != "IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=" {
		t.Errorf("PublicKey = %q, expected correct value", publicKey)
	}
}

func TestParseFileWithDefaultSection(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.conf")

	// Create a test config file with metadata at the top
	configContent := `Endpoint = vpn.example.com:52820
Network = 192.168.0.0/24
DNS = 8.8.8.8

[Interface]
PrivateKey = wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=
Address = 192.168.1.1/24
`

	err := os.WriteFile(filePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	f, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	// Check DEFAULT section
	def, err := f.GetSection(DEFAULT_SECTION)
	if err != nil {
		t.Fatalf("GetSection(DEFAULT_SECTION) failed: %v", err)
	}

	endpoint, err := def.Get("Endpoint")
	if err != nil {
		t.Fatalf("Get(\"Endpoint\") failed: %v", err)
	}
	if endpoint != "vpn.example.com:52820" {
		t.Errorf("Endpoint = %q, expected vpn.example.com:52820", endpoint)
	}
}

func TestParseFileWithComments(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.conf")

	// Create a test config with various comment styles
	configContent := `# Hash comment
; Semicolon comment
// Double slash comment
[Interface]
PrivateKey = wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=
# Another comment
Address = 192.168.1.1/24
`

	err := os.WriteFile(filePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	f, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	// Should parse successfully and find the Interface section
	iface, err := f.GetSection("Interface")
	if err != nil {
		t.Fatalf("GetSection(\"Interface\") failed: %v", err)
	}

	if !iface.HasKey("PrivateKey") {
		t.Error("Interface section missing PrivateKey")
	}
	if !iface.HasKey("Address") {
		t.Error("Interface section missing Address")
	}
}

func TestParseFileEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.conf")

	// Create a test config with empty lines
	configContent := `[Interface]

PrivateKey = wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=


Address = 192.168.1.1/24

`

	err := os.WriteFile(filePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	f, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	// Should parse successfully
	iface, err := f.GetSection("Interface")
	if err != nil {
		t.Fatalf("GetSection(\"Interface\") failed: %v", err)
	}

	if !iface.HasKey("PrivateKey") {
		t.Error("Interface section missing PrivateKey")
	}
	if !iface.HasKey("Address") {
		t.Error("Interface section missing Address")
	}
}

func TestParseFileNonExistent(t *testing.T) {
	_, err := ParseFile("/non/existent/file.conf")
	if err == nil {
		t.Error("ParseFile() on non-existent file should return error")
	}
}

func TestParseFileWithInvalidKeyValue(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.conf")

	// Create a test config with invalid key (contains dash which is not allowed)
	configContent := `[Interface]
Invalid-Key = value
`

	err := os.WriteFile(filePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file - should fail due to invalid key
	_, err = ParseFile(filePath)
	if err == nil {
		t.Error("ParseFile() with invalid key (containing dash) should return error")
	}
}

func TestFileWriteTo(t *testing.T) {
	f := NewFile()
	iface := f.AddSection("Interface")
	iface.Set("PrivateKey", "testkey123")

	var buf strings.Builder
	n, err := f.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	if n == 0 {
		t.Error("WriteTo() wrote 0 bytes")
	}

	result := buf.String()
	if !strings.Contains(result, "[Interface]") {
		t.Error("WriteTo() output does not contain [Interface]")
	}
	if !strings.Contains(result, "PrivateKey = testkey123") {
		t.Error("WriteTo() output does not contain PrivateKey")
	}
}
