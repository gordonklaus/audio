package audiogui

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"code.google.com/p/gordon-go/audio"
	. "code.google.com/p/gordon-go/gui"
)

type ScoreView struct {
	*ViewBase
	score      *audio.Score
	band       audio.Band
	parts      []*partView
	path       string
	transTime  float64
	scaleTime  float64
	timeGrid   *uniformGrid
	cursorTime float64

	player      *audio.ScorePlayer
	play, close chan bool
	oldFocus    View

	pattern *PatternView
}

func NewScoreView(score *audio.Score, band audio.Band) *ScoreView {
	s := &ScoreView{score: score, band: band, scaleTime: 32}
	s.ViewBase = NewView(s)
	instruments := audio.BandInstruments(band)
loop:
	for name := range instruments {
		for _, part := range score.Parts {
			if part.Name == name {
				continue loop
			}
		}
		score.Parts = append(score.Parts, &audio.Part{Name: name})
		fmt.Println("added part for instrument " + name)
	}
	// TODO: warn about parts without instruments
	for _, part := range score.Parts {
		p := newPartView(s, part)
		s.parts = append(s.parts, p)
		s.Add(p)
	}
	s.timeGrid = &uniformGrid{0, 1}

	s.player = audio.NewScorePlayer(score, band)
	s.play = make(chan bool)
	s.close = make(chan bool)
	go s.animate()

	return s
}

func (s *ScoreView) InitFocus() { SetKeyFocus(s.parts[0]) }

func (s *ScoreView) Close() {
	s.ViewBase.Close()
	s.close <- true
}

func (s *ScoreView) animate() {
	var next <-chan time.Time
	ctrl := &PlayControl{}
	for {
		select {
		case <-s.play:
			if next != nil {
				next = nil
				ctrl.Stop()
				break
			}
			for _, inst := range audio.BandInstruments(s.band) {
				inst.Stop()
			}
			ctrl = PlayAsync(s.player)
			next = time.After(time.Second / 60)
		case <-next:
			next = time.After(time.Second / 60)
			Do(s, func() {
				s.cursorTime = s.player.GetTime()
				Repaint(s)
			})
		case <-ctrl.Done:
			next = nil
			Do(s, func() {
				SetKeyFocus(s.oldFocus)
			})
		case <-s.close:
			return
		}
	}
}

func (s *ScoreView) reform() {
	w, h := Size(s)
	for _, p := range s.parts {
		ph := Height(p)
		h -= ph
		p.Resize(w, ph)
		p.Move(Pt(0, h))
		p.name.Move(Pt(w-Width(p.name), ph-Height(p.name)))
		for _, e := range p.events {
			e.reform()
		}
	}
	if s.pattern != nil {
		s.pattern.Resize(w, h)
	}
}

