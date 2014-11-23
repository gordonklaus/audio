package audiogui

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

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
		if <-sig == syscall.SIGQUIT {
			buf := make([]byte, 1<<10)
			for runtime.Stack(buf, true) == len(buf) {
				buf = make([]byte, 2*len(buf))
			}
			fmt.Fprintln(os.Stderr, string(buf))
		}
		for _, c := range playControls {
			c.Stop()
			<-c.Done
		}
		portaudio.Terminate()
	}()
}

func Play(v audio.Voice) {
	<-PlayAsync(v).Done
}

func PlayAsync(v audio.Voice) PlayControl {
	c := PlayControl{make(chan struct{}, 1), make(chan struct{}, 1)}
	params := audio.Params{SampleRate: 96000}
	audio.Init(v, params)
	s, err := portaudio.OpenDefaultStream(0, 1, params.SampleRate, 1024, func(out []float32) {
		for i := range out {
			out[i] = float32(v.Sing())
		}
		if v.Done() {
			c.Stop()
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

	go func() {
		<-c.stop
		if err := s.Close(); err != nil {
			fmt.Println(err)
		}
		c.Done <- struct{}{}
	}()
	playControls = append(playControls, c)
	return c
}

type PlayControl struct {
	stop, Done chan struct{}
}

func (c PlayControl) Stop() {
	select {
	case c.stop <- struct{}{}:
	default:
	}
}
