// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
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
	w.browser = newBrowser(fluxSourceOnly, nil)
	w.Add(w.browser)
	w.SetRect(Rect(w))
	w.browser.accepted = func(obj types.Object) {
		switch obj := obj.(type) {
		case *types.TypeName:
			// TODO: move type editing into browser (just like with localVar and when creating types for make, convert, etc)
			typ := obj.Type.(*types.Named)
			Hide(w.browser)
			v := newTypeView(&typ.UnderlyingT)
			w.Add(v)
			MoveCenter(v, Center(w))
			reset := func() {
				w.Remove(v)
				Show(w.browser)
				w.browser.text.SetText("")
				SetKeyFocus(w.browser.text)
			}
			if typ.UnderlyingT == nil {
				v.edit(func() {
					if typ.UnderlyingT == nil {
						delete(obj.Pkg.Scope().Objects, obj.Name)
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
		case *types.Func:
			Hide(w.browser)
			f := loadFunc(obj)
			w.Add(f)
			go animate(f.animate, f.stop)
			f.Move(Center(w))
			f.done = func() {
				Show(w.browser)
				w.browser.text.SetText("")
				SetKeyFocus(w.browser.text)
			}
			SetKeyFocus(f.inputsNode)
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
