package main

import (
	"fmt"
	"strings"
)

const (
	titleSize = 80
	titlePad  = "-"
	// docFile   = "wg-easy-vpn.ex"
)

func title(s string) {
	length := len(s) + 2
	var leftLength, rightLength int
	if (titleSize-length)%2 == 0 {
		leftLength = (titleSize - length) / 2

	} else {
		leftLength = (titleSize - length - 1) / 2
	}
	rightLength = titleSize - leftLength - length
	fmt.Println(strings.Repeat(titlePad, leftLength),
		s,
		strings.Repeat(titlePad, rightLength))
}

// func TestDoc(t *testing.T) {
// 	f, err := os.Create(docFile)
// 	if err != nil {
// 		t.Errorf("Error while opening %s (%v)", docFile, err)
// 	}
// 	if md, err := app.ToMan(); err == nil {
// 		f.WriteString(md)
// 	}

// 	// fmt.Println(app.ToMan())
// 	f.Close()
// }
