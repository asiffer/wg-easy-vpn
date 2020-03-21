//
//
//
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"strings"

	"github.com/boombuler/barcode"
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

func to16bitsGrayScale(img image.Image) image.Image {
	out := image.NewGray16(img.Bounds())
	for i := 0; i < img.Bounds().Max.X; i++ {
		for j := 0; j < img.Bounds().Max.Y; j++ {
			out.Set(i, j, color.Gray16Model.Convert(img.At(i, j)))
		}
	}
	return out
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

func standardImage(img image.Image) image.Image {
	return addBorder(to16bitsGrayScale(img), standardBorder)
}

// PNG encode an image through a PNG
func PNG(w io.Writer, b barcode.Barcode) error {
	// scale
	qr, err := barcode.Scale(b,
		standardWidth-2*standardBorder,
		standardWidth-2*standardBorder)
	if err != nil {
		return err
	}
	return png.Encode(w, standardImage(qr))
}

// JPG encode an image through a JPG
func JPG(w io.Writer, b barcode.Barcode) error {
	// scale
	qr, err := barcode.Scale(b,
		standardWidth-2*standardBorder,
		standardWidth-2*standardBorder)
	if err != nil {
		return err
	}
	return jpeg.Encode(w, standardImage(qr), nil)
}

// TXT encode an image through text
func TXT(w io.Writer, b barcode.Barcode) error {
	b, err := downScale(b)
	if err != nil {
		return err
	}

	var img image.Image
	if w == os.Stdout {
		img = addBorder(to16bitsGrayScale(b), terminalBorder)
	} else {
		img = standardImage(b)
	}

	str := QRCodeToString(img)
	_, err = w.Write([]byte(str))
	return err
}

// QRCodeToString encodes an image (a qrcode) to a
// printable string
func QRCodeToString(img image.Image) string {
	var output string
	bounds := img.Bounds()

	for i := 0; i < bounds.Max.X; i = i + 2 {
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

func downScale(bc barcode.Barcode) (barcode.Barcode, error) {
	bounds := bc.Bounds().Max
	width := bounds.X
	height := bounds.Y
	// Check that sizes are even
	if height%2 != 0 {
		height++
	}
	if width%2 != 0 {
		width++
	}
	// rescale
	bc, err := barcode.Scale(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("Error while scaling QRCode (%w)", err)
	}
	return bc, nil
}

func encodeInput(r io.Reader) (barcode.Barcode, error) {
	//  create en empty byte array
	p := make([]byte, RawConfigMaxLength)
	// read input
	n, err := r.Read(p)
	// check error
	if err != nil {
		return nil, fmt.Errorf("Error while exporting configuration file (%w)", err)
	}
	// encode input
	qr, err := qrc.Encode(string(p[:n]), qrc.L, qrc.Unicode)
	// check error
	if err != nil {
		return nil, fmt.Errorf("Error while encoding to QRCode (%w)", err)
	}
	return qr, nil
}

// ExportConfig prints a QRCode which embeds the configuration of a client
func ExportConfig(r io.Reader, file *os.File) error {
	// encoding input
	qr, err := encodeInput(r)
	if err != nil {
		return err
	}
	// retrieving file extension
	ext := strings.Replace(path.Ext(file.Name()), ".", "", -1)
	// saving the qr code to file (or stdout)
	switch ext {
	case "jpg", "JPG", "jpeg", "JPEG":
		return JPG(file, qr)
	case "png", "PNG":
		return PNG(file, qr)
	default:
		return TXT(file, qr)
	}
}
