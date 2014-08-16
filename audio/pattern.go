package audio

import (
	"fmt"
	"reflect"
	"sort"
)

type Pattern struct {
	Name  string
	Notes []*Note
}

type Note struct {
	Time       float64
	Attributes map[string][]*ControlPoint
}

type PatternPlayer struct {
	pattern *Pattern
	inst    Instrument
	i       int
	t, dt   float64
	play    reflect.Value
}

func NewPatternPlayer(pattern *Pattern, inst Instrument) *PatternPlayer {
	sort.Sort(byTime(pattern.Notes))
	return &PatternPlayer{pattern: pattern, inst: inst, play: InstrumentPlayMethod(inst)}
}

type byTime []*Note

func (n byTime) Len() int           { return len(n) }
func (n byTime) Less(i, j int) bool { return n[i].Time < n[j].Time }
func (n byTime) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

func (p *PatternPlayer) GetTime() float64 { return p.t }
func (p *PatternPlayer) SetTime(t float64) {
	p.t = t
	for p.i = 0; p.i < len(p.pattern.Notes) && p.pattern.Notes[p.i].Time < t; p.i++ {
	}
}

func (p *PatternPlayer) InitAudio(params Params) {
	Init(p.inst, params)
	p.dt = 1 / params.SampleRate
}

func (p *PatternPlayer) Sing() float64 {
	for ; p.i < len(p.pattern.Notes); p.i++ {
		n := p.pattern.Notes[p.i]
		if n.Time > p.t {
			break
		}
		p.play.Call([]reflect.Value{p.newNote(n)})
	}
	p.t += p.dt
	return p.inst.Sing()
}

func (p *PatternPlayer) Done() bool {
	return p.i == len(p.pattern.Notes) && p.inst.Done()
}

func (p *PatternPlayer) newNote(note *Note) reflect.Value {
	n := reflect.New(p.play.Type().In(0)).Elem()
	for name, val := range note.Attributes {
		f := n.FieldByName(name)
		if !f.IsValid() {
			fmt.Printf("audio.PatternPlayer: invalid note attribute '%s' in pattern '%s'\n", name, p.pattern.Name)
			continue
		}
		f.Set(reflect.ValueOf(val))
	}
	return n
}

// An Instrument must also have a method Play(noteType) where noteType is a struct with exported fields of type []*ControlPoint.
type Instrument interface {
	Voice
	Reset()
}

func InstrumentPlayMethod(inst Instrument) reflect.Value {
	m := reflect.ValueOf(inst).MethodByName("Play")
	if !m.IsValid() {
		panic(fmt.Sprintf("Type %T must have a method named Play.", inst))
	}
	if m.Type().NumIn() != 1 {
		panic(fmt.Sprintf("Method (%T).Play must have a single parameter.", inst))
	}
	n := m.Type().In(0)
	if n.Kind() != reflect.Struct {
		panic(fmt.Sprintf("The parameter to method (%T).Play must be a struct.", inst))
	}
	t := reflect.TypeOf([]*ControlPoint(nil))
	for i := 0; i < n.NumField(); i++ {
		f := n.Field(i)
		if f.Type != t || f.PkgPath != "" {
			panic(fmt.Sprintf("The parameter to method (%T).Play must only have exported fields of type %s.", inst, t))
		}
	}
	return m
}
