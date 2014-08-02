package audiogui

import (
	"math"
	"reflect"

	"code.google.com/p/gordon-go/audio"
	. "code.google.com/p/gordon-go/gui"
)

type PatternView struct {
	*ViewBase
	pattern   *audio.Pattern
	noteType  reflect.Type
	attrs     []*attributeView
	transTime float64
	scaleTime float64
	timeGrid  *uniformGrid
	focusTime float64
}

func NewPatternView(pattern *audio.Pattern) *PatternView {
	p := &PatternView{pattern: pattern, scaleTime: 32}
	p.ViewBase = NewView(p)
	p.noteType = audio.InstrumentPlayMethod(pattern.Instrument).Type().In(0).Elem()
	for i := 0; i < p.noteType.NumField(); i++ {
		f := p.noteType.Field(i)
		if f.Name == "Note" {
			continue
		}
		a := newAttributeView(p, f.Name)
		p.attrs = append(p.attrs, a)
		p.Add(a)
	}
	p.timeGrid = &uniformGrid{0, 1}
	return p
}

func (p *PatternView) reform() {
	r := InnerRect(p)
	x1, y1 := r.Min.XY()
	x2, y2 := r.Max.XY()
	dy := (y2 - y1) / float64(len(p.attrs))
	y2 -= dy
	for _, a := range p.attrs {
		a.Resize(x2-x1, dy)
		a.Move(Pt(x1, y2))
		y2 -= dy
	}
}

func (p *PatternView) newNote() audio.Note {
	n := reflect.New(p.noteType)
	n.Elem().FieldByName("Note").Set(reflect.ValueOf(audio.NewNote(p.focusTime)))
	for _, a := range p.attrs {
		n.Elem().FieldByName(a.name).Set(reflect.ValueOf([]*audio.ControlPoint{{0, a.focusVal}}))
	}
	note := n.Interface().(audio.Note)
	p.pattern.Notes = append(p.pattern.Notes, note)
	p.pattern.Sort()
	return note
}

func (p *PatternView) TookKeyFocus() { SetKeyFocus(p.attrs[0]) }

func (p *PatternView) Resize(width, height float64) {
	p.ViewBase.Resize(width, height)
	p.reform()
}

type attributeView struct {
	*ViewBase
	pattern   *PatternView
	name      string
	nameText  *Text
	notes     []*noteView
	transVal  float64
	scaleVal  float64
	valueGrid *uniformGrid
	focused   bool
	focusVal  float64
}

func newAttributeView(p *PatternView, name string) *attributeView {
	a := &attributeView{pattern: p, name: name, scaleVal: 32}
	a.ViewBase = NewView(a)
	a.nameText = NewText(name)
	a.nameText.SetBackgroundColor(Color{})
	a.Add(a.nameText)
	for _, note := range p.pattern.Notes {
		n := newNoteView(a, note)
		a.notes = append(a.notes, n)
		a.Add(n)
	}
	a.valueGrid = &uniformGrid{0, .2}
	return a
}

func (a *attributeView) TookKeyFocus() { a.focused = true; Repaint(a) }
func (a *attributeView) LostKeyFocus() { a.focused = false; Repaint(a) }

func (a *attributeView) KeyPress(k KeyEvent) {
	if k.Alt {
		switch k.Key {
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			a.focusNearest(a.to(Pt(a.pattern.focusTime, a.focusVal)), k.Key)
		}
		return
	}

	switch k.Key {
	case KeyLeft, KeyRight:
		a.pattern.focusTime = math.Max(0, a.pattern.timeGrid.next(a.pattern.focusTime, k.Key == KeyRight))
		Repaint(a)
	case KeyDown, KeyUp:
		a.focusVal = a.valueGrid.next(a.focusVal, k.Key == KeyUp)
		Repaint(a)
	case KeyTab:
		for i, a2 := range a.pattern.attrs {
			if a2 == a {
				if k.Shift {
					i--
					i += len(a.pattern.attrs)
				} else {
					i++
				}
				SetKeyFocus(a.pattern.attrs[i%len(a.pattern.attrs)])
				break
			}
		}
	case KeyEnter:
		note := a.pattern.newNote()
		for _, a2 := range a.pattern.attrs {
			n := newNoteView(a2, note)
			a2.notes = append(a2.notes, n)
			a2.Add(n)
			if a2 == a {
				SetKeyFocus(n)
			}
		}
	case KeySpace:
		go func() {
			Play(a.pattern.pattern)
			a.pattern.pattern.Reset()
		}()
	}
}

func (a *attributeView) to(p Point) Point {
	return Pt(p.X*a.pattern.scaleTime+a.pattern.transTime, p.Y*a.scaleVal+a.transVal)
}
func (a *attributeView) from(p Point) Point {
	return Pt((p.X-a.pattern.transTime)/a.pattern.scaleTime, (p.Y-a.transVal)/a.scaleVal)
}

func (a *attributeView) focusNearest(pt Point, dirKey int) {
	if n := a.nearest(pt, dirKey); n != nil {
		SetKeyFocus(n)
	}
}

