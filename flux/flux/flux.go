package main

import (
	"code.google.com/p/gordon-go/gui"
)

func main() {
	w := gui.NewWindow(nil, NewFunction())
	w.HandleEvents()
	w.Close()
}
