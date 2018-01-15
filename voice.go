package audio

type Voice interface {
	Sing() float64
	Done() bool
}

type MultiVoice struct {
	Params Params
	Voices []Voice
}

func (m *MultiVoice) Add(v Voice) {
	Init(v, m.Params)
	m.Voices = append(m.Voices, v)
}

func (m *MultiVoice) Sing() float64 {
	x := 0.0
	for i, n := 0, len(m.Voices); i < n; {
		v := m.Voices[i]
		x += v.Sing()
		if v.Done() {
			n--
			m.Voices[i] = m.Voices[n]
			m.Voices[n] = nil
			m.Voices = m.Voices[:n]
		} else {
			i++
		}
	}
	return x
}

func (m *MultiVoice) Done() bool {
	return len(m.Voices) == 0
}

func (m *MultiVoice) Stop() {
	m.Voices = nil
}
