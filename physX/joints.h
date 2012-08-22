#ifndef JOINTS_H
#define JOINTS_H


#include "geometry.h"

#ifdef __cplusplus
extern "C" {
#endif

void newIsoUniversalJoint(void* actor1, Transform transform1, void* actor2, Transform transform2);

#ifdef __cplusplus
}
#endif


#endif // JOINTS_H

