package main

type ratio struct {
	a, b int
}

func (r ratio) float() float64 { return float64(r.a) / float64(r.b) }

type melody struct {
	current float64
	history []int
	histlen int
}

func newMelody(start float64, histlen int) melody {
	return melody{start, []int{1}, histlen}
}

func (m *melody) add(r ratio) {
	m.current *= r.float()
	m.history = appendRatio(m.history, r)
	if len(m.history) > m.histlen {
		m.history = m.history[1:]
	}
	d := m.history[0]
	for _, x := range m.history[1:] {
		d = gcd(d, x)
	}
	for i := range m.history {
		m.history[i] /= d
	}
}

func appendRatio(history []int, r ratio) []int {
	a := r.a * history[len(history)-1]
	hist := make([]int, len(history))
	for i, x := range history {
		hist[i] = x * r.b
	}
	return append(hist, a)
}

func rats() []ratio {
	rats := []ratio{}
	pow := func(a, x int) int {
		y := 1
		for x > 0 {
			y *= a
			x--
		}
		return y
	}
	mul := func(n, d, a, x int) (int, int) {
		if x > 0 {
			return n * pow(a, x), d
		}
		return n, d * pow(a, -x)
	}
	x := []int{-2, -1, 0, 1, 2}
	for _, two := range []int{-3, -2, -1, 0, 1, 2, 3} {
		for _, three := range x {
			for _, five := range x {
				// for _, seven := range x {
				n, d := 1, 1
				n, d = mul(n, d, 2, two)
				n, d = mul(n, d, 3, three)
				n, d = mul(n, d, 5, five)
				// n, d = mul(n, d, 7, seven)
				rats = append(rats, ratio{n, d})
				// }
			}
		}
	}
	return rats
}

func complexity(n []int) int {
	c := 1
divisors:
	for d := 2; ; d++ {
		for {
			dividesAny := false
			dividesAll := true
			for i := range n {
				if n[i]%d == 0 {
					n[i] /= d
					dividesAny = true
				} else {
					dividesAll = false
				}
			}
			if !dividesAny {
				break
			}
			if !dividesAll {
				c += d - 1
			}
		}
		for _, n := range n {
			if n > 1 {
				continue divisors
			}
		}
		break
	}
	return c
}

func gcd(a, b int) int {
	for a > 0 {
		if a > b {
			a, b = b, a
		}
		b -= a
	}
	return b
}
