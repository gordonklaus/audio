#include "geometry.hpp"

PxVec3 v2P(Vector v) { return *(PxVec3*)&v; }
Vector P2v(PxVec3 v) { return *(Vector*)&v; }

PxQuat q2P(Quaternion q) { return *(PxQuat*)&q; }
Quaternion P2q(PxQuat q) { return *(Quaternion*)&q; }

PxMat33 m2P(Matrix m) { return *(PxMat33*)&m; }

PxTransform t2P(Transform t) { return PxTransform(v2P(t.p), q2P(t.o).getNormalized()); }
Transform P2t(PxTransform t) { Transform r; r.p = P2v(t.p); r.o = P2q(t.q); return r; }

Quaternion Matrix_toQuat(Matrix m) { return P2q(PxQuat(m2P(m))); }

