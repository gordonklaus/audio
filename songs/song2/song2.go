package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"

	"math"
)

func main() {
	audiogui.Main(score, &band{})
}

type band struct {
	Sines sines
}

func (b *band) Sing() float64 {
	return b.Sines.Sing()
}

func (b *band) Done() bool {
	return b.Sines.Done()
}

type sines struct {
	audio.MultiVoice
}

func (s *sines) Play(n struct{ Pitch, Amplitude []*audio.ControlPoint }) {
	s.Add(&sineVoice{
		Pitch: audio.NewControl(n.Pitch),
		Amp:   audio.NewControl(n.Amplitude),
		Env:   audio.NewAttackReleaseEnv(.05, 4),
	})
}

type sineVoice struct {
	Pitch, Amp *audio.Control
	Env        *audio.AttackReleaseEnv
	Sine       audio.SineOsc
}

func (v *sineVoice) Sing() float64 {
	f := math.Exp2(v.Pitch.Sing())
	g := math.Exp2(v.Amp.Sing()) * v.Env.Sing()
	if v.Pitch.Done() && v.Amp.Done() {
		v.Env.Release()
	}
	return g * math.Tanh(2*v.Sine.Sine(f))
}

func (v *sineVoice) Done() bool {
	return v.Pitch.Done() && v.Amp.Done() && v.Env.Done()
}
