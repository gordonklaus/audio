// +build !darwin

package gui

func commandKey(k KeyEvent) bool {
	return k.Ctrl
}
