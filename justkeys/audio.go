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

type pluckedTone struct {
	Amp *audio.Control
	Osc audio.FixedFreqSineOsc
}

func newPluckedTone(amp, freq float64) *pluckedTone {
	v := &pluckedTone{Amp: audio.NewControl([]*audio.ControlPoint{{0, -12}, {.05, amp}, {4, -12}})}
	v.Osc.SetFreq(freq)
	return v
}

func (v *pluckedTone) Sing() float64 {
	return math.Exp2(v.Amp.Sing()) * math.Tanh(2*v.Osc.Sine())
}

func (v *pluckedTone) Done() bool {
	return v.Amp.Done()
}

type bowedTone struct {
	amp, targetAmp float64
	ampChan        chan float64
	Osc            audio.FixedFreqSineOsc
}

func newBowedTone(freq float64) *bowedTone {
	v := &bowedTone{amp: -8, targetAmp: -8, ampChan: make(chan float64, 100)}
	v.Osc.SetFreq(freq)
	return v
}

func (v *bowedTone) attack(amp float64) {
	select {
	case v.ampChan <- amp:
	default:
	}
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
	v.amp += math.Min(da, math.Max(-da, v.targetAmp-v.amp))
	return math.Exp2(v.amp) * math.Tanh(2*v.Osc.Sine())
}

func (v *bowedTone) Done() bool {
	return v.amp < -12
}
