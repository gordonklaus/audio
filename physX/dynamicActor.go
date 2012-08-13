package physX

// #cgo LDFLAGS: CphysX.so
// #include "dynamicActor.h"
import "C"

import "unsafe"

type DynamicActor struct { a unsafe.Pointer }
func (a DynamicActor) AddSphere(pos Vector, radius float32, m Material) { C.DynamicActor_addSphere(a.a, pos.floatptr(), C.float(radius), unsafe.Pointer(m)) }

