package audiogui

import (
	"math"
	"sort"
)

type grid interface {
	defaultValue() float64
	next(float64, bool) float64
}

func defaultGrid(name string) grid {
	switch name {
	case "Pitch":
		return newPitchGrid(8, 7)
	case "Amplitude":
		return uniformGrid{1}
	}
	return &uniformGrid{1}
}

type uniformGrid struct {
	interval float64
}

func (g uniformGrid) defaultValue() float64 { return 0 }

func (g uniformGrid) next(t float64, next bool) float64 {
	x := t / g.interval
	i := math.Floor(x + .5)
	if i*g.interval == t {
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
	return i * g.interval
}

func (g uniformGrid) snap(t float64) float64 {
	x := t / g.interval
	i := math.Floor(x + .5)
	return i * g.interval
}

type valueGrid struct {
	values []float64
}

func (g valueGrid) next(x float64, next bool) float64 {
	i := sort.SearchFloat64s(g.values, x)
	n := len(g.values)
	if i == n {
		return g.values[n-1]
	}
	if g.values[i] == x {
		if next {
			i++
		} else {
			i--
		}
	} else {
		if !next {
			i--
		}
	}
	if i < 0 {
		i = 0
	}
	if i == n {
		i = n - 1
	}
	return g.values[i]
}

type pitchGrid struct {
	valueGrid
	center float64
}

func newPitchGrid(center float64, maxComplexity int) pitchGrid {
	vals := []float64{center}
	more := true
	for a := 2; more; a++ {
		more = false
		for b := 1; b < a; b++ {
			if gcd(a, b) > 1 || complexity(a, b) > maxComplexity {
				continue
			}
			p := math.Log2(float64(a) / float64(b))
			vals = append(vals, center-p, center+p)
			more = true
		}
	}
	sort.Float64s(vals)
	return pitchGrid{valueGrid{vals}, center}
}

func (g pitchGrid) defaultValue() float64 { return g.center }

func complexity(a, b int) int {
	lcm := a * b / gcd(a, b)
	c := 1
	for d := 2; lcm > 1; d++ {
		for lcm%d == 0 {
			c += d - 1
			lcm /= d
		}
	}
	return c
}

func gcd(a, b int) int {
	switch {
	case a > b:
		return gcd(b, a)
	case a == 0:
		return b
	}
	return gcd(b-a, a)
}
