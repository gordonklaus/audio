package audio

type Voice interface {
	Sing() (_ Audio, done bool)
}

type MultiVoice struct {
	Params Params
	voices map[Voice]struct{}
	Out    Audio
}

func (m *MultiVoice) Add(v Voice) {
	Init(v, m.Params)
	if m.voices == nil {
		m.voices = map[Voice]struct{}{}
	}
	m.voices[v] = struct{}{}
}

func (m *MultiVoice) Sing() (Audio, bool) {
	m.Out.Zero()
	for v := range m.voices {
		a, done := v.Sing()
		m.Out.Add(m.Out, a)
		if done {
			delete(m.voices, v)
		}
	}
	return m.Out, len(m.voices) == 0
}
