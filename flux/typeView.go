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
	name *TextBase  // typeView may also be used as a valueView, in which case this is non-nil
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

func newValueView(value *ValueInfo) *typeView {
	v := newTypeView(&value.typ)
	v.name = NewText(value.name)
	v.name.SetTextColor(getTextColor(value, .7))
	v.name.SetBackgroundColor(Color{0, 0, 0, .3})
	v.name.TextChanged.Connect(func(...interface{}) { v.reform() })
	v.AddChild(v.name)
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
	if v.name != nil { v.AddChild(v.name) }
	
	v.reform()
}

func (v *typeView) reform() {	
	const spacing = 2
	maxWidth := float64(0)
	h1 := float64(0); for i, c := range v.childTypes.left { h1 += c.Height(); if i > 0 { h1 += spacing }; if w := c.Width(); w > maxWidth { maxWidth = w } }
	h2 := float64(0); for i, c := range v.childTypes.right { h2 += c.Height(); if i > 0 { h2 += spacing } }
	x := 0.0
	if v.name != nil {
		v.name.Move(Pt(0, (Max(h1, h2) - v.name.Height()) / 2))
		x += v.name.Width() + spacing
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
	if v.name != nil {
		v.name.Accept.ConnectSingleShot(func(...interface{}) { v.editType(done) })
		v.name.Reject.ConnectSingleShot(func(...interface{}) { done() })
		v.name.TakeKeyboardFocus()
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
				f := v.Block().Outermost().function
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
	
	switch (*v.typ).(type) {
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
	case *StructType:
	case *FuncType:
	case *InterfaceType:
	}
	
	done()
}

func (v *typeView) focusNearest(child *typeView, dir int) {
	switch (*v.typ).(type) {
	case *MapType:
		switch dir {
		case KeyLeft:
			v.childTypes.left[0].TakeKeyboardFocus()
		case KeyRight:
			v.childTypes.right[0].TakeKeyboardFocus()
		}
	case *StructType:
	case *FuncType:
	case *InterfaceType:
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
		switch (*v.typ).(type) {
		case *PointerType, *ArrayType, *SliceType, *ChanType:
			v.childTypes.right[0].TakeKeyboardFocus()
		case *MapType:
			v.childTypes.left[0].TakeKeyboardFocus()
		case *StructType:
		case *FuncType:
		case *InterfaceType:
		}
	case KeyEsc:
		if v.done != nil {
			v.done()
		} else {
			v.Parent().TakeKeyboardFocus()
		}
	case KeyBackspace, KeyDel:
		oldTyp := *v.typ
		v.setType(nil)
		v.edit(func() {
			if *v.typ == nil { v.setType(oldTyp) }
			v.TakeKeyboardFocus()
		})
	}
}

func (v typeView) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[v.focused])
	SetLineWidth(1)
	DrawRect(v.Rect())
}
