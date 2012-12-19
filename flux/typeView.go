package main

import (
	."code.google.com/p/gordon-go/gui"
	."github.com/jteeuwen/glfw"
	."math"
)

type typeView struct {
	*ViewBase
	typ *Type
	text Text
	childTypes struct{left, right []*typeView}
	focused bool
	done func()
	
	// typeView is also used as a valueView, in which case these are non-nil
	nameText *TextBase
}

func newTypeView(t *Type) *typeView {
	v := &typeView{}
	v.ViewBase = NewView(v)
	v.typ = t
	v.text = NewText("")
	v.text.SetTextColor(getTextColor(&NamedType{}, .7))
	v.text.SetBackgroundColor(Color{0, 0, 0, .3})
	v.AddChild(v.text)
	v.setType(*t)
	return v
}

func newValueView(val *ValueInfo) *typeView {
	v := newTypeView(&val.typ)
	v.nameText = NewText(val.name)
	v.nameText.SetTextColor(getTextColor(val, .7))
	v.nameText.SetBackgroundColor(Color{0, 0, 0, .3})
	v.nameText.TextChanged.Connect(func(...interface{}) {
		val.name = v.nameText.GetText()
		v.reform()
	})
	v.AddChild(v.nameText)
	v.reform()
	return v
}

func (v *typeView) setType(t Type) {
	*v.typ = t
	s := ""
	v.childTypes.left, v.childTypes.right = nil, nil
	switch t := t.(type) {
	case *PointerType:
		s = "*"
		v.childTypes.right = []*typeView{newTypeView(&t.element)}
	case *ArrayType:
		s = "[]"
		v.childTypes.right = []*typeView{newTypeView(&t.element)}
	case *SliceType:
		s = "[]"
		v.childTypes.right = []*typeView{newTypeView(&t.element)}
	case *ChanType:
		s = "chan"
		v.childTypes.right = []*typeView{newTypeView(&t.element)}
	case *MapType:
		s = ":"
		v.childTypes.left = []*typeView{newTypeView(&t.key)}
		v.childTypes.right = []*typeView{newTypeView(&t.value)}
	case *StructType:
		s = "struct"
		for _, f := range t.fields { v.childTypes.right = append(v.childTypes.right, newValueView(f)) }
	case *FuncType:
		s = "func"
		for _, val := range t.parameters { v.childTypes.left = append(v.childTypes.left, newValueView(val)) }
		for _, val := range t.results { v.childTypes.right = append(v.childTypes.right, newValueView(val)) }
	case *InterfaceType:
		s = "interface"
		for _, m := range t.methods { v.childTypes.right = append(v.childTypes.right, newValueView(m)) }
	case *NamedType:
		s = t.name
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
		v.nameText.Move(Pt(0, (Max(h1, h2) - v.nameText.Height()) / 2))
		x += v.nameText.Width() + spacing
	}
	y := Max(0, h2 - h1) / 2
	for i := len(v.childTypes.left) - 1; i >= 0; i-- {
		c := v.childTypes.left[i]
		c.Move(Pt(x + maxWidth - c.Width(), y))
		y += c.Height() + spacing
	}
	x += maxWidth + spacing
	v.text.Move(Pt(x, (Max(h1, h2) - v.text.Height()) / 2))
	x += v.text.Width() + spacing
	y = Max(0, h1 - h2) / 2
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
		var pkg *PackageInfo
		var imports []*PackageInfo
l:		for v := View(v); v != nil; v = v.Parent() {
			switch v := v.(type) {
			case Node:
				f := v.Block().Func()
				pkg, imports = f.pkg(), f.imports()
				break l
			}
		}
		browser := NewBrowser(typesOnly, pkg, imports)
		v.AddChild(browser)
		browser.Move(v.Center())
		browser.accepted.Connect(func(info ...interface{}) {
			t := info[0].(Type)
			if nt, ok := t.(*NamedType); ok && protoType[nt] { // move this into Browser?
				t = newProtoType(nt)
			}
			v.setType(t)
			v.editType(done)
		})
		browser.canceled.Connect(func(...interface{}) { done() })
		browser.text.TakeKeyboardFocus()
		return
	}
	
	switch t := (*v.typ).(type) {
	case *PointerType, *ArrayType, *SliceType, *ChanType:
		if elt := v.childTypes.right[0]; *elt.typ == nil {
			elt.edit(func() {
				if *elt.typ == nil { v.setType(nil) }
				v.editType(done)
			})
			return
		}
	case *MapType:
		key := v.childTypes.left[0]
		val := v.childTypes.right[0]
		switch Type(nil) {
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
	case *StructType:
		v.addFields(&t.fields, false, &v.childTypes.right, done)
		return
	case *FuncType:
		v.addFields(&t.parameters, false, &v.childTypes.left, func() {
			v.addFields(&t.results, false, &v.childTypes.right, done)
		})
		return
	case *InterfaceType:
		v.addFields(&t.methods, true, &v.childTypes.right, done)
		return
	}
	
	done()
}
func (v *typeView) addFields(fields *[]*ValueInfo, funcVal bool, childTypes *[]*typeView, done func()) {
	val := &ValueInfo{}
	if funcVal {
		val.typ = &FuncType{}
	}
	*fields = append(*fields, val)
	v.setType(*v.typ)
	i := len(*fields) - 1
	f := (*childTypes)[i]
	f.edit(func() {
		if *f.typ == nil || funcVal && len(f.nameText.GetText()) == 0 {
			*fields = (*fields)[:i]
			v.setType(*v.typ)
			done()
		} else {
			v.addFields(fields, funcVal, childTypes, done)
		}
	})
}

func (v *typeView) focusNearest(child *typeView, dir int) {
	var views []View
	for _, v := range append(v.childTypes.left, v.childTypes.right...) {
		views = append(views, v)
	}
	nearest := nearestView(v, views, child.MapTo(child.Center(), v), dir)
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
		case *PointerType, *ArrayType, *SliceType, *ChanType:
			v.childTypes.right[0].TakeKeyboardFocus()
		case *MapType:
			v.childTypes.left[0].TakeKeyboardFocus()
		case *StructType:
			if len(t.fields) == 0 {
				v.edit(done)
			} else {
				v.childTypes.right[0].TakeKeyboardFocus()
			}
		case *FuncType:
			switch {
			case len(t.parameters) == 0 && len(t.results) == 0:
				v.edit(done)
			case len(t.parameters) == 0:
				v.addFields(&t.parameters, false, &v.childTypes.left, func() {
					if len(t.parameters) == 0 {
						v.childTypes.right[0].TakeKeyboardFocus()
					} else {
						v.childTypes.left[0].TakeKeyboardFocus()
					}
				})
			case len(t.results) == 0:
				v.addFields(&t.results, false, &v.childTypes.right, func() {
				if len(t.results) == 0 {
					v.childTypes.left[0].TakeKeyboardFocus()
				} else {
					v.childTypes.right[0].TakeKeyboardFocus()
				}
				})
			default:
				v.childTypes.left[0].TakeKeyboardFocus()
			}
		case *InterfaceType:
			if len(t.methods) == 0 {
				v.edit(done)
			} else {
				v.childTypes.right[0].TakeKeyboardFocus()
			}
		}
	case KeyEsc:
		if v.done != nil {
			v.done()
		} else {
			v.Parent().TakeKeyboardFocus()
		}
	case KeyBackspace, KeyDel:
		if p, ok := v.Parent().(*typeView); ok {
			if _, ok := (*p.typ).(*InterfaceType); ok {
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
