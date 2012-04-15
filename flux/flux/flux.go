package main

import (
	"code.google.com/p/gordon-go/gui"
	"code.google.com/p/gordon-go/flux"
)

func main() {
	w := gui.NewWindow(nil, flux.NewCompound())
	w.HandleEvents()
	w.Close()
}
