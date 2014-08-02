package audiogui

import "math"

type uniformGrid struct {
	offset, interval float64
}

func (g uniformGrid) next(t float64, next bool) float64 {
	x := (t - g.offset) / g.interval
	i := math.Floor(x + .5)
	if i*g.interval+g.offset == t {
		if next {
			i++
		} else {
			i--
		}
	} else {
		if next {
			i = math.Ceil(x)
		} else {
			i = math.Floor(x)
		}
	}
	return i*g.interval + g.offset
}

func (g uniformGrid) snap(t float64) float64 {
	x := (t - g.offset) / g.interval
	i := math.Floor(x + .5)
	return i*g.interval + g.offset
}
