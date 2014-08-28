package audio

import "fmt"

type Control struct {
	params  Params
	points  []*ControlPoint
	periods []*controlPeriod
	x       float64
}

type ControlPoint struct {
	Time, Value float64
}

func NewControl(points []*ControlPoint) *Control {
	return &Control{points: points}
}

func (c *Control) InitAudio(params Params) {
	c.params = params
	c.SetPoints(c.points)
}

func (c *Control) SetPoints(points []*ControlPoint) {
	for i := range points {
		if i > 0 && points[i].Time < points[i-1].Time {
			fmt.Printf("control points out of order:  %#s\n", points)
			break
		}
	}
	c.points = points
	c.SetTime(0)
}

func (c *Control) SetTime(t float64) {
	c.periods = make([]*controlPeriod, len(c.points))
	prev := &ControlPoint{}
	for i, p := range c.points {
		dn := (p.Time - prev.Time) * c.params.SampleRate
		dx := (p.Value - prev.Value) / dn
		c.periods[i] = &controlPeriod{int(dn), dx, p.Value}
		prev = p
	}

	c.x = 0
	n := int(t * c.params.SampleRate)
	for _, p := range c.periods {
		if p.n > n {
			p.n -= n
			c.x += float64(n) * p.dx
			break
		}
		n -= p.n
		c.x = p.value
		c.periods = c.periods[1:]
	}
}

func (c *Control) Sing() float64 {
	for _, p := range c.periods {
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
