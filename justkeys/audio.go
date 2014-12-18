package main

import (
	"math"

	"golang.org/x/mobile/sl"

	"code.google.com/p/gordon-go/audio"
)

var multivoice audio.MultiVoice

func startAudio() {
	audio.Init(&multivoice, audio.Params{48000})
	sl.Start(func(out []float32) {
		for i := range out {
			out[i] = float32(math.Tanh(multivoice.Sing()))
		}
	})
}

func stopAudio() {
	sl.Stop()
}

type sineVoice struct {
	Amp *audio.Control
	Osc audio.FixedFreqSineOsc
}

func newSineVoice(freq float64) *sineVoice {
	v := &sineVoice{Amp: audio.NewControl([]*audio.ControlPoint{{0, -12}, {.05, -2}, {8, -12}})}
	v.Osc.SetFreq(freq)
	return v
}

func (v *sineVoice) Sing() float64 {
	return v.Osc.Sine() * math.Exp2(v.Amp.Sing())
}

func (v *sineVoice) Done() bool {
	return v.Amp.Done()
}
