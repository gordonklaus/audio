package util

type Set map[interface{}]bool

func MakeSet(elements ...interface{}) Set {
	s := Set{}
	for _, x := range elements { s[x] = true }
	return s
}

func (s Set) Empty() bool { return len(s) == 0 }
func (s Set) Contains(x interface{}) bool { return s[x] }
func (s Set) Add(x interface{}) { s[x] = true }
func (s Set) Remove(x interface{}) { delete(s, x) }

func (s Set) Copy() Set {
	s2 := Set{}
	for x := range s { s2[x] = true }
	return s2
}