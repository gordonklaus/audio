package audio

import (
	"math"
	"math/cmplx"
)

type SineOsc struct {
	Params Params
	phase  float64
	Out    Audio
}

func (o *SineOsc) Sine(freq Audio) Audio {
	for i, f := range freq {
		o.Out[i] = math.Sin(2 * math.Pi * o.phase)
		_, o.phase = math.Modf(o.phase + f/o.Params.SampleRate)
	}
	return o.Out
}

type FixedFreqSineOsc struct {
	Params Params
	freq   float64
	x, d   complex128
	Out    Audio
}

func (o *FixedFreqSineOsc) InitAudio(p Params) {
	o.Params = p
	o.SetFreq(o.freq)
	o.Out.InitAudio(p)
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

func (o *FixedFreqSineOsc) Sine() Audio {
	for i := range o.Out {
		o.Out[i] = real(o.x)
		o.x *= o.d
	}
	return o.Out
}
