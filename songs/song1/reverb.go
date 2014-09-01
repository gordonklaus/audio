package main

import (
	"code.google.com/p/gordon-go/audio"

	"math"
	"math/rand"
	"time"
)

type reverb struct {
	params   audio.Params
	streams  []*grainStream
	buf      []float64
	i        int
	rand     *rand.Rand
	dcFilter audio.DCFilter
	ampMeter *audio.AmpMeter
}

func (r *reverb) InitAudio(p audio.Params) {
	r.params = p
	r.streams = nil
	for i := 0; i < 30; i++ {
		s := &grainStream{t: 1}
		r.streams = append(r.streams, s)
	}
	r.buf = make([]float64, int(p.SampleRate))
	r.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	r.dcFilter.InitAudio(p)
	r.ampMeter = audio.NewAmpMeter(.5)
	r.ampMeter.InitAudio(p)
}

func (r *reverb) reverb(dry float64) float64 {
	f := func(x float64) float64 { return x*math.Sqrt(x*(6*x*x-15*x+10)) }

	size := .05
	// decayTime := 4.0

	a := 1.0
	switch amp := r.ampMeter.Amplitude(); {
	case amp < .05:
		a = 1.1
	case .8 < amp:
		a = .8 / amp
	}

	wet := 0.0
	for _, s := range r.streams {
		if s.t >= 1 {
			s.t -= 1
			s.a1, s.i1 = s.a2, s.i2
			delay := math.Exp2(r.rand.Float64()-2.5)
			s.dt = 1 / (math.Exp2(r.rand.Float64()*4)*size*r.params.SampleRate)
			s.a2 = a//math.Pow(.01, delay / decayTime)
			s.i2 = (r.i-int(delay*r.params.SampleRate)+len(r.buf)) % len(r.buf)
		}
		wet += s.a1*f(1-s.t)*r.buf[s.i1] + s.a2*f(s.t)*r.buf[s.i2]
		s.i1 = (s.i1 + 1) % len(r.buf)
		s.i2 = (s.i2 + 1) % len(r.buf)
		s.t += s.dt
	}
	wet /= math.Sqrt(float64(len(r.streams)))
	r.buf[r.i] = r.dcFilter.Filter(dry + wet)
	r.i = (r.i + 1) % len(r.buf)

	r.ampMeter.Add(dry + wet)
	return dry + wet
}

func (r *reverb) Done() bool {
	return r.ampMeter.Amplitude() < .00001
}

type grainStream struct {
	t, dt  float64
	a1, a2 float64
	i1, i2 int
}
