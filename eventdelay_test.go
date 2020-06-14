package audio

import "testing"

func TestEventDelay(t *testing.T) {
	var d EventDelay
	Init(&d, Params{SampleRate: 1})

	var e, e2 delayedEvent
	d.Delay(4, e.f)
	e.test(t, &d, 4)

	e = false
	e2 = false
	d.Delay(4, e.f)
	d.Delay(8, e2.f)
	e.test(t, &d, 4)
	e2.test(t, &d, 4)

	e = false
	e2 = false
	d.Delay(8, e2.f)
	d.Delay(4, e.f)
	e.test(t, &d, 4)
	e2.test(t, &d, 4)

	e = false
	e2 = false
	d.Delay(4, e.f)
	d.Delay(4, e2.f)
	for i := 0; i < 4; i++ {
		if e == true {
			t.Fatalf("true before %d", i)
		}
		if e2 == true {
			t.Fatalf("e2 true before %d", i)
		}
		d.Step()
	}
	if e == false {
		t.Fatalf("false after %d", 4)
	}
	if e2 == false {
		t.Fatalf("e2 false after %d", 4)
	}
}

type delayedEvent bool

func (e *delayedEvent) f() { *e = true }

func (e *delayedEvent) test(t *testing.T, d *EventDelay, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if *e == true {
			t.Fatalf("true before %d", i)
		}
		d.Step()
	}
	if *e == false {
		t.Fatalf("false after %d", n)
	}
}
