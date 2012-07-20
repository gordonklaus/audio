package main

import (
	."code.google.com/p/gordon-go/gui"
)

type FluxWindow struct {
	*Window
	nodeCreator *NodeCreator
}

func NewFluxWindow() *FluxWindow {
	w := &FluxWindow{}
	w.Window = NewWindow(w)
	w.nodeCreator = NewNodeCreator()
	w.SetCentralView(w.nodeCreator)
	w.nodeCreator.created.Connect(func(n ...interface{}) {
		if _, ok := n[0].(*FunctionInfo); ok {
			w.SetCentralView(NewFunction())
		}
	})
	w.nodeCreator.canceled.Connect(func(...interface{}) {
		w.SetCentralView(w.nodeCreator)
		w.nodeCreator.text.TakeKeyboardFocus()
	})
	w.nodeCreator.text.TakeKeyboardFocus()
	return w
}

func (w *FluxWindow) Resize(width, height float64) {
	w.Window.Resize(width, height)
	w.nodeCreator.Move(w.Center())
}

func main() {
	w := NewFluxWindow()
	w.HandleEvents()
	w.Close()
}
