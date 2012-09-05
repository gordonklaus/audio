package physX

// #cgo LDFLAGS: CphysX.so
// #include "joints.h"
import "C"

func NewIsoUniversalJoint(actor1 DynamicActor, frame1 Transform, actor2 DynamicActor, frame2 Transform) { C.newIsoUniversalJoint(actor1.a, frame1.cTrans(), actor2.a, frame2.cTrans()) }

func NewIsoUniversalJointGlobal(actor1, actor2 DynamicActor, frame Transform) {
	NewIsoUniversalJoint(actor1, actor1.TransformFromGlobal(frame), actor2, actor2.TransformFromGlobal(frame))
}

