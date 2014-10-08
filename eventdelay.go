package audio

type EventDelay struct {
	Params Params
	events []*delayEvent
}

type delayEvent struct {
	n int
	f func()
}

func (d *EventDelay) Delay(t float64, f func()) {
	if d.Params.SampleRate == 0 {
		panic("EventDelay.Delay called before InitAudio")
	}
	n := int(t * d.Params.SampleRate)
	i := 0
	e := &delayEvent{}
	for i, e = range d.events {
		if n < e.n {
			break
		}
		n -= e.n
	}
	d.events = append(append(d.events[:i], &delayEvent{n, f}), d.events[i:]...)
	e.n -= n
}

func (d *EventDelay) Step() {
	for _, e := range d.events {
		e.n--
		if e.n > 0 {
			break
		}
		e.f()
		d.events = d.events[1:]
	}
}
