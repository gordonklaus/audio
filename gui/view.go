package gui

import (
	. "code.google.com/p/gordon-go/util"
	gl "github.com/chsc/gogl/gl21"
)

type KeyEvent struct {
	Key                     int
	Action                  int
	Text                    string // only present on Press and Repeat, not Release
	Shift, Ctrl, Alt, Super bool
}

type View interface {
	base() *ViewBase

	Parent() View
	SetParent(parent View)
	Children() []View
	AddChild(v View)
	RemoveChild(v View)

	Show()
	Hide()
	Visible() bool
	Close()

	Raise()
	RaiseChild(child View)
	Lower()
	LowerChild(child View)

	Position() Point
	Move(p Point)

	Rect() Rectangle
	Center() Point
	Size() Point
	Width() float64
	Height() float64
	Resize(width, height float64)
	Pan(d Point)

	SetKeyboardFocus(view View)
	GetKeyboardFocus() View
	TakeKeyboardFocus()
	TookKeyboardFocus()
	LostKeyboardFocus()
	KeyPressed(event KeyEvent)
	KeyReleased(event KeyEvent)

	SetMouseFocus(focus MouseHandlerView, button int)
	GetMouseFocus(button int, p Point) MouseHandlerView

	Repaint()
	Paint()

	Do(func())
}

type MouseHandlerView interface {
	View
	MouseHandler
}

func ResizeToFit(v View, margin float64) {
	if len(v.Children()) == 0 {
		v.Resize(0, 0)
		return
	}
	c1 := v.Children()[0]
	rect := MapRectToParent(c1, c1.Rect())
	for _, c := range v.Children() {
		rect = rect.Union(MapRectToParent(c, c.Rect()))
	}
	rect = rect.Inset(-margin)
	v.Resize(rect.Dx(), rect.Dy())
	v.Pan(rect.Min)
}

type ViewBase struct {
	Self     View
	parent   View
	children []View
	hidden   bool
	rect     Rectangle
	position Point
}

func NewView(self View) *ViewBase {
	v := &ViewBase{}
	if self == nil {
		self = v
	}
	v.Self = self
	return v
}

func (v *ViewBase) base() *ViewBase { return v }
func (v *ViewBase) Close() {
	if v.parent != nil {
		v.parent.RemoveChild(v.Self)
	}
}

func (v ViewBase) Parent() View           { return v.parent }
func (v *ViewBase) SetParent(parent View) { v.parent = parent }
func (v ViewBase) Children() []View       { return v.children }
func (v *ViewBase) AddChild(child View) {
	if child.Parent() != nil {
		child.Parent().RemoveChild(child)
	}
	v.children = append(v.children, child)
	child.SetParent(v.Self)
	child.Repaint()
}
func (v *ViewBase) RemoveChild(child View) {
	SliceRemove(&v.children, child)
	child.SetParent(nil)
	v.Self.Repaint()
}

func (v *ViewBase) Show()        { v.hidden = false; v.Self.Repaint() }
func (v *ViewBase) Hide()        { v.hidden = true; v.Self.Repaint() }
func (v ViewBase) Visible() bool { return !v.hidden }

func (v *ViewBase) Raise() {
	if v.parent != nil {
		v.parent.RaiseChild(v.Self)
	}
}
func (v *ViewBase) RaiseChild(child View) {
	for i, view := range v.children {
		if view == child {
			v.children = append(append(v.children[:i], v.children[i+1:]...), view)
			v.Self.Repaint()
			return
		}
	}
}
func (v *ViewBase) Lower() {
	if v.parent != nil {
		v.parent.LowerChild(v.Self)
	}
}
func (v *ViewBase) LowerChild(child View) {
	for i, view := range v.children {
		if view == child {
			v.children = append(v.children[i:i+1], append(v.children[:i], v.children[i+1:]...)...)
			v.Self.Repaint()
			return
		}
	}
}

func (v ViewBase) Position() Point { return v.position }
func (v *ViewBase) Move(p Point) {
	v.position = p
	v.Self.Repaint()
}
func MoveCenter(v View, p Point) { v.Move(p.Sub(v.Size().Div(2))) }
func MoveOrigin(v View, p Point) { v.Move(p.Add(v.Rect().Min)) }

