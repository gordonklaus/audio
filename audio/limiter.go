package audio

import "math"

// A soft limiter.  The RMS amplitude of the output will approach the supplied
// limit; this means that much of the signal will actually exceed the limit.
type Limiter struct {
	limit float64
	RMS   *RMS
	Delay *ConstDelay
}

func NewLimiter(limit, rmsWindowSize float64) *Limiter {
	return &Limiter{limit, NewRMS(rmsWindowSize), NewConstDelay(rmsWindowSize / 2)}
}

func (c *Limiter) Limit(x float64) float64 {
	c.RMS.Add(x)
	y := c.RMS.Amplitude() / c.limit
	return math.Tanh(y) / y * c.Delay.Delay(x)
}
