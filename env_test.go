package audio

import (
	"testing"
)

func BenchmarkExpEnv(b *testing.B) {
	var e ExpEnv
	Init(&e, Params{96000})
	for i := 0; i < b.N; i++ {
		if e.Done() {
			e.AttackHoldRelease(.1, 1, 2)
		}
		e.Sing()
	}
}
