package audio

import "math"

type ExpEnv struct {
	p Params
	s []*expSeg
	y float64
}

func (e *ExpEnv) InitAudio(p Params) {
	e.p = p
	Init(e.s, p)
}

func (e *ExpEnv) Go(x, t float64) *ExpEnv {
	s := newExpEnv(x, t)
	Init(s, e.p)
	e.s = append(e.s, s)
	return e
}

func (e *ExpEnv) AttackHoldRelease(a, h, r float64) *ExpEnv {
	return e.Go(1, a).Go(1, h).Go(0, r)
}

func (e *ExpEnv) Release(t float64) {
	s := newExpEnv(0, t)
	Init(s, e.p)
	e.s = []*expSeg{s}
}

func (e *ExpEnv) Sing() float64 {
	if len(e.s) > 0 {
		s := e.s[0]
		e.y = s.do(e.y)
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
	b float64
}

func newExpEnv(x, t float64) *expSeg {
	return &expSeg{x: x, t: t}
}

func (s *expSeg) InitAudio(p Params) {
	n := p.SampleRate * s.t
	s.n = int(n) - 1
	s.b = math.Pow(.01, 1/n)
}

func (s *expSeg) do(x float64) float64 {
	s.n--
	return s.x - (s.x-x)*s.b
}

func (s *expSeg) done() bool {
	return s.n < 0
}
