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
				v := audiogui.NewScoreView(score, newBand())
				w.SetCentralView(v)
				v.InitFocus()
			})
		})
	} else {
		audiogui.Play(audio.NewScorePlayer(score, newBand()))
	}
}

type band struct {
	Sines      sines
	Distortion distortion
}

func newBand() *band {
	return &band{}
}

func (b *band) Sing() float64 {
	return b.Distortion.distort(b.Sines.Sing())
}

func (b *band) Done() bool {
	return b.Sines.Done()
}

type sines struct{ audio.MultiVoice }

func (s *sines) Play(n struct{ Pitch, Amplitude []*audio.ControlPoint }) {
	s.Add(&sineVoice{
		Pitch: audio.NewControl(n.Pitch),
		Amp:   audio.NewControl(n.Amplitude),
	})
}

type sineVoice struct {
	Pitch, Amp *audio.Control
	Sine       audio.SineOsc
}

func (v *sineVoice) Sing() float64 {
	return v.Sine.Sine(math.Exp2(v.Pitch.Sing())) * math.Exp2(v.Amp.Sing())
}

func (v *sineVoice) Done() bool {
	return v.Pitch.Done() && v.Amp.Done()
}

type distortion struct{ Gain audio.Control }

func (d *distortion) Play(n struct{ Gain []*audio.ControlPoint }) {
	d.Gain.SetPoints(n.Gain)
}

func (d *distortion) Stop() {}

func (d *distortion) distort(x float64) float64 {
	return math.Tanh(x * d.Gain.Sing())
}

func (d *distortion) Sing() float64 { return 0 }
func (d *distortion) Done() bool    { return true }

type noiseForcedSines struct{ audio.MultiVoice }

func (s *noiseForcedSines) Play(n struct{ Pitch, Amplitude []*audio.ControlPoint }) {
	releaseTime := 1.0
	s.Add(&voice{
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
