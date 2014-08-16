package audiogui

import (
	"fmt"
	"go/build"
	"math"
	"os"
	"path/filepath"
	"time"

	"code.google.com/p/gordon-go/audio"
	. "code.google.com/p/gordon-go/gui"
)

type PatternView struct {
	*ViewBase
	pattern    *audio.Pattern
	inst       audio.Instrument
	attrs      []*attributeView
	transTime  float64
	scaleTime  float64
	timeGrid   *uniformGrid
	cursorTime float64

	player      *audio.PatternPlayer
	play, close chan bool
	oldFocus    View
}

func NewPatternView(pattern *audio.Pattern, inst audio.Instrument) *PatternView {
	p := &PatternView{pattern: pattern, inst: inst, scaleTime: 32}
	p.ViewBase = NewView(p)
	noteType := audio.InstrumentPlayMethod(inst).Type().In(0)
	for i := 0; i < noteType.NumField(); i++ {
		a := newAttributeView(p, noteType.Field(i).Name)
		p.attrs = append(p.attrs, a)
		p.Add(a)
	}
	p.timeGrid = &uniformGrid{1}

	p.player = audio.NewPatternPlayer(pattern, inst)
	p.play = make(chan bool)
	p.close = make(chan bool)
	go p.animate()

	return p
}

func (p *PatternView) Close() {
	p.ViewBase.Close()
	p.close <- true
}

func (p *PatternView) animate() {
	var next <-chan time.Time
	ctrl := &PlayControl{}
	for {
		select {
		case <-p.play:
			if next != nil {
				next = nil
				ctrl.Stop()
				break
			}
			p.inst.Reset()
			ctrl = PlayAsync(p.player)
			next = time.After(time.Second / 60)
		case <-next:
			next = time.After(time.Second / 60)
			Do(p, func() {
				p.cursorTime = p.player.GetTime()
				Repaint(p)
			})
		case <-ctrl.Done:
			next = nil
			Do(p, func() {
				SetKeyFocus(p.oldFocus)
			})
		case <-p.close:
			return
		}
	}
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

func (p *PatternView) newNote() *audio.Note {
	n := &audio.Note{p.cursorTime, map[string][]*audio.ControlPoint{}}
	for _, a := range p.attrs {
		n.Attributes[a.name] = []*audio.ControlPoint{{0, a.cursorVal}}
	}
	p.pattern.Notes = append(p.pattern.Notes, n)
	return n
}

func (p *PatternView) InitFocus() { SetKeyFocus(p.attrs[0]) }

func (p *PatternView) KeyPress(k KeyEvent) {
	switch k.Key {
	case KeySpace:
		p.play <- false
		SetKeyFocus(p.oldFocus)
	}
}

func (p *PatternView) Resize(width, height float64) {
	p.transTime += (width - Width(p)) / 2
	for _, a := range p.attrs {
		for _, n := range a.notes {
			for _, p := range n.points {
				p.reform()
			}
		}
	}
	p.ViewBase.Resize(width, height)
	p.reform()
}

func (p *PatternView) save() {
	var f *os.File
	for _, dir := range build.Default.SrcDirs() {
		var err error
		f, err = os.Create(filepath.Join(dir, "code.google.com/p/gordon-go/songs", os.Args[0], p.pattern.Name+".go"))
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			fmt.Println("error saving pattern '%s':  ", err)
			return
		}
	}
	if f == nil {
		fmt.Printf("error saving pattern '%s':  unable to open file\n", p.pattern.Name)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "package main\n\nimport \"code.google.com/p/gordon-go/audio\"\n\nvar %s = &audio.Pattern{%#v, []*audio.Note{\n", p.pattern.Name, p.pattern.Name)
	for _, n := range p.pattern.Notes {
		fmt.Fprintf(f, "\t{%#v, map[string][]*audio.ControlPoint{\n", n.Time)
		for name, attr := range n.Attributes {
			fmt.Fprintf(f, "\t\t%#v: {\n", name)
			for _, p := range attr {
				fmt.Fprintf(f, "\t\t\t{%#v, %#v},\n", p.Time, p.Value)
			}
			fmt.Fprint(f, "\t\t},\n")
		}
		fmt.Fprint(f, "\t}},\n")
	}
	fmt.Fprint(f, "}}\n")
}

type attributeView struct {
	*ViewBase
	pattern   *PatternView
	name      string
	nameText  *Text
	notes     map[*audio.Note]*noteView
	transVal  float64
	scaleVal  float64
	valueGrid grid
	focused   bool
	cursorVal float64
}

