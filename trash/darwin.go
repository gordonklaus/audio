package trash

import (
	"os"
	"path/filepath"
)

func getDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".Trash")
}
