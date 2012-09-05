#ifndef SCENE_H
#define SCENE_H


#include "geometry.h"

#ifdef __cplusplus
extern "C" {
#endif

void* newScene();
void* Scene_newDynamicActor(void* s, Transform pose);
void Scene_simulate(void* s, float dt);

#ifdef __cplusplus
}
#endif


#endif // SCENE_H

