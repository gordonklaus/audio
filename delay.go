package audio

import "math"

type Delay struct {
	Params Params
	x      []float64
	i      int
}

func (d *Delay) Read(t float64) float64 {
	i_, f := math.Modf(t * d.Params.SampleRate)
	i := int(i_)
	return Interp3(f, d.ReadSample(i-1), d.ReadSample(i), d.ReadSample(i+1), d.ReadSample(i+2))
}

func (d *Delay) ReadSample(i int) float64 {
	for i >= len(d.x)/2 {
		if len(d.x) == 0 {
			d.x = make([]float64, 1)
		} else {
			d.x = append(d.x, d.x...)
		}
	}
	i = d.i - i
	if i < 0 {
		i += len(d.x)
	}
	return d.x[i]
}

func (d *Delay) Write(x float64) {
	d.x[d.i] = x
	d.i++
	if d.i == len(d.x) {
		d.i = 0
	}
}

// Hermite cubic interpolation between x1 and x2 (t=0..1).
func Interp3(t, x0, x1, x2, x3 float64) float64 {
	c0 := x1
	c1 := (x2 - x0) / 2
	c2 := x0 - 2.5*x1 + 2*x2 - x3/2
	c3 := 1.5*(x1-x2) + (x3-x0)/2
	return c0 + t*(c1+t*(c2+t*c3))
}

type ConstDelay struct {
	t float64
	x []float64
	i int
}

func NewConstDelay(t float64) *ConstDelay {
	return &ConstDelay{t: t}
}

func (d *ConstDelay) InitAudio(p Params) {
	d.x = make([]float64, int(d.t*p.SampleRate))
	d.i = 0
}

func (d *ConstDelay) Delay(x float64) float64 {
	y := d.x[d.i]
	d.x[d.i] = x
	d.i++
	if d.i == len(d.x) {
		d.i = 0
	}
	return y
}
