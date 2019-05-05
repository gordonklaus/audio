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

type SinePM struct {
	pidt  float64
	freq  float64
	step  float64
	phase float64
	pm    float64
}

func (o *SinePM) InitAudio(p Params) {
	o.pidt = math.Pi / p.SampleRate
	o.Freq(o.freq)
}

func (o *SinePM) Freq(freq float64) *SinePM {
	o.freq = freq
	o.step = 2 * freq * o.pidt
	return o
}

func (o *SinePM) PM(pm float64) *SinePM {
	o.pm = pm * math.Pi
	return o
}

func (o *SinePM) Phase(phase float64) *SinePM {
	o.phase = phase
	return o
}

func (o *SinePM) Sing() float64 {
	o.phase += o.step
	if o.phase > math.Pi {
		o.phase -= 2 * math.Pi
	}
	return math.Sin(o.phase + o.pm)
}

type SineSelfPM struct {
	pidt  float64
	freq  float64
	index float64
	x, d  complex128
}

func (o *SineSelfPM) InitAudio(p Params) {
	o.pidt = math.Pi / p.SampleRate
	if o.x == 0 {
		o.x = 1
	}
	o.Freq(o.freq)
}

func (o *SineSelfPM) Freq(freq float64) *SineSelfPM {
	o.freq = freq
	return o
}

func (o *SineSelfPM) Index(i float64) *SineSelfPM {
	o.index = i
	return o
}

func (o *SineSelfPM) Phase(phase float64) *SineSelfPM {
	o.x = cmplx.Exp(complex(0, 2*math.Pi*phase))
	return o
}

func (o *SineSelfPM) Sing() float64 {
	const (
		maxStepSize = 0.05
		maxSteps    = 100
	)

	wdt_2 := o.freq * o.pidt / (1 - o.index*real(o.x))
	if wdt_2 < maxStepSize {
		o.x *= o.expwdt_approx(wdt_2)
		return imag(o.x)
	}

	steps := math.Min(maxSteps, math.Ceil(wdt_2/maxStepSize))
	for i := 0.; i < steps; i++ {
		wdt_2 := o.freq * o.pidt / ((1 - o.index*real(o.x)) * steps)
		o.x *= o.expwdt_approx(wdt_2)
	}

	return imag(o.x)
}

func (o *SineSelfPM) expwdt_approx(wdt_2 float64) complex128 {
	// an optimization of complex(1, wdt_2) / complex(1, -wdt_2):
	wdt_22 := wdt_2 * wdt_2
	_1wdt_22 := 1 / (1 + wdt_22)
	return complex((1-wdt_22)*_1wdt_22, 2*wdt_2*_1wdt_22)
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

func (o *SawOsc) Phase(x float64) *SawOsc {
	o.x = 2*x - 1
	return o
}

func (o *SawOsc) Sing() float64 {
	o.x += o.d
	if o.x > 1 {
		o.x -= 2
	}
	return o.x
}
