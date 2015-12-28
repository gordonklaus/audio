package audio

import (
	"math"
	"math/cmplx"
)

type SineOsc struct {
	pidt float64
	freq float64
	x, d complex128
}

func (o *SineOsc) InitAudio(p Params) {
	o.pidt = math.Pi / p.SampleRate
	if o.x == 0 {
		o.x = 1
	}
	o.Freq(o.freq)
}

func (o *SineOsc) Freq(freq float64) *SineOsc {
	o.freq = freq
	o.d = o.expwdt_approx(freq)
	return o
}

// expwdt_approx approximates cmplx.Exp(complex(0, w*dt)) as complex(1, w*dt/2) / complex(1, -w*dt/2), where w=2*pi*freq and dt=1/sampleRate.
// At 96000 samples per second, it results in sine waves with frequencies >99% accurate up to 2048 Hz.  At higher frequences, the errors are typically imperceptible.  See TestSineOsc_expwdt_approx for details.
func (o *SineOsc) expwdt_approx(freq float64) complex128 {
	wdt_2 := freq * o.pidt

	// an optimization of complex(1, wdt_2) / complex(1, -wdt_2):
	wdt_22 := wdt_2 * wdt_2
	_1wdt_22 := 1 / (1 + wdt_22)
	return complex((1-wdt_22)*_1wdt_22, 2*wdt_2*_1wdt_22)
}

func (o *SineOsc) Phase(phase float64) *SineOsc {
	o.x = cmplx.Exp(complex(0, 2*math.Pi*phase))
	return o
}

func (o *SineOsc) Sing() float64 {
	o.x *= o.d
	return imag(o.x)
}

type SawOsc struct {
	dt   float64
	freq float64
	x, d float64
}

func (o *SawOsc) InitAudio(p Params) {
	o.dt = 1 / p.SampleRate
	o.Freq(o.freq)
}

func (o *SawOsc) Freq(freq float64) *SawOsc {
	o.freq = freq
	o.d = 2 * freq * o.dt
	return o
}

func (o *SawOsc) Sing() float64 {
	o.x += o.d
	if o.x > 1 {
		o.x -= 2
	}
	return o.x
}
