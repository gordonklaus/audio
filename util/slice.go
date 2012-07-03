package util

import ."reflect"

func SliceContains(slice interface{}, value interface{}) bool {
	s := ValueOf(slice)
	for i := 0; i < s.Len(); i++ {
		if DeepEqual(s.Index(i).Interface(), value) {
			return true
		}
	}
	return false
}

func SliceAdd(slicePtr interface{}, values ...interface{}) {
	s := ValueOf(slicePtr).Elem()
	for _, value := range values {
		s.Set(Append(s, ValueOf(value)))
	}
}

func SliceRemove(slicePtr interface{}, value interface{}) {
	s := ValueOf(slicePtr).Elem()
	for i := 0; i < s.Len(); i++ {
		if DeepEqual(s.Index(i).Interface(), value) {
			s.Set(AppendSlice(s.Slice(0, i), s.Slice(i + 1, s.Len())))
			return
		}
	}
}
