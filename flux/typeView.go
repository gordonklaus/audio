package main

import (
	."code.google.com/p/gordon-go/gui"
	"code.google.com/p/go.exp/go/types"
	"fmt"
	"math"
)

type typeView struct {
	*ViewBase
	typ *types.Type
	text Text
	childTypes struct{left, right []*typeView}
	focused bool
	done func()
	
	// typeView is also used as a valueView, in which case this is non-nil
	nameText *TextBase
}

func newTypeView(t *types.Type) *typeView {
	v := &typeView{}
	v.ViewBase = NewView(v)
	v.typ = t
	v.text = NewText("")
	v.text.SetTextColor(getTextColor(&types.TypeName{}, .7))
	v.text.SetBackgroundColor(Color{0, 0, 0, .3})
	v.AddChild(v.text)
	v.setType(*t)
	return v
}

func newValueView(val types.Object) *typeView {
	var t *types.Type
	var name *string
	switch val := val.(type) {
	case *types.Var: t, name = &val.Type, &val.Name
	case method:     m := types.Type(val.Type)
	                 t, name = &m, &val.Name
	case field:      t, name = &val.Type, &val.Name
	}
	v := newTypeView(t)
	v.nameText = NewText(val.GetName())
	v.nameText.SetTextColor(getTextColor(val, .7))
	v.nameText.SetBackgroundColor(Color{0, 0, 0, .3})
	v.nameText.TextChanged.Connect(func(...interface{}) {
		*name = v.nameText.GetText()
		v.reform()
	})
	v.AddChild(v.nameText)
	v.reform()
	return v
}

func (v *typeView) setType(t types.Type) {
	*v.typ = t
	s := ""
	v.childTypes.left, v.childTypes.right = nil, nil
	switch t := t.(type) {
	case *types.Basic:
		s = t.Name
	case *types.NamedType:
		s = t.Obj.Name
	case *types.Pointer:
		s = "*"
		v.childTypes.right = []*typeView{newTypeView(&t.Base)}
	case *types.Array:
		s = fmt.Sprintf("[%d]", t.Len)
		v.childTypes.right = []*typeView{newTypeView(&t.Elt)}
	case *types.Slice:
		s = "[]"
		v.childTypes.right = []*typeView{newTypeView(&t.Elt)}
	case *types.Chan:
		s = "chan"
		v.childTypes.right = []*typeView{newTypeView(&t.Elt)}
	case *types.Map:
		s = ":"
		v.childTypes.left = []*typeView{newTypeView(&t.Key)}
		v.childTypes.right = []*typeView{newTypeView(&t.Elt)}
	case *types.Struct:
		s = "struct"
		for _, f := range t.Fields { v.childTypes.right = append(v.childTypes.right, newValueView(field{nil, f})) }
	case *types.Signature:
		s = "func"
		for _, val := range t.Params { v.childTypes.left = append(v.childTypes.left, newValueView(val)) }
		for _, val := range t.Results { v.childTypes.right = append(v.childTypes.right, newValueView(val)) }
	case *types.Interface:
		s = "interface"
		for _, m := range t.Methods { v.childTypes.right = append(v.childTypes.right, newValueView(method{nil, m})) }
	}
	v.text.SetText(s)
	
	for len(v.Children()) > 0 { v.RemoveChild(v.Children()[0]) }
	v.AddChild(v.text)
	for _, c := range append(v.childTypes.left, v.childTypes.right...) { v.AddChild(c) }
	if v.nameText != nil { v.AddChild(v.nameText) }
	
	v.reform()
}

func (v *typeView) reform() {	
	const spacing = 2
	maxWidth := float64(0)
	h1 := float64(0); for i, c := range v.childTypes.left { h1 += c.Height(); if i > 0 { h1 += spacing }; if w := c.Width(); w > maxWidth { maxWidth = w } }
	h2 := float64(0); for i, c := range v.childTypes.right { h2 += c.Height(); if i > 0 { h2 += spacing } }
	x := 0.0
	if v.nameText != nil {
		v.nameText.Move(Pt(0, (math.Max(h1, h2) - v.nameText.Height()) / 2))
		x += v.nameText.Width() + spacing
	}
	y := math.Max(0, h2 - h1) / 2
	for i := len(v.childTypes.left) - 1; i >= 0; i-- {
		c := v.childTypes.left[i]
		c.Move(Pt(x + maxWidth - c.Width(), y))
		y += c.Height() + spacing
	}
	x += maxWidth + spacing
	v.text.Move(Pt(x, (math.Max(h1, h2) - v.text.Height()) / 2))
	x += v.text.Width() + spacing
	y = math.Max(0, h1 - h2) / 2
	for i := len(v.childTypes.right) - 1; i >= 0; i-- {
		c := v.childTypes.right[i]
		c.Move(Pt(x, y))
		y += c.Height() + spacing
	}
	
	ResizeToFit(v, 2)
	if p, ok := v.Parent().(*typeView); ok {
		p.reform()
	}
}

