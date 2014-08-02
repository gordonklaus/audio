package gui

type Mouser interface {
	Mouse(MouseEvent)
}

type MouseEvent struct {
	Pos                        Point
	Enter, Leave               bool
	Move, Press, Release, Drag bool
	Button                     int
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
	return func(m MouseEvent) {
		if m.Press {
			SetKeyFocus(view)
		}
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
	switch {
	case m.Press:
		Raise(d.v)
		d.p = m.Pos
	case m.Drag, m.Release:
		d.v.Move(Pos(d.v).Add(m.Pos.Sub(d.p)))
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
	switch {
	case m.Press:
		p.p = m.Pos
	case m.Drag, m.Release:
		p.v.Pan(InnerRect(p.v).Min.Add(p.p.Sub(m.Pos)))
	}
}
