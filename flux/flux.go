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
	w.nodeCreator = NewNodeCreator(true)
	w.SetCentralView(w.nodeCreator)
	w.nodeCreator.created.Connect(func(info ...interface{}) {
		switch info := info[0].(type) {
		case *FunctionInfo:
			w.SetCentralView(NewFunction(info))
		default:
			w.SetCentralView(w.nodeCreator)
			w.nodeCreator.text.TakeKeyboardFocus()
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
