package audiogui

import (
	"code.google.com/p/gordon-go/audio"
	"github.com/oov/audio/wave"

	"os"
)

func Write(v audio.Voice, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	params := audio.Params{96000}
	const channels = 1
	const bitsPerSample = 32
	w, err := wave.NewWriter(f, &wave.WaveFormatExtensible{Format: wave.WaveFormatEx{
		FormatTag:      wave.WAVE_FORMAT_IEEE_FLOAT,
		Channels:       channels,
		SamplesPerSec:  uint32(params.SampleRate),
		BitsPerSample:  bitsPerSample,
		AvgBytesPerSec: uint32(params.SampleRate * channels * bitsPerSample / 8),
		BlockAlign:     uint16(channels * bitsPerSample / 8),
	}})
	if err != nil {
		panic(err)
	}
	defer w.Close()

	audio.Init(v, params)
	for !v.Done() {
		n, err := w.WriteFloat64Interleaved([][]float64{{v.Sing()}})
		if err != nil {
			panic(err)
		}
		if n != 1 {
			panic("short write")
		}
	}
}
