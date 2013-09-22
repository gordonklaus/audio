package gui

type MouseHandler interface {
	MousePressed(button int, p Point)
	MouseDragged(button int, p Point)
	MouseReleased(button int, p Point)
}

type AggregateMouseHandler []MouseHandler

func (a AggregateMouseHandler) MousePressed(button int, p Point) {
	for _, h := range a {
		h.MousePressed(button, p)
	}
}
func (a AggregateMouseHandler) MouseDragged(button int, p Point) {
	for _, h := range a {
		h.MouseDragged(button, p)
	}
}
func (a AggregateMouseHandler) MouseReleased(button int, p Point) {
	for _, h := range a {
		h.MouseReleased(button, p)
	}
}

type ClickHandler func(int, Point)

func (c ClickHandler) MousePressed(button int, p Point) {
	c(button, p)
}
func (c ClickHandler) MouseDragged(button int, p Point)  {}
func (c ClickHandler) MouseReleased(button int, p Point) {}

func NewClickKeyboardFocuser(view View) ClickHandler {
	return func(int, Point) {
		view.TakeKeyboardFocus()
	}
}

type ViewDragger struct {
	view View
	p    Point
}

func NewViewDragger(view View) *ViewDragger {
	return &ViewDragger{view: view}
}
func (d *ViewDragger) MousePressed(button int, p Point) {
	d.view.Raise()
	d.p = p
}
func drag(v View, p1, p2 Point)                         { v.Move(v.Position().Add(p2.Sub(p1))) }
func (d ViewDragger) MouseDragged(button int, p Point)  { drag(d.view, d.p, p) }
func (d ViewDragger) MouseReleased(button int, p Point) { drag(d.view, d.p, p) }

type ViewPanner struct {
	view View
	p    Point
}

func NewViewPanner(view View) *ViewPanner {
	return &ViewPanner{view: view}
}
func (vp *ViewPanner) MousePressed(button int, p Point) {
	vp.p = p
}
func pan(v View, p1, p2 Point)                           { v.Pan(v.Rect().Min.Add(p1).Sub(p2)) }
func (vp *ViewPanner) MouseDragged(button int, p Point)  { pan(vp.view, vp.p, p) }
func (vp *ViewPanner) MouseReleased(button int, p Point) { pan(vp.view, vp.p, p) }
