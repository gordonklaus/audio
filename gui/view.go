package gui

import (
	. "code.google.com/p/gordon-go/util"
	gl "github.com/chsc/gogl/gl21"
)

type View interface {
	base() *ViewBase

	Parent() View
	Children() []View
	AddChild(View)
	RemoveChild(View)

	Show()
	Hide()

	Raise()
	Lower()

	Pos() Point
	Move(Point)

	Rect() Rectangle
	SetRect(Rectangle)

	setKeyFocus(View)
	KeyFocus() View
	TookKeyFocus()
	LostKeyFocus()

	KeyPress(KeyEvent)
	KeyRelease(KeyEvent)

	setMouser(m MouserView, button int)

	Repaint()
	Paint()

	Do(func())
}

type KeyEvent struct {
	Key                     int
	Action                  int
	Text                    string // only present on Press and Repeat, not Release
	Shift, Ctrl, Alt, Super bool
}

type MouserView interface {
	Mouser
	View
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

func (v ViewBase) Parent() View     { return v.parent }
func (v ViewBase) Children() []View { return v.children }
func (v *ViewBase) AddChild(child View) {
	if child.Parent() != nil {
		child.Parent().RemoveChild(child)
	}
	v.children = append(v.children, child)
	child.base().parent = v.Self
	child.Repaint()
}
func (v *ViewBase) RemoveChild(child View) {
	SliceRemove(&v.children, child)
	child.base().parent = nil
	v.Self.Repaint()
}
func Close(v View) {
	if v.Parent() != nil {
		v.Parent().RemoveChild(v)
	}
}

func (v *ViewBase) Show() { v.hidden = false; v.Self.Repaint() }
func (v *ViewBase) Hide() { v.hidden = true; v.Self.Repaint() }

func (v *ViewBase) Raise() {
	if v.parent != nil {
		p := v.parent.base()
		for i, view := range p.children {
			if view == v.Self {
				p.children = append(append(p.children[:i], p.children[i+1:]...), view)
				p.Self.Repaint()
				return
			}
		}
	}
}
func (v *ViewBase) Lower() {
	if v.parent != nil {
		p := v.parent.base()
		for i, view := range p.children {
			if view == v.Self {
				p.children = append(p.children[i:i+1], append(p.children[:i], p.children[i+1:]...)...)
				p.Self.Repaint()
				return
			}
		}
	}
}

func (v ViewBase) Pos() Point { return v.position }
func (v *ViewBase) Move(p Point) {
	v.position = p
	v.Self.Repaint()
}
func MoveCenter(v View, p Point) { v.Move(p.Sub(Size(v).Div(2))) }
func MoveOrigin(v View, p Point) { v.Move(p.Add(v.Rect().Min)) }

func (v ViewBase) Rect() Rectangle      { return v.rect }
func Center(v View) Point               { return v.Rect().Min.Add(Size(v).Div(2)) }
func Size(v View) Point                 { return v.Rect().Size() }
func Width(v View) float64              { return v.Rect().Dx() }
func Height(v View) float64             { return v.Rect().Dy() }
func (v *ViewBase) SetRect(r Rectangle) { v.rect = r; v.Repaint() }
func Pan(v View, p Point) {
	r := v.Rect()
	v.SetRect(r.Add(p.Sub(r.Min)))
}
func Resize(v View, s Point) {
	r := v.Rect()
	r.Max = r.Min.Add(s)
	v.SetRect(r)
}
func ResizeToFit(v View, margin float64) {
	if len(v.Children()) == 0 {
		v.SetRect(ZR)
		return
	}
	c1 := v.Children()[0]
	rect := MapRectToParent(c1, c1.Rect())
	for _, c := range v.Children() {
		rect = rect.Union(MapRectToParent(c, c.Rect()))
	}
	v.SetRect(rect.Inset(-margin))
}

func SetKeyFocus(v View) { v.setKeyFocus(v) }
func (v *ViewBase) setKeyFocus(view View) {
	if v.parent != nil {
		v.parent.setKeyFocus(view)
	}
}
func (v ViewBase) KeyFocus() View {
	if v.parent != nil {
		return v.parent.KeyFocus()
	}
	return nil
}
func (v *ViewBase) TookKeyFocus() {}
func (v *ViewBase) LostKeyFocus() {}

func (v *ViewBase) KeyPress(event KeyEvent) {
	if v.parent != nil {
		v.parent.KeyPress(event)
	}
}
func (v *ViewBase) KeyRelease(event KeyEvent) {
	if v.parent != nil {
		v.parent.KeyRelease(event)
	}
}

func SetMouser(m MouserView, button int) { m.setMouser(m, button) }
func (v *ViewBase) setMouser(m MouserView, button int) {
	if v.parent != nil {
		v.parent.setMouser(m, button)
	}
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
	delta := v.Pos().Sub(v.Rect().Min)
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

func ViewAt(v View, p Point) View { return viewAtFunc(v, p, func(v View) View { return v }) }
func viewAtFunc(v View, p Point, f func(View) View) View {
	if !p.In(v.Rect()) {
		return nil
	}
	children := v.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		view := viewAtFunc(child, MapFromParent(child, p), f)
		if view != nil {
			return view
		}
	}
	return f(v)
}

func MapFromParent(v View, p Point) Point {
	return p.Sub(v.Pos()).Add(v.Rect().Min)
}
func MapFrom(v View, p Point, parent View) Point {
	if v == parent || v.Parent() == nil {
		return p
	}
	return MapFromParent(v, MapFrom(v.Parent(), p, parent))
}
func MapToParent(v View, p Point) Point {
	return p.Sub(v.Rect().Min).Add(v.Pos())
}
func MapTo(v View, p Point, parent View) Point {
	if v == parent || v.Parent() == nil {
		return p
	}
	return MapTo(v.Parent(), MapToParent(v, p), parent)
}

func MapRectFromParent(v View, r Rectangle) Rectangle {
	return Rectangle{MapFromParent(v, r.Min), MapFromParent(v, r.Max)}
}
func MapRectFrom(v View, r Rectangle, parent View) Rectangle {
	return Rectangle{MapFrom(v, r.Min, parent), MapFrom(v, r.Max, parent)}
}
func MapRectToParent(v View, r Rectangle) Rectangle {
	return Rectangle{MapToParent(v, r.Min), MapToParent(v, r.Max)}
}
func MapRectTo(v View, r Rectangle, parent View) Rectangle {
	return Rectangle{MapTo(v, r.Min, parent), MapTo(v, r.Max, parent)}
}
