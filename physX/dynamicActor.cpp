#include <PxPhysicsAPI.h>
#include "dynamicActor.h"
#include "geometry.hpp"

using namespace physx;

#define A (*(PxRigidDynamic*)actor)

Transform DynamicActor_globalPose(void* actor) { return P2t(A.getGlobalPose()); }

void DynamicActor_addSphere(void* actor, Vector pos, float radius, void* material) {
	A.createShape(PxSphereGeometry(radius), *(PxMaterial*)material, PxTransform(v2P(pos)));
	PxRigidBodyExt::updateMassAndInertia(A, 1);
}

void DynamicActor_addCapsule(void* actor, Vector start, Vector end, float radius, void* material) {
	PxReal halfHeight;
	PxTransform pose = PxTransformFromSegment(v2P(start), v2P(end), &halfHeight);
	A.createShape(PxCapsuleGeometry(radius, halfHeight), *(PxMaterial*)material, pose);
	PxRigidBodyExt::updateMassAndInertia(A, 1);
}

