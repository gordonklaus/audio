package audio

import (
	"math"
	"math/cmplx"
	"testing"
)

func TestSineOsc_expwdt_approx(t *testing.T) {
	const sampleRate = 96000

	var osc SineOsc
	Init(&osc, Params{sampleRate})
	for freq, err := range map[float64]float64{
		32:    0.00,
		64:    0.00,
		128:   0.00,
		256:   0.00,
		512:   0.00,
		1024:  0.00,
		2048:  0.00,
		4096:  0.01,
		8192:  0.02,
		16384: 0.09,
	} {
		_, θ := cmplx.Polar(osc.expwdt_approx(freq))
		actualFreq := θ / 2 / math.Pi * sampleRate
		actualErr := freq/actualFreq - 1
		if math.Abs(err-actualErr) > .01 {
			t.Errorf("freq=%5.f:, expected %.2f, got %.2f", freq, err, actualErr)
		}
	}
}

func BenchmarkSineOsc(b *testing.B) {
	o := new(SineOsc)
	Init(o, Params{96000})
	for i := 0; i < b.N; i++ {
		o.Freq(1234)
		o.Sing()
	}
}

func BenchmarkSawOsc(b *testing.B) {
	o := new(SawOsc)
	Init(o, Params{96000})
	for i := 0; i < b.N; i++ {
		o.Freq(1234)
		o.Sing()
	}
}
