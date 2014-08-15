package audio

import "fmt"

type Control struct {
	points  []*ControlPoint
	periods []controlPeriod
	x       float64
}

type ControlPoint struct {
	Time, Value float64
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
	c.periods = make([]controlPeriod, len(c.points))
	prev := &ControlPoint{}
	for i, p := range c.points {
		dn := (p.Time - prev.Time) * params.SampleRate
		dx := (p.Value - prev.Value) / dn
		c.periods[i] = controlPeriod{int(dn), dx, p.Value}
		prev = p
	}
}

func (c *Control) Sing() float64 {
	for len(c.periods) > 0 {
		p := &c.periods[0]
		if p.n > 0 {
			p.n--
			c.x += p.dx
			break
		}
		c.x = p.value // this is necessary for zero-length controlPeriods that mark discontinuities
		c.periods = c.periods[1:]
	}
	return c.x
}

func (c *Control) Done() bool {
	return len(c.periods) == 0
}

type controlPeriod struct {
	n     int
	dx    float64
	value float64
}
