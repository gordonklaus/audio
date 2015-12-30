package audio

import "testing"

func TestInit(t *testing.T) {
	var i audioIniter
	didPanic := false
	defer func() {
		if x := recover(); x != nil {
			didPanic = true
		}
	}()

	Init(i, Params{})
	if !didPanic {
		t.Error("expected panic")
	}
	if i.inited {
		t.Error("expected not inited")
	}

	Init(&i, Params{})
	if !i.inited {
		t.Error("expected inited")
	}
}

type audioIniter struct {
	inited bool
}

func (i *audioIniter) InitAudio(p Params) { i.inited = true }
