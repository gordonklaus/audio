package audio

import (
	"fmt"
	"reflect"
	"sort"
)

type Pattern struct {
	Params     Params
	Notes      []Note
	i          int
	t          float64
	Instrument Instrument
	play       reflect.Value
}

func NewPattern(notes []Note, i Instrument) *Pattern {
	return &Pattern{Notes: notes, Instrument: i, play: InstrumentPlayMethod(i)}
}

func (p *Pattern) Sort() { sort.Sort(byTime(p.Notes)) }

type byTime []Note

func (n byTime) Len() int           { return len(n) }
func (n byTime) Less(i, j int) bool { return n[i].Time() < n[j].Time() }
func (n byTime) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

func (p *Pattern) Sing() (a Audio, done bool) {
	p.t += float64(p.Params.BufferSize) / p.Params.SampleRate
	for {
		if p.i >= len(p.Notes) {
			done = true
			break
		}
		n := p.Notes[p.i]
		if n.Time() > p.t {
			break
		}
		p.play.Call([]reflect.Value{reflect.ValueOf(n)})
		p.i++
	}
	a, done2 := p.Instrument.Sing()
	return a, done && done2
}

func (p *Pattern) Reset() {
	p.i, p.t = 0, 0
}

type Note interface {
	Time() float64
	SetTime(float64)
}

type note struct{ time float64 }

func (n *note) Time() float64     { return n.time }
func (n *note) SetTime(t float64) { n.time = t }

func NewNote(time float64) Note {
	return &note{time}
}

// An Instrument must also have a method Play(NoteType) where NoteType implements the Note interface.
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
	noteInterface := reflect.TypeOf(new(Note)).Elem()
	noteType := m.Type().In(0)
	if !noteType.Implements(noteInterface) {
		panic(fmt.Sprintf("Parameter type %s to method (%T).Play does not implement %s.", noteType, i, noteInterface))
	}
	return m
}
