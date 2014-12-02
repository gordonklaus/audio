package audio

func (c *Control) sing() (_ Audio, done bool) {
	i := 0
	max := len(c.Out)
	for c.i < len(c.Periods) && i < max {
		p := c.Periods[c.i]
		n := int((c.elapsed[c.i] - c.t) * c.Params.SampleRate)
		if n > max {
			n = max
		} else {
			c.i++
		}
		p.Control(c.Out[i:n])
		i = n
	}
	if i < max {
		x := c.Out[len(c.Out)-1]
		if i > 0 {
			x = c.Out[i-1]
		}
		for ; i < max; i++ {
			c.Out[i] = x
		}
	}
	c.t += float64(c.Params.BufferSize) / c.Params.SampleRate
	if c.i == len(c.Periods) {
		done = true
	}
	return c.Out, done
}
