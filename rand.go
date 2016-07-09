package audio

import (
	"math/rand"
	"time"
)

type SlowRand struct {
	freq  float64
	i, n  int
	x     [4]float64
	rand  *rand.Rand
	t, dt float64
}

func NewSlowRand(freq float64) *SlowRand {
	return &SlowRand{
		freq: freq,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *SlowRand) InitAudio(p Params) {
	n := p.SampleRate / r.freq / 2
	r.n = int(n)
	r.dt = 1 / n
}

func (r *SlowRand) Sing() float64 {
	r.i--
	if r.i < 0 {
		r.i = r.n - 1
		r.x[0] = r.x[1]
		r.x[1] = r.x[2]
		r.x[2] = r.x[3]
		r.x[3] = .8 * (2*r.rand.Float64() - 1)
		r.t = 0
	}
	r.t += r.dt
	return Interp3(r.t, r.x[0], r.x[1], r.x[2], r.x[3])
}
