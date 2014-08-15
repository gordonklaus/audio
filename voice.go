package audio

type Voice interface {
	Sing() float64
	Done() bool
}

type MultiVoice struct {
	Params Params
	voices map[Voice]struct{}
}

func (m *MultiVoice) Add(v Voice) {
	Init(v, m.Params)
	if m.voices == nil {
		m.voices = map[Voice]struct{}{}
	}
	m.voices[v] = struct{}{}
}

func (m *MultiVoice) Sing() float64 {
	x := 0.0
	for v := range m.voices {
		x += v.Sing()
		if v.Done() {
			delete(m.voices, v)
		}
	}
	return x
}

func (m *MultiVoice) Done() bool {
	return len(m.voices) == 0
}
