package main

import (
	"code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	"code.google.com/p/go.exp/go/types"
)

func main() {
	newFluxWindow()
	gui.Run()
}

type fluxWindow struct {
	*gui.Window
	browser *browser
}

func newFluxWindow() *fluxWindow {
	w := &fluxWindow{}
	w.Window = gui.NewWindow(w)
	w.browser = newBrowser(fluxSourceOnly, nil, nil)
	w.AddChild(w.browser)
	w.Resize(w.Size().XY())
	w.browser.accepted = func(obj types.Object) {
		switch obj := obj.(type) {
		case *types.TypeName:
			typ := obj.Type.(*types.NamedType)
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
			if typ.Underlying == nil {
				v.edit(func() {
					if typ.Underlying == nil {
						SliceRemove(&obj.Pkg.Scope.Entries, obj) // this won't remove it from Scope.map if it has one (Scope needs a Remove() method)
					} else {
						saveType(typ)
					}
					reset()
				})
			} else {
				v.done = func() {
					saveType(typ)
					reset()
				}
				v.TakeKeyboardFocus()
			}
		case *types.Func, method:
			n := newFuncNode(obj)
			go n.animate()
			w.browser.Hide()
			w.AddChild(n)
			n.Move(w.Center())
			n.done = func() {
				w.browser.Show()
				w.browser.text.SetText("")
				w.browser.text.TakeKeyboardFocus()
			}
			n.TakeKeyboardFocus()
		}
	}
	w.browser.text.TakeKeyboardFocus()
	return w
}

func (w *fluxWindow) Resize(width, height float64) {
	w.Window.Resize(width, height)
	w.browser.Move(w.Center())
}
