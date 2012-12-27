package main

import (
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
)

type fluxWindow struct {
	*Window
	browser *browser
}

func newFluxWindow() *fluxWindow {
	w := &fluxWindow{}
	w.Window = NewWindow(w)
	w.browser = newBrowser(fluxSourceOnly, nil, nil)
	w.AddChild(w.browser)
	w.browser.accepted.Connect(func(info ...interface{}) {
		switch info := info[0].(type) {
		case *NamedType:
			w.browser.Hide()
			v := w.browser.typeView
			w.AddChild(v)
			v.MoveCenter(w.Center())
			reset := func() {
				w.browser.AddChild(v)
				w.browser.Show()
				w.browser.text.SetText("")
				w.browser.text.TakeKeyboardFocus()
			}
			if info.underlying == nil {
				v.edit(func() {
					if info.underlying == nil {
						SliceRemove(&info.parent.(*Package).types, info)
					} else {
						saveType(info)
					}
					reset()
				})
			} else {
				v.done = func() {
					saveType(info)
					reset()
				}
				v.TakeKeyboardFocus()
			}
		case *Func:
			n := newFuncNode(info)
			w.browser.Hide()
			w.AddChild(n)
			n.Move(w.Center())
			n.TakeKeyboardFocus()
		}
	})
	w.browser.text.TakeKeyboardFocus()
	return w
}

func (w *fluxWindow) Resize(width, height float64) {
	w.Window.Resize(width, height)
	w.browser.Move(w.Center())
}

func main() {
	w := newFluxWindow()
	w.HandleEvents()
	w.Close()
}
