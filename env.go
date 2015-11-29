package audio

import "math"

type AttackReleaseEnv struct {
	Params                  Params
	attackTime, releaseTime float64
	up, down                float64
	release                 bool
	x                       float64
}

func NewAttackReleaseEnv(attackTime, releaseTime float64) *AttackReleaseEnv {
	return &AttackReleaseEnv{attackTime: attackTime, releaseTime: releaseTime}
}

func (e *AttackReleaseEnv) InitAudio(p Params) {
	e.Params = p
	e.SetAttackTime(e.attackTime)
	e.SetReleaseTime(e.releaseTime)
}

func (e *AttackReleaseEnv) SetAttackTime(t float64) {
	e.attackTime = t
	e.up = math.Pow(.01, 1/(e.Params.SampleRate*t))
}

func (e *AttackReleaseEnv) SetReleaseTime(t float64) {
	e.releaseTime = t
	e.down = math.Pow(.01, 1/(e.Params.SampleRate*t))
}

func (e *AttackReleaseEnv) Attack()  { e.release = false }
func (e *AttackReleaseEnv) Release() { e.release = true }

func (e *AttackReleaseEnv) Sing() float64 {
	if e.release {
		e.x *= e.down
	} else {
		e.x = 1 - (1-e.x)*e.up
	}
	return e.x
}

func (e *AttackReleaseEnv) Done() bool {
	return e.release && e.x < .0001
}

type LinAttackReleaseEnv struct {
	Params                  Params
	attackTime, releaseTime float64
	up, down                float64
	release                 bool
	x                       float64
}

func NewLinAttackReleaseEnv(attackTime, releaseTime float64) *LinAttackReleaseEnv {
	return &LinAttackReleaseEnv{attackTime: attackTime, releaseTime: releaseTime}
}

func (e *LinAttackReleaseEnv) InitAudio(p Params) {
	e.Params = p
	e.SetAttackTime(e.attackTime)
	e.SetReleaseTime(e.releaseTime)
}

func (e *LinAttackReleaseEnv) SetAttackTime(t float64) {
	e.attackTime = t
	e.up = 1 / (e.Params.SampleRate * t)
}

func (e *LinAttackReleaseEnv) SetReleaseTime(t float64) {
	e.releaseTime = t
	e.down = 1 / (e.Params.SampleRate * t)
}

func (e *LinAttackReleaseEnv) Attack()  { e.release = false }
func (e *LinAttackReleaseEnv) Release() { e.release = true }

func (e *LinAttackReleaseEnv) Sing() float64 {
	if e.release {
		e.x = math.Max(0, e.x-e.down)
	} else {
		e.x = math.Min(1, e.x+e.up)
	}
	return e.x
}

func (e *LinAttackReleaseEnv) Done() bool {
	return e.release && e.x < .0001
}
