#include <PxPhysicsAPI.h>
#include "joints.h"
#include "physics.hpp"
#include "geometry.hpp"

using namespace physx;

#define A1 ((PxRigidDynamic*)actor1)
#define A2 ((PxRigidDynamic*)actor2)

void newIsoUniversalJoint(void* actor1, Transform transform1, void* actor2, Transform transform2) {
	PxD6JointCreate(*physics, A1, t2P(transform1), A2, t2P(transform2));
}

