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
	params := audio.Params{SampleRate: 96000}
	audio.Init(v, params)
	s, err := portaudio.OpenDefaultStream(0, 1, params.SampleRate, 64, func(out []float32) {
		for i := range out {
			out[i] = float32(v.Sing())
			if v.Done() {
				select {
				case stop <- true:
				default:
				}
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
