#include <stdio.h>
#include <malloc.h>
#include <PxPhysicsAPI.h>
#include "physics.hpp"

void fatalError(const char* s) { printf("%s\n", s); exit(1); }

using namespace physx;

class Allocator : public PxAllocatorCallback {
public:
	void* allocate(size_t size, const char*, const char*, int) { return ::memalign(16, size); }
	void deallocate(void* ptr) { ::free(ptr); }
};

static PxDefaultErrorCallback gDefaultErrorCallback;
static Allocator gDefaultAllocatorCallback;

static PxFoundation* foundation;
PxPhysics* physics;

void Physics_init() {
	foundation = PxCreateFoundation(PX_PHYSICS_VERSION, gDefaultAllocatorCallback, gDefaultErrorCallback);
	if(!foundation)
		fatalError("PxCreateFoundation failed!");
	bool recordMemoryAllocations = true;
	PxProfileZoneManager* profileZoneManager = &PxProfileZoneManager::createProfileZoneManager(foundation);
	if(!profileZoneManager)
		fatalError("PxProfileZoneManager::createProfileZoneManager failed!");

	physics = PxCreatePhysics(PX_PHYSICS_VERSION, *foundation, PxTolerancesScale(), recordMemoryAllocations, profileZoneManager);
	if(!physics)
		fatalError("PxCreatePhysics failed!");
//	if(!PxInitExtensions(*physics))
//		fatalError("PxInitExtensions failed!");
}

void Physics_close() {
	physics->release();
	foundation->release();
}

void* Physics_newMaterial(float staticFriction, float dynamicFriction, float restitution) {
	return physics->createMaterial(staticFriction, dynamicFriction, restitution);
}

