package physX

// #cgo LDFLAGS: CphysX.so
// #include "physics.h"
import "C"

import "unsafe"

func Init() { C.Physics_init() }
func Close() { C.Physics_close() }

type Material unsafe.Pointer
func NewMaterial(staticFriction, dynamicFriction, restitution float32) Material { return Material(C.Physics_newMaterial(C.float(staticFriction), C.float(dynamicFriction), C.float(restitution))) }

