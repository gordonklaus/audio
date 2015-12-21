package audio

import "math"

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
