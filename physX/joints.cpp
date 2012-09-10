#include <PxPhysicsAPI.h>
#include "joints.h"
#include "physics.hpp"
#include "geometry.hpp"

using namespace physx;

#define A1 ((PxRigidDynamic*)actor1)
#define A2 ((PxRigidDynamic*)actor2)
#define J (*(PxD6Joint*)j)

void* newJoint(void* actor1, Transform transform1, void* actor2, Transform transform2) {
	return PxD6JointCreate(*physics, A1, t2P(transform1), A2, t2P(transform2));
}

void Joint_enableCollision(void* j) {
	J.setConstraintFlag(PxConstraintFlag::eCOLLISION_ENABLED, true);
}

void Joint_setMotion(void* j, int axis, int type) {
	J.setMotion(PxD6Axis::Enum(axis), PxD6Motion::Enum(type));
}

void Joint_setDrive(void* j, int index, float spring, float damping, float forceLimit) {
	J.setDrive(PxD6Drive::Enum(index), PxD6JointDrive(spring, damping, forceLimit));
}

void Joint_setDrivePose(void* j, Transform pose) {
	J.setDrivePosition(t2P(pose));
}

void Joint_setDriveVelocity(void* j, Vector linear, Vector angular) {
	J.setDriveVelocity(v2P(linear), v2P(angular));
}

