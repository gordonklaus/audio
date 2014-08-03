package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
	"code.google.com/p/gordon-go/gui"
	"math/rand"
	"os"
)

func main() {
	p1 := []*audio.ControlPoint{
		{0, 1},
		{1, 0},
	}
	amp1 := []*audio.ControlPoint{
		{0, 0},
		{.25, 1},
		{1, 0},
	}
	p2 := []*audio.ControlPoint{
		{0, 3},
		{2, 4},
	}
	amp2 := []*audio.ControlPoint{
		{0, 0},
		{.25, 1},
		{1, 0},
	}
	p := audio.NewPattern([]audio.Note{&note{audio.NewNote(.25), p1, amp1}, &note{audio.NewNote(4), p2, amp2}}, inst)
	if len(os.Args) > 1 && os.Args[1] == "edit" {
		gui.Run(func() {
			gui.NewWindow(nil, "song1", func(w *gui.Window) {
				pv := audiogui.NewPatternView(p)
				w.SetCentralView(pv)
				gui.SetKeyFocus(pv)
			})
		})
	} else {
		audiogui.Play(p)
	}
}

var inst = &instrument{}

type instrument struct { audio.MultiVoice }
func (i *instrument) Play(n *note) {
	i.Add(&voice{Amp: audio.NewControl(n.Amplitude)})
}

type note struct {
	audio.Note
	Pitch     []*audio.ControlPoint
	Amplitude []*audio.ControlPoint
}

type voice struct {
	Amp *audio.Control
	Out audio.Audio
}

func (v *voice) Sing() (a audio.Audio, done bool) {
	for i := range v.Out {
		v.Out[i] = rand.Float64()
	}
	a, done = v.Amp.Sing()
	a = a.Mul(a, v.Out)
	return
}
