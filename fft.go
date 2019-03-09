package audio

import (
	"math"

	"github.com/ktye/fft"
)

type FFT struct {
	buf  [2]fftBuffer
	fft  fft.FFT
	filt func([]complex128)
	env  []float64
}

type fftBuffer struct {
	buf []complex128
	i   int
}

func NewFFT(size int, filter func([]complex128)) *FFT {
	env := make([]float64, size)
	for i := range env {
		env[i] = (1 - math.Cos(2*math.Pi*float64(i)/float64(size))) / 2
	}
	fft, err := fft.New(size)
	if err != nil {
		panic(err)
	}
	return &FFT{
		buf: [2]fftBuffer{
			{buf: make([]complex128, size)},
			{buf: make([]complex128, size), i: size / 2},
		},
		fft:  fft,
		filt: filter,
		env:  env,
	}
}

func (f *FFT) Filter(x float64) float64 {
	return f.filter(x, 0) + f.filter(x, 1)
}

func (f *FFT) filter(x float64, buf int) float64 {
	b := &f.buf[buf]
	y := real(b.buf[b.i]) * f.env[b.i]
	b.buf[b.i] = complex(x, 0)
	b.i++
	if b.i == len(b.buf) {
		b.i = 0
		b.buf = f.fft.Transform(b.buf)
		f.filt(b.buf)
		b.buf = f.fft.Inverse(b.buf)
	}
	return y
}
