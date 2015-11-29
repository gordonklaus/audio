package main

import (
	"math"
	"math/rand"
	"time"

	"code.google.com/p/gordon-go/audio"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	audio.Play(&song{center: 256})
}

type song struct {
	Params audio.Params
	tones  []*tone
	n      int

	center float64
	last   float64
}

func (s *song) Sing() float64 {
	s.n--
	if s.n < 0 {
		s.n = int(s.Params.SampleRate / 20)

		for i, t := range s.tones {
			if t.Smoother.Value() < -16 {
				s.removeTone(i)
				break
			}
		}
		for len(s.tones) < 6 {
			s.addTone()
			print(".")
		}

		for _, t := range s.tones {
			t.dAmp = math.Max(-8, math.Min(8, t.dAmp + 4*(rand.NormFloat64()+1/t.amp) / 20))
			// t.dAmp *= .98
			t.amp = math.Min(0, t.amp + t.dAmp / 20)
		}
		// c := s.complexity()
		// for _, t := range s.tones {
		// 	t.amp += math.Log2(1 / c)
		// }
	}
	x := 0.0
	for _, t := range s.tones {
		x += t.sing()
	}
	return math.Tanh(x)
}

func (s *song) Done() bool {
	return false
}

func (s *song) addTone() {
	// TODO: avoid duplicate freqs
	t := newTone(s.next())
	audio.Init(t, s.Params)
	s.tones = append(s.tones, t)
}

func (s *song) next() (n int, freq float64) {
	if len(s.tones) == 0 {
		s.last = s.center
		return 1, s.last
	}
	cSum, ampSum := s.partialComplexity()
	sum := 0.0
	sums := make([]float64, len(rats))
	for i, r := range rats {
		p := math.Log2(s.last * r.float() / s.center)
		sum += math.Exp2(-p*p/2) * math.Exp2(-s.complexityWith(cSum, ampSum, r))
		sums[i] = sum
	}
	i := 0
	x := sum * rand.Float64()
	for i = range sums {
		if x < sums[i] {
			break
		}
	}
	r := rats[i]
	s.last *= r.float()
	r.a *= s.tones[len(s.tones)-1].n
	for _, t := range s.tones {
		t.n *= r.b
	}
	return r.a, s.last
}

func (s *song) removeTone(i int) {
	s.tones = append(s.tones[:i], s.tones[i+1:]...)
	d := s.tones[0].n
	for _, t := range s.tones[1:] {
		d = gcd(d, t.n)
	}
	for _, t := range s.tones {
		t.n /= d
	}
}

type ratio struct {
	a, b int
}

func (r ratio) float() float64 { return float64(r.a) / float64(r.b) }

var rats []ratio

func init() {
	pow := func(a, x int) int {
		y := 1
		for x > 0 {
			y *= a
			x--
		}
		return y
	}
	mul := func(n, d, a, x int) (int, int) {
		if x > 0 {
			return n * pow(a, x), d
		}
		return n, d * pow(a, -x)
	}
	for _, two := range []int{-3, -2, -1, 0, 1, 2, 3} {
		for _, three := range []int{-2, -1, 0, 1, 2} {
			for _, five := range []int{-1, 0, 1} {
				for _, seven := range []int{-1, 0, 1} {
					n, d := 1, 1
					n, d = mul(n, d, 2, two)
					n, d = mul(n, d, 3, three)
					n, d = mul(n, d, 5, five)
					n, d = mul(n, d, 7, seven)
					if complexity(n, d) < 12 {
						rats = append(rats, ratio{n, d})
					}
				}
			}
		}
	}
}

func (s *song) complexity() float64 {
	cSum, ampSum := s.partialComplexity()
	return cSum / ampSum
}

func (s *song) partialComplexity() (cSum, ampSum float64) {
	for _, t1 := range s.tones {
		a1 := math.Exp2(t1.amp)
		for _, t2 := range s.tones {
			a2 := math.Exp2(t2.amp)
			cSum += a1 * a2 * float64(complexity(t1.n, t2.n))
		}
		ampSum += a1
	}
	return
}

func (s *song) complexityWith(cSum, ampSum float64, r ratio) float64 {
	a1 := 1.0
	t1n := r.a * s.tones[len(s.tones)-1].n
	for _, t2 := range s.tones {
		cSum += a1 * math.Exp2(t2.amp) * float64(complexity(t1n, t2.n*r.b))
	}
	return (cSum + a1*a1) / (ampSum + a1)
}

func complexity(a, b int) int {
	c := 1
	for d := 2; a != b; {
		d1 := a%d == 0
		d2 := b%d == 0
		if d1 != d2 {
			c += d - 1
		}
		if d1 {
			a /= d
		}
		if d2 {
			b /= d
		}
		if !(d1 || d2) {
			d++
		}
	}
	return c
}

func gcd(a, b int) int {
	for a > 0 {
		if a > b {
			a, b = b, a
		}
		b -= a
	}
	return b
}

type tone struct {
	attenuate float64
	amp, dAmp float64
	n         int
	Smoother  audio.LinSmoother
	Osc       audio.FixedFreqSineOsc
}

func newTone(n int, freq float64) *tone {
	t := &tone{attenuate: 4 / math.Log2(freq), amp: -16, dAmp: 4, n: n}
	t.Smoother.SetValue(t.amp)
	t.Smoother.SetAttackSpeed(2)
	t.Smoother.SetReleaseSpeed(2)
	t.Osc.SetFreq(freq)
	return t
}

func (t *tone) sing() float64 {
	return t.attenuate * math.Exp2(t.Smoother.Smooth(t.amp)) * math.Tanh(2*t.Osc.Sine())
}
