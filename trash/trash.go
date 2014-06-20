package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

var dir string

func init() {
	dir = getDir()
	if dir == "" {
		fmt.Println("WARNING:  Trash directory not found or not supported.  Removed files will be permanently deleted.")
	}
}

func Trash(path string) error {
	if dir == "" {
		return os.Remove(path)
	}

	name := filepath.Base(path)
	for i := 0; ; i++ {
		dst := filepath.Join(dir, name)
		if i > 0 {
			dst += "." + strconv.Itoa(i)
		}
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			return os.Rename(path, dst)
		}
	}
}
