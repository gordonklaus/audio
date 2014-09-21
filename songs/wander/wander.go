package main

import (
	"math"
	"math/rand"
	"time"

	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	audiogui.Play(&song{freq: 256})
}

type song struct {
	Params     audio.Params
	next       int
	hist       []ratio
	freq       float64
	MultiVoice audio.MultiVoice
}

func (s *song) Sing() float64 {
	s.next--
	if s.next <= 0 {
		s.next = int(s.Params.SampleRate)
		n := len(s.hist) - 2
		if n < 0 {
			n = 0
		}
		r := rat(s.freq, s.hist[n:])
		s.hist = append(s.hist, r)
		s.freq = s.freq*r.float() + rand.Float64()/64
		s.MultiVoice.Add(newSineVoice(s.freq))
	}
	return math.Tanh(s.MultiVoice.Sing() / 8)
}

func (s *song) Done() bool {
	return false
}

type sineVoice struct {
	Osc audio.FixedFreqSineOsc
	Env *audio.AttackReleaseEnv
	n   int
}

func newSineVoice(freq float64) *sineVoice {
	v := &sineVoice{}
	v.Osc.SetFreq(freq)
	v.Env = audio.NewAttackReleaseEnv(8, 8)
	return v
}

func (v *sineVoice) InitAudio(p audio.Params) {
	v.Osc.InitAudio(p)
	v.Env.InitAudio(p)
	v.n = int(p.SampleRate * 8)
}

func (v *sineVoice) Sing() float64 {
	v.n--
	if v.n < 0 {
		v.Env.Release()
	}
	return v.Osc.Sine() * v.Env.Sing()
}

func (v *sineVoice) Done() bool {
	return v.Env.Done()
}

type ratio struct {
	a, b int
}

func (r ratio) float() float64 { return float64(r.a) / float64(r.b) }

func rat(freq float64, hist []ratio) ratio {
	sum := 0.0
	sums := []float64{}
	rats := []ratio{}
	for a := 1; a < 16; a++ {
		for b := 1; b < 16; b++ {
			if gcd(a, b) != 1 {
				continue
			}
			r := ratio{a, b}
			p := math.Log2(freq*r.float()) - 8
			sum += math.Exp2(-p*p/2) * math.Exp2(-float64(complexity(cumrat(append(hist, r)))))
			sums = append(sums, sum)
			rats = append(rats, r)
		}
	}
	x := sum * rand.Float64()
	for i := range sums {
		if x < sums[i] {
			return rats[i]
		}
	}
	panic("unreachable")
}

func cumrat(rats []ratio) []ratio {
	rats = append([]ratio{{1, 1}}, rats...)
	r := ratio{1, 1}
	for i, s := range rats {
		r.a *= s.a
		r.b *= s.b
		rats[i] = r
	}
	return rats
}

func complexity(rats []ratio) int {
	n := make([]int, len(rats))
	for j := range n {
		n[j] = rats[j].a
		for i, r := range rats {
			if i != j {
				n[j] *= r.b
			}
		}
	}

	c := 1
divisors:
	for d := 2; ; d++ {
		for {
			dividesAny := false
			dividesAll := true
			for i := range n {
				if n[i]%d == 0 {
					n[i] /= d
					dividesAny = true
				} else {
					dividesAll = false
				}
			}
			if !dividesAny {
				break
			}
			if !dividesAll {
				c += d - 1
			}
		}
		for _, n := range n {
			if n > 1 {
				continue divisors
			}
		}
		break
	}
	return c
}

func gcd(a, b int) int {
	switch {
	case a > b:
		return gcd(b, a)
	case a == 0:
		return b
	}
	return gcd(b-a, a)
}