func newAttributeView(p *PatternView, name string) *attributeView {
	a := &attributeView{pattern: p, name: name, scaleVal: 32}
	a.ViewBase = NewView(a)
	a.nameText = NewText(name)
	a.nameText.SetBackgroundColor(Color{})
	a.Add(a.nameText)
	a.notes = map[*audio.Note]*noteView{}
	for _, note := range p.pattern.Notes {
		n := newNoteView(a, note)
		a.notes[note] = n
		a.Add(n)
	}
	a.valueGrid = defaultGrid(name)
	a.cursorVal = a.valueGrid.defaultValue()
	a.transVal = -a.cursorVal * a.scaleVal
	return a
}

func (a *attributeView) TookKeyFocus() { a.focused = true; Repaint(a) }
func (a *attributeView) LostKeyFocus() { a.focused = false; Repaint(a) }

func (a *attributeView) KeyPress(k KeyEvent) {
	if k.Command && k.Key == KeyS {
		a.pattern.save()
		return
	}

	if k.Alt {
		switch k.Key {
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			a.focusNearest(a.to(Pt(a.pattern.cursorTime, a.cursorVal)), k.Key)
		}
		return
	}

	switch k.Key {
	case KeyLeft, KeyRight:
		a.pattern.cursorTime = math.Max(0, a.pattern.timeGrid.next(a.pattern.cursorTime, k.Key == KeyRight))
		Repaint(a.pattern)
	case KeyDown, KeyUp:
		a.cursorVal = a.valueGrid.next(a.cursorVal, k.Key == KeyUp)
		Repaint(a)
	case KeyTab:
		SetKeyFocus(a.next(k.Shift))
	case KeyEnter:
		note := a.pattern.newNote()
		var n *noteView
		for _, a2 := range a.pattern.attrs {
			n2 := newNoteView(a2, note)
			a2.notes[note] = n2
			a2.Add(n2)
			if a2 == a {
				n = n2
			}
		}
		SetKeyFocus(n)
	case KeySpace:
		a.pattern.oldFocus = a
		SetKeyFocus(a.pattern)
		a.pattern.player.SetTime(a.pattern.cursorTime)
		a.pattern.play <- true
	}
}

func (a *attributeView) next(prev bool) *attributeView {
	attrs := a.pattern.attrs
	for i, a2 := range attrs {
		if a2 == a {
			if prev {
				i--
			} else {
				i++
			}
			return attrs[(i+len(attrs))%len(attrs)]
		}
	}
	panic("unreachable")
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
	a.transVal += (height - Height(a)) / 2
	for _, n := range a.notes {
		for _, p := range n.points {
			p.reform()
		}
	}
	Repaint(a)
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
	Repaint(a)
}

func (a *attributeView) Paint() {
	r := InnerRect(a)
	min := a.from(r.Min)
	max := a.from(r.Max)
	for t := a.pattern.timeGrid.next(math.Nextafter(math.Max(0, min.X), -math.MaxFloat64), true); t < max.X; t = a.pattern.timeGrid.next(t, true) {
		SetColor(Color{.2, .2, .2, 1})
		SetLineWidth(2)
		if t == 0 {
			SetColor(Color{.3, .3, .3, 1})
			SetLineWidth(5)
		}
		DrawLine(a.to(Pt(t, min.Y)), a.to(Pt(t, max.Y)))
	}
	i := 0
	prev := -math.MaxFloat64
	for v := a.valueGrid.next(math.Nextafter(min.Y, -math.MaxFloat64), true); i < 1000 && v < max.Y && v != prev; v = a.valueGrid.next(v, true) {
		SetColor(Color{.2, .2, .2, 1})
		SetLineWidth(2)
		if v == a.valueGrid.defaultValue() {
			SetColor(Color{.3, .3, .3, 1})
			SetLineWidth(5)
		}
		DrawLine(a.to(Pt(math.Max(0, min.X), v)), a.to(Pt(max.X, v)))
		prev = v
		i++
	}
	if i == 1000 {
		println("too much")
	}

	if p, ok := KeyFocus(a).(*controlPointView); !ok || p.note.attr == a {
		SetLineWidth(3)
		SetColor(Color{.2, .2, .35, 1})
		if a.focused {
			SetColor(Color{.3, .3, .5, 1})
		}
		DrawLine(a.to(Pt(a.pattern.cursorTime, min.Y)), a.to(Pt(a.pattern.cursorTime, max.Y)))
		if n, ok := KeyFocus(a).(*noteView); !ok || n.attr == a {
			DrawLine(a.to(Pt(math.Max(0, min.X), a.cursorVal)), a.to(Pt(max.X, a.cursorVal)))
		}
	}

	SetLineWidth(1)
	SetColor(Color{1, 1, 1, 1})
	DrawLine(r.Min, Pt(r.Max.X, r.Min.Y))
}

