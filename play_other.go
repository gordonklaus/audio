// +build !android,!js

package audio

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gordonklaus/portaudio"
)

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

var stream *portaudio.Stream

func startPlaying(v Voice, c PlayControl) error {
	const sampleRate = 96000
	Init(v, Params{SampleRate: sampleRate})
	var err error
	stream, err = portaudio.OpenDefaultStream(0, 1, sampleRate, 1024, func(out []float32) {
		for i := range out {
			out[i] = float32(v.Sing())
		}
		if v.Done() {
			c.Stop()
		}
	})
	if err != nil {
		return err
	}
	return stream.Start()
}

func stopPlaying() error {
	return stream.Close()
}
