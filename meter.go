package audio

import "math"

type AmpMeter struct {
	windowSize float64
	buf        []float64
	i          int
	sum        float64
}

func NewAmpMeter(windowSize float64) *AmpMeter {
	return &AmpMeter{windowSize: windowSize}
}

func (a *AmpMeter) InitAudio(p Params) {
	a.buf = make([]float64, int(p.SampleRate*a.windowSize))
}

func (a *AmpMeter) Add(x float64) {
	a.sum -= a.buf[a.i]
	a.buf[a.i] = x * x
	a.sum += x * x
	a.i = (a.i + 1) % len(a.buf)
}

func (a *AmpMeter) Amplitude() float64 {
	return math.Sqrt(a.sum / float64(len(a.buf)))
}
