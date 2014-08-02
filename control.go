package audio

import "fmt"

type Control struct {
	Params  Params
	points  []*ControlPoint
	periods []controlPeriod
	x       float64
	Out     Audio
}

func NewControl(points []*ControlPoint) *Control {
	for i := range points {
		if i > 0 && points[i].Time < points[i-1].Time {
			fmt.Printf("control points out of order:  %#s\n", points)
			break
		}
	}
	return &Control{points: points}
}

func (c *Control) InitAudio(params Params) {
	c.Params = params
	c.periods = make([]controlPeriod, len(c.points))
	prev := &ControlPoint{}
	for i, p := range c.points {
		dn := (p.Time - prev.Time) * params.SampleRate
		dx := (p.Value - prev.Value) / dn
		c.periods[i] = controlPeriod{int(dn), dx, p.Value}
		prev = p
	}
	c.Out.InitAudio(params)
}

func (c *Control) Sing() (_ Audio, done bool) {
	i := 0
	n := len(c.Out)
	for len(c.periods) > 0 && i < n {
		p := &c.periods[0]
		for i < n && p.n > 0 {
			c.Out[i] = c.x
			c.x += p.dx
			i++
			p.n--
		}
		if p.n == 0 {
			c.x = p.value // this is necessary for zero-length controlPeriods that mark discontinuities (and may also help correct for rounding errors)
			c.periods = c.periods[1:]
		}
	}
	for ; i < n; i++ {
		c.Out[i] = c.x
	}
	return c.Out, len(c.periods) == 0
}

type ControlPoint struct {
	Time, Value float64
}

type controlPeriod struct {
	n     int
	dx    float64
	value float64
}
