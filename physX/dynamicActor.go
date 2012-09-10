package physX

// #cgo LDFLAGS: CphysX.so
// #include "dynamicActor.h"
import "C"

import "unsafe"

type DynamicActor struct { a unsafe.Pointer }
func (a DynamicActor) GlobalPose() Transform { return C2t(C.DynamicActor_globalPose(a.a)) }
func (a DynamicActor) Position() Vector { return a.GlobalPose().Position }
func (a DynamicActor) Orientation() Quaternion { return a.GlobalPose().Orientation }
func (a DynamicActor) VectorToGlobal(v Vector) Vector { return a.GlobalPose().TransformVector(v) }
func (a DynamicActor) TransformToGlobal(t Transform) Transform { return a.GlobalPose().Transform(t) }
func (a DynamicActor) TransformFromGlobal(t Transform) Transform { return a.GlobalPose().TransformInv(t) }
func (a DynamicActor) XAxis() Vector { return a.GlobalPose().Orientation.XAxis() }
func (a DynamicActor) YAxis() Vector { return a.GlobalPose().Orientation.YAxis() }
func (a DynamicActor) ZAxis() Vector { return a.GlobalPose().Orientation.ZAxis() }
func (a DynamicActor) AddSphere(pos Vector, radius float32, m Material) { C.DynamicActor_addSphere(a.a, pos.c(), C.float(radius), unsafe.Pointer(m)) }
func (a DynamicActor) AddCapsule(start, end Vector, radius float32, m Material) { C.DynamicActor_addCapsule(a.a, start.c(), end.c(), C.float(radius), unsafe.Pointer(m)) }

