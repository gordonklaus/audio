package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/portaudio-go/portaudio"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

var (
	seed   = flag.Int64("seed", math.MinInt64, "seed for random phases")
	random = flag.Bool("random", false, "randomly choose seed for random phases")
)

func main() {
	flag.Parse()
	if *random {
		*seed = time.Now().UnixNano()
		fmt.Println("seed =", *seed)
	}
	if *seed != math.MinInt64 {
		*random = true
	}
	rand.Seed(*seed)

	portaudio.Initialize()
	defer portaudio.Terminate()

	p := audio.Params{SampleRate: 48000}

	m := audio.MultiVoice{}
	audio.Init(&m, p)
	// for _, x := range [][2]float64{{1, 1}, {2, 1}, {3, 1}, {4, 1}, {5, 1}, {3, 2}, {4, 3}, {5, 4}, {7, 4}, {6, 5}, {7, 5}, {8, 5}, {9, 5}} {
	for xy := 1; xy <= 60; xy++ {
		for x := 2; x <= int(math.Sqrt(float64(xy))); x++ {
			if y := xy / x; x*y == xy && relPrime(x, y) {
				x, y, xy := float64(x), float64(y), float64(xy)
				c := math.Exp(-xy * math.Log2(y/x) / 12)
				f := y / x
				phase := 0.0
				if *random {
					phase = rand.Float64()
				}
				m.Add(newSineBeat(.5*c, 128*f, 1/xy, phase, .1/f))
			}
		}
	}

	s, err := portaudio.OpenDefaultStream(0, 1, p.SampleRate, 512, func(out []float32) {
		for i := range out {
			out[i] = float32(math.Tanh(m.Sing()))
		}
	})
	if err != nil {
		panic(err)
	}
	s.Start()
	fmt.Println("Press Enter to stop.")
	fmt.Scanln()
	s.Stop()
}

func relPrime(x, y int) bool {
	for i := 2; i < x && i < y; i++ {
		if x%i == 0 && y%i == 0 {
			return false
		}
	}
	return true
}

type sineBeat struct {
	amp  float64
	Sine audio.FixedFreqSineOsc
	Env  normalOsc
}

func newSineBeat(amp, sineFreq, beatFreq, beatPhase, beatWidth float64) *sineBeat {
	b := &sineBeat{amp: amp}
	b.Sine.SetFreq(sineFreq)
	b.Env.Sine.SetFreq(beatFreq / 2)
	b.Env.Sine.SetPhase(beatPhase)
	b.Env.width = math.Log(beatWidth)
	return b
}

func (b *sineBeat) Sing() float64 {
	return b.amp * b.Sine.Sine() * b.Env.osc()
}

func (b *sineBeat) Done() bool { return false }

type normalOsc struct {
	Sine  audio.FixedFreqSineOsc
	width float64
}

func (o *normalOsc) osc() float64 {
	x := o.Sine.Sine() * o.width
	return math.Exp(-x * x)
}
