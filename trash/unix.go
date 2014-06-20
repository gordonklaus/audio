// +build dragonfly freebsd linux nacl netbsd openbsd solaris

package trash

import (
	"os"
	"path/filepath"
)

func getDir() string {
	x := os.Getenv("XDG_DATA_HOME")
	if x == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return ""
		}
		x = filepath.Join(home, ".local.share")
	}
	return filepath.Join(x, ".Trash/files")
}
