#ifndef DYNAMICACTOR_H
#define DYNAMICACTOR_H


#include "geometry.h"

#ifdef __cplusplus
extern "C" {
#endif

Transform DynamicActor_globalPose(void* actor);

void DynamicActor_addSphere(void* actor, Vector pos, float radius, void* material);
void DynamicActor_addCapsule(void* actor, Vector start, Vector end, float radius, void* material);

#ifdef __cplusplus
}
#endif


#endif // DYNAMICACTOR_H

