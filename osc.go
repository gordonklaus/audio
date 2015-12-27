package audio

import (
	"math"
	"math/cmplx"
)

type SineOsc struct {
	pidt float64
	x    complex128
}

func (o *SineOsc) InitAudio(p Params) {
	o.pidt = math.Pi / p.SampleRate
	o.x = 1
}

func (o *SineOsc) Sine(freq float64) float64 {
	o.x *= o.expwdt_approx(freq)
	return imag(o.x)
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
