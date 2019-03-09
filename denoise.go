package audio

import (
	"math"
	"math/cmplx"
	"sync"
)

type Denoiser struct {
	fft *FFT

	mu    sync.Mutex
	noise []float64

	collecting   []complex128
	collected    []float64
	numCollected int
}

func NewDenoiser() *Denoiser {
	const size = 512

	d := &Denoiser{
		noise:      make([]float64, size),
		collecting: make([]complex128, 0, size),
		collected:  make([]float64, size),
	}
	d.fft = NewFFT(size, d.denoise)
	return d
}

func (d *Denoiser) Filter(x float64) float64 {
	const numToCollect = 256

	if d.numCollected < numToCollect {
		d.collecting = append(d.collecting, complex(x, 0))
		if len(d.collecting) == len(d.collected) {
			d.collecting = d.fft.fft.Transform(d.collecting)
			for i, x := range d.collecting {
				d.collected[i] += cmplx.Abs(x)
			}

			d.collecting = d.collecting[:0]
			d.numCollected++

			if d.numCollected == numToCollect {
				for i := range d.collected {
					d.collected[i] = math.Pow(d.collected[i]/numToCollect, 2)
				}
				d.mu.Lock()
				d.noise = d.collected
				d.mu.Unlock()
			}
		}
	}

	return d.fft.Filter(x)
}

func (d *Denoiser) denoise(x []complex128) {
	d.mu.Lock()
	noise := d.noise
	d.mu.Unlock()

	for i := range x {
		s := real(x[i])*real(x[i]) + imag(x[i])*imag(x[i])
		if s != 0 {
			x[i] *= complex(math.Sqrt(math.Max(0, 1-noise[i]/s)), 0)
		}
	}
}