func (a *attributeView) nearest(pt Point, dirKey int) (nearest *noteView) {
	dir := map[int]Point{KeyLeft: {-1, 0}, KeyRight: {1, 0}, KeyUp: {0, 1}, KeyDown: {0, -1}}[dirKey]
	best := 0.0
	for _, n := range a.notes {
		d := Map(Pos(n), Parent(n), a).Sub(pt)
		score := (dir.X*d.X + dir.Y*d.Y) / (d.X*d.X + d.Y*d.Y)
		if score > best {
			best = score
			nearest = n
		}
	}
	return
}

func (a *attributeView) Resize(width, height float64) {
	a.ViewBase.Resize(width, height)
	a.nameText.Move(InnerRect(a).Max.Sub(Pt(Size(a.nameText))))
}

func (a *attributeView) Scroll(s ScrollEvent) {
	if s.Shift {
		dt := math.Pow(1.05, -s.Delta.X)
		dv := math.Pow(1.05, -s.Delta.Y)
		a.pattern.scaleTime = math.Max(10, math.Min(1000, a.pattern.scaleTime*dt))
		if a.pattern.scaleTime == 10 || a.pattern.scaleTime == 1000 {
			dt = 1
		}
		a.scaleVal = math.Max(10, math.Min(1000, a.scaleVal*dv))
		if a.scaleVal == 10 || a.scaleVal == 1000 {
			dv = 1
		}
		a.pattern.transTime = s.Pos.X + (a.pattern.transTime-s.Pos.X)*dt
		a.transVal = s.Pos.Y + (a.transVal-s.Pos.Y)*dv
	} else {
		a.pattern.transTime += 8 * s.Delta.X
		a.transVal += 8 * s.Delta.Y
	}
	for _, a := range a.pattern.attrs {
		for _, n := range a.notes {
			for _, p := range n.points {
				p.reform()
			}
		}
	}
}

func (a *attributeView) Paint() {
	r := InnerRect(a)
	SetLineWidth(2)
	SetColor(Color{.2, .2, .2, 1})
	min := a.from(r.Min)
	max := a.from(r.Max)
	for t := a.pattern.timeGrid.next(math.Nextafter(math.Max(0, min.X), -math.MaxFloat64), true); t < max.X; t = a.pattern.timeGrid.next(t, true) {
		DrawLine(a.to(Pt(t, min.Y)), a.to(Pt(t, max.Y)))
	}
	for v := a.valueGrid.next(math.Nextafter(min.Y, -math.MaxFloat64), true); v < max.Y; v = a.valueGrid.next(v, true) {
		DrawLine(a.to(Pt(math.Max(0, min.X), v)), a.to(Pt(max.X, v)))
	}

	SetLineWidth(3)
	SetColor(Color{1, 1, 1, .2})
	DrawLine(a.to(Pt(a.pattern.focusTime, min.Y)), a.to(Pt(a.pattern.focusTime, max.Y)))
	if !a.focused {
		SetColor(Color{1, 1, 1, .1})
	}
	DrawLine(a.to(Pt(math.Max(0, min.X), a.focusVal)), a.to(Pt(max.X, a.focusVal)))

	SetLineWidth(1)
	SetColor(Color{1, 1, 1, 1})
	DrawLine(r.Min, Pt(r.Max.X, r.Min.Y))
}

type noteView struct {
	*ViewBase
	attr    *attributeView
	note    audio.Note
	points  []*controlPointView
	focused bool
}

func newNoteView(attr *attributeView, note audio.Note) *noteView {
	n := &noteView{attr: attr, note: note}
	n.ViewBase = NewView(n)
	for _, point := range n.getPoints() {
		p := newControlPointView(n, point)
		n.points = append(n.points, p)
		n.Add(p)
		p.reform()
	}
	n.reform()
	return n
}

func (n *noteView) getPoints() []*audio.ControlPoint {
	return reflect.Indirect(reflect.ValueOf(n.note)).FieldByName(n.attr.name).Interface().([]*audio.ControlPoint)
}

func (n *noteView) setPoints(value []*audio.ControlPoint) {
	reflect.Indirect(reflect.ValueOf(n.note)).FieldByName(n.attr.name).Set(reflect.ValueOf(value))
}

func (n *noteView) TookKeyFocus() { n.focused = true; Repaint(n) }
func (n *noteView) LostKeyFocus() { n.focused = false; Repaint(n) }

func (n *noteView) KeyPress(k KeyEvent) {
	if k.Alt {
		switch k.Key {
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			n.attr.focusNearest(Pos(n), k.Key)
		}
		return
	}

	if k.Shift {
		switch k.Key {
		case KeyLeft, KeyRight:
			n.setTime(n.attr.pattern.timeGrid.next(n.note.Time(), k.Key == KeyRight))
		case KeyDown, KeyUp:
			p0 := n.points[0].point
			d := n.attr.valueGrid.next(p0.Value, k.Key == KeyUp) - p0.Value
			for _, p := range n.points {
				p.setValue(p.point.Value + d)
			}
		}
		return
	}

	switch k.Key {
	case KeyLeft:
		n.attr.pattern.focusTime = n.note.Time()
		SetKeyFocus(n.attr)
	case KeyRight:
		n.attr.pattern.focusTime = n.attr.pattern.timeGrid.next(n.note.Time(), true)
		SetKeyFocus(n.attr)
	case KeyComma:
		n.newPoint(0)
	case KeyPeriod:
		n.newPoint(len(n.points))
	case KeyEnter:
		SetKeyFocus(n.points[0])
	}
}

