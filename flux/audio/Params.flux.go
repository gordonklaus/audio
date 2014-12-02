package audio

type Params struct{SampleRate float64; BufferSize int; visited map[interface{}]struct{}}