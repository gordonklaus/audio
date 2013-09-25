package gui

import (
	. "code.google.com/p/gordon-go/util"
	gl "github.com/chsc/gogl/gl21"
)

type View interface {
	base() *ViewBase
	win() *Window

	Moved(Point)
	RectSet(Rectangle)

	TookKeyFocus()
	LostKeyFocus()

	KeyPress(KeyEvent)
	KeyRelease(KeyEvent)

	Paint()
}

type KeyEvent struct {
	Key                     int
	action                  int
	Repeat                  bool
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
func (v *ViewBase) win() *Window {
	if v.parent == nil {
		return nil
	}
	return v.parent.win()
}

func Parent(v View) View     { return v.base().parent }
func Children(v View) []View { return v.base().children }
func AddChild(v, child View) {
	if Parent(child) != nil {
		RemoveChild(Parent(child), child)
	}
	b := v.base()
	b.children = append(b.children, child)
	child.base().parent = v
	Repaint(child)
}
func RemoveChild(v, child View) {
	SliceRemove(&v.base().children, child)
	child.base().parent = nil
	Repaint(v)
}
func Close(v View) {
	if Parent(v) != nil {
		RemoveChild(Parent(v), v)
	}
}

func Show(v View) { v.base().hidden = false; Repaint(v) }
func Hide(v View) { v.base().hidden = true; Repaint(v) }

func Raise(v View) {
	if Parent(v) != nil {
		p := Parent(v).base()
		for i, view := range p.children {
			if view == v {
				p.children = append(append(p.children[:i], p.children[i+1:]...), view)
				Repaint(p)
				return
			}
		}
	}
}
func Lower(v View) {
	if Parent(v) != nil {
		p := Parent(v).base()
		for i, view := range p.children {
			if view == v {
				p.children = append(p.children[i:i+1], append(p.children[:i], p.children[i+1:]...)...)
				Repaint(p)
				return
			}
		}
	}
}

func Pos(v View) Point { return v.base().position }
func Move(v View, p Point) {
	v.base().position = p
	v.Moved(p)
	Repaint(v)
}
func MoveCenter(v View, p Point) { Move(v, p.Sub(Size(v).Div(2))) }
func MoveOrigin(v View, p Point) { Move(v, p.Add(Rect(v).Min)) }
func (v ViewBase) Moved(p Point) {}

func Rect(v View) Rectangle { return v.base().rect }
func Center(v View) Point   { return Rect(v).Min.Add(Size(v).Div(2)) }
func Size(v View) Point     { return Rect(v).Size() }
func Width(v View) float64  { return Rect(v).Dx() }
func Height(v View) float64 { return Rect(v).Dy() }
func SetRect(v View, r Rectangle) {
	v.base().rect = r
	v.RectSet(r)
	Repaint(v)
}
func (v ViewBase) RectSet(r Rectangle) {}
func Pan(v View, p Point) {
	r := Rect(v)
	SetRect(v, r.Add(p.Sub(r.Min)))
}
func Resize(v View, s Point) {
	r := Rect(v)
	r.Max = r.Min.Add(s)
	SetRect(v, r)
}
func ResizeToFit(v View, margin float64) {
	if len(Children(v)) == 0 {
		SetRect(v, ZR)
		return
	}
	c1 := Children(v)[0]
	rect := MapRectToParent(c1, Rect(c1))
	for _, c := range Children(v) {
		rect = rect.Union(MapRectToParent(c, Rect(c)))
	}
	SetRect(v, rect.Inset(-margin))
}

func SetKeyFocus(v View) {
	if w := v.win(); w != nil {
		w.setKeyFocus(v)
	}
}
func KeyFocus(v View) View {
	if w := v.win(); w != nil {
		return w.keyFocus
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

func SetMouser(m MouserView, button int) {
	if w := m.win(); w != nil {
		w.setMouser(m, button)
	}
}

func Repaint(v View) {
	if w := v.win(); w != nil {
		w.repaint()
	}
}

func (v ViewBase) paint() {
	if v.hidden {
		return
	}
	gl.PushMatrix()
	defer gl.PopMatrix()
	delta := v.position.Sub(v.rect.Min)
	gl.Translated(gl.Double(delta.X), gl.Double(delta.Y), 0)
	v.Self.Paint()
	for _, child := range v.children {
		child.base().paint()
	}
}
func (v ViewBase) Paint() {}

func Do(v View, f func()) {
	if w := v.win(); w != nil {
		w.do <- f
	}
}

func ViewAt(v View, p Point) View { return viewAtFunc(v, p, func(v View) View { return v }) }
func viewAtFunc(v View, p Point, f func(View) View) View {
	if !p.In(Rect(v)) {
		return nil
	}
	children := Children(v)
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
	return p.Sub(Pos(v)).Add(Rect(v).Min)
}
func MapFrom(v View, p Point, parent View) Point {
	if v == parent || Parent(v) == nil {
		return p
	}
	return MapFromParent(v, MapFrom(Parent(v), p, parent))
}
func MapToParent(v View, p Point) Point {
	return p.Sub(Rect(v).Min).Add(Pos(v))
}
func MapTo(v View, p Point, parent View) Point {
	if v == parent || Parent(v) == nil {
		return p
	}
	return MapTo(Parent(v), MapToParent(v, p), parent)
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
