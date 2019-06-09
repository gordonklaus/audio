package audio

import (
	"log"
)

var playControls []PlayControl

func Play(v interface{}) {
	switch v.(type) {
	case Voice, StereoVoice:
	default:
		panic("can only play Voice or StereoVoice")
	}
	<-PlayAsync(v).Done
}

func PlayAsync(v interface{}) PlayControl {
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
