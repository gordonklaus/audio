package audio

import (
	"math"
	"math/cmplx"
)

type SineOsc struct {
	Params Params
	freq   float64
	x, d   complex128
	Out    Audio
}

func (o *SineOsc) InitAudio(p Params) {
	o.Params = p
	o.SetFreq(o.freq)
	o.Out.InitAudio(p)
}

func (o *SineOsc) SetFreq(freq float64) {
	o.freq = freq
	if o.x == 0 {
		o.x = 1
	}
	o.d = cmplx.Exp(complex(0, 2*math.Pi*freq/o.Params.SampleRate))
}

func (o *SineOsc) SetPhase(phase float64) {
	o.x = cmplx.Exp(complex(0, 2*math.Pi*phase))
}

func (o *SineOsc) Sine() Audio {
	for i := range o.Out {
		o.Out[i] = real(o.x)
		o.x *= o.d
	}
	return o.Out
}
