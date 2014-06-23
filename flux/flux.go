// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	"code.google.com/p/gordon-go/refactor"
	"fmt"
	"math"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	go refactor.ReportShadowedPackages()
	if err := Run(newFluxWindow); err != nil {
		fmt.Println(err)
	}
}

type fluxWindow struct {
	*Window
	*Panner
	browser *browser

	target chan Point
	pause  chan bool
}

func newFluxWindow() {
	w := &fluxWindow{}
	NewWindow(w, "Flux", func(win *Window) {
		w.Window = win
		w.Panner = NewPanner(w)
		w.browser = newBrowser(browserOptions{objFilter: isFluxObj, acceptTypes: true, enterTypes: true, mutable: true}, nil)
		w.Add(w.browser)
		w.SetRect(Rect(w))
		w.browser.accepted = func(obj types.Object) {
			switch obj := obj.(type) {
			case *types.TypeName:
				w.SetTitle(obj.Pkg.Path + "." + obj.Name)
				typ := obj.Type.(*types.Named)
				Hide(w.browser)
				v := newTypeView(&typ.UnderlyingT, obj.Pkg)
				w.Add(v)
				MoveCenter(v, Center(w))
				reset := func() {
					w.Remove(v)
					Show(w.browser)
					w.browser.clearText()
					SetKeyFocus(w.browser)
					w.SetTitle("Flux")
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
				prefix := obj.Pkg.Path + "."
				if recv := obj.Type.(*types.Signature).Recv; recv != nil {
					t, _ := indirect(recv.Type)
					prefix += t.(*types.Named).Obj.Name + "."
				}
				w.SetTitle(prefix + obj.Name)
				Hide(w.browser)
				f := loadFunc(obj)
				w.Add(f)
				go animate(f.animate, f.stop)
				f.Move(Center(w))
				f.done = func() {
					Show(w.browser)
					w.browser.clearText()
					SetKeyFocus(w.browser)
					w.SetTitle("Flux")
				}
				SetKeyFocus(f.inputsNode)
			}
		}
		w.browser.canceled = func() {}
		SetKeyFocus(w.browser)

		w.target = make(chan Point)
		w.pause = make(chan bool)
		go w.animate()
	})
}

func panTo(v View, p Point) {
	w := window(v)
	if w == nil {
		return
	}
	p = MapTo(v, p, w)
	go func() { w.target <- p }()
}

func window(v View) *fluxWindow {
	switch v := v.(type) {
	case nil:
		return nil
	case *fluxWindow:
		return v
	}
	return window(Parent(v))
}

func (w *fluxWindow) animate() {
	target := <-w.target
	vel := ZP
	rect := make(chan Rectangle)
	for {
		next := time.After(time.Second / fps)
		Do(w) <- func() {
			Pan(w, Rect(w).Min.Add(vel.Div(fps)))
			rect <- Rect(w)
		}
		r := <-rect
		d := target.Sub(r.Center())
		d.X = math.Copysign(math.Max(0, math.Abs(d.X)-r.Dx()/4), d.X)
		d.Y = math.Copysign(math.Max(0, math.Abs(d.Y)-r.Dy()/4), d.Y)
		vel = vel.Add(d).Mul(.8)
		if vel.Len() < .1 {
			next = nil
		}
		select {
		case <-next:
		case target = <-w.target:
		case <-w.pause:
			target = <-w.target
		}
	}
}

func (w *fluxWindow) SetRect(r Rectangle) {
	w.Window.SetRect(r)
	w.browser.Move(Center(w))
}

func (w *fluxWindow) Scroll(s ScrollEvent) {
	select {
	case w.pause <- true:
	default:
	}
	Pan(w, Rect(w).Min.Sub(s.Delta.Mul(4)))
}
