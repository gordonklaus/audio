package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
	"code.google.com/p/gordon-go/gui"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "edit" {
		gui.Run(func() {
			gui.NewWindow(nil, "song1", func(w *gui.Window) {
				v := audiogui.NewPatternView(melody, inst)
				w.SetCentralView(v)
				gui.SetKeyFocus(v)
			})
		})
	} else {
		audiogui.Play(audio.NewPatternPlayer(melody, inst))
	}
}

var inst = &instrument{}

type instrument struct{ audio.MultiVoice }

func (i *instrument) Play(n struct{ Pitch, Amplitude []*audio.ControlPoint }) {
	i.Add(&voice{Pitch: audio.NewControl(n.Pitch), Amp: audio.NewControl(n.Amplitude)})
}

type voice struct {
	Sine       audio.SineOsc
	Pitch, Amp *audio.Control
}

func (v *voice) Sing() (audio.Audio, bool) {
	p, done1 := v.Pitch.Sing()
	a, done2 := v.Amp.Sing()
	a.Mul(a, v.Sine.Sine(p.Pow2(p.AddX(p, 8))))
	return a, done1 && done2
}
