//
// This file has a weird name to ensure its init function
// will be run the first. It initializes the working directory.
//
package main

import (
	"fmt"
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
