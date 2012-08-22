package main

import (
	."code.google.com/p/gordon-go/gui"
)

type FluxWindow struct {
	*Window
	browser *Browser
}

func NewFluxWindow() *FluxWindow {
	w := &FluxWindow{}
	w.Window = NewWindow(w)
	w.browser = NewBrowser(fluxSourceOnly)
	w.SetCentralView(w.browser)
	w.browser.created.Connect(func(info ...interface{}) {
		switch info := info[0].(type) {
		case *FunctionInfo:
			w.SetCentralView(NewFunction(info))
		default:
			w.SetCentralView(w.browser)
			w.browser.text.TakeKeyboardFocus()
		}
	})
	w.browser.canceled.Connect(func(...interface{}) {
		w.SetCentralView(w.browser)
		w.browser.text.TakeKeyboardFocus()
	})
	w.browser.text.TakeKeyboardFocus()
	return w
}

func (w *FluxWindow) Resize(width, height float64) {
	w.Window.Resize(width, height)
	w.browser.Move(w.Center())
}

func main() {
	w := NewFluxWindow()
	w.HandleEvents()
	w.Close()
}
