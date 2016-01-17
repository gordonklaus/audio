package audio

import (
	"log"
)

var playControls []PlayControl

func Play(v Voice) {
	<-PlayAsync(v).Done
}

func PlayAsync(v Voice) PlayControl {
	c := PlayControl{make(chan struct{}, 1), make(chan struct{}, 1)}
	if err := startPlaying(v, c); err != nil {
		log.Println(err)
		close(c.Done)
		return c
	}

	go func() {
		<-c.stop
		if err := stopPlaying(); err != nil {
			log.Println(err)
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
