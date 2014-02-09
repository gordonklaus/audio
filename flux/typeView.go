// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	"fmt"
	"math"
)

type typeView struct {
	*ViewBase
	typ     *types.Type
	mode    browserMode
	text    Text
	elems   struct{ left, right []*typeView }
	focused bool
	done    func()

	// typeView is also used as a valueView, in which case this is non-nil
	nameText *TextBase
}

func newTypeView(t *types.Type) *typeView {
	v := &typeView{typ: t, mode: typesOnly}
	v.ViewBase = NewView(v)
	v.text = NewText("")
	v.text.SetTextColor(color(&types.TypeName{}, false, false))
	v.text.SetBackgroundColor(Color{0, 0, 0, .3})
	v.Add(v.text)
	v.setType(*t)
	return v
}

func newValueView(val types.Object) *typeView {
	var t *types.Type
	name := new(string)
	switch val := val.(type) {
	case *types.Var:
		t, name = &val.Type, &val.Name
	case *types.Func:
		if isMethod(val) {
			m := types.Type(val.Type)
			t, name = &m, &val.Name
		}
	case field:
		t = &val.Type
		if !val.Anonymous {
			name = &val.Name
		}
	}
	v := newTypeView(t)
	v.nameText = NewText(*name)
	v.nameText.SetTextColor(color(val, false, false))
	v.nameText.SetBackgroundColor(Color{0, 0, 0, .3})
	v.nameText.TextChanged = func(text string) {
		*name = text
		v.reform()
	}
	v.Add(v.nameText)
	v.reform()
	return v
}

func (v *typeView) setType(t types.Type) {
	*v.typ = t

	v.elems.left = nil
	v.elems.right = nil
	for NumChildren(v) > 0 {
		v.Remove(Child(v, 0))
	}
	v.Add(v.text)
	if v.nameText != nil {
		v.Add(v.nameText)
	}

	s := ""
	switch t := t.(type) {
	case generic:
		s = "<T>"
	case *types.Basic:
		s = t.Name
	case *types.Named:
		s = t.Obj.Name
	case *types.Pointer:
		s = "*"
		v.elems.right = []*typeView{newTypeView(&t.Elem)}
	case *types.Array:
		s = fmt.Sprintf("[%d]", t.Len)
		v.elems.right = []*typeView{newTypeView(&t.Elem)}
	case *types.Slice:
		s = "[]"
		v.elems.right = []*typeView{newTypeView(&t.Elem)}
	case *types.Chan:
		s = "chan"
		v.elems.right = []*typeView{newTypeView(&t.Elem)}
	case *types.Map:
		s = ":"
		v.elems.left = []*typeView{newTypeView(&t.Key)}
		v.elems.right = []*typeView{newTypeView(&t.Elem)}
	case *types.Struct:
		s = "struct"
		for _, f := range t.Fields {
			v.elems.right = append(v.elems.right, newValueView(field{f, nil}))
		}
	case *types.Signature:
		s = "func"
		for _, val := range t.Params {
			v.elems.left = append(v.elems.left, newValueView(val))
		}
		if t.IsVariadic {
			v.elems.left[len(v.elems.left)-1].setVariadic()
		}
		for _, val := range t.Results {
			v.elems.right = append(v.elems.right, newValueView(val))
		}
	case *types.Interface:
		s = "interface"
		for _, m := range t.Methods {
			v.elems.right = append(v.elems.right, newValueView(m))
		}
	}
	v.text.SetText(s)
	for _, c := range append(v.elems.left, v.elems.right...) {
		v.Add(c)
	}

	if _, ok := t.(*types.Pointer); ok && v.mode == compositeOrPtrType {
		v.elems.right[0].mode = compositeType
	}

	v.reform()
}

func (v *typeView) setVariadic() {
	v.text.SetText("…")
	v.reform()
}

