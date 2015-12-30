package audio

import (
	"fmt"
	"reflect"
)

type Initer interface {
	InitAudio(Params)
}

type Params struct {
	SampleRate float64
}

func (p *Params) InitAudio(q Params) { *p = q }

func Init(x interface{}, p Params) {
	if err := initVal(reflect.ValueOf(x), p); err != nil {
		panic("audio.Init: " + err.Error())
	}
}

var initerType = reflect.TypeOf(new(Initer)).Elem()

func initVal(v reflect.Value, p Params) (err error) {
	if v.Kind() == reflect.Ptr && v.IsNil() || !v.CanInterface() {
		return
	}

	v = reflect.Indirect(v)
	if v.CanAddr() && v.Type().Name() != "" && v.Kind() != reflect.Interface {
		v = v.Addr()
	}
	if x, ok := v.Interface().(Initer); ok {
		x.InitAudio(p)
		return
	}

	defer func() {
		if err != nil {
			// append v to the Init stack trace
			err = fmt.Errorf("%s\n\t%#v", err, v)
		}
	}()
	if t := v.Type(); reflect.PtrTo(t).Implements(initerType) {
		return fmt.Errorf("%s does not implement audio.Initer but *%s does.\nInit stack:", t, t)
	}

	v = reflect.Indirect(v)
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err = initVal(v.Field(i), p); err != nil {
				return
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err = initVal(v.Index(i), p); err != nil {
				return
			}
		}
	}

	return
}
