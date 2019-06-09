package audio

import (
	"math"
)

type Reverb struct {
	Delay, AllPass1, AllPass2 *reverbAllPass
	LowPass                   *LowPass1
}

func (r *Reverb) InitAudio(p Params) {
	r.Delay = newReverbAllPass([]float64{.15011404})
	r.AllPass1 = newReverbAllPass([]float64{.1350392713})
	r.AllPass2 = newReverbAllPass([]float64{.14302603, .114801574})
	r.LowPass = new(LowPass1).Freq(22000)
	Init(r.Delay, p)
	Init(r.AllPass1, p)
	Init(r.AllPass2, p)
	Init(r.LowPass, p)
}

func (r *Reverb) Filter(x float64) float64 {
	decay := 4.0
	t := r.Delay.Taps[0].t * (1 + r.Delay.Taps[0].Rand.Sing()/150)
	c := math.Pow(.01, t/decay) * 5 / 6
	x = c * r.LowPass.Filter(r.AllPass2.Filter(r.AllPass1.Filter(x-r.Delay.Delay.Read(t))))
	r.Delay.Delay.Write(x)
	return x
}

// reverbAllPass is an allpass filter with slightly modulating taps.
type reverbAllPass struct {
	Taps  []reverbTap
	Delay Delay
}

type reverbTap struct {
	t, c float64
	Rand *SlowRand
}

func newReverbAllPass(taps []float64) *reverbAllPass {
	decay := 8.0
	r := &reverbAllPass{Taps: make([]reverbTap, len(taps))}
	tMax := taps[0]
	for i, t := range taps {
		if t > tMax {
			panic("The first tap must be the largest.")
		}
		r.Taps[i] = reverbTap{
			t:    t,
			c:    math.Pow(.01, t/decay) / float64(len(taps)),
			Rand: NewSlowRand(1),
		}
	}
	return r
}

func (f *reverbAllPass) Filter(x float64) float64 {
	// Direct form II implementation (accumulate feedback into x, feedforward into y):
	y := 0.0
	tap0 := f.Taps[0]
	t0 := tap0.t * (1 + tap0.Rand.Sing()/150)
	c0 := tap0.c
	x -= c0 * f.Delay.Read(t0)
	for _, tap := range f.Taps[1:] {
		t := tap.t * (1 + tap.Rand.Sing()/150)
		x -= tap.c * f.Delay.Read(t)
		y += tap.c * f.Delay.Read(t0-t)
	}
	y += c0*x + f.Delay.Read(t0)
	f.Delay.Write(x)
	return y
}
