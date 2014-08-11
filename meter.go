package audio

import "math"

type AmpMeter struct {
	windowSize float64
	buf        Audio
	i          int
	sum        float64
}

func NewAmpMeter(windowSize float64) *AmpMeter {
	return &AmpMeter{windowSize: windowSize}
}

func (a *AmpMeter) InitAudio(p Params) {
	a.buf = make(Audio, int(p.SampleRate*a.windowSize))
}

func (a *AmpMeter) Amplitude(x Audio) float64 {
	for _, x := range x {
		a.sum -= a.buf[a.i]
		a.buf[a.i] = x * x
		a.sum += a.buf[a.i]
		a.i = (a.i + 1) % len(a.buf)
	}
	return math.Sqrt(a.sum / float64(len(a.buf)))
}
