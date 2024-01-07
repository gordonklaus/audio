package audio

import "math"

type ExpEnv struct {
	p Params
	s []expSeg
	y float64
}

func (e *ExpEnv) InitAudio(p Params) {
	e.p = p
	Init(e.s, p)
}

func (e *ExpEnv) Go(x, t float64) *ExpEnv {
	s := expSeg{x: x, t: t}
	Init(&s, e.p)
	e.s = append(e.s, s)
	return e
}

func (e *ExpEnv) GoNow(x, t float64) *ExpEnv {
	e.s = nil
	e.Go(x, t)
	return e
}

func (e *ExpEnv) AttackHoldRelease(a, h, r float64) *ExpEnv {
	return e.Go(1, a).Go(1, h).Go(0, r)
}

func (e *ExpEnv) ReleaseNow(t float64) {
	e.s = nil
	e.Go(0, t)
}

func (e *ExpEnv) Sing() float64 {
	if len(e.s) > 0 {
		s := &e.s[0]
		e.y += s.do(e.y)
		if s.done() && len(e.s) > 1 {
			e.s = e.s[1:]
		}
	}
	return e.y
}

func (e *ExpEnv) Done() bool {
	return len(e.s) == 0 || len(e.s) == 1 && math.Abs(e.y-e.s[0].x) < .0001
}

type expSeg struct {
	x float64
	t float64
	n int
	a float64
}

func (s *expSeg) InitAudio(p Params) {
	n := p.SampleRate * s.t
	s.n = int(n) - 1
	s.a = 1 - math.Pow(.01, 1/n)
}

func (s *expSeg) do(y float64) float64 {
	s.n--
	return s.a * (s.x - y)
}

func (s *expSeg) done() bool {
	return s.n < 0
}