func (v *typeView) edit(done func()) {
	if v.nameText != nil {
		v.nameText.Accept.ConnectSingleShot(func(...interface{}) { v.editType(done) })
		v.nameText.Reject.ConnectSingleShot(func(...interface{}) { done() })
		v.nameText.TakeKeyboardFocus()
		return
	}
	v.editType(done)
}
func (v *typeView) editType(done func()) {
	if *v.typ == nil {
		var pkg *types.Package
		var imports []*types.Package
l:		for v := View(v); v != nil; v = v.Parent() {
			switch v := v.(type) {
			case node:
				f := v.block().func_()
				pkg, imports = f.pkg(), f.imports()
				break l
			}
		}
		b := newBrowser(typesOnly, pkg, imports)
		v.AddChild(b)
		b.Move(v.Center())
		b.accepted.Connect(func(obj ...interface{}) {
			b.Close()
			n := obj[0].(*types.TypeName)
			if n.Type != nil {
				v.setType(n.Type)
			} else {
				v.setType(newProtoType(n))
			}
			v.editType(done)
		})
		b.canceled.Connect(func(...interface{}) {
			b.Close()
			done()
		})
		b.text.TakeKeyboardFocus()
		return
	}
	
	switch t := (*v.typ).(type) {
	case *types.Pointer, *types.Array, *types.Slice, *types.Chan:
		if elt := v.childTypes.right[0]; *elt.typ == nil {
			elt.edit(func() {
				if *elt.typ == nil { v.setType(nil) }
				v.editType(done)
			})
			return
		}
	case *types.Map:
		key := v.childTypes.left[0]
		val := v.childTypes.right[0]
		switch types.Type(nil) {
		case *key.typ:
			key.edit(func() {
				if *key.typ == nil { v.setType(nil) }
				v.editType(done)
			})
			return
		case *val.typ:
			val.edit(func() {
				if *val.typ == nil { key.setType(nil) }
				v.editType(done)
			})
			return
		}
	// next three cases assume the type is brand-new, and thus empty (otherwise we need special handling of protoTypes above)
	case *types.Struct:
		v.addFields(&t.Fields, &v.childTypes.right, done)
		return
	case *types.Signature:
		v.addVars(&t.Params, &v.childTypes.left, func() {
			v.addVars(&t.Results, &v.childTypes.right, done)
		})
		return
	case *types.Interface:
		v.addMethods(&t.Methods, &v.childTypes.right, done)
		return
	}
	
	done()
}

func (v *typeView) addFields(f *[]*types.Field, childTypes *[]*typeView, done func()) {
	*f = append(*f, &types.Field{})
	v.setType(*v.typ)
	i := len(*f) - 1
	t := (*childTypes)[i]
	t.edit(func() {
		if *t.typ == nil {
			*f = (*f)[:i]
			v.setType(*v.typ)
			done()
		} else {
			v.addFields(f, childTypes, done)
		}
	})
}

func (v *typeView) addVars(vs *[]*types.Var, childTypes *[]*typeView, done func()) {
	*vs = append(*vs, &types.Var{})
	v.setType(*v.typ)
	i := len(*vs) - 1
	t := (*childTypes)[i]
	t.edit(func() {
		if *t.typ == nil {
			*vs = (*vs)[:i]
			v.setType(*v.typ)
			done()
		} else {
			v.addVars(vs, childTypes, done)
		}
	})
}

func (v *typeView) addMethods(m *[]*types.Method, childTypes *[]*typeView, done func()) {
	*m = append(*m, &types.Method{Type:&types.Signature{}})
	v.setType(*v.typ)
	i := len(*m) - 1
	t := (*childTypes)[i]
	t.edit(func() {
		if *t.typ == nil || len(t.nameText.GetText()) == 0 {
			*m = (*m)[:i]
			v.setType(*v.typ)
			done()
		} else {
			v.addMethods(m, childTypes, done)
		}
	})
}

func (v *typeView) focusNearest(child *typeView, dirKey int) {
	var views []View
	for _, v := range append(v.childTypes.left, v.childTypes.right...) {
		views = append(views, v)
	}
	nearest := nearestView(v, views, child.MapTo(child.Center(), v), dirKey)
	if nearest != nil {
		nearest.TakeKeyboardFocus()
	}
}

func (v *typeView) TookKeyboardFocus() { v.focused = true; v.Repaint() }
func (v *typeView) LostKeyboardFocus() { v.focused = false; v.Repaint() }

func (v *typeView) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		if p, ok := v.Parent().(*typeView); ok {
			p.focusNearest(v, event.Key)
		}
	case KeyEnter:
		done := func() { v.TakeKeyboardFocus() }
		switch t := (*v.typ).(type) {
		case *types.Pointer, *types.Array, *types.Slice, *types.Chan:
			v.childTypes.right[0].TakeKeyboardFocus()
		case *types.Map:
			v.childTypes.left[0].TakeKeyboardFocus()
		case *types.Struct:
			if len(t.Fields) == 0 {
				v.edit(done)
			} else {
				v.childTypes.right[0].TakeKeyboardFocus()
			}
		case *types.Signature:
			switch {
			case len(t.Params) == 0 && len(t.Results) == 0:
				v.edit(done)
			case len(t.Params) == 0:
				v.addVars(&t.Params, &v.childTypes.left, func() {
					if len(t.Params) == 0 {
						v.childTypes.right[0].TakeKeyboardFocus()
					} else {
						v.childTypes.left[0].TakeKeyboardFocus()
					}
				})
			case len(t.Results) == 0:
				v.addVars(&t.Results, &v.childTypes.right, func() {
					if len(t.Results) == 0 {
						v.childTypes.left[0].TakeKeyboardFocus()
					} else {
						v.childTypes.right[0].TakeKeyboardFocus()
					}
				})
			default:
				v.childTypes.left[0].TakeKeyboardFocus()
			}
		case *types.Interface:
			if len(t.Methods) == 0 {
				v.edit(done)
			} else {
				v.childTypes.right[0].TakeKeyboardFocus()
			}
		}
	case KeyEscape:
		if v.done != nil {
			v.done()
		} else {
			v.Parent().TakeKeyboardFocus()
		}
	case KeyBackspace, KeyDelete:
		if p, ok := v.Parent().(*typeView); ok {
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
			v.TakeKeyboardFocus()
		})
	}
}

func (v typeView) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[v.focused])
	SetLineWidth(1)
	DrawRect(v.Rect())
}
