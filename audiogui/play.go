package audiogui

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/portaudio-go/portaudio"
)

var playControls []PlayControl

func init() {
	portaudio.Initialize()
	go func() {
		defer os.Exit(0)
		sig := make(chan os.Signal, 1)
		signal.Notify(sig)
		<-sig
		for _, c := range playControls {
			c.Stop()
			c.Wait()
		}
		portaudio.Terminate()
	}()
}

func Play(v audio.Voice) {
	PlayAsync(v).Wait()
}

func PlayAsync(v audio.Voice) PlayControl {
	c := PlayControl{}
	done := make(chan struct{}, 1)
	params := audio.Params{SampleRate: 96000}
	audio.Init(v, params)
	s, err := portaudio.OpenDefaultStream(0, 1, params.SampleRate, 1024, func(out []float32) {
		defer func() {
			if x := recover(); x != nil {
				for i := range out {
					out[i] = 0
				}
				select {
				case done <- struct{}{}:
					fmt.Println("panic in stream callback:", x)
				default:
				}
			}
		}()
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
		return c
	}
	if err := s.Start(); err != nil {
		fmt.Println(err)
		return c
	}

	c.stop = make(chan struct{}, 1)
	c.wg = new(sync.WaitGroup)
	c.wg.Add(1)
	go func() {
		select {
		case <-done:
		case <-c.stop:
		}
		if err := s.Close(); err != nil {
			fmt.Println(err)
		}
		c.wg.Done()
	}()
	playControls = append(playControls, c)
	return c
}

type PlayControl struct {
	stop chan struct{}
	wg   *sync.WaitGroup
}

func (c PlayControl) Stop() {
	select {
	case c.stop <- struct{}{}:
	default:
	}
}

func (c PlayControl) Wait() {
	c.wg.Wait()
}

func (c PlayControl) WaitChan() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		c.wg.Wait()
		ch <- struct{}{}
	}()
	return ch
}
