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
	*Panner
	browser *browser
}

func newFluxWindow() *fluxWindow {
	w := &fluxWindow{}
	w.Window = NewWindow(w)
	w.Panner = NewPanner(w)
	w.browser = newBrowser(fluxSourceOnly, nil, nil)
	w.Add(w.browser)
	w.SetRect(Rect(w))
	w.browser.accepted = func(obj types.Object) {
		switch obj := obj.(type) {
		case *types.TypeName:
			typ := obj.Type.(*types.NamedType)
			Hide(w.browser)
			v := w.browser.typeView
			w.Add(v)
			MoveCenter(v, Center(w))
			reset := func() {
				w.browser.Add(v)
				Show(w.browser)
				w.browser.text.SetText("")
				SetKeyFocus(w.browser.text)
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
				SetKeyFocus(v)
			}
		case *types.Func, method:
			n := newFuncNode(obj)
			go n.animate()
			Hide(w.browser)
			w.Add(n)
			n.Move(Center(w))
			n.done = func() {
				Show(w.browser)
				w.browser.text.SetText("")
				SetKeyFocus(w.browser.text)
			}
			SetKeyFocus(n)
		}
	}
	SetKeyFocus(w.browser.text)
	return w
}

func (w *fluxWindow) SetRect(r Rectangle) {
	w.Window.SetRect(r)
	w.browser.Move(Center(w))
}

func (w *fluxWindow) Scroll(s ScrollEvent) {
	w.SetRect(Rect(w).Sub(s.Delta.Mul(4)))
}
