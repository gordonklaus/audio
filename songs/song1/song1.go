package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
	"code.google.com/p/gordon-go/gui"

	"math"
	"math/rand"
	"os"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "edit" {
		gui.Run(func() {
			gui.NewWindow(nil, "song1", func(w *gui.Window) {
				v := audiogui.NewPatternView(melody, inst)
				w.SetCentralView(v)
				v.InitFocus()
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
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
		b:        -math.Log(.001) / 2 / releaseTime,
		AmpMeter: audio.NewAmpMeter(.05),
	})
}

type voice struct {
	Pitch, Amp        *audio.Control
	rand              *rand.Rand
	u, v, b, dt, sqdt float64
	AmpMeter          *audio.AmpMeter
}

func (v *voice) InitAudio(p audio.Params) {
	audio.Init(v.Pitch, p)
	audio.Init(v.Amp, p)
	v.dt = 1 / p.SampleRate
	v.sqdt = math.Sqrt(v.dt)
	audio.Init(v.AmpMeter, p)
}

func (v *voice) Sing() float64 {
	p := math.Exp2(v.Pitch.Sing())
	r := math.Exp2(v.Amp.Sing()) * v.rand.NormFloat64()
	if v.Pitch.Done() && v.Amp.Done() {
		r = 0
	}
	c := v.b*v.b/4 + 4*math.Pi*math.Pi*p*p
	v.u += v.dt*v.v + v.sqdt*r
	v.v -= v.dt * (v.b*v.v + c*v.u)
	v.AmpMeter.Add(v.u)
	return v.u
}

func (v *voice) Done() bool {
	return v.Pitch.Done() && v.Amp.Done() && v.AmpMeter.Amplitude() < .00001
}
