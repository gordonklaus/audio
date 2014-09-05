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
	fbRMS    *audio.RMS
	limiter  *audio.Limiter
	outRMS   *audio.RMS

	Decay, Sustain, Dry, Wet audio.Control
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
	r.fbRMS = audio.NewRMS(.5)
	r.fbRMS.InitAudio(p)
	r.outRMS = audio.NewRMS(.5)
	r.outRMS.InitAudio(p)
	r.limiter = audio.NewLimiter(.25, 1, 1)
	r.limiter.InitAudio(p)

	r.Decay.InitAudio(p)
	r.Sustain.InitAudio(p)
	r.Dry.InitAudio(p)
	r.Wet.InitAudio(p)
}

func (r *reverb) Play(_ struct{}) {}
func (r *reverb) Sing() float64   { return 0 }

func (r *reverb) Done() bool {
	return r.Dry.Done() && r.Wet.Done() && r.outRMS.Amplitude() < .001
}

func (r *reverb) Stop() {
	r.InitAudio(r.params)
}

func (r *reverb) reverb(dry float64) float64 {
	f := func(x float64) float64 { return x * math.Sqrt(x*(6*x*x-15*x+10)) }

	size := .05

	decayTime := r.Decay.Sing()
	sustain := r.Sustain.Sing()

	wet := 0.0
	for _, s := range r.streams {
		if s.t >= 1 {
			s.t -= 1
			s.a1, s.i1 = s.a2, s.i2
			delay := math.Exp2(r.rand.Float64() - 2.5)
			s.dt = 1 / (math.Exp2(r.rand.Float64()*4) * size * r.params.SampleRate)
			s.a2 = math.Pow(.01, delay/decayTime)
			if amp, sustain := r.fbRMS.Amplitude(), math.Exp2(sustain); amp < sustain {
				s.a2 = math.Min(sustain/amp, 1.1)
			} else if amp > 1 {
				s.a2 = math.Max(1/amp, .9)
			}
			s.i2 = (r.i - int(delay*r.params.SampleRate) + len(r.buf)) % len(r.buf)
		}
		wet += s.a1*f(1-s.t)*r.buf[s.i1] + s.a2*f(s.t)*r.buf[s.i2]
		s.i1 = (s.i1 + 1) % len(r.buf)
		s.i2 = (s.i2 + 1) % len(r.buf)
		s.t += s.dt
	}
	wet /= math.Sqrt(float64(len(r.streams)))
	fb := r.dcFilter.Filter(dry + wet)
	r.fbRMS.Add(fb)
	r.buf[r.i] = fb
	r.i = (r.i + 1) % len(r.buf)

	out := r.limiter.Limit(math.Exp2(r.Dry.Sing())*dry + math.Exp2(r.Wet.Sing())*wet)
	r.outRMS.Add(out)
	return out
}

type grainStream struct {
	t, dt  float64
	a1, a2 float64
	i1, i2 int
}
