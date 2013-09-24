package gui

type Mouser interface {
	Mouse(MouseEvent)
}

type MouseEvent struct {
	Pos    Point
	Action int
	Button int
}

type AggregateMouser []Mouser

func (a AggregateMouser) Mouse(m MouseEvent) {
	for _, h := range a {
		h.Mouse(m)
	}
}

type Clicker func(MouseEvent)

func (c Clicker) Mouse(m MouseEvent) { c(m) }

func NewClickFocuser(view View) Clicker {
	return func(MouseEvent) {
		SetKeyFocus(view)
	}
}

type Mover struct {
	v View
	p Point
}

func NewMover(v View) *Mover {
	return &Mover{v: v}
}

func (d *Mover) Mouse(m MouseEvent) {
	switch m.Action {
	case Press:
		d.v.Raise()
		d.p = m.Pos
	case Drag, Release:
		d.v.Move(d.v.Pos().Add(m.Pos.Sub(d.p)))
	}
}

type Panner struct {
	v View
	p Point
}

func NewPanner(v View) *Panner {
	return &Panner{v: v}
}

func (p *Panner) Mouse(m MouseEvent) {
	switch m.Action {
	case Press:
		p.p = m.Pos
	case Drag, Release:
		Pan(p.v, p.p.Sub(m.Pos))
	}
}
