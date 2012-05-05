package gui

import (
	gl "github.com/chsc/gogl/gl21"
)

type KeyEvent struct {
	Key int
	Text string
}

type View interface {
	GetViewBase() *ViewBase
	
	Parent() View
	SetParent(parent View)
	Children() []View
	AddChild(v View)
	RemoveChild(v View)
	ViewAt(point Point) View
	
	Close()
	
	Raise()
	RaiseChild(child View)
	Lower()
	LowerChild(child View)
	
	Position() Point
	Move(p Point)
	MoveCenter(p Point)
	Moved(p Point)
	
	Rect() Rectangle
	Center() Point
	Size() Point
	Width() float64
	Height() float64
	Resize(width, height float64)
	Pan(d Point)
	
	MapFromParent(point Point) Point
	MapFrom(point Point, parent View) Point
	MapToParent(point Point) Point
	MapTo(point Point, parent View) Point
	
	SetKeyboardFocus(view View)
	TakeKeyboardFocus()
	TookKeyboardFocus()
	LostKeyboardFocus()
	KeyPressed(event KeyEvent)
	KeyReleased(event KeyEvent)
	
	SetMouseFocus(focus MouseHandlerView, button int)
	GetMouseFocus(button int, p Point) MouseHandlerView
	
	Repaint()
	Paint()
}

type MouseHandlerView interface {
	View
	MouseHandler
}

type ViewBase struct {
	self View
	parent View
	children []View
	rect Rectangle
	position Point
}

func NewView(self View) *ViewBase {
	return &ViewBase{self, nil, make([]View, 0), ZR, ZP}
}

func (v *ViewBase) GetViewBase() *ViewBase { return v }
func (v ViewBase) Parent() View { return v.parent }
func (v *ViewBase) SetParent(parent View) { v.parent = parent }
func (v ViewBase) Children() []View { return v.children }
func (v *ViewBase) AddChild(childView View) {
	v.children = append(v.children, childView)
	childView.SetParent(v.self)
	childView.Repaint()
}
func (v *ViewBase) RemoveChild(view View) {
	for i, child := range v.children {
		if child == view {
			v.children = append(v.children[:i], v.children[i+1:]...)
			view.SetParent(nil)
			v.self.Repaint()
			return
		}
	}
}
func (v ViewBase) ViewAt(point Point) View {
	if !point.In(v.self.Rect()) { return nil }
	children := v.self.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		view := child.ViewAt(child.MapFromParent(point))
		if view != nil { return view }
	}
	return v.self
}

func (v *ViewBase) Close() { if v.parent != nil { v.parent.RemoveChild(v.self) } }

func (v *ViewBase) Raise() { if v.parent != nil { v.parent.RaiseChild(v.self) } }
func (v *ViewBase) RaiseChild(child View) {
	for i, view := range v.children {
		if view == child {
			v.children = append(append(v.children[:i], v.children[i+1:]...), view)
			v.self.Repaint()
			return
		}
	}
}
func (v *ViewBase) Lower() { if v.parent != nil { v.parent.LowerChild(v.self) } }
func (v *ViewBase) LowerChild(child View) {
	for i, view := range v.children {
		if view == child {
			v.children = append(v.children[i:i+1], append(v.children[:i], v.children[i+1:]...)...)
			v.self.Repaint()
			return
		}
	}
}

func (v ViewBase) Position() Point { return v.position }
func (v *ViewBase) Move(p Point) {
	v.position = p
	v.self.Moved(p)
	v.self.Repaint()
}
func (v *ViewBase) MoveCenter(p Point) { v.Move(p.Sub(v.Size().Div(2))) }
func (v *ViewBase) Moved(p Point) {}

func (v ViewBase) Rect() Rectangle { return v.rect }
func (v ViewBase) Center() Point { return v.rect.Min.Add(v.Size().Div(2)) }
func (v ViewBase) Size() Point { return v.rect.Size() }
func (v ViewBase) Width() float64 { return v.rect.Dx() }
func (v ViewBase) Height() float64 { return v.rect.Dy() }
func (v *ViewBase) Resize(width, height float64) {
	v.rect.Max = v.rect.Min.Add(Pt(width, height))
	v.self.Repaint()
}
func (v *ViewBase) Pan(p Point) {
	v.rect = v.rect.Add(p.Sub(v.rect.Min))
	v.self.Repaint()
}

func (v ViewBase) MapFromParent(point Point) Point {
	return point.Sub(v.Position()).Add(v.Rect().Min)
}
func (v ViewBase) MapFrom(point Point, parent View) Point {
	if v.self == parent { return point }
	return v.MapFromParent(v.parent.MapFrom(point, parent))
}
func (v ViewBase) MapToParent(point Point) Point {
	return point.Sub(v.Rect().Min).Add(v.Position())
}
func (v ViewBase) MapTo(point Point, parent View) Point {
	if v.self == parent { return point }
	return v.parent.MapTo(v.MapToParent(point), parent)
}

func (v *ViewBase) SetKeyboardFocus(view View) { if v.parent != nil { v.parent.SetKeyboardFocus(view) } }
func (v *ViewBase) TakeKeyboardFocus() { v.self.SetKeyboardFocus(v.self) }
func (v *ViewBase) TookKeyboardFocus() {}
func (v *ViewBase) LostKeyboardFocus() {}
func (v *ViewBase) KeyPressed(event KeyEvent) { if v.parent != nil { v.parent.KeyPressed(event) } }
func (v *ViewBase) KeyReleased(event KeyEvent) { if v.parent != nil { v.parent.KeyReleased(event) } }

func (v *ViewBase) SetMouseFocus(focus MouseHandlerView, button int) { if v.parent != nil { v.parent.SetMouseFocus(focus, button) } }
func (v *ViewBase) GetMouseFocus(button int, p Point) MouseHandlerView {
	if !p.In(v.Rect()) { return nil }
	children := v.self.Children()
	for i := len(children) - 1; i >= 0; i-- {
		if c := children[i].GetMouseFocus(button, children[i].MapFromParent(p)); c != nil { return c }
	}
	f, _ := v.self.(MouseHandlerView)
	return f
}

func (v ViewBase) Repaint() { if v.parent != nil { v.parent.Repaint() } }
func (v ViewBase) paintBase() {
	gl.PushMatrix(); defer gl.PopMatrix()
	gl.PushAttrib(gl.ALL_ATTRIB_BITS); defer gl.PopAttrib()
	delta := v.Position().Sub(v.Rect().Min)
	gl.Translated(gl.Double(delta.X), gl.Double(delta.Y), 0)
	v.self.Paint()
	for _, child := range v.self.Children() {
		child.GetViewBase().paintBase()
	}
}
func (v ViewBase) Paint() {}
