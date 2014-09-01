package audio

import "math"

type DCFilter struct {
	a, x, y float64
}

func (f *DCFilter) InitAudio(p Params) {
	rc := 1 / (2*math.Pi*10)
	f.a = rc / (rc + 1/p.SampleRate)
}

func (f *DCFilter) Filter(x float64) float64 {
	f.y = f.a * (f.y + x - f.x)
	f.x = x
	return f.y
}
