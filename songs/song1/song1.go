package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
	"code.google.com/p/gordon-go/gui"

	"math"
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
	releaseTime := 1.0
	i.Add(&voice{
		Pitch:    audio.NewControl(n.Pitch),
		Amp:      audio.NewControl(n.Amplitude),
		Rand:     audio.NewRand(),
		b:        -math.Log(.001) / 2 / releaseTime,
		AmpMeter: audio.NewAmpMeter(.05),
	})
}

type voice struct {
	Pitch, Amp        *audio.Control
	Rand              *audio.Rand
	u, v, b, dt, sqdt float64
	AmpMeter          *audio.AmpMeter
}

func (v *voice) InitAudio(p audio.Params) {
	audio.Init(v.Pitch, p)
	audio.Init(v.Amp, p)
	audio.Init(v.Rand, p)
	v.dt = 1 / p.SampleRate
	v.sqdt = math.Sqrt(v.dt)
	audio.Init(v.AmpMeter, p)
}

func (v *voice) Sing() (audio.Audio, bool) {
	p, done1 := v.Pitch.Sing()
	a, done2 := v.Amp.Sing()
	r := v.Rand.NormRand()
	r.Amplify(r, a)
	if done1 && done2 {
		r.Zero()
	}
	p.Exp2(p)
	for i := range a {
		c := v.b*v.b/4 + 4*math.Pi*math.Pi*p[i]*p[i]
		v.u += v.dt*v.v + v.sqdt*r[i]
		v.v -= v.dt * (v.b*v.v + c*v.u)
		a[i] = v.u
	}
	amp := v.AmpMeter.Amplitude(a)
	return a, done1 && done2 && amp < .00001
}
