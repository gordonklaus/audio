// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/flux/go/types"
	. "code.google.com/p/gordon-go/flux/gui"
	"fmt"
	"math"
)

type typeView struct {
	*ViewBase
	mode       typeViewMode
	typ        *types.Type
	val        types.Object // non-nil if this is a valueView
	currentPkg *types.Package
	done       func()

	name       *Text // non-nil if this is a valueView
	pkg        *pkgText
	text       *Text
	elems      struct{ left, right []*typeView }
	unexported *Text
	ellipsis   bool
	focused    bool
}

type typeViewMode int

const (
	anyType typeViewMode = iota
	comparableType
	compositeOrPtrType
	compositeType
	makeableType
)

func newTypeView(t *types.Type, currentPkg *types.Package) *typeView {
	v := &typeView{typ: t, currentPkg: currentPkg, mode: anyType}
	v.ViewBase = NewView(v)
	v.text = NewText("")
	v.text.SetTextColor(color(&types.TypeName{}, true, false))
	v.text.SetBackgroundColor(noColor)
	v.Add(v.text)
	v.setType(*t)
	return v
}

func newValueView(val types.Object, currentPkg *types.Package) *typeView {
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
	v := newTypeView(t, currentPkg)
	v.val = val
	v.name = NewText(*name)
	v.name.SetTextColor(color(val, true, false))
	v.name.SetBackgroundColor(noColor)
	v.name.Validate = validateID
	v.name.TextChanged = func(text string) {
		*name = text
		v.reform()
	}
	v.Add(v.name)
	v.reform()
	return v
}

func (v *typeView) setType(t types.Type) {
	*v.typ = t

	v.elems.left = nil
	v.elems.right = nil
	v.unexported = nil
	for NumChildren(v) > 0 {
		v.Remove(Child(v, 0))
	}
	v.Add(v.text)
	if v.name != nil {
		v.Add(v.name)
	}

	s := ""
	switch t := t.(type) {
	case *types.Basic:
		s = t.Name
	case *types.Named:
		s = t.Obj.Name
		if p := t.Obj.Pkg; p != v.currentPkg && p != nil {
			v.pkg = newPkgText()
			v.pkg.setPkg(p)
			v.Add(v.pkg)
		}
	case *types.Pointer:
		s = "*"
		elem := newTypeView(&t.Elem, v.currentPkg)
		if v.mode == compositeOrPtrType {
			elem.mode = compositeType
		}
		v.elems.right = []*typeView{elem}
	case *types.Array:
		s = fmt.Sprintf("[%d]", t.Len)
		v.elems.right = []*typeView{newTypeView(&t.Elem, v.currentPkg)}
	case *types.Slice:
		s = "[]"
		if v.ellipsis {
			s = "…"
		}
		v.elems.right = []*typeView{newTypeView(&t.Elem, v.currentPkg)}
	case *types.Chan:
		s = "chan"
		v.elems.right = []*typeView{newTypeView(&t.Elem, v.currentPkg)}
	case *types.Map:
		s = ":"
		key := newTypeView(&t.Key, v.currentPkg)
		key.mode = comparableType
		v.elems.left = []*typeView{key}
		v.elems.right = []*typeView{newTypeView(&t.Elem, v.currentPkg)}
	case *types.Struct:
		s = "struct"
		for _, f := range t.Fields {
			if invisible(f, v.currentPkg) {
				v.unexported = NewText("contains unexported fields")
				continue
			}
			v.elems.right = append(v.elems.right, newValueView(field{Var: f}, v.currentPkg))
		}
	case *types.Signature:
		s = "func"
		for _, val := range t.Params {
			v.elems.left = append(v.elems.left, newValueView(val, v.currentPkg))
		}
		if t.IsVariadic {
			v.elems.left[len(v.elems.left)-1].setEllipsis()
		}
		for _, val := range t.Results {
			v.elems.right = append(v.elems.right, newValueView(val, v.currentPkg))
		}
	case *types.Interface:
		s = "interface"
		for _, m := range t.Methods {
			if invisible(m, v.currentPkg) {
				v.unexported = NewText("contains unexported methods")
				continue
			}
			v.elems.right = append(v.elems.right, newValueView(m, v.currentPkg))
		}
	}
	v.text.SetText(s)
	if v.unexported != nil {
		// TODO: small font
		v.unexported.SetTextColor(Color{.3, .3, .3, 1})
		v.unexported.SetBackgroundColor(Color{0, 0, 0, .3})
		v.Add(v.unexported)
	}
	for _, c := range append(v.elems.left, v.elems.right...) {
		if v.mode == comparableType {
			c.mode = comparableType
		}
		v.Add(c)
	}

	v.reform()
}

