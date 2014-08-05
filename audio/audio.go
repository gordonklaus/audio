package audio

import "math"

type Audio []float64

func (a *Audio) InitAudio(p Params) {
	*a = make(Audio, p.BufferSize)
}

func (z Audio) Zero() Audio {
	for i := range z {
		z[i] = 0
	}
	return z
}

func (z Audio) Add(x Audio, y Audio) Audio {
	for i := range z {
		z[i] = x[i] + y[i]
	}
	return z
}

func (z Audio) Sub(x Audio, y Audio) Audio {
	for i := range z {
		z[i] = x[i] - y[i]
	}
	return z
}

func (z Audio) Mul(x Audio, y Audio) Audio {
	for i := range z {
		z[i] = x[i] * y[i]
	}
	return z
}

func (z Audio) Div(x Audio, y Audio) Audio {
	for i := range z {
		z[i] = x[i] / y[i]
	}
	return z
}

func (z Audio) AddX(x Audio, f float64) Audio {
	for i := range z {
		z[i] = x[i] + f
	}
	return z
}

func (z Audio) MulX(x Audio, f float64) Audio {
	for i := range z {
		z[i] = x[i] * f
	}
	return z
}

func (z Audio) Pow2(x Audio) Audio {
	for i := range z {
		z[i] = math.Pow(2, x[i])
	}
	return z
}

func (z Audio) Tanh(x Audio) Audio {
	for i := range z {
		z[i] = math.Tanh(x[i])
	}
	return z
}

func Add(x Audio, y Audio) Audio { return make(Audio, len(x)).Add(x, y) }
func Sub(x Audio, y Audio) Audio { return make(Audio, len(x)).Sub(x, y) }
func Mul(x Audio, y Audio) Audio { return make(Audio, len(x)).Mul(x, y) }
func Div(x Audio, y Audio) Audio { return make(Audio, len(x)).Div(x, y) }
