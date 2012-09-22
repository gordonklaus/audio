package main

import (
	."code.google.com/p/gordon-go/gui"
	."math"
)

type typeView struct {
	*ViewBase
	typ Type
	text Text
}

func newTypeView(typ Type) *typeView {
	v := &typeView{typ:typ}
	v.ViewBase = NewView(v)
	v.text = NewText("")
	v.text.SetTextColor(getTextColor(&NamedType{}, .7))
	v.text.SetBackgroundColor(Color{0, 0, 0, .3})
	v.AddChild(v.text)
	s := ""
	var col1, col2 []View
	switch t := typ.(type) {
	case PointerType:
		s = "*"
		col2 = []View{newTypeView(t.element)}
	case ArrayType:
		s = "[]"
		col2 = []View{newTypeView(t.element)}
	case SliceType:
		s = "[]"
		col2 = []View{newTypeView(t.element)}
	case ChanType:
		s = "chan"
		col2 = []View{newTypeView(t.element)}
	case MapType:
		s = ":"
		col1 = []View{newTypeView(t.key)}
		col2 = []View{newTypeView(t.value)}
	case StructType:
		s = "struct"
		for _, f := range t.fields { col2 = append(col2, newValueView(f)) }
	case FuncType:
		s = "func"
		for _, v := range t.parameters { col1 = append(col1, newValueView(v)) }
		for _, v := range t.results { col2 = append(col2, newValueView(v)) }
	case InterfaceType:
		s = "interface"
		for _, m := range t.methods { col2 = append(col2, newValueView(m)) }
	case *NamedType:
		s = t.name
	}
	v.text.SetText(s)
	
	const spacing = 2
	maxWidth := float64(0)
	h1 := float64(0); for i, c := range col1 { h1 += c.Height(); if i > 0 { h1 += spacing }; if w := c.Width(); w > maxWidth { maxWidth = w } }
	h2 := float64(0); for i, c := range col2 { h2 += c.Height(); if i > 0 { h2 += spacing } }
	y := Max(0, h2 - h1) / 2
	for i := len(col1) - 1; i >= 0; i-- {
		c := col1[i]
		v.AddChild(c)
		c.Move(Pt(maxWidth - c.Width(), y))
		y += c.Height() + spacing
	}
	x := maxWidth + spacing
	v.text.Move(Pt(x, (Max(h1, h2) - v.text.Height()) / 2))
	x += v.text.Width() + spacing
	y = Max(0, h1 - h2) / 2
	for i := len(col2) - 1; i >= 0; i-- {
		c := col2[i]
		v.AddChild(c)
		c.Move(Pt(x, y))
		y += c.Height() + spacing
	}
	
	ResizeToFit(v, 2)
	return v
}

func (v typeView) Paint() {
	SetColor(Color{1, 1, 1, .3})
	SetLineWidth(1)
	DrawRect(v.Rect())
}
