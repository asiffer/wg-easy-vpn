package export

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"strings"
	"testing"
)

func TestQRCodeToString(t *testing.T) {
	t.Run("small image with mixed colors", func(t *testing.T) {
		// Create a 4x4 test image
		img := image.NewGray16(image.Rect(0, 0, 4, 4))

		// Set pixels: black = 0, white = 65535
		// Row 0-1: black-black, white-white (should produce " " and "█")
		img.SetGray16(0, 0, color.Gray16{Y: 0})     // up black
		img.SetGray16(1, 0, color.Gray16{Y: 0})     // down black
		img.SetGray16(0, 1, color.Gray16{Y: 65535}) // up white
		img.SetGray16(1, 1, color.Gray16{Y: 65535}) // down white

		// Row 2-3: white-black, black-white (should produce "▀" and "▄")
		img.SetGray16(2, 0, color.Gray16{Y: 65535}) // up white
		img.SetGray16(3, 0, color.Gray16{Y: 0})     // down black
		img.SetGray16(2, 1, color.Gray16{Y: 0})     // up black
		img.SetGray16(3, 1, color.Gray16{Y: 65535}) // down white

		output := QRCodeToString(img)

		// Should have newlines
		if !strings.Contains(output, "\n") {
			t.Error("expected output to contain newlines")
		}

		// Output should not be empty
		if len(output) == 0 {
			t.Error("expected non-empty output")
		}
	})

	t.Run("all black image", func(t *testing.T) {
		img := image.NewGray16(image.Rect(0, 0, 4, 4))
		// All pixels default to black (0)

		output := QRCodeToString(img)

		// All black should produce spaces
		lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
		for _, line := range lines {
			for _, r := range line {
				if r != ' ' {
					t.Errorf("expected space for all-black image, got '%c'", r)
				}
			}
		}
	})

	t.Run("all white image", func(t *testing.T) {
		img := image.NewGray16(image.Rect(0, 0, 4, 4))
		bounds := img.Bounds()
		for i := 0; i < bounds.Max.X; i++ {
			for j := 0; j < bounds.Max.Y; j++ {
				img.SetGray16(i, j, gray16White)
			}
		}

		output := QRCodeToString(img)

		// All white should produce full blocks
		lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
		for _, line := range lines {
			for _, r := range line {
				if r != '█' {
					t.Errorf("expected '█' for all-white image, got '%c'", r)
				}
			}
		}
	})

	t.Run("output dimensions", func(t *testing.T) {
		// Create 8x6 image (height 8, width 6)
		img := image.NewGray16(image.Rect(0, 0, 8, 6))

		output := QRCodeToString(img)
		lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")

		// Height is processed 2 rows at a time, so 8/2 = 4 lines
		// But the loop goes to Max.X-1, so (8-1)/2 = 3 iterations + potential partial
		// Actually: for i := 0; i < 8-1; i += 2 gives i=0,2,4,6 = 4 iterations
		expectedLines := 4
		if len(lines) != expectedLines {
			t.Errorf("expected %d lines, got %d", expectedLines, len(lines))
		}

		// Each line should have width characters (6)
		for i, line := range lines {
			if len(line) != 6 {
				t.Errorf("line %d: expected 6 characters, got %d", i, len(line))
			}
		}
	})
}

func TestEncode(t *testing.T) {
	t.Run("encode simple text", func(t *testing.T) {
		input := "Hello, World!"
		reader := strings.NewReader(input)

		img, err := Encode(reader)
		if err != nil {
			t.Fatalf("failed to encode: %v", err)
		}

		if img == nil {
			t.Fatal("expected non-nil image")
		}

		bounds := img.Bounds()
		if bounds.Max.X == 0 || bounds.Max.Y == 0 {
			t.Error("expected non-zero image dimensions")
		}

		// Image should be square (QR codes are square) plus border
		// The border is added symmetrically, so width should equal height
		if bounds.Max.X != bounds.Max.Y {
			t.Errorf("expected square image, got %dx%d", bounds.Max.X, bounds.Max.Y)
		}
	})

	t.Run("encode WireGuard config", func(t *testing.T) {
		config := `[Interface]
Address = 10.0.0.2/24
PrivateKey = wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=
DNS = 1.1.1.1

[Peer]
PublicKey = IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=
AllowedIPs = 0.0.0.0/0
Endpoint = vpn.example.com:51820`

		reader := strings.NewReader(config)
		img, err := Encode(reader)
		if err != nil {
			t.Fatalf("failed to encode WireGuard config: %v", err)
		}

		if img == nil {
			t.Fatal("expected non-nil image")
		}

		// Should be able to convert to string
		output := QRCodeToString(img)
		if len(output) == 0 {
			t.Error("expected non-empty string output")
		}
	})

	t.Run("encode empty reader returns error", func(t *testing.T) {
		reader := strings.NewReader("")

		_, err := Encode(reader)
		// Empty read should return io.EOF error
		if err == nil {
			t.Error("expected error for empty reader")
		}
	})

	t.Run("encode reader error propagates", func(t *testing.T) {
		reader := &errorReader{err: io.ErrUnexpectedEOF}

		_, err := Encode(reader)
		if err == nil {
			t.Error("expected error to propagate")
		}
	})

	t.Run("encode max length content", func(t *testing.T) {
		// Create content just under max length
		content := strings.Repeat("A", RawConfigMaxLength-100)
		reader := strings.NewReader(content)

		img, err := Encode(reader)
		if err != nil {
			t.Fatalf("failed to encode large content: %v", err)
		}

		if img == nil {
			t.Fatal("expected non-nil image")
		}
	})

	t.Run("image has border", func(t *testing.T) {
		input := "Test"
		reader := strings.NewReader(input)

		img, err := Encode(reader)
		if err != nil {
			t.Fatalf("failed to encode: %v", err)
		}

		bounds := img.Bounds()

		// Check corners are white (border)
		// Top-left corner
		r, g, b, _ := img.At(0, 0).RGBA()
		if r == 0 && g == 0 && b == 0 {
			t.Error("expected white border at top-left corner")
		}

		// Bottom-right corner
		r, g, b, _ = img.At(bounds.Max.X-1, bounds.Max.Y-1).RGBA()
		if r == 0 && g == 0 && b == 0 {
			t.Error("expected white border at bottom-right corner")
		}
	})
}

