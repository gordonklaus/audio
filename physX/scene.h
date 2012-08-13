#ifndef SCENE_H
#define SCENE_H


#ifdef __cplusplus
extern "C" {
#endif

void* newScene();
void* Scene_newDynamicActor(void* s, float* pos);
void Scene_simulate(void* s, float dt);

#ifdef __cplusplus
}
#endif


#endif // SCENE_H

