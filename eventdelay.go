package audio

type EventDelay struct {
	Params Params
	events []delayEvent
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
	for ; i < len(d.events); i++ {
		e := &d.events[i]
		if n < e.n {
			e.n -= n
			break
		}
		n -= e.n
	}
	d.events = append(d.events, delayEvent{})
	copy(d.events[i+1:], d.events[i:])
	d.events[i] = delayEvent{n, f}
}

func (d *EventDelay) Step() {
	if len(d.events) > 0 {
		d.events[0].n--
		for len(d.events) > 0 {
			e := &d.events[0]
			if e.n > 0 {
				break
			}
			e.f()
			d.events = d.events[1:]
		}
	}
}
