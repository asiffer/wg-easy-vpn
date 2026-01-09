package export

import (
	"fmt"
	"image"
	"image/color"
	"io"

	qrc "github.com/boombuler/barcode/qr"
	// qrcode "github.com/skip2/go-qrcode"
)

const (
	// QRCodeSize is the width/height of the exported QR code
	QRCodeSize = 76
	// RawConfigMaxLength is the maximum length (in bytes) of a config
	// file to export to QRCode
	RawConfigMaxLength = 2048
	standardWidth      = 200
	standardBorder     = 16
	terminalBorder     = 4
)

var (
	gray16White = color.Gray16{Y: 65535}
)

// QRCodeToString encodes an image (a qrcode) to a
// printable string
func QRCodeToString(img image.Image) string {
	var output string
	bounds := img.Bounds()

	for i := 0; i < bounds.Max.X-1; i = i + 2 {
		for j := 0; j < bounds.Max.Y; j++ {
			up, _, _, _ := img.At(i, j).RGBA()
			down, _, _, _ := img.At(i+1, j).RGBA()

			if up == 0 {
				if down == 0 {
					output += " "
				} else {
					output += "▄"
				}
			} else {
				if down == 0 {
					output += "▀"
				} else {
					output += "█"
				}
			}
		}
		output += "\n"
	}
	return output
}

func Encode(r io.Reader) (image.Image, error) {
	//  create en empty byte array
	p := make([]byte, RawConfigMaxLength)
	// read input
	n, err := r.Read(p)
	// check error
	if err != nil {
		return nil, fmt.Errorf("error while exporting configuration file (%w)", err)
	}
	// encode input
	qr, err := qrc.Encode(string(p[:n]), qrc.L, qrc.Unicode)
	// check error
	if err != nil {
		return nil, fmt.Errorf("error while encoding to QRCode (%w)", err)
	}

	return addBorder(to16bitsGrayScale(qr), terminalBorder), nil
}

func newGray16White(r image.Rectangle) *image.Gray16 {
	img := image.NewGray16(r)
	bounds := img.Bounds()
	for i := 0; i < bounds.Max.X; i++ {
		for j := 0; j < bounds.Max.Y; j++ {
			img.SetGray16(i, j, gray16White)
		}
	}
	return img
}

func addBorder(b image.Image, width int) image.Image {
	bounds := b.Bounds()
	newWidth := bounds.Max.X + 2*width
	newHeight := bounds.Max.Y + 2*width
	newImage := newGray16White(image.Rect(0, 0, newWidth, newHeight))
	for i := 0; i < bounds.Max.X; i++ {
		for j := 0; j < bounds.Max.Y; j++ {
			newImage.Set(i+width, j+width, b.At(i, j))
		}
	}
	return newImage
}

func to16bitsGrayScale(img image.Image) image.Image {
	out := image.NewGray16(img.Bounds())
	for i := 0; i < img.Bounds().Max.X; i++ {
		for j := 0; j < img.Bounds().Max.Y; j++ {
			out.Set(i, j, color.Gray16Model.Convert(img.At(i, j)))
		}
	}
	return out
}