func (s *ScoreView) save() {
	f, err := os.Create(s.path)
	if err != nil {
		fmt.Println("error saving score:", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "package main\n\nimport %q\n\nvar score = &audio.Score{[]*audio.Part{\n", audioPkgPath)
	for _, p := range s.score.Parts {
		fmt.Fprintf(f, "\t{%q, []*audio.PatternEvent{\n", p.Name)
		for _, e := range p.Events {
			fmt.Fprintf(f, "\t\t{%v, %s_pattern},\n", e.Time, e.Pattern.Name)
		}
		fmt.Fprint(f, "\t}},\n")
	}
	fmt.Fprint(f, "}}\n")
}

func (s *ScoreView) editPattern(e *patternEventView) {
	p := NewPatternView(e.event.Pattern, audio.BandInstruments(s.band)[e.part.part.Name])
	p.closed = func() {
		s.pattern = nil
		s.reform()
		SetKeyFocus(e)
	}
	s.pattern = p
	s.Add(p)
	s.reform()
	p.InitFocus()
}

func (s *ScoreView) Resize(width, height float64) {
	s.transTime += (width - Width(s)) / 2
	s.ViewBase.Resize(width, height)
	s.reform()
}

func (s *ScoreView) KeyPress(k KeyEvent) {
	switch k.Key {
	case KeySpace:
		s.play <- false
		SetKeyFocus(s.oldFocus)
	}
}

func (s *ScoreView) Scroll(e ScrollEvent) {
	if e.Shift {
		prevScale := s.scaleTime
		s.scaleTime = math.Max(10, math.Min(1000, s.scaleTime*math.Pow(1.05, -e.Delta.X)))
		s.transTime = e.Pos.X + (s.transTime-e.Pos.X)*s.scaleTime/prevScale
	} else {
		s.transTime += 8 * e.Delta.X
	}
	for _, p := range s.parts {
		for _, e := range p.events {
			e.reform()
		}
	}
	Repaint(s)
}

type partView struct {
	*ViewBase
	score   *ScoreView
	part    *audio.Part
	name    *Text
	events  []*patternEventView
	focused bool
}

func newPartView(score *ScoreView, part *audio.Part) *partView {
	p := &partView{score: score, part: part}
	p.ViewBase = NewView(p)
	p.name = NewText(part.Name)
	p.name.SetBackgroundColor(Color{})
	p.Add(p.name)
	for _, event := range part.Events {
		e := newPatternEventView(p, event)
		p.events = append(p.events, e)
		p.Add(e)
	}
	p.Resize(0, 32)
	return p
}

func (p *partView) to(t float64) float64 {
	return t*p.score.scaleTime + p.score.transTime
}
func (p *partView) from(t float64) float64 {
	return (t - p.score.transTime) / p.score.scaleTime
}

func (p *partView) TookKeyFocus() { p.focused = true; Repaint(p) }
func (p *partView) LostKeyFocus() { p.focused = false; Repaint(p) }

func (p *partView) KeyPress(k KeyEvent) {
	if k.Command && k.Key == KeyS {
		p.score.save()
		return
	}

	if k.Alt {
		switch k.Key {
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			p.focusNearest(MapToParent(Pt(p.to(p.score.cursorTime), Height(p)/2), p), k.Key)
		}
		return
	}

	switch k.Key {
	case KeyLeft, KeyRight:
		p.score.cursorTime = math.Max(0, p.score.timeGrid.next(p.score.cursorTime, k.Key == KeyRight))
		Repaint(p.score)
	case KeyDown, KeyUp:
		SetKeyFocus(p.next(k.Key == KeyUp))
	case KeyEnter:
		event := &audio.PatternEvent{p.score.cursorTime, &audio.Pattern{Attributes: map[string][]*audio.ControlPoint{}}}
		e := newPatternEventView(p, event)
		p.Add(e)
		e.name.accepted = func(name string) {
			p.events = append(p.events, e)
			p.part.Events = append(p.part.Events, event)
			if info, ok := Patterns[name]; ok {
				event.Pattern = info.p
			} else {
				event.Pattern.Name = name
				Patterns[name] = patternInfo{event.Pattern, filepath.Join(filepath.Dir(p.score.path), name) + "_pattern.go"}
				savePattern(event.Pattern)
			}
			e.reform()
			SetKeyFocus(e)
		}
		e.name.canceled = func() {
			p.Remove(e)
			SetKeyFocus(p)
		}
		SetKeyFocus(e.name)
	case KeySpace:
		p.score.oldFocus = p
		SetKeyFocus(p.score)
		p.score.player.SetTime(p.score.cursorTime)
		p.score.play <- true
	}
}

func (p *partView) focusNearest(pt Point, dirKey int) {
	views := []View{}
	for _, p := range p.score.parts {
		for _, e := range p.events {
			views = append(views, e)
		}
	}
	v := nearestView(p.score, views, pt, dirKey)
	if v != nil {
		SetKeyFocus(v)
	}
}

func nearestView(parent View, views []View, p Point, dirKey int) (nearest View) {
	dir := map[int]Point{KeyLeft: {-1, 0}, KeyRight: {1, 0}, KeyUp: {0, 1}, KeyDown: {0, -1}}[dirKey]
	best := 0.0
	for _, v := range views {
		d := Map(Center(v), Parent(v), parent).Sub(p)
		score := (dir.X*d.X + dir.Y*d.Y) / (d.X*d.X + d.Y*d.Y)
		if score > best {
			best = score
			nearest = v
		}
	}
	return
}

func (p *partView) next(next bool) *partView {
	parts := p.score.parts
	for i, p2 := range parts {
		if p2 == p {
			if next {
				i++
			} else {
				i--
			}
			return parts[(i+len(parts))%len(parts)]
		}
	}
	panic("unreachable")
}

func (p *partView) Paint() {
	r := InnerRect(p)
	for t := p.score.timeGrid.next(math.Nextafter(math.Max(0, p.from(r.Min.X)), -math.MaxFloat64), true); t < p.from(r.Max.X); t = p.score.timeGrid.next(t, true) {
		SetColor(Color{.2, .2, .2, 1})
		SetLineWidth(2)
		if t == 0 {
			SetColor(Color{.3, .3, .3, 1})
			SetLineWidth(5)
		}
		DrawLine(Pt(p.to(t), r.Min.Y), Pt(p.to(t), r.Max.Y))
	}

	SetLineWidth(3)
	SetColor(Color{.2, .2, .35, 1})
	if p.focused {
		SetColor(Color{.3, .3, .5, 1})
	}
	DrawLine(Pt(p.to(p.score.cursorTime), r.Min.Y), Pt(p.to(p.score.cursorTime), r.Max.Y))

	SetColor(Color{1, 1, 1, 1})
	SetLineWidth(1)
	DrawLine(ZP, Pt(Width(p), 0))
}

type patternEventView struct {
	*ViewBase
	part    *partView
	event   *audio.PatternEvent
	name    *patternName
	focused bool
}

func newPatternEventView(part *partView, event *audio.PatternEvent) *patternEventView {
	e := &patternEventView{part: part, event: event}
	e.ViewBase = NewView(e)
	e.name = newPatternName(e, event.Pattern.Name)
	e.name.Move(Pt(3, 3))
	e.Add(e.name)
	e.reform()
	return e
}

func (e *patternEventView) reform() {
	t := 0.0
	for _, n := range e.event.Pattern.Notes {
		for _, a := range n.Attributes {
			t = math.Max(t, a[len(a)-1].Time)
		}
	}
	for _, a := range e.event.Pattern.Attributes {
		t = math.Max(t, a[len(a)-1].Time)
	}
	max := OuterRect(e.name).Max
	e.Resize(math.Max(max.X+3, t*e.part.score.scaleTime), max.Y)
	e.Move(Pt(e.part.to(e.event.Time), 4))
}

func (e *patternEventView) TookKeyFocus() {
	e.focused = true
	e.part.score.cursorTime = e.event.Time
	Raise(e)
}
func (e *patternEventView) LostKeyFocus() { e.focused = false; Repaint(e) }

func (e *patternEventView) KeyPress(k KeyEvent) {
	if k.Command && k.Key == KeyS {
		e.part.score.save()
		return
	}

	if k.Alt {
		switch k.Key {
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			e.part.focusNearest(MapToParent(Center(e), e.part), k.Key)
		}
		return
	}

	if k.Shift {
		switch k.Key {
		case KeyLeft, KeyRight:
			e.event.Time = e.part.score.timeGrid.next(e.event.Time, k.Key == KeyRight)
			e.reform()
		}
		return
	}

	switch k.Key {
	case KeyLeft, KeyRight:
		e.part.score.cursorTime = e.part.score.timeGrid.next(e.part.score.cursorTime, k.Key == KeyRight)
		SetKeyFocus(e.part)
	case KeyDown, KeyUp:
		SetKeyFocus(e.part.next(k.Key == KeyUp))
	case KeyEnter:
		e.part.score.editPattern(e)
	case KeyEscape:
		SetKeyFocus(e.part)
	case KeyBackspace, KeyDelete:
		part := e.part.part
		for i, e2 := range part.Events {
			if e2 == e.event {
				part.Events = append(part.Events[:i], part.Events[i+1:]...)
				break
			}
		}
		for i, e2 := range e.part.events {
			if e2 == e {
				e.part.events = append(e.part.events[:i], e.part.events[i+1:]...)
				break
			}
		}
		e.part.Remove(e)
		SetKeyFocus(e.part)
	}
}

func (e *patternEventView) Paint() {
	SetLineWidth(2)
	SetColor(Color{1, 1, 1, 1})
	if e.focused {
		SetLineWidth(3)
		SetColor(Color{.4, .4, .9, 1})
	}
	r := InnerRect(e)
	x0, x1 := r.Min.X, r.Max.X
	y0, y1, y2 := r.Min.Y, r.Min.Y+4, r.Min.Y+8
	DrawLine(Pt(x0, y0), Pt(x0, y2))
	DrawLine(Pt(x0, y1), Pt(x1, y1))
	DrawLine(Pt(x1, y0), Pt(x1, y2))
}

type patternName struct {
	*ViewBase
	e        *patternEventView
	text     *Text
	names    []*Text
	i        int
	accepted func(string)
	canceled func()
}

func newPatternName(e *patternEventView, name string) *patternName {
	n := &patternName{e: e, i: -1}
	n.ViewBase = NewView(n)
	n.ViewBase.NoClip = true
	n.text = NewText("")
	n.text.SetBackgroundColor(Color{})
	n.text.TextChanged = func(text string) {
		n.Resize(Size(n.text))
		n.e.Resize(Width(n)+6, Height(n))
	}
	n.Add(n.text)
	n.text.SetText(name)
	return n
}

func (n *patternName) TookKeyFocus() {
	n.text.ShowCursor()
	n.addNames()
}

func (n *patternName) LostKeyFocus() {
	n.text.HideCursor()
	n.removeNames()
}

func (n *patternName) addNames() {
	names := []string{}
	for name := range Patterns {
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(n.text.Text())) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	y := 0.0
	for i := range names {
		name := NewText(names[i])
		if names[i] == n.text.Text() {
			name.SetBackgroundColor(Color{0, .1, .4, 1})
			n.i = i
		}
		y -= Height(name)
		name.Move(Pt(0, y))
		n.names = append(n.names, name)
		n.Add(name)
	}
}

func (n *patternName) removeNames() {
	for _, name := range n.names {
		n.Remove(name)
	}
	n.names = nil
}

func (n *patternName) KeyPress(k KeyEvent) {
	switch k.Key {
	case KeyDown, KeyUp:
		len := len(n.names)
		if len == 0 {
			break
		}
		if n.i >= 0 {
			n.names[n.i].SetBackgroundColor(Color{})
		} else if k.Key == KeyUp {
			n.i = 0
		}
		if k.Key == KeyDown {
			n.i = (n.i + 1) % len
		} else {
			n.i = (n.i - 1 + len) % len
		}
		n.names[n.i].SetBackgroundColor(Color{0, .1, .4, 1})
		n.text.SetText(n.names[n.i].Text())
	case KeyEnter:
		n.accepted(n.text.Text())
	case KeyEscape:
		n.canceled()
	default:
		if k.Text != "" || k.Key == KeyBackspace {
			n.i = -1
			n.text.KeyPress(k)
			n.removeNames()
			n.addNames()
		}
	}
}

func (n *patternName) Paint() {
	r := ZR
	for _, name := range n.names {
		r = r.Union(OuterRect(name))
	}
	if r != ZR {
		r = r.Inset(-1)
		SetColor(Color{0, 0, 0, 1})
		FillRect(r)
		SetLineWidth(1)
		SetColor(Color{1, 1, 1, 1})
		DrawRect(r)
	}
}

var Patterns = map[string]patternInfo{}

type patternInfo struct {
	p    *audio.Pattern
	path string
}

func NewPattern(notes []*audio.Note, attributes map[string][]*audio.ControlPoint) *audio.Pattern {
	_, path, _, _ := runtime.Caller(1)
	name := strings.TrimSuffix(filepath.Base(path), "_pattern.go")
	p := &audio.Pattern{name, notes, attributes}
	Patterns[name] = patternInfo{p, path}
	return p
}
