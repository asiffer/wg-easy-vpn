//
//
//
package main

import (
	"bytes"
	"crypto/sha256"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var (
	pngBaseQrCode  string
	pngTruthQrCode string
	pngTestQrCode  string
	jpgBaseQrCode  string
	jpgTruthQrCode string
	jpgTestQrCode  string
)

func Test0(t *testing.T) {
	pngBaseQrCode = path.Join(testpath, "hello_world0.png")
	pngTruthQrCode = path.Join(testpath, "hello_world1.png")
	pngTestQrCode = path.Join(testpath, "hello_world2.png")
	jpgBaseQrCode = path.Join(testpath, "hello_world0.jpg")
	jpgTruthQrCode = path.Join(testpath, "hello_world1.jpg")
	jpgTestQrCode = path.Join(testpath, "hello_world2.jpg")
}

// func Test0(t *testing.T) {
// 	pngBaseQrCode = path.Join(testpath, "hello_world0.png")
// 	pngTruthQrCode = path.Join(testpath, "hello_world1.png")
// 	pngTestQrCode = path.Join(testpath, "hello_world2.png")
// }

func Test16GSConvert(t *testing.T) {
	file, err := os.Open(pngBaseQrCode)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(file)
	if err != nil {
		t.Error(err)
	}
	img16, err := os.Create(pngTruthQrCode)
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(img16, to16bitsGrayScale(img))
}

func TestQRCode(t *testing.T) {
	title("exporting to QRCode (PNG)")
	msg := "Hello World!"
	reader := bytes.NewReader([]byte(msg))

	out, err := os.Create(pngTestQrCode)
	if err != nil {
		t.Errorf("error while opening PNG test file (%w)", err)
	}

	ExportConfig(reader, out)
	// fmt.Println(err)
	out.Close()

	rawTruth, _ := ioutil.ReadFile(pngTruthQrCode)
	hashTruth := sha256.Sum256(rawTruth)

	rawTest, _ := ioutil.ReadFile(pngTestQrCode)
	hashTest := sha256.Sum256(rawTest)

	if !bytes.Equal(hashTest[:], hashTruth[:]) {
		t.Error("Bad hash")
	}
}

func TestPlainTextQRCode(t *testing.T) {
	title("exporting to QRCode (TXT)")
	msg := "Hello World!"
	reader := bytes.NewReader([]byte(msg))
	ExportConfig(reader, os.Stdout)
}

func TestJPGQRCode(t *testing.T) {
	title("exporting to QRCode (JPG)")
	msg := "Hello World!"
	reader := bytes.NewReader([]byte(msg))
	out, err := os.Create(jpgTestQrCode)
	if err != nil {
		t.Errorf("error while opening PNG test file (%w)", err)
	}
	ExportConfig(reader, out)
}
