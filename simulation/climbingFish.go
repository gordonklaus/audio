package simulation

import (
	."code.google.com/p/gordon-go/physX"
	."math"
)

type Body struct {
	a DynamicActor
	limbAttachmentPoints [4][3]Vector
}

func newBody(scene Scene, position Vector) Body {
	b := Body{scene.NewDynamicActor(Transform{position, IQ}), [4][3]Vector{}}
	const sideLength = 2
	ne := Vector{sideLength / 2, sideLength / 2, 0}
	nw := Vector{-sideLength / 2, sideLength / 2, 0}
	sw := Vector{-sideLength / 2, -sideLength / 2, 0}
	se := Vector{sideLength / 2, -sideLength / 2, 0}
	bottom := Vector{0, 0, -1}

	b.limbAttachmentPoints = [4][3]Vector{{bottom, ne, se}, {bottom, nw, ne}, {bottom, sw, nw}, {bottom, se, sw}}

	material := NewMaterial(1, .7, .5)
	for _, limb := range b.limbAttachmentPoints {
		b.a.AddCapsule(limb[0], limb[1], .1, material)
		b.a.AddCapsule(limb[1], limb[2], .1, material)
	}
	return b
}

func (b Body) getLimbAttachmentPoints(iLimb int) (p [3]Vector) {
	for i := 0; i < 3; i++ {
		p[i] = b.a.VectorToGlobal(b.limbAttachmentPoints[iLimb][i])
	}
	return
}


type TripartiteLimb struct {
	thigh1, thigh2, thigh3 Thigh
	calf1, calf2, calf3 Calf
	foot Foot
	actuator1, actuator2, actuator3 CylindricalJoint
	controller1, controller2, controller3 JointController
}

func newTripartiteLimb(scene Scene, body Body, iLimb int, params [3]JointControllerParam) TripartiteLimb {
	l := TripartiteLimb{}
	p := body.getLimbAttachmentPoints(iLimb)

	center := p[0].Add(p[1]).Add(p[2]).Div(3)
	normal := p[1].Sub(p[0]).Cross(p[2].Sub(p[1])); normal.Normalize()
	normal = normal.Mul((p[1].Sub(p[0]).Len() + p[2].Sub(p[1]).Len() + p[0].Sub(p[2]).Len()) / 3)
	footPosition := center.Add(normal)

	l.thigh1 = newThigh(scene, p[0], footPosition, center)
	l.thigh2 = newThigh(scene, p[1], footPosition, center)
	l.thigh3 = newThigh(scene, p[2], footPosition, center)
	l.calf1 = newCalf(scene, p[0], footPosition, center)
	l.calf2 = newCalf(scene, p[1], footPosition, center)
	l.calf3 = newCalf(scene, p[2], footPosition, center)
	l.foot = newFoot(scene, footPosition, center)

	NewIsoUniversalJointGlobal(body.a, l.thigh1.a, l.thigh1.a.GlobalPose())
	NewIsoUniversalJointGlobal(body.a, l.thigh2.a, l.thigh2.a.GlobalPose())
	NewIsoUniversalJointGlobal(body.a, l.thigh3.a, l.thigh3.a.GlobalPose())

	l.actuator1 = NewCylindricalJointGlobal(l.thigh1.a, l.calf1.a, Transform{l.thigh1.End(), l.thigh1.a.Orientation()})
	l.actuator2 = NewCylindricalJointGlobal(l.thigh2.a, l.calf2.a, Transform{l.thigh2.End(), l.thigh2.a.Orientation()})
	l.actuator3 = NewCylindricalJointGlobal(l.thigh3.a, l.calf3.a, Transform{l.thigh3.End(), l.thigh3.a.Orientation()})
	
	const x = MaxFloat32
	l.actuator1.SetDrive(x, x, x)
	l.actuator2.SetDrive(x, x, x)
	l.actuator3.SetDrive(x, x, x)

	NewRevoluteJointGlobal(l.calf1.a, l.foot.a, Transform{footPosition, AxisOrientation(l.calf1.a.ZAxis())})
	NewRevoluteJointGlobal(l.calf2.a, l.foot.a, Transform{footPosition, AxisOrientation(l.calf2.a.ZAxis())})
	NewRevoluteJointGlobal(l.calf3.a, l.foot.a, Transform{footPosition, AxisOrientation(l.calf3.a.ZAxis())})

	l.controller1.p = params[0]
	l.controller2.p = params[1]
	l.controller3.p = params[2]
	
	return l
}