func TestNewGray16White(t *testing.T) {
	t.Run("creates white image", func(t *testing.T) {
		rect := image.Rect(0, 0, 10, 10)
		img := newGray16White(rect)

		if img == nil {
			t.Fatal("expected non-nil image")
		}

		bounds := img.Bounds()
		if bounds.Max.X != 10 || bounds.Max.Y != 10 {
			t.Errorf("expected 10x10 image, got %dx%d", bounds.Max.X, bounds.Max.Y)
		}

		// All pixels should be white
		for i := 0; i < bounds.Max.X; i++ {
			for j := 0; j < bounds.Max.Y; j++ {
				c := img.Gray16At(i, j)
				if c.Y != 65535 {
					t.Errorf("pixel (%d,%d) expected white (65535), got %d", i, j, c.Y)
				}
			}
		}
	})

	t.Run("zero size rectangle", func(t *testing.T) {
		rect := image.Rect(0, 0, 0, 0)
		img := newGray16White(rect)

		if img == nil {
			t.Fatal("expected non-nil image")
		}

		bounds := img.Bounds()
		if bounds.Max.X != 0 || bounds.Max.Y != 0 {
			t.Errorf("expected 0x0 image, got %dx%d", bounds.Max.X, bounds.Max.Y)
		}
	})
}

func TestAddBorder(t *testing.T) {
	t.Run("adds correct border width", func(t *testing.T) {
		original := image.NewGray16(image.Rect(0, 0, 10, 10))
		borderWidth := 5

		result := addBorder(original, borderWidth)

		bounds := result.Bounds()
		expectedSize := 10 + 2*borderWidth // 20

		if bounds.Max.X != expectedSize {
			t.Errorf("expected width %d, got %d", expectedSize, bounds.Max.X)
		}
		if bounds.Max.Y != expectedSize {
			t.Errorf("expected height %d, got %d", expectedSize, bounds.Max.Y)
		}
	})

	t.Run("preserves original content", func(t *testing.T) {
		original := image.NewGray16(image.Rect(0, 0, 4, 4))
		// Set a specific pixel to black
		original.SetGray16(2, 2, color.Gray16{Y: 0})

		borderWidth := 2
		result := addBorder(original, borderWidth)

		// The pixel at (2,2) in original should be at (4,4) in result
		r, _, _, _ := result.At(4, 4).RGBA()
		if r != 0 {
			t.Error("expected black pixel to be preserved at offset position")
		}
	})

	t.Run("border is white", func(t *testing.T) {
		original := image.NewGray16(image.Rect(0, 0, 4, 4))
		// Set all original pixels to black
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				original.SetGray16(i, j, color.Gray16{Y: 0})
			}
		}

		borderWidth := 2
		result := addBorder(original, borderWidth)

		// Check border pixels are white
		// Top border
		for i := 0; i < 8; i++ {
			for j := 0; j < 2; j++ {
				r, _, _, _ := result.At(i, j).RGBA()
				if r == 0 {
					t.Errorf("expected white border at (%d,%d)", i, j)
				}
			}
		}

		// Left border
		for i := 0; i < 2; i++ {
			for j := 0; j < 8; j++ {
				r, _, _, _ := result.At(i, j).RGBA()
				if r == 0 {
					t.Errorf("expected white border at (%d,%d)", i, j)
				}
			}
		}
	})

	t.Run("zero border width", func(t *testing.T) {
		original := image.NewGray16(image.Rect(0, 0, 5, 5))
		result := addBorder(original, 0)

		bounds := result.Bounds()
		if bounds.Max.X != 5 || bounds.Max.Y != 5 {
			t.Errorf("expected same size with zero border, got %dx%d", bounds.Max.X, bounds.Max.Y)
		}
	})
}