func (v *typeView) reform() {
	const spacing = 2
	maxWidth := float64(0)
	h1 := float64(0)
	for i, c := range v.elems.left {
		h1 += Height(c)
		if i > 0 {
			h1 += spacing
		}
		if w := Width(c); w > maxWidth {
			maxWidth = w
		}
	}
	h2 := float64(0)
	for i, c := range v.elems.right {
		h2 += Height(c)
		if i > 0 {
			h2 += spacing
		}
	}
	x := 0.0
	if v.nameText != nil {
		v.nameText.Move(Pt(0, (math.Max(h1, h2)-Height(v.nameText))/2))
		x += Width(v.nameText) + spacing
	}
	y := math.Max(0, h2-h1) / 2
	for i := len(v.elems.left) - 1; i >= 0; i-- {
		c := v.elems.left[i]
		c.Move(Pt(x+maxWidth-Width(c), y))
		y += Height(c) + spacing
	}
	x += maxWidth + spacing
	v.text.Move(Pt(x, (math.Max(h1, h2)-Height(v.text))/2))
	x += Width(v.text) + spacing
	y = math.Max(0, h1-h2) / 2
	for i := len(v.elems.right) - 1; i >= 0; i-- {
		c := v.elems.right[i]
		c.Move(Pt(x, y))
		y += Height(c) + spacing
	}

	ResizeToFit(v, 2)
	if p, ok := Parent(v).(*typeView); ok {
		p.reform()
	}
}

func (v *typeView) edit(done func()) {
	if v.nameText != nil {
		v.nameText.Accept = func(string) { v.editType(done) }
		v.nameText.Reject = done
		SetKeyFocus(v.nameText)
		return
	}
	v.editType(done)
}
func (v *typeView) editType(done func()) {
	switch t := (*v.typ).(type) {
	case nil:
		b := newBrowser(v.mode, v)
		v.Add(b)
		b.Move(Center(v))
		b.accepted = func(obj types.Object) {
			b.Close()
			n := obj.(*types.TypeName)
			if n.Type != nil {
				v.setType(n.Type)
			} else {
				v.setType(newProtoType(n))
			}
			v.editType(done)
		}
		b.canceled = func() {
			b.Close()
			done()
		}
		SetKeyFocus(b)
	case *types.Basic, *types.Named:
		done()
	case *types.Pointer, *types.Array, *types.Slice, *types.Chan:
		if elt := v.elems.right[0]; *elt.typ == nil {
			elt.editType(func() {
				if *elt.typ == nil {
					v.setType(nil)
				}
				v.editType(done)
			})
		} else {
			done()
		}
	case *types.Map:
		key := v.elems.left[0]
		val := v.elems.right[0]
		switch types.Type(nil) {
		case *key.typ:
			key.editType(func() {
				if *key.typ == nil {
					v.setType(nil)
				}
				v.editType(done)
			})
		case *val.typ:
			val.editType(func() {
				if *val.typ == nil {
					key.setType(nil)
				}
				v.editType(done)
			})
		default:
			done()
		}
	case *types.Struct:
		v.addVars(&t.Fields, &v.elems.right, done)
	case *types.Signature:
		v.addVars(&t.Params, &v.elems.left, func() {
			v.addVars(&t.Results, &v.elems.right, done)
		})
	case *types.Interface:
		v.addMethods(&t.Methods, &v.elems.right, done)
	}
}

func (v *typeView) insertVar(vs *[]*types.Var, elems *[]*typeView, before bool, i int, success, fail func()) {
	if !before {
		i++
	}
	*vs = append((*vs)[:i], append([]*types.Var{{}}, (*vs)[i:]...)...)
	v.refresh()
	t := (*elems)[i]
	t.edit(func() {
		if *t.typ == nil {
			*vs = append((*vs)[:i], (*vs)[i+1:]...)
			v.refresh()
			if fail != nil {
				fail()
			} else {
				if !before {
					i--
				}
				SetKeyFocus((*elems)[i])
			}
		} else {
			if success != nil {
				success()
			} else {
				SetKeyFocus(t)
			}
		}
	})
}