func (l TripartiteLimb) update(simulationTime float64) {
	l.actuator1.SetDriveLength(l.controller1.position(simulationTime))
	l.actuator2.SetDriveLength(l.controller2.position(simulationTime))
	l.actuator3.SetDriveLength(l.controller3.position(simulationTime))
}

type Thigh struct {
	a DynamicActor
	start, end Vector
}

func newThigh(scene Scene, bodyAttachmentPoint, footPosition, limbCenter Vector) Thigh {
	t := Thigh{}
	t.a = scene.NewDynamicActor(Transform{bodyAttachmentPoint, AxisNormalOrientation(footPosition.Sub(bodyAttachmentPoint), limbCenter.Sub(bodyAttachmentPoint))})
	d := footPosition.Sub(bodyAttachmentPoint).Len()
	t.start = Vector{.3, 0, 0}
	t.end = Vector{d / 2, 0, 0}
	material := NewMaterial(1, .7, .5)
	t.a.AddCapsule(t.start, t.end, .1, material)
	return t
}

func (t Thigh) Start() Vector { return t.a.VectorToGlobal(t.start) }
func (t Thigh) End() Vector { return t.a.VectorToGlobal(t.end) }


type Calf struct { a DynamicActor }
func newCalf(scene Scene, bodyAttachmentPoint, footPosition, limbCenter Vector) Calf {
	c := Calf{scene.NewDynamicActor(Transform{bodyAttachmentPoint, AxisOrientation(footPosition.Sub(bodyAttachmentPoint))})}
	d := footPosition.Sub(bodyAttachmentPoint).Len()
	material := NewMaterial(1, .7, .5)
	c.a.AddCapsule(Vector{d / 2 + .3, 0, 0}, Vector{d - .3, 0, 0}, .05, material)
	return c
}

type Foot struct { a DynamicActor }
func newFoot(scene Scene, position, limbCenter Vector) Foot {
	f := Foot{scene.NewDynamicActor(Transform{position, AxisOrientation(limbCenter.Sub(position))})}
	material := NewMaterial(1, .7, .5)
	f.a.AddCapsule(Vector{}, Vector{.01, 0, 0}, .3, material)
	return f
}

type JointController struct { p JointControllerParam }
const frequency = .3
func (c JointController) position(simulationTime float64) float64 {
	pulseWidth := c.p.ReleasePhase - c.p.AttackPhase
	pulseWidth -= Floor(pulseWidth)
	phase := simulationTime * frequency - c.p.AttackPhase
	phase -= Floor(phase)
	if phase < pulseWidth {
		return 1
	}
	return 0
}

type ClimbingFish struct {
	body Body
	limbs [4]TripartiteLimb
}

type JointControllerParam struct { AttackPhase, ReleasePhase float64 }

func NewClimbingFish(scene Scene, params [4][3]JointControllerParam) ClimbingFish {
	c := ClimbingFish{}
	c.body = newBody(scene, Vector{0, 0, 2})
	for i := 0; i < 4; i++ {
		c.limbs[i] = newTripartiteLimb(scene, c.body, i, params[i])
	}
	return c
}

/*
NxVec3 ClimbingFish::getPosition() { return body.getGlobalPose().t }
NxVec3 ClimbingFish::getFootPosition(int i) { return limbs[i].foot.getGlobalPose().t }
NxVec3 ClimbingFish::getZAxis() { return body.getZAxis() }
*/

func (f ClimbingFish) Update(simulationTime float64) {
	for _, limb := range f.limbs {
		limb.update(simulationTime)
	}
}

