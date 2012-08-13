#include <PxPhysicsAPI.h>
#include "dynamicActor.h"

using namespace physx;

void DynamicActor_addSphere(void* actor, float* pos, float radius, void* material) {
	PxRigidDynamic* a = (PxRigidDynamic*)actor;
	a->createShape(PxSphereGeometry(radius), *(PxMaterial*)material);
	PxRigidBodyExt::updateMassAndInertia(*a, 1);
}