func (v *typeView) addVars(vs *[]*types.Var, elems *[]*typeView, done func()) {
	v.insertVar(vs, elems, true, len(*vs), func() {
		v.addVars(vs, elems, done)
	}, done)
}

func (v *typeView) insertMethod(m *[]*types.Func, elems *[]*typeView, before bool, i int, success, fail func()) {
	if !before {
		i++
	}
	*m = append((*m)[:i], append([]*types.Func{types.NewFunc(0, nil, "", &types.Signature{Recv: newVar("", *v.typ)})}, (*m)[i:]...)...)
	v.refresh()
	t := (*elems)[i]
	t.edit(func() {
		if *t.typ == nil || len(t.nameText.GetText()) == 0 {
			*m = append((*m)[:i], (*m)[i+1:]...)
			v.refresh()
			if fail != nil {
				fail()
			} else {
				if !before {
					i--
				}
				SetKeyFocus((*elems)[i])
			}
		} else {
			if success != nil {
				success()
			} else {
				SetKeyFocus(t)
			}
		}
	})
}

func (v *typeView) addMethods(m *[]*types.Func, elems *[]*typeView, done func()) {
	v.insertMethod(m, elems, true, len(*m), func() {
		v.addMethods(m, elems, done)
	}, done)
}

func (v *typeView) refresh() {
	v.setType(*v.typ)
}

func (v *typeView) TookKeyFocus() { v.focused = true; Repaint(v) }
func (v *typeView) LostKeyFocus() { v.focused = false; Repaint(v) }

