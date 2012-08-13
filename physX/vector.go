package physX

import "C"

type Vector [3]float32

func (v Vector) floatptr() *C.float { return (*C.float)(&v[0]) }

