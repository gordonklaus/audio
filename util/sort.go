package util

import (
	"sort"
	."reflect"
	."fmt"
)

// data and SortBy taken from go/doc
type data struct {
	n int
	less func(i, j int) bool
	swap func(i, j int)
}
func (d *data) Len() int { return d.n }
func (d *data) Less(i, j int) bool { return d.less(i, j) }
func (d *data) Swap(i, j int) { d.swap(i, j) }

func SortBy(n int, less func(i, j int) bool, swap func(i, j int)) {
	sort.Sort(&data{n, less, swap})
}

func Sort(slice interface{}, key interface{}) {
	s, k := ValueOf(slice), ValueOf(key)
	if s.Kind() != Slice && s.Kind() != Array { panic(Sprintf("Can't sort '%#v'.", slice)) }
	elemType := s.Type().Elem()
	
	var less func(i, j int) bool
	switch k.Kind() {
	case String:
		if method, ok := elemType.MethodByName(key.(string)); ok && method.Type.NumIn() == 1 && method.Type.NumOut() >= 1 {
			call := func(i int) Value { return method.Func.Call([]Value{s.Index(i)})[0] }
			switch method.Type.Out(0).Kind() {
			case Bool:
				less = func(i, j int) bool { return !call(i).Bool() && call(j).Bool() }
			case Int, Int8, Int16, Int32, Int64:
				less = func(i, j int) bool { return call(i).Int() < call(j).Int() }
			case Uint, Uint8, Uint16, Uint32, Uint64:
				less = func(i, j int) bool { return call(i).Uint() < call(j).Uint() }
			case Float32, Float64:
				less = func(i, j int) bool { return call(i).Float() < call(j).Float() }
			case String:
				less = func(i, j int) bool { return call(i).String() < call(j).String() }
			}
		}
	default:
		panic(Sprintf("Can't sort using key '%#v'.", key))
	}
	
	tmp := New(elemType).Elem()
	SortBy(s.Len(), less, func(i, j int) {
		tmp.Set(s.Index(i))
		s.Index(i).Set(s.Index(j))
		s.Index(j).Set(tmp)
	})
}