func (n *noteView) newPoint(i int) {
	t, v := 0.0, 0.0
	switch {
	case i == 0:
		t = n.attr.pattern.timeGrid.next(n.note.Time(), false) - n.note.Time()
		v = n.points[0].point.Value
	default:
		p, q := n.points[i-1].point, n.points[i].point
		t = (p.Time + q.Time) / 2
		v = (p.Value + q.Value) / 2
	case i == len(n.points):
		p := n.points[len(n.points)-1].point
		t = n.attr.pattern.timeGrid.next(p.Time+n.note.Time(), true) - n.note.Time()
		v = p.Value
	}
	p := n.insertPoint(&audio.ControlPoint{t, v}, i)
	n.normalizePoints()
	SetKeyFocus(p)
}

func (n *noteView) insertPoint(point *audio.ControlPoint, i int) *controlPointView {
	points := n.getPoints()
	n.setPoints(append(points[:i], append([]*audio.ControlPoint{point}, points[i:]...)...))
	p := newControlPointView(n, point)
	n.points = append(n.points[:i], append([]*controlPointView{p}, n.points[i:]...)...)
	n.Add(p)
	return p
}

func (n *noteView) setTime(t float64) {
	t = math.Max(0, t)
	n.note.SetTime(t)
	n.attr.pattern.pattern.Sort()
	for _, p := range n.points {
		p.reform()
	}
}

func (n *noteView) reform() {
	ResizeToFit(n, 0)
	n.Move(InnerRect(n).Min)
}

func (n *noteView) normalizePoints() {
	t := n.points[0].point.Time
	for _, p := range n.points {
		p.point.Time -= t
		p.reform()
	}
	n.setTime(n.note.Time() + t)
}

func (n *noteView) Paint() {
	SetLineWidth(2)
	SetColor(Color{.75, .75, .75, 1})
	if n.focused {
		SetLineWidth(3)
		SetColor(Color{1, 1, 1, 1})
	}
	for i, p := range n.points[1:] {
		DrawLine(Center(n.points[i]), Center(p))
	}
}

type controlPointView struct {
	*ViewBase
	note    *noteView
	point   *audio.ControlPoint
	focused bool
}

func newControlPointView(note *noteView, point *audio.ControlPoint) *controlPointView {
	p := &controlPointView{note: note, point: point}
	p.ViewBase = NewView(p)
	p.Resize(10, 10)
	p.Pan(Pt(-5, -5))
	return p
}

func (p *controlPointView) TookKeyFocus() { p.focused = true; Repaint(p) }
func (p *controlPointView) LostKeyFocus() { p.focused = false; Repaint(p) }

func (p *controlPointView) KeyPress(k KeyEvent) {
	if k.Shift {
		switch k.Key {
		case KeyLeft, KeyRight:
			p.setTime(p.note.attr.pattern.timeGrid.next(p.point.Time+p.note.note.Time(), k.Key == KeyRight) - p.note.note.Time())
		case KeyDown, KeyUp:
			p.setValue(p.note.attr.valueGrid.next(p.point.Value, k.Key == KeyUp))
		}
		return
	}

	switch k.Key {
	case KeyLeft, KeyRight:
		p.focusNext(k.Key)
	case KeyComma:
		p.note.newPoint(p.index())
	case KeyPeriod:
		p.note.newPoint(p.index() + 1)
	case KeyEscape:
		SetKeyFocus(p.note)
	}
}

func (p *controlPointView) setTime(t float64) {
	i := p.index()
	points := p.note.points
	if i == 0 {
		t = math.Max(-p.note.note.Time(), t)
	} else {
		t = math.Max(points[i-1].point.Time, t)
	}
	if i+1 < len(points) {
		t = math.Min(t, points[i+1].point.Time)
	}
	p.point.Time = t
	p.note.normalizePoints()
	p.reform()
}

func (p *controlPointView) setValue(v float64) {
	p.point.Value = v
	p.reform()
}

func (p *controlPointView) reform() {
	MoveOrigin(p, p.note.attr.to(Pt(p.note.note.Time()+p.point.Time, p.point.Value)))
	p.note.reform()
}

func (p *controlPointView) focusNext(dir int) {
	i := p.index()
	if dir == KeyLeft {
		i--
	} else {
		i++
	}
	if i >= 0 && i < len(p.note.points) {
		SetKeyFocus(p.note.points[i])
	}
}

func (p *controlPointView) index() int {
	for i, p2 := range p.note.points {
		if p2 == p {
			return i
		}
	}
	panic("unreachable")
}

func (p *controlPointView) Paint() {
	SetPointSize(5)
	if p.note.focused {
		SetPointSize(7)
	}
	if p.focused {
		SetPointSize(10)
	}
	DrawPoint(ZP)
}