func invisible(obj types.Object, p *types.Package) bool {
	p2 := obj.GetPkg()
	return !(p2 == nil || p == nil || p2 == p || obj.IsExported())
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
	if v.unexported != nil {
		if h2 > 0 {
			h2 += spacing
		}
		h2 += Height(v.unexported)
	}
	x := 0.0
	if v.name != nil {
		v.name.Move(Pt(0, (math.Max(h1, h2)-Height(v.name))/2))
		x += Width(v.name) + spacing
	}
	y := math.Max(0, h2-h1) / 2
	for i := len(v.elems.left) - 1; i >= 0; i-- {
		c := v.elems.left[i]
		c.Move(Pt(x+maxWidth-Width(c), y))
		y += Height(c) + spacing
	}
	x += maxWidth + spacing
	if v.pkg != nil {
		v.pkg.Move(Pt(x, (math.Max(h1, h2)-Height(v.pkg))/2))
		x += Width(v.pkg)
	}
	v.text.Move(Pt(x, (math.Max(h1, h2)-Height(v.text))/2))
	x += Width(v.text) + spacing
	y = math.Max(0, h1-h2) / 2
	if v.unexported != nil {
		v.unexported.Move(Pt(x, y))
		y += Height(v.unexported) + spacing
	}
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
	if v.name == nil {
		v.editType(done)
		return
	}

	done2 := func() {
		done()
		if f, ok := v.val.(field); ok {
			f.Anonymous = len(v.name.Text()) == 0
			if f.Anonymous && *v.typ != nil {
				t, _ := indirect(*v.typ)
				f.Name = t.(*types.Named).Obj.GetName()
			}
		}
	}
	v.name.Accept = func(name string) {
		if v.unique(name) {
			v.editType(done2)
		} else {
			SetKeyFocus(v.name)
		}
	}
	v.name.Reject = done2
	SetKeyFocus(v.name)
}

func (v *typeView) unique(name string) bool {
	if p, ok := Parent(v).(*port); ok {
		ports := append(ins(p.node), outs(p.node)...)
		if n, ok := p.node.(*portsNode); ok {
			f := n.blk.node.(*funcNode)
			ports = append(f.inputsNode.outs, f.outputsNode.ins...)
		}
		for _, p2 := range ports {
			if p2 != p && p2.obj.Name == name {
				return false
			}
		}
		return true
	}
	switch t := (*Parent(v).(*typeView).typ).(type) {
	case *types.Struct:
		if name == "" {
			return true
		}
		for _, f := range t.Fields {
			if f.Name == name && f != v.val.(field).Var {
				return false
			}
		}
		// TODO: if this struct is a Named.UnderlyingT, check for collisions with method names
	case *types.Signature:
		if name == "" {
			return true
		}
		for _, pr := range append(t.Params, t.Results...) {
			if pr != v.val && pr.Name == name {
				return false
			}
		}
	case *types.Interface:
		if name == "" {
			return false
		}
		for _, m := range t.Methods {
			if m.Name == name && m != v.val {
				return false
			}
		}
	}
	return true
}

func (v *typeView) editType(done func()) {
	switch t := (*v.typ).(type) {
	case nil:
		opts := browserOptions{acceptTypes: true}
		opts.objFilter = map[typeViewMode]func(types.Object) bool{
			anyType:            isType,
			comparableType:     isComparableType,
			compositeOrPtrType: isCompositeOrPtrType,
			compositeType:      isCompositeType,
			makeableType:       isMakeableType,
		}[v.mode]
		b := newBrowser(opts, v)
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
	*vs = append((*vs)[:i], append([]*types.Var{types.NewVar(0, v.currentPkg, "", nil)}, (*vs)[i:]...)...)
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
	sig := &types.Signature{Recv: types.NewVar(0, v.currentPkg, "", *v.typ)}
	*m = append((*m)[:i], append([]*types.Func{types.NewFunc(0, v.currentPkg, "", sig)}, (*m)[i:]...)...)
	v.refresh()
	t := (*elems)[i]
	t.edit(func() {
		if *t.typ == nil || len(t.name.Text()) == 0 {
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

func (v *typeView) refresh() { v.setType(*v.typ) }

func (v *typeView) setEllipsis() { v.ellipsis = true; v.refresh() }

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
		if v.name != nil {
			oldName = v.name.Text()
			v.name.SetText("")
		}
		v.setType(nil)
		v.edit(func() {
			if *v.typ == nil {
				v.setType(oldTyp)
				if v.name != nil {
					v.name.SetText(oldName)
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
	default:
		v.ViewBase.KeyPress(event)
	}
}

func (v *typeView) Paint() {
	if v.focused {
		SetColor(Color{.25, .25, .25, 1})
		FillRect(Rect(v))
	}
	if _, ok := Parent(v).(*typeView); ok {
		SetColor(lineColor)
		SetLineWidth(1)
		DrawRect(Rect(v))
	}
	if t, ok := (*v.typ).(*types.Named); ok && unknown(t.Obj) {
		SetColor(Color{1, 0, 0, 1})
		SetLineWidth(3)
		r := Rect(v)
		DrawLine(r.Min, r.Max)
		DrawLine(Pt(r.Min.X, r.Max.Y), Pt(r.Max.X, r.Min.Y))
	}
}

func underlying(t types.Type) types.Type {
	if nt, ok := t.(*types.Named); ok {
		return nt.UnderlyingT
	}
	return t
}

func indirect(t types.Type) (types.Type, bool) {
	if p, ok := t.(*types.Pointer); ok {
		return p.Elem, true
	}
	return t, false
}