func (v *typeView) KeyPress(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		foc := KeyFocus(v)
		if foc == v {
			v.ViewBase.KeyPress(event)
			return
		}
		moveFocus := func(these, others []*typeView, dir int) bool {
			for i, c := range these {
				if c == foc {
					switch event.Key {
					case dir:
						min := math.MaxFloat64
						nearest := (*typeView)(nil)
						cy := CenterInParent(c).Y
						for _, c2 := range others {
							d := math.Abs(cy - CenterInParent(c2).Y)
							if d < min {
								min, nearest = d, c2
							}
						}
						if nearest != nil {
							SetKeyFocus(nearest)
						}
					case KeyUp:
						if i > 0 {
							SetKeyFocus(these[i-1])
						}
					case KeyDown:
						if i < len(these)-1 {
							SetKeyFocus(these[i+1])
						}
					}
					return true
				}
			}
			return false
		}
		_ = moveFocus(v.elems.left, v.elems.right, KeyRight) || moveFocus(v.elems.right, v.elems.left, KeyLeft)
	case KeyEnter:
		done := func() { SetKeyFocus(v) }
		switch t := (*v.typ).(type) {
		case *types.Pointer, *types.Array, *types.Slice, *types.Chan:
			SetKeyFocus(v.elems.right[0])
		case *types.Map:
			SetKeyFocus(v.elems.left[0])
		case *types.Struct:
			if len(t.Fields) == 0 {
				v.edit(done)
			} else {
				SetKeyFocus(v.elems.right[0])
			}
		case *types.Signature:
			switch {
			case len(t.Params) == 0 && len(t.Results) == 0:
				v.edit(done)
			case len(t.Params) == 0:
				v.addVars(&t.Params, &v.elems.left, func() {
					if len(t.Params) == 0 {
						SetKeyFocus(v.elems.right[0])
					} else {
						SetKeyFocus(v.elems.left[0])
					}
				})
			case len(t.Results) == 0:
				v.addVars(&t.Results, &v.elems.right, func() {
					if len(t.Results) == 0 {
						SetKeyFocus(v.elems.left[0])
					} else {
						SetKeyFocus(v.elems.right[0])
					}
				})
			default:
				SetKeyFocus(v.elems.left[0])
			}
		case *types.Interface:
			if len(t.Methods) == 0 {
				v.edit(done)
			} else {
				SetKeyFocus(v.elems.right[0])
			}
		}
	case KeyEscape:
		if v.done != nil {
			v.done()
		} else {
			SetKeyFocus(Parent(v))
		}
	case KeyBackspace:
		if p, ok := Parent(v).(*typeView); ok {
			if _, ok := (*p.typ).(*types.Interface); ok {
				break
			}
		}
		oldTyp, oldName := *v.typ, ""
		if v.nameText != nil {
			oldName = v.nameText.GetText()
			v.nameText.SetText("")
		}
		v.setType(nil)
		v.edit(func() {
			if *v.typ == nil {
				v.setType(oldTyp)
				if v.nameText != nil && len(v.nameText.GetText()) == 0 {
					v.nameText.SetText(oldName)
				}
			}
			SetKeyFocus(v)
		})
	case KeyComma:
		if p, ok := Parent(v).(*typeView); ok {
			switch t := (*p.typ).(type) {
			case *types.Struct:
				for i, c := range p.elems.right {
					if c == v {
						p.insertVar(&t.Fields, &p.elems.right, event.Shift, i, nil, nil)
						break
					}
				}
			case *types.Signature:
				for i, c := range p.elems.left {
					if c == v {
						p.insertVar(&t.Params, &p.elems.left, event.Shift, i, nil, nil)
						break
					}
				}
				for i, c := range p.elems.right {
					if c == v {
						p.insertVar(&t.Results, &p.elems.right, event.Shift, i, nil, nil)
						break
					}
				}
			case *types.Interface:
				for i, c := range p.elems.right {
					if c == v {
						p.insertMethod(&t.Methods, &p.elems.right, event.Shift, i, nil, nil)
						break
					}
				}
			}
		}
	case KeyDelete:
		if p, ok := Parent(v).(*typeView); ok {
			switch t := (*p.typ).(type) {
			case *types.Struct:
				for i, c := range p.elems.right {
					if c == v {
						t.Fields = append(t.Fields[:i], t.Fields[i+1:]...)
						p.refresh()
						if len := len(t.Fields); len > 0 {
							if i == len {
								i--
							}
							SetKeyFocus(p.elems.right[i])
						} else {
							SetKeyFocus(p)
						}
						break
					}
				}
			case *types.Signature:
				for i, c := range p.elems.left {
					if c == v {
						t.Params = append(t.Params[:i], t.Params[i+1:]...)
						p.refresh()
						if len := len(t.Params); len > 0 {
							if i == len {
								i--
							}
							SetKeyFocus(p.elems.left[i])
						} else {
							SetKeyFocus(p)
						}
						break
					}
				}
				for i, c := range p.elems.right {
					if c == v {
						t.Results = append(t.Results[:i], t.Results[i+1:]...)
						p.refresh()
						if len := len(t.Results); len > 0 {
							if i == len {
								i--
							}
							SetKeyFocus(p.elems.right[i])
						} else {
							SetKeyFocus(p)
						}
						break
					}
				}
			case *types.Interface:
				for i, c := range p.elems.right {
					if c == v {
						t.Methods = append(t.Methods[:i], t.Methods[i+1:]...)
						p.refresh()
						if len := len(t.Methods); len > 0 {
							if i == len {
								i--
							}
							SetKeyFocus(p.elems.right[i])
						} else {
							SetKeyFocus(p)
						}
						break
					}
				}
			}
		}
	}
}

func (v *typeView) Paint() {
	SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[v.focused])
	SetLineWidth(1)
	DrawRect(Rect(v))
}

type generic struct{ types.Type }

func underlying(t types.Type) types.Type {
	if nt, ok := t.(*types.Named); ok {
		return nt.UnderlyingT
	}
	return t
}

func indirect(t types.Type) (types.Type, bool) {
	if p, ok := underlying(t).(*types.Pointer); ok {
		return p.Elem, true
	}
	return t, false
}
