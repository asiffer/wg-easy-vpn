//
// This file has a weird name to ensure its init function
// will be run the first. It initializes the working directory.
//
package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
)

var (
	workdir  string
	testpath string
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	workdir = path.Dir(filename)
	testpath = path.Join(workdir, "test")
	fmt.Println("Working directory:", workdir)
}

func readFileContent(p string) string {
	content, err := ioutil.ReadFile(p)
	if err != nil {
		fmt.Printf("Error while reading %s: %v\n", p, err)
		return ""
	}
	return string(content)
}
