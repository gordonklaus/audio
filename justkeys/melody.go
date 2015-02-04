package main

type ratio struct {
	a, b int
}

func (r ratio) mul(s ratio) ratio {
	r.a *= s.a
	r.b *= s.b
	d := gcd(r.a, r.b)
	r.a /= d
	r.b /= d
	return r
}

func (r ratio) div(s ratio) ratio { return r.mul(ratio{s.b, s.a}) }

func (r ratio) less(s ratio) bool { return r.a*s.b < s.a*r.b }
func (r ratio) float() float64    { return float64(r.a) / float64(r.b) }

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
	for _, two := range []int{-3, -2, -1, 0, 1, 2, 3} {
		for _, three := range []int{-2, -1, 0, 1, 2} {
			for _, five := range []int{-1, 0, 1} {
				for _, seven := range []int{-1, 0, 1} {
					n, d := 1, 1
					n, d = mul(n, d, 2, two)
					n, d = mul(n, d, 3, three)
					n, d = mul(n, d, 5, five)
					n, d = mul(n, d, 7, seven)
					rats = append(rats, ratio{n, d})
				}
			}
		}
	}
	return rats
}

func complexity(r []ratio, amp []float64) float64 {
	c := 0.0
	ampSum := 0.0
	for i := range r {
		for j := range r[:i] {
			c += amp[i] * amp[j] * float64(r[i].div(r[j]).complexity())
		}
		ampSum += amp[i]
	}
	return c / ampSum
}

func (r ratio) complexity() int {
	c := 0
	for d := 2; r.a != r.b; {
		d1 := r.a%d == 0
		d2 := r.b%d == 0
		if d1 != d2 {
			c += d - 1
		}
		if d1 {
			r.a /= d
		}
		if d2 {
			r.b /= d
		}
		if !(d1 || d2) {
			d++
		}
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
