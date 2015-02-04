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
			out[i] = float32(math.Tanh(multivoice.Sing() / 4))
		}
	})
}

func stopAudio() {
	sl.Stop()
}

type pressedTone struct {
	Amp audio.Control
	Osc audio.FixedFreqSineOsc
}

func newPressedTone(freq float64) *pressedTone {
	v := &pressedTone{}
	v.Amp.SetPoints([]*audio.ControlPoint{{0, -12}, {9999, -12}})
	v.Osc.SetFreq(freq)
	return v
}

func (v *pressedTone) attack(amp float64) {
	a := v.Amp.Sing()
	t := .15
	if amp < a {
		t = 4
	}
	v.Amp.SetPoints([]*audio.ControlPoint{{0, a}, {t, amp}, {9999, amp}})
}

func (v *pressedTone) release() {
	a := v.Amp.Sing()
	v.Amp.SetPoints([]*audio.ControlPoint{{0, a}, {4, -12}})
}

func (v *pressedTone) amp() float64 {
	return math.Exp2(v.Amp.Sing())
}

func (v *pressedTone) Sing() float64 {
	return math.Exp2(v.Amp.Sing()) * math.Tanh(2*v.Osc.Sine())
}

func (v *pressedTone) Done() bool {
	return v.Amp.Done()
}

type pluckedTone struct {
	Amp *audio.Control
	Osc audio.FixedFreqSineOsc
}

func newPluckedTone(amp, freq float64) *pluckedTone {
	v := &pluckedTone{Amp: audio.NewControl([]*audio.ControlPoint{{0, -12}, {.05, amp}, {4, -12}})}
	v.Osc.SetFreq(freq)
	return v
}

func (v *pluckedTone) amp() float64 {
	return math.Exp2(v.Amp.Sing())
}

func (v *pluckedTone) Sing() float64 {
	return math.Exp2(v.Amp.Sing()) * math.Tanh(2*v.Osc.Sine())
}

func (v *pluckedTone) Done() bool {
	return v.Amp.Done()
}

type bowedTone struct {
	amp_, targetAmp float64
	ampChan         chan float64
	Osc             audio.FixedFreqSineOsc
}

func newBowedTone(freq float64) *bowedTone {
	v := &bowedTone{amp_: -8, targetAmp: -8, ampChan: make(chan float64, 100)}
	v.Osc.SetFreq(freq)
	return v
}

func (v *bowedTone) attack(amp float64) {
	select {
	case v.ampChan <- amp:
	default:
	}
}

func (v *bowedTone) amp() float64 {
	return math.Exp2(v.amp_)
}

func (v *bowedTone) Sing() float64 {
	select {
	case targetAmp := <-v.ampChan:
		v.targetAmp = math.Max(v.targetAmp, targetAmp)
	default:
	}
	decay := 3.0 / 48000
	v.targetAmp -= decay
	da := 16.0 / 48000
	v.amp_ += math.Min(da, math.Max(-da, v.targetAmp-v.amp_))
	return math.Exp2(v.amp_) * math.Tanh(2*v.Osc.Sine())
}

func (v *bowedTone) Done() bool {
	return v.amp_ < -12
}
