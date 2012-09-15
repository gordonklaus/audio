package physX

// #cgo LDFLAGS: CphysX.so
// #include "joints.h"
import "C"

import (
	"unsafe"
	."math"
)

type JointAxis int
const (
	XAxis = JointAxis(iota)
	YAxis
	ZAxis
	Twist
	Swing1
	Swing2
)

type JointMotion int
const (
	Locked = JointMotion(iota)
	Limited
	Free
)

type Drive int
const (
	XDrive = Drive(iota)
	YDrive
	ZDrive
	SwingDrive
	TwistDrive
	SlerpDrive
)

type Joint struct {
	j unsafe.Pointer
	actor1, actor2 DynamicActor
	frame1, frame2 Transform
}

func NewJoint(actor1 DynamicActor, frame1 Transform, actor2 DynamicActor, frame2 Transform) Joint {
	return Joint{C.newJoint(actor1.a, frame1.c(), actor2.a, frame2.c()), actor1, actor2, frame1, frame2}
}

func NewJointGlobal(actor1, actor2 DynamicActor, frame Transform) Joint {
	return NewJoint(actor1, actor1.TransformFromGlobal(frame), actor2, actor2.TransformFromGlobal(frame))
}

func (j Joint) EnableCollision() {
	C.Joint_enableCollision(j.j)
}
func (j Joint) SetMotion(axis JointAxis, typ JointMotion) {
	C.Joint_setMotion(j.j, C.int(axis), C.int(typ))
}
func (j Joint) SetDrive(drive Drive, spring, damping, forceLimit float64) {
	C.Joint_setDrive(j.j, C.int(drive), C.float(spring), C.float(damping), C.float(forceLimit))
}
func (j Joint) SetDrivePose(pose Transform) {
	C.Joint_setDrivePose(j.j, pose.c())
}
func (j Joint) SetDriveVelocity(linear, angular Vector) {
	C.Joint_setDriveVelocity(j.j, linear.c(), angular.c())
}

func NewRevoluteJoint(actor1 DynamicActor, frame1 Transform, actor2 DynamicActor, frame2 Transform) Joint {
	j := NewJoint(actor1, frame1, actor2, frame2)
	j.SetMotion(Twist, Free)
	return j
}
func NewRevoluteJointGlobal(actor1, actor2 DynamicActor, frame Transform) Joint {
	return NewRevoluteJoint(actor1, actor1.TransformFromGlobal(frame), actor2, actor2.TransformFromGlobal(frame))
}

type CylindricalJoint struct { Joint }
func NewCylindricalJoint(actor1 DynamicActor, frame1 Transform, actor2 DynamicActor, frame2 Transform) CylindricalJoint {
	j := CylindricalJoint{NewJoint(actor1, frame1, actor2, frame2)}
	j.SetMotion(XAxis, Free)
	j.SetMotion(Twist, Free)
	return j
}
func NewCylindricalJointGlobal(actor1, actor2 DynamicActor, frame Transform) CylindricalJoint {
	j := NewCylindricalJoint(actor1, actor1.TransformFromGlobal(frame), actor2, actor2.TransformFromGlobal(frame))
	j.EnableCollision()
	return j
}
func (j CylindricalJoint) SetDrive(spring, damping, forceLimit float64) {
	j.Joint.SetDrive(XDrive, spring, damping, forceLimit)
}
func (j CylindricalJoint) SetDriveLength(x float64) {
	pos2 := j.actor2.VectorToGlobal(j.frame2.Position)
	frame1 := j.actor1.TransformToGlobal(j.frame1)
	d := frame1.TransformVectorInv(pos2).X
	
	j.SetDriveVelocity(Vector{Copysign(1, d - x), 0, 0}, Vector{})
	j.SetDrivePose(Transform{Vector{x, 0, 0}, IQ})
}

func NewIsoUniversalJoint(actor1 DynamicActor, frame1 Transform, actor2 DynamicActor, frame2 Transform) Joint {
	j := NewJoint(actor1, frame1, actor2, frame2)
	j.SetMotion(Swing1, Free)
	j.SetMotion(Swing2, Free)
	return j
}
func NewIsoUniversalJointGlobal(actor1, actor2 DynamicActor, frame Transform) Joint {
	return NewIsoUniversalJoint(actor1, actor1.TransformFromGlobal(frame), actor2, actor2.TransformFromGlobal(frame))
}

