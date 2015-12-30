package audio

import (
	"math/rand"
	"testing"
)

func BenchmarkSaturate(b *testing.B) {
	x := make([]float64, 1024)
	for i := range x {
		x[i] = 8*rand.Float64() - 4
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Saturate(x[i&1<<9])
	}
}
