package audio

import (
	"fmt"
	"reflect"
	"sort"
)

type Pattern struct {
	Params     Params
	Notes      []*Note
	i          int
	t          float64
	Instrument Instrument
	play       reflect.Value
}

func NewPattern(notes []*Note, i Instrument) *Pattern {
	return &Pattern{Notes: notes, Instrument: i, play: InstrumentPlayMethod(i)}
}

func (p *Pattern) Sort() { sort.Sort(byTime(p.Notes)) }

type byTime []*Note

func (n byTime) Len() int           { return len(n) }
func (n byTime) Less(i, j int) bool { return n[i].Time < n[j].Time }
func (n byTime) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

func (p *Pattern) Sing() (a Audio, done bool) {
	p.t += float64(p.Params.BufferSize) / p.Params.SampleRate
	for {
		if p.i >= len(p.Notes) {
			done = true
			break
		}
		n := p.Notes[p.i]
		if n.Time > p.t {
			break
		}
		p.play.Call([]reflect.Value{p.newNote(n)})
		p.i++
	}
	a, done2 := p.Instrument.Sing()
	return a, done && done2
}

func (p *Pattern) newNote(note *Note) reflect.Value {
	n := reflect.New(p.play.Type().In(0)).Elem()
	for name, val := range note.Attributes {
		f := n.FieldByName(name)
		if !f.IsValid() {
			fmt.Printf("audio.Pattern: invalid note attribute '%s'\n", name)
			continue
		}
		f.Set(reflect.ValueOf(val))
	}
	return n
}

func (p *Pattern) Reset() {
	p.i, p.t = 0, 0
}

type Note struct {
	Time       float64
	Attributes map[string][]*ControlPoint
}

// An Instrument must also have a method Play(noteType) where noteType is a struct with exported fields of type []*ControlPoint.
type Instrument interface {
	Voice
}

func InstrumentPlayMethod(i Instrument) reflect.Value {
	m := reflect.ValueOf(i).MethodByName("Play")
	if !m.IsValid() {
		panic(fmt.Sprintf("Type %T must have a method named Play.", i))
	}
	if m.Type().NumIn() != 1 {
		panic(fmt.Sprintf("Method (%T).Play must have a single parameter.", i))
	}
	n := m.Type().In(0)
	if n.Kind() != reflect.Struct {
		panic(fmt.Sprintf("The parameter to method (%T).Play must be a struct.", i))
	}
	t := reflect.TypeOf([]*ControlPoint(nil))
	for j := 0; j < n.NumField(); j++ {
		f := n.Field(j)
		if f.Type != t || f.PkgPath != "" {
			panic(fmt.Sprintf("The parameter to method (%T).Play must only have exported fields of type %s.", i, t))
		}
	}
	return m
}
