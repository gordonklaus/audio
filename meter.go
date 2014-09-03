package audio

import "math"

type RMS struct {
	windowSize float64
	buf        []float64
	i          int
	sum        float64
}

func NewRMS(windowSize float64) *RMS {
	return &RMS{windowSize: windowSize}
}

func (a *RMS) InitAudio(p Params) {
	a.buf = make([]float64, int(p.SampleRate*a.windowSize))
}

func (a *RMS) Add(x float64) {
	a.sum -= a.buf[a.i]
	a.buf[a.i] = x * x
	a.sum += x * x
	a.i = (a.i + 1) % len(a.buf)
}

func (a *RMS) Amplitude() float64 {
	return math.Sqrt(a.sum / float64(len(a.buf)))
}
