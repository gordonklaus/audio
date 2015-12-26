package audio

import (
	"math"
	"math/cmplx"
)

type SineOsc struct {
	Params Params
	phase  float64
}

func (o *SineOsc) Sine(freq float64) float64 {
	_, o.phase = math.Modf(o.phase + freq/o.Params.SampleRate)
	return math.Sin(2 * math.Pi * o.phase)
}

type FixedFreqSineOsc struct {
	Params Params
	freq   float64
	x, d   complex128
}

func (o *FixedFreqSineOsc) InitAudio(p Params) {
	o.Params = p
	o.SetFreq(o.freq)
}

func (o *FixedFreqSineOsc) SetFreq(freq float64) {
	o.freq = freq
	if o.x == 0 {
		o.SetPhase(0)
	}
	o.d = cmplx.Exp(complex(0, 2*math.Pi*freq/o.Params.SampleRate))
}

func (o *FixedFreqSineOsc) SetPhase(phase float64) {
	o.x = cmplx.Exp(complex(0, 2*math.Pi*phase))
}

func (o *FixedFreqSineOsc) Sine() float64 {
	o.x *= o.d
	return real(o.x)
}

type SawOsc struct {
	Params Params
	freq   float64
	x, d   float64
}

func (o *SawOsc) InitAudio(p Params) {
	o.Params = p
	o.Freq(o.freq)
}

func (o *SawOsc) Freq(freq float64) *SawOsc {
	o.freq = freq
	o.d = 2 * freq / o.Params.SampleRate
	return o
}

func (o *SawOsc) Saw() float64 {
	o.x += o.d
	if o.x > 1 {
		o.x -= 2
	}
	return o.x
}
