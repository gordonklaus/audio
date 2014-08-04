package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
	"code.google.com/p/gordon-go/gui"
	"math/rand"
	"os"
)

func main() {
	p := audio.NewPattern("melody", melody, inst)
	if len(os.Args) > 1 && os.Args[1] == "edit" {
		gui.Run(func() {
			gui.NewWindow(nil, "song1", func(w *gui.Window) {
				v := audiogui.NewPatternView(p)
				w.SetCentralView(v)
				gui.SetKeyFocus(v)
			})
		})
	} else {
		audiogui.Play(p)
	}
}

var inst = &instrument{}

type instrument struct { audio.MultiVoice }
func (i *instrument) Play(n struct{ Pitch, Amplitude []*audio.ControlPoint }) {
	i.Add(&voice{Amp: audio.NewControl(n.Amplitude)})
}

type voice struct {
	Amp *audio.Control
	Out audio.Audio
}

func (v *voice) Sing() (audio.Audio, bool) {
	for i := range v.Out {
		v.Out[i] = rand.Float64()
	}
	a, done := v.Amp.Sing()
	a.Mul(a, v.Out)
	return a, done
}
