package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var (
	print   = fmt.Print
	printf  = fmt.Printf
	println = fmt.Println
)

func main() {
	rand.Seed(time.Now().UnixNano())
	audiogui.Play(&song{repPeriod: newMelody(ratio{1, 4}), beatFreq: newMelody(ratio{6, 1}), sineFreq: newMelody(ratio{256, 1})})
}

type song struct {
	EventDelay audio.EventDelay
	repPeriod  melody
	beatCount  int
	beatFreq   melody
	sineFreq   melody
	MultiVoice audio.MultiVoice
}

func (s *song) InitAudio(p audio.Params) {
	audio.Init(&s.EventDelay, p)
	s.EventDelay.Delay(0, s.beat)
	audio.Init(&s.MultiVoice, p)
}

func (s *song) beat() {
	s.beatCount--
	if s.beatCount <= 0 {
		f := s.beatFreq.new(rats())
		rats := make([]ratio, 12)
		for i := range rats {
			rats[i] = ratio{i + 1, 1}.div(f).div(s.repPeriod.last())
		}
		c := s.repPeriod.new(rats).mul(f)
		if c.b != 1 {
			panic(c.b)
		}
		s.beatCount = c.a
	}
	s.MultiVoice.Add(newSineVoice(s.sineFreq.new(rats()).float()))
	s.EventDelay.Delay(1/s.beatFreq.float(), s.beat)
}

func (s *song) Sing() float64 {
	s.EventDelay.Step()
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
	v.Env = audio.NewAttackReleaseEnv(.1, 4)
	return v
}

func (v *sineVoice) InitAudio(p audio.Params) {
	v.Osc.InitAudio(p)
	v.Env.InitAudio(p)
	v.n = int(p.SampleRate * .1)
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

func (r ratio) mul(s ratio) ratio { return ratio{r.a * s.a, r.b * s.b}.reduced() }
func (r ratio) div(s ratio) ratio { return ratio{r.a * s.b, r.b * s.a}.reduced() }
func (r ratio) reduced() ratio    { gcd := gcd(r.a, r.b); return ratio{r.a / gcd, r.b / gcd} }
func (r ratio) float() float64    { return float64(r.a) / float64(r.b) }

type melody struct {
	center  float64
	history []ratio
}

func newMelody(center ratio) melody {
	return melody{center: center.float(), history: []ratio{center}}
}

func (m *melody) last() ratio    { return m.history[len(m.history)-1] }
func (m *melody) float() float64 { return m.last().float() }

func (m *melody) new(rats []ratio) ratio {
	n := len(m.history) - 5
	if n < 0 {
		n = 0
	}
	hist := m.history[n:]
	normHist := make([]ratio, len(hist))
	r := hist[0]
	for i := range hist {
		normHist[i] = hist[i].div(r)
	}
	hist = normHist
	last := hist[len(hist)-1]

	sum := 0.0
	sums := make([]float64, len(rats))
	for i, r := range rats {
		p := math.Log2(m.float() * r.float() / m.center)
		sum += math.Exp2(-p*p/2) * math.Exp2(-float64(complexity(append(hist, r.mul(last)))))
		sums[i] = sum
	}
	i := 0
	x := sum * rand.Float64()
	for i = range sums {
		if x < sums[i] {
			break
		}
	}
	m.history = append(m.history, m.last().mul(rats[i]))
	return m.last()
}

func rats() []ratio {
	rats := []ratio{}
	for a := 1; a < 16; a++ {
		for b := 1; b < 16; b++ {
			if gcd(a, b) != 1 {
				continue
			}
			rats = append(rats, ratio{a, b})
		}
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
	for a > 0 {
		if a > b {
			a, b = b, a
		}
		b -= a
	}
	return b
}
