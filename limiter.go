package audio

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
	return Saturate(y) / y * c.Delay.Delay(x)
}

// Saturate is a cheap approximation of math.Tanh.
func Saturate(x float64) float64 {
	if x > 3 {
		return 1
	}
	if x < -3 {
		return -1
	}
	return x * (27 + x*x) / (27 + 9*x*x)
}
