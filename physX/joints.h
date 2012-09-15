#ifndef JOINTS_H
#define JOINTS_H


#include "geometry.h"

#ifdef __cplusplus
extern "C" {
#endif

void* newJoint(void* actor1, Transform transform1, void* actor2, Transform transform2);
void Joint_enableCollision(void* j);
void Joint_setMotion(void* j, int axis, int type);
void Joint_setDrive(void* j, int index, float spring, float damping, float forceLimit);
void Joint_setDrivePose(void* j, Transform pose);
void Joint_setDriveVelocity(void* j, Vector linear, Vector angular);

#ifdef __cplusplus
}
#endif


#endif // JOINTS_H

