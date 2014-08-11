package audio

import (
	"math/rand"
	"time"
)

type Rand struct {
	Out  Audio
	rand *rand.Rand
}

func NewRand() *Rand {
	return &Rand{rand: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (r *Rand) Rand() Audio {
	for i := range r.Out {
		r.Out[i] = r.rand.Float64()
	}
	return r.Out
}

func (r *Rand) NormRand() Audio {
	for i := range r.Out {
		r.Out[i] = r.rand.NormFloat64()
	}
	return r.Out
}
