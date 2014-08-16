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
	PlayAsync(v).Wait()
}

func PlayAsync(v audio.Voice) *PlayControl {
	done := make(chan struct{}, 1)
	params := audio.Params{SampleRate: 96000}
	audio.Init(v, params)
	s, err := portaudio.OpenDefaultStream(0, 1, params.SampleRate, 1024, func(out []float32) {
		for i := range out {
			out[i] = float32(v.Sing())
		}
		if v.Done() {
			select {
			case done <- struct{}{}:
			default:
			}
		}
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if err := s.Start(); err != nil {
		fmt.Println(err)
		return nil
	}

	Done := make(chan struct{})
	stop := make(chan struct{}, 1)
	go func() {
		select {
		case <-done:
		case <-stop:
		}
		if err := s.Close(); err != nil {
			fmt.Println(err)
		}
		Done <- struct{}{}
	}()
	return &PlayControl{Done, stop}
}

type PlayControl struct {
	Done, stop chan struct{}
}

func (s PlayControl) Stop() { s.stop <- struct{}{} }
func (s PlayControl) Wait() { <-s.Done }