func (v ViewBase) Rect() Rectangle { return v.rect }
func (v ViewBase) Center() Point   { return v.rect.Min.Add(v.Size().Div(2)) }
func (v ViewBase) Size() Point     { return v.rect.Size() }
func (v ViewBase) Width() float64  { return v.rect.Dx() }
func (v ViewBase) Height() float64 { return v.rect.Dy() }
func (v *ViewBase) Resize(width, height float64) {
	v.rect.Max = v.rect.Min.Add(Pt(width, height))
	v.Self.Repaint()
}
func (v *ViewBase) Pan(p Point) {
	v.rect = v.rect.Add(p.Sub(v.rect.Min))
	v.Self.Repaint()
}

func (v *ViewBase) SetKeyboardFocus(view View) {
	if v.parent != nil {
		v.parent.SetKeyboardFocus(view)
	}
}
func (v ViewBase) GetKeyboardFocus() View {
	if v.parent != nil {
		return v.parent.GetKeyboardFocus()
	}
	return nil
}
func (v *ViewBase) TakeKeyboardFocus() { v.Self.SetKeyboardFocus(v.Self) }
func (v *ViewBase) TookKeyboardFocus() {}
func (v *ViewBase) LostKeyboardFocus() {}
func (v *ViewBase) KeyPressed(event KeyEvent) {
	if v.parent != nil {
		v.parent.KeyPressed(event)
	}
}
func (v *ViewBase) KeyReleased(event KeyEvent) {
	if v.parent != nil {
		v.parent.KeyReleased(event)
	}
}

func (v *ViewBase) SetMouseFocus(focus MouseHandlerView, button int) {
	if v.parent != nil {
		v.parent.SetMouseFocus(focus, button)
	}
}
func (v *ViewBase) GetMouseFocus(button int, p Point) MouseHandlerView {
	if !p.In(v.Rect()) {
		return nil
	}
	children := v.Self.Children()
	for i := len(children) - 1; i >= 0; i-- {
		if c := children[i].GetMouseFocus(button, MapFromParent(children[i], p)); c != nil {
			return c
		}
	}
	f, _ := v.Self.(MouseHandlerView)
	return f
}

func (v ViewBase) Repaint() {
	if v.parent != nil {
		v.parent.Repaint()
	}
}
func (v ViewBase) paint() {
	if v.hidden {
		return
	}
	gl.PushMatrix()
	defer gl.PopMatrix()
	delta := v.Position().Sub(v.Rect().Min)
	gl.Translated(gl.Double(delta.X), gl.Double(delta.Y), 0)
	v.Self.Paint()
	for _, child := range v.Self.Children() {
		child.base().paint()
	}
}
func (v ViewBase) Paint() {}

func (v ViewBase) Do(f func()) {
	if v.parent != nil {
		v.parent.Do(f)
	}
}

func ViewAt(v View, point Point) View {
	if !point.In(v.Rect()) {
		return nil
	}
	children := v.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		view := ViewAt(child, MapFromParent(child, point))
		if view != nil {
			return view
		}
	}
	return v
}

func MapFromParent(v View, point Point) Point {
	return point.Sub(v.Position()).Add(v.Rect().Min)
}
func MapFrom(v View, point Point, parent View) Point {
	if v == parent || v.Parent() == nil {
		return point
	}
	return MapFromParent(v, MapFrom(v.Parent(), point, parent))
}
func MapToParent(v View, point Point) Point {
	return point.Sub(v.Rect().Min).Add(v.Position())
}
func MapTo(v View, point Point, parent View) Point {
	if v == parent || v.Parent() == nil {
		return point
	}
	return MapTo(v.Parent(), MapToParent(v, point), parent)
}

func MapRectFromParent(v View, rect Rectangle) Rectangle {
	return Rectangle{MapFromParent(v, rect.Min), MapFromParent(v, rect.Max)}
}
func MapRectFrom(v View, rect Rectangle, parent View) Rectangle {
	return Rectangle{MapFrom(v, rect.Min, parent), MapFrom(v, rect.Max, parent)}
}
func MapRectToParent(v View, rect Rectangle) Rectangle {
	return Rectangle{MapToParent(v, rect.Min), MapToParent(v, rect.Max)}
}
func MapRectTo(v View, rect Rectangle, parent View) Rectangle {
	return Rectangle{MapTo(v, rect.Min, parent), MapTo(v, rect.Max, parent)}
}
