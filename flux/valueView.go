package main

import (
	."code.google.com/p/gordon-go/gui"
)

type valueView struct {
	*ViewBase
	value *ValueInfo
	name Text
	typ *typeView
}

func newValueView(value *ValueInfo) *valueView {
	v := &valueView{value:value}
	v.ViewBase = NewView(v)
	v.name = NewText(value.name)
	v.name.SetTextColor(getTextColor(v.value, .7))
	v.name.SetBackgroundColor(Color{0, 0, 0, .3})
	v.AddChild(v.name)
	v.typ = newTypeView(value.typ)
	v.typ.Move(Pt(v.name.Width() + 5, 0))
	v.AddChild(v.typ)
	v.name.Move(Pt(0, (v.typ.Height() - v.name.Height()) / 2))
	ResizeToFit(v, 2)
	return v
}

func (v valueView) Paint() {
	SetColor(Color{1, 1, 1, .3})
	SetLineWidth(1)
	DrawRect(v.Rect())
}
