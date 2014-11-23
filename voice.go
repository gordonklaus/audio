package audio

type Voice interface {
	Sing() float64
	Done() bool
}

type MultiVoice struct {
	Params Params
	voices []Voice
}

func (m *MultiVoice) Add(v Voice) {
	Init(v, m.Params)
	m.voices = append(m.voices, v)
}

func (m *MultiVoice) Sing() float64 {
	x := 0.0
	for i, n := 0, len(m.voices); i < n; {
		v := m.voices[i]
		x += v.Sing()
		if v.Done() {
			n--
			m.voices[i] = m.voices[n]
			m.voices[n] = nil
			m.voices = m.voices[:n]
		} else {
			i++
		}
	}
	return x
}

func (m *MultiVoice) Done() bool {
	return len(m.voices) == 0
}

func (m *MultiVoice) Stop() {
	m.voices = nil
}
