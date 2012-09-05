#ifndef PHYSICS_H
#define PHYSICS_H


#ifdef __cplusplus
extern "C" {
#endif

void Physics_init();
void Physics_close();

void* Physics_newMaterial(float staticFriction, float dynamicFriction, float restitution);

#ifdef __cplusplus
}
#endif


#endif // PHYSICS_H

