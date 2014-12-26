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

type pluckedSine struct {
	Amp *audio.Control
	Osc audio.FixedFreqSineOsc
}

func newPluckedSine(amp, freq float64) *pluckedSine {
	v := &pluckedSine{Amp: audio.NewControl([]*audio.ControlPoint{{0, -12}, {.05, amp}, {4, -12}})}
	v.Osc.SetFreq(freq)
	return v
}

func (v *pluckedSine) Sing() float64 {
	return v.Osc.Sine() * math.Exp2(v.Amp.Sing())
}

func (v *pluckedSine) Done() bool {
	return v.Amp.Done()
}

type bowedSine struct {
	amp, targetAmp float64
	ampChan        chan float64
	Osc            audio.FixedFreqSineOsc
}

func newBowedSine(freq float64) *bowedSine {
	v := &bowedSine{amp: -8, targetAmp: -8, ampChan: make(chan float64, 100)}
	v.Osc.SetFreq(freq)
	return v
}

func (v *bowedSine) attack(amp float64) {
	select {
	case v.ampChan <- amp:
	default:
	}
}

func (v *bowedSine) Sing() float64 {
	select {
	case targetAmp := <-v.ampChan:
		v.targetAmp = math.Max(v.targetAmp, targetAmp)
	default:
	}
	decay := 3.0 / 48000
	v.targetAmp -= decay
	da := 16.0 / 48000
	v.amp += math.Min(da, math.Max(-da, v.targetAmp-v.amp))
	return v.Osc.Sine() * math.Exp2(v.amp)
}

func (v *bowedSine) Done() bool {
	return v.amp < -12
}
