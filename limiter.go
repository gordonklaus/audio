package audio

import "math"

// A soft limiter.  The RMS amplitude of the output (averaged over the attack
// time) will approach the supplied limit; this means that much of the signal
// will actually exceed the limit.
type Limiter struct {
	limit         float64
	attack, decay float64
	down, up      float64
	amp           float64
	rms           *RMS
	delay         *ConstDelay
}

func NewLimiter(limit, attack, decay float64) *Limiter {
	return &Limiter{limit: limit, attack: attack, decay: decay, rms: NewRMS(attack), delay: NewConstDelay(attack)}
}

func (c *Limiter) InitAudio(p Params) {
	c.down = -1 / (c.attack * p.SampleRate)
	c.up = 1 / (c.decay * p.SampleRate)
	c.rms.InitAudio(p)
	c.delay.InitAudio(p)
}

func (c *Limiter) Limit(x float64) float64 {
	gain := math.Exp2(c.amp)
	c.rms.Add(x)
	if y := c.rms.Amplitude() / c.limit; math.Tanh(y)/y < gain {
		c.amp += c.down
	} else {
		c.amp += c.up
	}
	return gain * c.delay.Delay(x)
}
