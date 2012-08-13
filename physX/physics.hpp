#ifndef PHYSICS_HPP
#define PHYSICS_HPP


#include <PxPhysicsAPI.h>
#include "physics.h"

void fatalError(const char* s);

using namespace physx;

extern PxPhysics* physics;


#endif // PHYSICS_HPP