func TestTo16bitsGrayScale(t *testing.T) {
	t.Run("converts RGBA to grayscale", func(t *testing.T) {
		original := image.NewRGBA(image.Rect(0, 0, 4, 4))
		// Set some colored pixels
		original.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})     // Red
		original.Set(1, 1, color.RGBA{R: 0, G: 255, B: 0, A: 255})     // Green
		original.Set(2, 2, color.RGBA{R: 0, G: 0, B: 255, A: 255})     // Blue
		original.Set(3, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255}) // White

		result := to16bitsGrayScale(original)

		// Result should be Gray16
		gray16, ok := result.(*image.Gray16)
		if !ok {
			t.Fatal("expected *image.Gray16")
		}

		// White should remain bright
		white := gray16.Gray16At(3, 3)
		if white.Y < 60000 { // Should be close to max
			t.Errorf("expected white to be bright, got %d", white.Y)
		}
	})

	t.Run("preserves dimensions", func(t *testing.T) {
		original := image.NewRGBA(image.Rect(0, 0, 7, 11))
		result := to16bitsGrayScale(original)

		bounds := result.Bounds()
		if bounds.Max.X != 7 || bounds.Max.Y != 11 {
			t.Errorf("expected 7x11, got %dx%d", bounds.Max.X, bounds.Max.Y)
		}
	})

	t.Run("black stays black", func(t *testing.T) {
		original := image.NewRGBA(image.Rect(0, 0, 2, 2))
		// Default is black (0,0,0,0)

		result := to16bitsGrayScale(original)
		gray16 := result.(*image.Gray16)

		for i := 0; i < 2; i++ {
			for j := 0; j < 2; j++ {
				c := gray16.Gray16At(i, j)
				if c.Y != 0 {
					t.Errorf("expected black (0), got %d at (%d,%d)", c.Y, i, j)
				}
			}
		}
	})
}

func TestConstants(t *testing.T) {
	t.Run("QRCodeSize is reasonable", func(t *testing.T) {
		if QRCodeSize < 10 || QRCodeSize > 500 {
			t.Errorf("QRCodeSize %d seems unreasonable", QRCodeSize)
		}
	})

	t.Run("RawConfigMaxLength is sufficient for WireGuard configs", func(t *testing.T) {
		// A typical WireGuard config is around 300-500 bytes
		if RawConfigMaxLength < 1000 {
			t.Errorf("RawConfigMaxLength %d may be too small for WireGuard configs", RawConfigMaxLength)
		}
	})

	t.Run("terminal border is smaller than standard border", func(t *testing.T) {
		if terminalBorder >= standardBorder {
			t.Errorf("terminal border (%d) should be smaller than standard border (%d)",
				terminalBorder, standardBorder)
		}
	})
}

func TestEndToEnd(t *testing.T) {
	t.Run("full pipeline: encode then convert to string", func(t *testing.T) {
		config := `[Interface]
Address = 10.0.0.2/24
PrivateKey = wDx8ruBJgk2ZmDwgHkZfnoaSdfCgXUb4MwJ87psOJGE=

[Peer]
PublicKey = IYIgnBITiOdCJUyg/c0jpPi0+OWVhcWw/CS5FIpG024=
AllowedIPs = 0.0.0.0/0
Endpoint = vpn.example.com:51820`

		img, err := Encode(strings.NewReader(config))
		if err != nil {
			t.Fatalf("failed to encode: %v", err)
		}

		output := QRCodeToString(img)

		// Output should have content
		if len(output) == 0 {
			t.Error("expected non-empty output")
		}

		// Should contain only valid characters
		validChars := " ▀▄█\n"
		for _, r := range output {
			if !strings.ContainsRune(validChars, r) {
				t.Errorf("unexpected character in output: '%c' (U+%04X)", r, r)
			}
		}

		// Should have multiple lines
		lines := strings.Split(output, "\n")
		if len(lines) < 5 {
			t.Errorf("expected multiple lines, got %d", len(lines))
		}
	})

	t.Run("different inputs produce different outputs", func(t *testing.T) {
		img1, _ := Encode(strings.NewReader("Input One"))
		img2, _ := Encode(strings.NewReader("Input Two"))

		output1 := QRCodeToString(img1)
		output2 := QRCodeToString(img2)

		if output1 == output2 {
			t.Error("expected different inputs to produce different outputs")
		}
	})
}

// errorReader is a helper that always returns an error
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

// Verify errorReader implements io.Reader
var _ io.Reader = (*errorReader)(nil)

func TestEncodeWithBytesReader(t *testing.T) {
	t.Run("works with bytes.Reader", func(t *testing.T) {
		data := []byte("Test data for QR code")
		reader := bytes.NewReader(data)

		img, err := Encode(reader)
		if err != nil {
			t.Fatalf("failed to encode from bytes.Reader: %v", err)
		}

		if img == nil {
			t.Fatal("expected non-nil image")
		}
	})

	t.Run("works with bytes.Buffer", func(t *testing.T) {
		var buf bytes.Buffer
		buf.WriteString("Test data for QR code")

		img, err := Encode(&buf)
		if err != nil {
			t.Fatalf("failed to encode from bytes.Buffer: %v", err)
		}

		if img == nil {
			t.Fatal("expected non-nil image")
		}
	})
}
