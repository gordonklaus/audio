package audio

type ConstDelay struct {
	delay float64
	buf   []float64
	i     int
}

func NewConstDelay(delay float64) *ConstDelay {
	return &ConstDelay{delay: delay}
}

func (d *ConstDelay) InitAudio(p Params) {
	d.buf = make([]float64, int(d.delay*p.SampleRate))
	d.i = 0
}

func (d *ConstDelay) Delay(x float64) float64 {
	y := d.buf[d.i]
	d.buf[d.i] = x
	d.i = (d.i + 1) % len(d.buf)
	return y
}
