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
	mode browserMode
	text Text
	childTypes struct{left, right []*typeView}
	focused bool
	done func()
	
	// typeView is also used as a valueView, in which case this is non-nil
	nameText *TextBase
}

func newTypeView(t *types.Type) *typeView {
	v := &typeView{typ: t, mode: typesOnly}
	v.ViewBase = NewView(v)
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
	v.nameText.TextChanged = func(text string) {
		*name = text
		v.reform()
	}
	v.AddChild(v.nameText)
	v.reform()
	return v
}

func (v *typeView) setType(t types.Type) {
	*v.typ = t
	s := ""
	v.childTypes.left, v.childTypes.right = nil, nil
	switch t := t.(type) {
	case generic:
		s = "<T>"
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
		for _, f := range t.Fields { v.childTypes.right = append(v.childTypes.right, newValueView(field{nil, f, nil})) }
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
	
	if _, ok := t.(*types.Pointer); ok && v.mode == compositeOrPtrType {
		v.childTypes.right[0].mode = compositeType
	}
	
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
		v.nameText.Accept = func(string) { v.editType(done) }
		v.nameText.Reject = done
		v.nameText.TakeKeyboardFocus()
		return
	}
	v.editType(done)
}
func (v *typeView) editType(done func()) {
	switch t := (*v.typ).(type) {
	case nil:
		var pkg *types.Package
		var imports []*types.Package
		for v := View(v); v != nil; v = v.Parent() {
			if n, ok := v.(node); ok {
				f := n.block().func_()
				pkg, imports = f.pkg(), f.imports()
				break
			}
		}
		if pkg == nil {
			// TODO: get pkg and imports for the type being edited
		}
		b := newBrowser(v.mode, pkg, imports)
		v.AddChild(b)
		b.Move(v.Center())
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
		b.text.TakeKeyboardFocus()
	case *types.Basic, *types.NamedType:
		done()
	case *types.Pointer, *types.Array, *types.Slice, *types.Chan:
		if elt := v.childTypes.right[0]; *elt.typ == nil {
			elt.editType(func() {
				if *elt.typ == nil { v.setType(nil) }
				v.editType(done)
			})
		} else {
			done()
		}
	case *types.Map:
		key := v.childTypes.left[0]
		val := v.childTypes.right[0]
		switch types.Type(nil) {
		case *key.typ:
			key.editType(func() {
				if *key.typ == nil { v.setType(nil) }
				v.editType(done)
			})
		case *val.typ:
			val.editType(func() {
				if *val.typ == nil { key.setType(nil) }
				v.editType(done)
			})
		default:
			done()
		}
	case *types.Struct:
		v.addFields(&t.Fields, &v.childTypes.right, done)
	case *types.Signature:
		v.addVars(&t.Params, &v.childTypes.left, func() {
			v.addVars(&t.Results, &v.childTypes.right, done)
		})
	case *types.Interface:
		v.addMethods(&t.Methods, &v.childTypes.right, done)
	}
}

func (v *typeView) insertField(f *[]*types.Field, childTypes *[]*typeView, before bool, i int, success, fail func()) {
	if !before {
		i++
	}
	*f = append((*f)[:i], append([]*types.Field{{}}, (*f)[i:]...)...)
	v.setType(*v.typ)
	t := (*childTypes)[i]
	t.edit(func() {
		if *t.typ == nil {
			*f = append((*f)[:i], (*f)[i+1:]...)
			v.setType(*v.typ)
			if fail != nil {
				fail()
			} else {
				if !before {
					i--
				}
				(*childTypes)[i].TakeKeyboardFocus()
			}
		} else {
			if success != nil {
				success()
			} else {
				t.TakeKeyboardFocus()
			}
		}
	})
}

func (v *typeView) addFields(f *[]*types.Field, childTypes *[]*typeView, done func()) {
	v.insertField(f, childTypes, true, len(*f), func() {
		v.addFields(f, childTypes, done)
	}, done)
}

func (v *typeView) insertVar(vs *[]*types.Var, childTypes *[]*typeView, before bool, i int, success, fail func()) {
	if !before {
		i++
	}
	*vs = append((*vs)[:i], append([]*types.Var{{}}, (*vs)[i:]...)...)
	v.setType(*v.typ)
	t := (*childTypes)[i]
	t.edit(func() {
		if *t.typ == nil {
			*vs = append((*vs)[:i], (*vs)[i+1:]...)
			v.setType(*v.typ)
			if fail != nil {
				fail()
			} else {
				if !before {
					i--
				}
				(*childTypes)[i].TakeKeyboardFocus()
			}
		} else {
			if success != nil {
				success()
			} else {
				t.TakeKeyboardFocus()
			}
		}
	})
}

