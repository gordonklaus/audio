package simulation

import (
	."code.google.com/p/gordon-go/physX"
)

type Body struct {
	a DynamicActor
	limbAttachmentPoints [4][3]Vector
}

func newBody(scene Scene, position Vector) Body {
	b := Body{scene.NewDynamicActor(Transform{position, IQ}), [4][3]Vector{}}
	const sideLength = 2
	ne := Vector{sideLength / 2, sideLength / 2}
	nw := Vector{-sideLength / 2, sideLength / 2}
	sw := Vector{-sideLength / 2, -sideLength / 2}
	se := Vector{sideLength / 2, -sideLength / 2}
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
//	actuator1, actuator2, actuator3 CylindricalJoint
//	controller1, controller2, controller3 JointController
}

func newTripartiteLimb(scene Scene, body Body, iLimb int, params [3]JointControllerParams) TripartiteLimb {
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

	NewIsoUniversalJointGlobal(body.a, l.thigh1.a, Transform{p[0], l.thigh1.a.GlobalPose().Orientation})
	NewIsoUniversalJointGlobal(body.a, l.thigh2.a, Transform{p[1], l.thigh2.a.GlobalPose().Orientation})
	NewIsoUniversalJointGlobal(body.a, l.thigh3.a, Transform{p[2], l.thigh3.a.GlobalPose().Orientation})

/*	actuator1 = new CylindricalJoint(scene, *thigh1, *calf1, thigh1->getEnd(), thigh1->getXAxis())
	actuator2 = new CylindricalJoint(scene, *thigh2, *calf2, thigh2->getEnd(), thigh2->getXAxis())
	actuator3 = new CylindricalJoint(scene, *thigh3, *calf3, thigh3->getEnd(), thigh3->getXAxis())

	addRevoluteJoint(scene, *calf1, *foot, footPosition, calf1->getZAxis())
	addRevoluteJoint(scene, *calf2, *foot, footPosition, calf2->getZAxis())
	addRevoluteJoint(scene, *calf3, *foot, footPosition, calf3->getZAxis())

	controller1.attackPhase = params[0]["attackPhase"]
	controller1.releasePhase = params[0]["releasePhase"]
	controller2.attackPhase = params[1]["attackPhase"]
	controller2.releasePhase = params[1]["releasePhase"]
	controller3.attackPhase = params[2]["attackPhase"]
	controller3.releasePhase = params[2]["releasePhase"]*/
	
	return l
}

/*CylindricalJoint& getActuator1() { return *actuator1 }
CylindricalJoint& getActuator2() { return *actuator2 }
CylindricalJoint& getActuator3() { return *actuator3 }

void update(double simulationTime) {
	actuator1->setTargetDistance(controller1.getPosition(simulationTime))
	actuator2->setTargetDistance(controller2.getPosition(simulationTime))
	actuator3->setTargetDistance(controller3.getPosition(simulationTime))

	actuator1->update(simulationTime)
	actuator2->update(simulationTime)
	actuator3->update(simulationTime)
}
*/

type Thigh struct {
	a DynamicActor
	start, end Vector
}

func newThigh(scene Scene, bodyAttachmentPoint, footPosition, limbCenter Vector) Thigh {
	t := Thigh{}
	t.a = scene.NewDynamicActor(TransformFromSegment(bodyAttachmentPoint, footPosition))
//formerly	t.a = scene.NewDynamicActor(bodyAttachmentPoint, AxisNormalOrientation(footPosition.Sub(bodyAttachmentPoint), limbCenter.Sub(bodyAttachmentPoint)))
	d := footPosition.Sub(bodyAttachmentPoint).Len()
	t.start = Vector{.3}
	t.end = Vector{d / 2}
	material := NewMaterial(1, .7, .5)
	t.a.AddCapsule(t.start, t.end, .1, material)
	return t
}

//	NxVec3 getStart() { return transformToGlobal(start) }
//	NxVec3 getEnd() { return transformToGlobal(end) }


type Calf struct { DynamicActor }
func newCalf(scene Scene, bodyAttachmentPoint, footPosition, limbCenter Vector) Calf {
	c := Calf{scene.NewDynamicActor(TransformFromSegment(bodyAttachmentPoint, footPosition))}
	d := footPosition.Sub(bodyAttachmentPoint).Len()
	material := NewMaterial(1, .7, .5)
	c.AddCapsule(Vector{d / 2 + .3}, Vector{d - .3}, .05, material)
	return c
}

type Foot struct { DynamicActor }
func newFoot(scene Scene, position, limbCenter Vector) Foot {
	f := Foot{scene.NewDynamicActor(TransformFromSegment(position, limbCenter))}
	material := NewMaterial(1, .7, .5)
	f.AddCapsule(Vector{}, Vector{.01}, .3, material)
	return f
}

/*struct JointController {
	JointController() : frequency(.3), attackPhase(0), releasePhase(0) {}

	double getPosition(double simulationTime) {
		double pulseWidth = releasePhase - attackPhase
		pulseWidth -= floor(pulseWidth)
		double phase = simulationTime * frequency - attackPhase
		phase -= floor(phase)
		return phase < pulseWidth
	}

	double frequency
	double attackPhase
	double releasePhase
}
*/
type ClimbingFish struct {
	body Body
	limbs [4]TripartiteLimb
}

type JointControllerParams struct { attackPhase, releasePhase float64 }

func NewClimbingFish(scene Scene, params [4][3]JointControllerParams) ClimbingFish {
	c := ClimbingFish{}
	c.body = newBody(scene, Vector{0, 0, 2})
	for i := 0; i < 4; i++ {
		c.limbs[i] = newTripartiteLimb(scene, c.body, i, params[i])
	}
	return c
}

/*
NxVec3 ClimbingFish::getPosition() { return body->getGlobalPose().t }
NxVec3 ClimbingFish::getFootPosition(int i) { return limbs[i]->foot->getGlobalPose().t }
NxVec3 ClimbingFish::getZAxis() { return body->getZAxis() }

void ClimbingFish::update(double simulationTime) {
	for(int i = 0 i < limbs.size() ++i)
		limbs[i]->update(simulationTime)
}
*/
