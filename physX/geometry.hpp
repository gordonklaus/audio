#include <PxPhysicsAPI.h>
#include "geometry.h"

using namespace physx;

PxVec3 v2P(Vector v);
Vector P2v(PxVec3 v);

PxQuat q2P(Quaternion q);
Quaternion P2q(PxQuat q);

PxMat33 m2P(Matrix m);

PxTransform t2P(Transform t);
Transform P2t(PxTransform t);

