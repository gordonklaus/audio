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
	audiogui.Play(&song{rhythm: newRhythm(1./4, 5), melody: newMelody(256, 5)})
}

type song struct {
	EventDelay audio.EventDelay
	rhythm     melody
	melody     melody
	MultiVoice audio.MultiVoice
}

func (s *song) InitAudio(p audio.Params) {
	audio.Init(&s.EventDelay, p)
	s.EventDelay.Delay(0, s.beat)
	audio.Init(&s.MultiVoice, p)
}

func (s *song) beat() {
	s.MultiVoice.Add(newSineVoice(s.melody.next(rats)))
	s.EventDelay.Delay(s.rhythm.next(rats), s.beat)
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

func (r ratio) float() float64 { return float64(r.a) / float64(r.b) }

type melody struct {
	rhythm  bool
	last    float64
	center  float64
	history []int
	histlen int
}

func newMelody(center float64, histlen int) melody {
	return melody{false, center, center, []int{1}, histlen}
}

func newRhythm(center float64, histlen int) melody {
	return melody{true, center, center, []int{1}, histlen}
}

func (m *melody) next(rats []ratio) float64 {
	sum := 0.0
	sums := make([]float64, len(rats))
	for i, r := range rats {
		p := math.Log2(m.last * r.float() / m.center)
		sum += math.Exp2(-p*p/2) * math.Exp2(-float64(complexity(appendRatio(m.history, r))))
		sums[i] = sum
	}
	if m.rhythm {
		sum += math.Exp2(-float64(complexity(m.history)))
		sums = append(sums, sum)
	}
	i := 0
	x := sum * rand.Float64()
	for i = range sums {
		if x < sums[i] {
			break
		}
	}
	if i == len(rats) {
		return 0
	}
	m.last *= rats[i].float()
	m.history = appendRatio(m.history, rats[i])
	if len(m.history) > m.histlen {
		m.history = m.history[1:]
	}

	d := m.history[0]
	for _, x := range m.history[1:] {
		d = gcd(d, x)
	}
	for i := range m.history {
		m.history[i] /= d
	}

	return m.last
}

func appendRatio(history []int, r ratio) []int {
	r.a *= history[len(history)-1]
	hist := make([]int, len(history))
	for i, x := range history {
		hist[i] = x * r.b
	}
	return append(hist, r.a)
}

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
					rats = append(rats, ratio{n, d})
				}
			}
		}
	}
}

func complexity(n []int) int {
	n = append([]int{}, n...)
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
