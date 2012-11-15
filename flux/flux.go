package main

import (
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
)

type FluxWindow struct {
	*Window
	browser *Browser
}

func NewFluxWindow() *FluxWindow {
	w := &FluxWindow{}
	w.Window = NewWindow(w)
	w.browser = NewBrowser(fluxSourceOnly, nil, nil)
	w.SetCentralView(w.browser)
	w.browser.accepted.Connect(func(info ...interface{}) {
		switch info := info[0].(type) {
		case *NamedType:
			vv := NewView(nil)
			w.SetCentralView(vv)
			v := w.browser.typeView
			vv.AddChild(v)
			v.MoveCenter(vv.Center())
			reset := func() {
				w.browser.AddChild(v)
				w.SetCentralView(w.browser)
				w.browser.text.SetText("")
				w.browser.text.TakeKeyboardFocus()
			}
			if info.underlying == nil {
				v.edit(func() {
					if info.underlying == nil {
						SliceRemove(&info.parent.(*PackageInfo).types, info)
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
		case *FuncInfo:
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
