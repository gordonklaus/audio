package gui

import (
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/util"
)

type KeyEvent struct {
	Key int
	Text string
	Shift, Ctrl, Alt, Super bool
}

type View interface {
	GetViewBase() *ViewBase
	
	Parent() View
	SetParent(parent View)
	Children() []View
	AddChild(v View)
	RemoveChild(v View)
	ViewAt(point Point) View
	
	Show()
	Hide()
	Close()
	
	Raise()
	RaiseChild(child View)
	Lower()
	LowerChild(child View)
	
	Position() Point
	Move(p Point)
	MoveCenter(p Point)
	
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
	
	MapRectFromParent(rect Rectangle) Rectangle
	MapRectFrom(rect Rectangle, parent View) Rectangle
	MapRectToParent(rect Rectangle) Rectangle
	MapRectTo(rect Rectangle, parent View) Rectangle
	
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

func ResizeToFit(v View, margin float64) {
	if len(v.Children()) == 0 {
		v.Resize(0, 0)
		return
	}
	c1 := v.Children()[0]
	rect := c1.MapRectToParent(c1.Rect())
	for _, c := range v.Children() {
		rect = rect.Union(c.MapRectToParent(c.Rect()))
	}
	rect = rect.Inset(-margin)
	v.Resize(rect.Dx(), rect.Dy())
	v.Pan(rect.Min)
}

type ViewBase struct {
	Self View
	parent View
	children []View
	hidden bool
	rect Rectangle
	position Point
}

func NewView(self View) *ViewBase {
	return &ViewBase{self, nil, make([]View, 0), false, ZR, ZP}
}

func (v *ViewBase) GetViewBase() *ViewBase { return v }
func (v ViewBase) Parent() View { return v.parent }
func (v *ViewBase) SetParent(parent View) { v.parent = parent }
func (v ViewBase) Children() []View { return v.children }
func (v *ViewBase) AddChild(childView View) {
	if childView.Parent() != nil {
		childView.Parent().RemoveChild(childView)
	}
	v.children = append(v.children, childView)
	childView.SetParent(v.Self)
	childView.Repaint()
}
func (v *ViewBase) RemoveChild(view View) {
	SliceRemove(&v.children, view)
	view.SetParent(nil)
	v.Self.Repaint()
}
func (v ViewBase) ViewAt(point Point) View {
	if !point.In(v.Self.Rect()) { return nil }
	children := v.Self.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		view := child.ViewAt(child.MapFromParent(point))
		if view != nil { return view }
	}
	return v.Self
}

func (v *ViewBase) Show() { v.hidden = false; v.Self.Repaint() }
func (v *ViewBase) Hide() { v.hidden = true; v.Self.Repaint() }
func (v *ViewBase) Close() { if v.parent != nil { v.parent.RemoveChild(v.Self) } }

func (v *ViewBase) Raise() { if v.parent != nil { v.parent.RaiseChild(v.Self) } }
func (v *ViewBase) RaiseChild(child View) {
	for i, view := range v.children {
		if view == child {
			v.children = append(append(v.children[:i], v.children[i+1:]...), view)
			v.Self.Repaint()
			return
		}
	}
}
func (v *ViewBase) Lower() { if v.parent != nil { v.parent.LowerChild(v.Self) } }
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
func (v *ViewBase) MoveCenter(p Point) { v.Self.Move(p.Sub(v.Size().Div(2))) }

func (v ViewBase) Rect() Rectangle { return v.rect }
func (v ViewBase) Center() Point { return v.rect.Min.Add(v.Size().Div(2)) }
func (v ViewBase) Size() Point { return v.rect.Size() }
func (v ViewBase) Width() float64 { return v.rect.Dx() }
func (v ViewBase) Height() float64 { return v.rect.Dy() }
func (v *ViewBase) Resize(width, height float64) {
	v.rect.Max = v.rect.Min.Add(Pt(width, height))
	v.Self.Repaint()
}
func (v *ViewBase) Pan(p Point) {
	v.rect = v.rect.Add(p.Sub(v.rect.Min))
	v.Self.Repaint()
}

func (v ViewBase) MapFromParent(point Point) Point {
	return point.Sub(v.Position()).Add(v.Rect().Min)
}
func (v ViewBase) MapFrom(point Point, parent View) Point {
	if v.Self == parent { return point }
	return v.MapFromParent(v.parent.MapFrom(point, parent))
}
func (v ViewBase) MapToParent(point Point) Point {
	return point.Sub(v.Rect().Min).Add(v.Position())
}
func (v ViewBase) MapTo(point Point, parent View) Point {
	if v.Self == parent { return point }
	return v.parent.MapTo(v.MapToParent(point), parent)
}

func (v ViewBase) MapRectFromParent(rect Rectangle) Rectangle { return Rectangle{v.MapFromParent(rect.Min), v.MapFromParent(rect.Max)} }
func (v ViewBase) MapRectFrom(rect Rectangle, parent View) Rectangle { return Rectangle{v.MapFrom(rect.Min, parent), v.MapFrom(rect.Max, parent)} }
func (v ViewBase) MapRectToParent(rect Rectangle) Rectangle { return Rectangle{v.MapToParent(rect.Min), v.MapToParent(rect.Max)} }
func (v ViewBase) MapRectTo(rect Rectangle, parent View) Rectangle { return Rectangle{v.MapTo(rect.Min, parent), v.MapTo(rect.Max, parent)} }

func (v *ViewBase) SetKeyboardFocus(view View) { if v.parent != nil { v.parent.SetKeyboardFocus(view) } }
func (v *ViewBase) TakeKeyboardFocus() { v.Self.SetKeyboardFocus(v.Self) }
func (v *ViewBase) TookKeyboardFocus() {}
func (v *ViewBase) LostKeyboardFocus() {}
func (v *ViewBase) KeyPressed(event KeyEvent) { if v.parent != nil { v.parent.KeyPressed(event) } }
func (v *ViewBase) KeyReleased(event KeyEvent) { if v.parent != nil { v.parent.KeyReleased(event) } }

func (v *ViewBase) SetMouseFocus(focus MouseHandlerView, button int) { if v.parent != nil { v.parent.SetMouseFocus(focus, button) } }
func (v *ViewBase) GetMouseFocus(button int, p Point) MouseHandlerView {
	if !p.In(v.Rect()) { return nil }
	children := v.Self.Children()
	for i := len(children) - 1; i >= 0; i-- {
		if c := children[i].GetMouseFocus(button, children[i].MapFromParent(p)); c != nil { return c }
	}
	f, _ := v.Self.(MouseHandlerView)
	return f
}

func (v ViewBase) Repaint() { if v.parent != nil { v.parent.Repaint() } }
func (v ViewBase) paintBase() {
	if v.hidden { return }
	gl.PushMatrix(); defer gl.PopMatrix()
	gl.PushAttrib(gl.ALL_ATTRIB_BITS); defer gl.PopAttrib()
	delta := v.Position().Sub(v.Rect().Min)
	gl.Translated(gl.Double(delta.X), gl.Double(delta.Y), 0)
	v.Self.Paint()
	for _, child := range v.Self.Children() {
		child.GetViewBase().paintBase()
	}
}
func (v ViewBase) Paint() {}
