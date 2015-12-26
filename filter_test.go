package audio

import (
	"testing"
)

func BenchmarkLowPass1(b *testing.B) {
	f := new(LowPass1).Freq(1234)
	Init(&f, Params{96000})
	x := 1.0
	for i := 0; i < b.N; i++ {
		x = f.Filter(x)
	}
}