type noteView struct {
	*ViewBase
	attr    *attributeView
	note    *audio.Note
	points  []*controlPointView
	focused bool
}

func newNoteView(attr *attributeView, note *audio.Note) *noteView {
	n := &noteView{attr: attr, note: note}
	n.ViewBase = NewView(n)
	for _, point := range note.Attributes[attr.name] {
		p := newControlPointView(n, point)
		n.points = append(n.points, p)
		n.Add(p)
		p.reform()
	}
	n.reform()
	return n
}

func (n *noteView) getpts() []*audio.ControlPoint    { return n.note.Attributes[n.attr.name] }
func (n *noteView) setpts(pts []*audio.ControlPoint) { n.note.Attributes[n.attr.name] = pts }

func (n *noteView) TookKeyFocus() {
	for _, a := range n.attr.pattern.attrs {
		n := a.notes[n.note]
		n.focused = true
		Raise(n)
		n.updateCursor()
	}
}
func (n *noteView) LostKeyFocus() {
	for _, a := range n.attr.pattern.attrs {
		n := a.notes[n.note]
		n.focused = false
		Raise(n)
	}
}

func (n *noteView) updateCursor() {
	n.attr.pattern.cursorTime = n.note.Time
	n.attr.cursorVal = n.points[0].point.Value
	Repaint(n.attr.pattern)
}

func (n *noteView) KeyPress(k KeyEvent) {
	if k.Command && k.Key == KeyS {
		n.attr.pattern.save()
		return
	}

	if k.Alt {
		switch k.Key {
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			n.attr.focusNearest(Pos(n), k.Key)
		}
		return
	}

	if k.Shift && k.Key != KeyTab {
		switch k.Key {
		case KeyLeft, KeyRight:
			n.setTime(n.attr.pattern.timeGrid.next(n.note.Time, k.Key == KeyRight))
			n.updateCursor()
		case KeyDown, KeyUp:
			p0 := n.points[0].point
			d := n.attr.valueGrid.next(p0.Value, k.Key == KeyUp) - p0.Value
			for _, p := range n.points {
				p.setValue(p.point.Value + d)
			}
			n.updateCursor()
		}
		return
	}

	switch k.Key {
	case KeyLeft, KeyRight:
		n.attr.pattern.cursorTime = n.attr.pattern.timeGrid.next(n.attr.pattern.cursorTime, k.Key == KeyRight)
		SetKeyFocus(n.attr)
	case KeyDown, KeyUp:
		n.attr.cursorVal = n.attr.valueGrid.next(n.attr.cursorVal, k.Key == KeyUp)
		SetKeyFocus(n.attr)
	case KeyTab:
		SetKeyFocus(n.attr.next(k.Shift).notes[n.note])
	case KeyEnter:
		SetKeyFocus(n.points[0])
	case KeyEscape:
		SetKeyFocus(n.attr)
	case KeyComma:
		n.newPoint(0)
	case KeyPeriod:
		n.newPoint(len(n.points))
	case KeyBackspace, KeyDelete:
		pattern := n.attr.pattern.pattern
		for i, n2 := range pattern.Notes {
			if n2 == n.note {
				pattern.Notes = append(pattern.Notes[:i], pattern.Notes[i+1:]...)
				break
			}
		}
		SetKeyFocus(n.attr)
		for _, a := range n.attr.pattern.attrs {
			a.Remove(a.notes[n.note])
			delete(a.notes, n.note)
		}
	}
}

func (n *noteView) newPoint(i int) {
	t, v := 0.0, 0.0
	switch {
	case i == 0:
		t = n.attr.pattern.timeGrid.next(n.note.Time, false) - n.note.Time
		v = n.points[0].point.Value
	default:
		p, q := n.points[i-1].point, n.points[i].point
		t = (p.Time + q.Time) / 2
		v = (p.Value + q.Value) / 2
	case i == len(n.points):
		p := n.points[len(n.points)-1].point
		t = n.attr.pattern.timeGrid.next(p.Time+n.note.Time, true) - n.note.Time
		v = p.Value
	}
	point := &audio.ControlPoint{t, v}
	n.setpts(append(n.getpts()[:i], append([]*audio.ControlPoint{point}, n.getpts()[i:]...)...))
	p := newControlPointView(n, point)
	n.points = append(n.points[:i], append([]*controlPointView{p}, n.points[i:]...)...)
	n.Add(p)
	n.normalizePoints()
	SetKeyFocus(p)
}

