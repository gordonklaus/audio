#include <PxPhysicsAPI.h>
#include "scene.h"
#include "physics.hpp"
#include "geometry.hpp"

using namespace physx;

void* newScene() {
	PxSceneDesc sceneDesc(physics->getTolerancesScale());
	sceneDesc.gravity = PxVec3(0, 0, -9.8);
	if(!sceneDesc.cpuDispatcher) {
		PxDefaultCpuDispatcher* cpuDispatcher = PxDefaultCpuDispatcherCreate(1);
		if(!cpuDispatcher)
			fatalError("PxDefaultCpuDispatcherCreate failed!");
		sceneDesc.cpuDispatcher = cpuDispatcher;
	}
	if(!sceneDesc.filterShader)
		sceneDesc.filterShader = &PxDefaultSimulationFilterShader;
	PxScene* scene = physics->createScene(sceneDesc);
	if(!scene)
		fatalError("createScene failed!");
	
	scene->addActor(*PxCreatePlane(*physics, PxPlane(PxVec3(0, 0, 1), 0), *physics->createMaterial(1, 0.7, .5)));

	return scene;
}

void* Scene_newDynamicActor(void* s, Transform pose) {
	PxRigidDynamic* a = physics->createRigidDynamic(t2P(pose));
	a->setSolverIterationCounts(16, 4);
	((PxScene*)s)->addActor(*a);
	return a;
}

void Scene_simulate(void* s, float dt) {
	PxScene* scene = (PxScene*)s;
	scene->simulate(dt);
	scene->fetchResults(true);
}

