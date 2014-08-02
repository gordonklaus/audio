package audiogui

import (
	"fmt"

	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/portaudio-go/portaudio"
)

func init() {
	portaudio.Initialize()
}

func Play(v audio.Voice) {
	portaudio.Initialize()
	defer portaudio.Terminate()
	stop := make(chan bool, 1)
	params := audio.Params{SampleRate: 96000, BufferSize: 64}
	audio.Init(v, params)
	s, err := portaudio.OpenDefaultStream(0, 1, params.SampleRate, params.BufferSize, func(in [][]float32, out [][]float32) {
		a, done := v.Sing()
		for i := range out[0] {
			out[0][i] = float32(a[i])
		}
		if done {
			select {
			case stop <- true:
			default:
			}
		}
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := s.Start(); err != nil {
		fmt.Println(err)
		return
	}
	<-stop
	if err := s.Stop(); err != nil {
		fmt.Println(err)
		return
	}
}
