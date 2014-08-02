package audio

import "reflect"

type AudioIniter interface {
	InitAudio(Params)
}

type Params struct {
	SampleRate float64
	BufferSize int
	visited    map[interface{}]struct{}
}

func (p *Params) InitAudio(q Params) { *p = q }

func Init(x interface{}, p Params) {
	if p.visited == nil {
		p.visited = map[interface{}]struct{}{}
	}
	if _, ok := p.visited[x]; ok {
		return
	}
	p.visited[x] = struct{}{}

	if x, ok := x.(AudioIniter); ok {
		x.InitAudio(p)
		return
	}

	initVal := func(v reflect.Value) {
		if v.CanAddr() && v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
			v = v.Addr()
		}
		if v.CanInterface() {
			Init(v.Interface(), p)
		}
	}
	v := reflect.Indirect(reflect.ValueOf(x))
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			initVal(v.Field(i))
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			initVal(v.Index(i))
		}
	}
}