func (n *noteView) setTime(t float64) {
	t = math.Max(0, t)
	n.note.Time = t
	for _, a := range n.attr.pattern.attrs {
		for _, p := range a.notes[n.note].points {
			p.reform()
		}
	}
}

func (n *noteView) reform() {
	ResizeToFit(n, 0)
	n.Move(InnerRect(n).Min)

	// make space for a tail after the final control point
	dx := n.attr.to(Pt(n.note.Time+n.duration(), 0)).X - InnerRect(n).Max.X
	if dx > 0 {
		width, height := Size(n)
		width += dx
		n.Resize(width, height)
	}
}

func (n *noteView) duration() float64 {
	t := 0.0
	for _, a := range n.attr.pattern.attrs {
		n := a.notes[n.note]
		if n == nil {
			// may be nil during newNoteView
			continue
		}
		t = math.Max(t, n.points[len(n.points)-1].point.Time)
	}
	return t
}

func (n *noteView) normalizePoints() {
	t := n.points[0].point.Time
	for _, p := range n.points {
		p.point.Time -= t
		p.reform()
	}
	n.setTime(n.note.Time + t)
}

func (n *noteView) Paint() {
	SetLineWidth(2)
	SetColor(Color{1, 1, 1, 1})
	if n.focused {
		SetLineWidth(3)
		SetColor(Color{.6, .6, .9, 1})
		if KeyFocus(n) == n {
			SetColor(Color{.4, .4, .9, 1})
		}
	}
	for i, p := range n.points[1:] {
		DrawLine(Center(n.points[i]), Center(p))
	}

	// draw a tail after the final control point
	p := n.points[len(n.points)-1]
	DrawLine(Center(p), n.attr.to(Pt(n.note.Time+n.duration(), p.point.Value)))
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

func (p *controlPointView) TookKeyFocus() { p.focused = true; Raise(p); p.updateCursor() }
func (p *controlPointView) LostKeyFocus() { p.focused = false; Raise(p) }

func (p *controlPointView) updateCursor() {
	p.note.attr.pattern.cursorTime = p.note.note.Time + p.point.Time
	p.note.attr.cursorVal = p.point.Value
	Repaint(p.note.attr.pattern)
}

func (p *controlPointView) KeyPress(k KeyEvent) {
	if k.Command && k.Key == KeyS {
		p.note.attr.pattern.save()
		return
	}

	if k.Shift && k.Key != KeyTab {
		switch k.Key {
		case KeyLeft, KeyRight:
			p.setTime(p.note.attr.pattern.timeGrid.next(p.point.Time+p.note.note.Time, k.Key == KeyRight) - p.note.note.Time)
			p.updateCursor()
		case KeyDown, KeyUp:
			p.setValue(p.note.attr.valueGrid.next(p.point.Value, k.Key == KeyUp))
			p.updateCursor()
		}
		return
	}

	switch k.Key {
	case KeyLeft, KeyRight:
		p.focusNext(k.Key == KeyRight)
	case KeyTab:
		SetKeyFocus(p.note.attr.next(k.Shift).notes[p.note.note].points[0])
	case KeyEscape:
		SetKeyFocus(p.note)
	case KeyComma:
		p.note.newPoint(p.index())
	case KeyPeriod:
		p.note.newPoint(p.index() + 1)
	case KeyBackspace, KeyDelete:
		if len(p.note.points) == 1 {
			break
		}
		i := p.index()
		p.focusNext((i == 0 || k.Key == KeyDelete) && i < len(p.note.points)-1)
		p.note.setpts(append(p.note.getpts()[:i], p.note.getpts()[i+1:]...))
		p.note.points = append(p.note.points[:i], p.note.points[i+1:]...)
		p.note.Remove(p)
		p.note.normalizePoints()
	}
}

func (p *controlPointView) setTime(t float64) {
	i := p.index()
	points := p.note.points
	if i == 0 {
		t = math.Max(-p.note.note.Time, t)
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
	MoveOrigin(p, p.note.attr.to(Pt(p.note.note.Time+p.point.Time, p.point.Value)))
	p.note.reform()
}

func (p *controlPointView) focusNext(next bool) {
	i := p.index()
	if next {
		i++
	} else {
		i--
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
	SetColor(Color{1, 1, 1, 1})
	if p.note.focused {
		SetColor(Color{.6, .6, .9, 1})
		if KeyFocus(p) == p.note {
			SetPointSize(7)
			SetColor(Color{.4, .4, .9, 1})
		}
	}
	if p.focused {
		SetPointSize(10)
		SetColor(Color{.4, .4, .9, 1})
	}
	DrawPoint(ZP)
}
