package audio

import "math"

type Filter struct {
	a, b []FilterTap
	buf  []float64
	i    int
}

type FilterTap struct {
	Delay int
	Coef  float64
}

func (f *Filter) Taps(a, b []FilterTap) *Filter {
	n := 0
	a0 := 0.0
	for _, a := range a {
		if a.Delay < 0 {
			panic(a.Delay)
		}
		if n < a.Delay {
			n = a.Delay
		}
		if a.Delay == 0 {
			a0 = a.Coef
		}
	}
	for _, b := range b {
		if b.Delay < 0 {
			panic(b.Delay)
		}
		if n < b.Delay {
			n = b.Delay
		}
	}
	if a0 == 0 {
		panic(a0)
	}
	for i := range a {
		a[i].Coef /= a0
	}
	for i := range b {
		b[i].Coef /= a0
	}

	f.a = a
	f.b = b
	f.buf = make([]float64, n)
	f.i = 0
	return f
}

func (f *Filter) Filter(x float64) float64 {
	y := 0.0
	n := len(f.buf)

	// Direct form II implementation:
	// First, accumulate feedback into x.
	for _, a := range f.a {
		if a.Delay > 0 {
			x -= a.Coef * f.buf[(f.i-a.Delay+n)%n]
		}
	}
	// Then, accumulate feedforward into y.
	for _, b := range f.b {
		if b.Delay > 0 {
			y += b.Coef * f.buf[(f.i-b.Delay+n)%n]
		} else {
			y += b.Coef * x
		}
	}

	f.buf[f.i] = x
	f.i = (f.i + 1) % n
	return y
}

func AllPassFilterTaps(a []FilterTap) (_, _ []FilterTap) {
	a_ := make([]FilterTap, len(a))
	N := 0
	for _, a := range a {
		if N < a.Delay {
			N = a.Delay
		}
	}
	for i := range a {
		a_[len(a)-i-1] = FilterTap{N - a[i].Delay, a[i].Coef}
	}
	return a, a_
}

func TapsFromRoots(r []complex128) []FilterTap {
	p := poly{1}
	for _, r := range r {
		p = p.mul(poly{1, -2 * real(r), real(r) * real(r) + imag(r) * imag(r)})
	}
	t := make([]FilterTap, len(p))
	for i := range p {
		t[i] = FilterTap{i, p[i]}
	}
	return t
}

type poly []float64

func (p poly) mul(q poly) poly {
	r := make(poly, len(p) + len(q) - 1)
	for i := range p {
		for j := range q {
			r[i+j] += p[i] * q[j]
		}
	}
	return r
}

type LowPass1 struct {
	p          Params
	freq       float64
	a0, b1, y1 float64
}

func (f *LowPass1) Freq(freq float64) *LowPass1 {
	f.freq = freq
	f.b1 = math.Exp(-2 * math.Pi * freq / f.p.SampleRate)
	f.a0 = 1 - f.b1
	return f
}

func (f *LowPass1) InitAudio(p Params) {
	f.p = p
	f.Freq(f.freq)
}

func (f *LowPass1) Filter(x float64) float64 {
	f.y1 = f.a0*x + f.b1*f.y1
	return f.y1
}

type DCFilter struct {
	a, x, y float64
}

func (f *DCFilter) InitAudio(p Params) {
	rc := 1 / (2 * math.Pi * 10)
	f.a = rc / (rc + 1/p.SampleRate)
}

func (f *DCFilter) Filter(x float64) float64 {
	f.y = f.a * (f.y + x - f.x)
	f.x = x
	return f.y
}

type LinSmoother struct {
	params                    Params
	attackSpeed, releaseSpeed float64
	up, down                  float64
	x, y                      float64
}

func NewLinSmoother(attackSpeed, releaseSpeed, startValue float64) *LinSmoother {
	return &LinSmoother{attackSpeed: attackSpeed, releaseSpeed: releaseSpeed, y: startValue}
}

func (e *LinSmoother) InitAudio(p Params) {
	e.params = p
	e.SetAttackSpeed(e.attackSpeed)
	e.SetReleaseSpeed(e.releaseSpeed)
}

func (e *LinSmoother) Value() float64 {
	return e.y
}

func (e *LinSmoother) SetValue(v float64) {
	e.y = v
}

func (e *LinSmoother) SetAttackSpeed(s float64) {
	e.attackSpeed = s
	e.up = s / e.params.SampleRate
}

func (e *LinSmoother) SetReleaseSpeed(s float64) {
	e.releaseSpeed = s
	e.down = -s / e.params.SampleRate
}

func (e *LinSmoother) Smooth(x float64) float64 {
	e.x = x
	d := e.up
	if e.x < e.y {
		d = e.down
	}
	e.y += d
	return e.y
}

func (e *LinSmoother) Done() bool {
	return math.Abs(e.x-e.y) < .0001
}

type ExpSmoother struct {
	params                  Params
	attackTime, releaseTime float64
	up, down                float64
	x, y                    float64
}

func NewExpSmoother(attackTime, releaseTime float64) *ExpSmoother {
	return &ExpSmoother{attackTime: attackTime, releaseTime: releaseTime}
}

func (e *ExpSmoother) InitAudio(p Params) {
	e.params = p
	e.SetAttackTime(e.attackTime)
	e.SetReleaseTime(e.releaseTime)
}

func (e *ExpSmoother) SetAttackTime(t float64) {
	e.attackTime = t
	e.up = 1 - math.Pow(.01, 1/(e.params.SampleRate*t))
}

func (e *ExpSmoother) SetReleaseTime(t float64) {
	e.releaseTime = t
	e.down = 1 - math.Pow(.01, 1/(e.params.SampleRate*t))
}

func (e *ExpSmoother) Smooth(x float64) float64 {
	e.x = x
	a := e.up
	if e.x < e.y {
		a = e.down
	}
	e.y += a * (e.x - e.y)
	return e.y
}

func (e *ExpSmoother) Done() bool {
	return math.Abs(e.x-e.y) < .0001
}
