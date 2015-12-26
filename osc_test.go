package audio

import (
	"testing"
)

func BenchmarkSawOsc(b *testing.B) {
	o := new(SawOsc).Freq(1234)
	Init(o, Params{96000})
	for i := 0; i < b.N; i++ {
		o.Saw()
	}
}
