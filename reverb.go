package audio

import (
	"math"
	"math/rand"
	"time"
)

type Reverb struct {
	params  Params
	streams []*grainStream
	rand    *rand.Rand
}

func (r *Reverb) InitAudio(p Params) {
	r.params = p
	for i := 0; i < 10; i++ {
		s := &grainStream{buf: make([]float64, int(p.SampleRate)), t: 1}
		s.dcFilter.InitAudio(p)
		r.streams = append(r.streams, s)
	}
	r.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (r *Reverb) Reverb(dry float64) float64 {
	size := .2
	decayTime := 4.0

	wet := 0.0
	x := dry
	for _, s := range r.streams {
		if s.t >= 1 {
			s.t -= 1
			s.a1, s.i1 = s.a2, s.i2
			delay := math.Exp2(r.rand.Float64()-2.5)
			s.dt = 1/math.Exp2(r.rand.Float64())/size/r.params.SampleRate
			s.a2 = math.Pow(.01, delay / decayTime)
			s.i2 = (s.i-int(delay*r.params.SampleRate)+len(s.buf)) % len(s.buf)
		}
		sin2 := math.Sin(math.Pi/2*s.t); sin2 *= sin2
		y := s.dcFilter.Filter(s.a1*(1-sin2)*s.buf[s.i1] + s.a2*sin2*s.buf[s.i2])
		s.i1 = (s.i1 + 1) % len(s.buf)
		s.i2 = (s.i2 + 1) % len(s.buf)
		s.t += s.dt
		s.buf[s.i] = x + y
		s.i = (s.i + 1) % len(s.buf)
		x = y // feed wet output into next stream's input
		wet += y
	}

	return (wet + dry) / 2
}

type grainStream struct {
	buf    []float64
	i      int
	t, dt  float64
	a1, a2 float64
	i1, i2 int
	dcFilter DCFilter
}

type DCFilter struct {
	a, x, y float64
}

func (f *DCFilter) InitAudio(p Params) {
	rc := 1 / (2*math.Pi*10)
	f.a = rc / (rc + 1/p.SampleRate)
}

func (f *DCFilter) Filter(x float64) float64 {
	f.y = f.a * (f.y + x - f.x)
	f.x = x
	return f.y
}

/*

y1[i] = x1[i] + f1*y1[i-n1]
y2[i] = x2[i] + f2*y2[i-n2]
x2 = y1
y2[i] = x1[i] + f1*y1[i-n1] + f2*y2[i-n2]

*/


