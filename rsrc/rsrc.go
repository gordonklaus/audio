package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	path := os.Args[1]
	varName := os.Args[2]
	b, err := ioutil.ReadFile(path)
	chk(err)

	s := fmt.Sprintf(`//generated from %s
package rsrc
var %s = []byte(%s)`, filepath.Base(path), varName, strconv.Quote(string(b)))

	chk(ioutil.WriteFile(path+".go", []byte(s), 0666))
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