func (v *typeView) addVars(vs *[]*types.Var, childTypes *[]*typeView, done func()) {
	v.insertVar(vs, childTypes, true, len(*vs), func() {
		v.addVars(vs, childTypes, done)
	}, done)
}

func (v *typeView) insertMethod(m *[]*types.Method, childTypes *[]*typeView, before bool, i int, success, fail func()) {
	if !before {
		i++
	}
	*m = append((*m)[:i], append([]*types.Method{{Type: &types.Signature{}}}, (*m)[i:]...)...)
	v.setType(*v.typ)
	t := (*childTypes)[i]
	t.edit(func() {
		if *t.typ == nil || len(t.nameText.GetText()) == 0 {
			*m = append((*m)[:i], (*m)[i+1:]...)
			v.setType(*v.typ)
			if fail != nil {
				fail()
			} else {
				if !before {
					i--
				}
				(*childTypes)[i].TakeKeyboardFocus()
			}
		} else {
			if success != nil {
				success()
			} else {
				t.TakeKeyboardFocus()
			}
		}
	})
}

func (v *typeView) addMethods(m *[]*types.Method, childTypes *[]*typeView, done func()) {
	v.insertMethod(m, childTypes, true, len(*m), func() {
		v.addMethods(m, childTypes, done)
	}, done)
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
		done := v.TakeKeyboardFocus
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
	case KeyBackspace:
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
	case KeyComma:
		if p, ok := v.Parent().(*typeView); ok {
			switch t := (*p.typ).(type) {
			case *types.Struct:
				for i, c := range p.childTypes.right {
					if c == v {
						p.insertField(&t.Fields, &p.childTypes.right, event.Shift, i, nil, nil)
						break
					}
				}
			case *types.Signature:
				for i, c := range p.childTypes.left {
					if c == v {
						p.insertVar(&t.Params, &p.childTypes.left, event.Shift, i, nil, nil)
						break
					}
				}
				for i, c := range p.childTypes.right {
					if c == v {
						p.insertVar(&t.Results, &p.childTypes.right, event.Shift, i, nil, nil)
						break
					}
				}
			case *types.Interface:
				for i, c := range p.childTypes.right {
					if c == v {
						p.insertMethod(&t.Methods, &p.childTypes.right, event.Shift, i, nil, nil)
						break
					}
				}
			}
		}
	case KeyDelete:
		if p, ok := v.Parent().(*typeView); ok {
			switch t := (*p.typ).(type) {
			case *types.Struct:
				for i, c := range p.childTypes.right {
					if c == v {
						t.Fields = append(t.Fields[:i], t.Fields[i+1:]...)
						p.setType(*p.typ)
						if len := len(t.Fields); len > 0 {
							if i == len {
								i--
							}
							p.childTypes.right[i].TakeKeyboardFocus()
						} else {
							p.TakeKeyboardFocus()
						}
						break
					}
				}
			case *types.Signature:
				for i, c := range p.childTypes.left {
					if c == v {
						t.Params = append(t.Params[:i], t.Params[i+1:]...)
						p.setType(*p.typ)
						if len := len(t.Params); len > 0 {
							if i == len {
								i--
							}
							p.childTypes.left[i].TakeKeyboardFocus()
						} else {
							p.TakeKeyboardFocus()
						}
						break
					}
				}
				for i, c := range p.childTypes.right {
					if c == v {
						t.Results = append(t.Results[:i], t.Results[i+1:]...)
						p.setType(*p.typ)
						if len := len(t.Results); len > 0 {
							if i == len {
								i--
							}
							p.childTypes.right[i].TakeKeyboardFocus()
						} else {
							p.TakeKeyboardFocus()
						}
						break
					}
				}
			case *types.Interface:
				for i, c := range p.childTypes.right {
					if c == v {
						t.Methods = append(t.Methods[:i], t.Methods[i+1:]...)
						p.setType(*p.typ)
						if len := len(t.Methods); len > 0 {
							if i == len {
								i--
							}
							p.childTypes.right[i].TakeKeyboardFocus()
						} else {
							p.TakeKeyboardFocus()
						}
						break
					}
				}
			}
		}
	}
}

func (v typeView) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[v.focused])
	SetLineWidth(1)
	DrawRect(v.Rect())
}


type generic struct { types.Type }

func underlying(t types.Type) types.Type {
	if nt, ok := t.(*types.NamedType); ok {
		return nt.Underlying
	}
	return t
}

func indirect(t types.Type) (types.Type, bool) {
	if p, ok := underlying(t).(*types.Pointer); ok {
		return p.Base, true
	}
	return t, false
}
