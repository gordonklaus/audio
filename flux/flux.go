package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
	. "code.google.com/p/gordon-go/util"
)

func main() {
	newFluxWindow()
	Run()
}

type fluxWindow struct {
	*Window
	browser *browser
}

func newFluxWindow() *fluxWindow {
	w := &fluxWindow{}
	w.Window = NewWindow(w)
	w.browser = newBrowser(fluxSourceOnly, nil, nil)
	w.AddChild(w.browser)
	w.SetRect(w.Rect())
	w.browser.accepted = func(obj types.Object) {
		switch obj := obj.(type) {
		case *types.TypeName:
			typ := obj.Type.(*types.NamedType)
			w.browser.Hide()
			v := w.browser.typeView
			w.AddChild(v)
			MoveCenter(v, Center(w))
			reset := func() {
				w.browser.AddChild(v)
				w.browser.Show()
				w.browser.text.SetText("")
				SetKeyboardFocus(w.browser.text)
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
				SetKeyboardFocus(v)
			}
		case *types.Func, method:
			n := newFuncNode(obj)
			go n.animate()
			w.browser.Hide()
			w.AddChild(n)
			n.Move(Center(w))
			n.done = func() {
				w.browser.Show()
				w.browser.text.SetText("")
				SetKeyboardFocus(w.browser.text)
			}
			SetKeyboardFocus(n)
		}
	}
	SetKeyboardFocus(w.browser.text)
	return w
}

func (w *fluxWindow) SetRect(r Rectangle) {
	w.Window.SetRect(r)
	w.browser.Move(Center(w))
}
