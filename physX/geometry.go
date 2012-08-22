package physX

// #include "geometry.h"
import "C"

import ."math"

type Vector [3]float64
func (v Vector) Len() float64 { return Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2]) }
func (v Vector) Add(w Vector) Vector { return Vector{v[0] + w[0], v[1] + w[1], v[2] + w[2]} }
func (v Vector) Sub(w Vector) Vector { return Vector{v[0] - w[0], v[1] - w[1], v[2] - w[2]} }
func (v Vector) Mul(x float64) Vector { return Vector{v[0] * x, v[1] * x, v[2] * x} }
func (v Vector) Div(x float64) Vector { return Vector{v[0] / x, v[1] / x, v[2] / x} }
func (v Vector) Dot(w Vector) float64 { return v[0]*w[0] + v[1]*w[1] + v[2]*w[2] }
func (v Vector) Cross(w Vector) Vector { return Vector{v[1]*w[2] - v[2]*w[1], v[2]*w[0] - v[0]*w[2], v[0]*w[1] - v[1]*w[0]} }
func (v Vector) Normalized() Vector { return v.Div(v.Len()) }
func (v *Vector) Normalize() { *v = v.Div(v.Len()) }
func (v Vector) cVec() C.Vector { return C.Vector{C.float(v[0]), C.float(v[1]), C.float(v[2])} }
func vectorFromCVec(v C.Vector) Vector { return Vector{float64(v.x), float64(v.y), float64(v.z)} }

type Quaternion [4]float64
var IQ = Quaternion{0, 0, 0, 1}
func (q Quaternion) Mul(r Quaternion) Quaternion {
	return Quaternion{q[3]*r[0] + q[0]*r[3] + q[1]*r[2] - q[2]*r[1],
					  q[3]*r[1] + q[1]*r[3] + q[2]*r[0] - q[0]*r[2],
					  q[3]*r[2] + q[2]*r[3] + q[0]*r[1] - q[1]*r[0],
					  q[3]*r[3] - q[0]*r[0] - q[1]*r[1] - q[2]*r[2]}
}
func (q Quaternion) Conjugate() Quaternion { return Quaternion{-q[0], -q[1], -q[2], q[3]} }
func (q Quaternion) Rotate(v Vector) Vector {
	qv := Vector{q[0], q[1], q[2]}
	return v.Mul(q[3]*q[3] - .5).Add(qv.Cross(v).Mul(q[3])).Add(qv.Mul(qv.Dot(v))).Mul(2)
}
func (q Quaternion) cQuat() C.Quaternion { return C.Quaternion{C.float(q[0]), C.float(q[1]), C.float(q[2]), C.float(q[3])} }
func quaternionFromCQuat(q C.Quaternion) Quaternion { return Quaternion{float64(q.x), float64(q.y), float64(q.z), float64(q.w)} }

type Matrix [3]Vector
var IM = Matrix{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
func (m Matrix) Transposed() Matrix { return Matrix{{m[0][0], m[1][0], m[2][0]}, {m[0][1], m[1][1], m[2][1]}, {m[0][2], m[1][2], m[2][2]}} }
func (m *Matrix) Transpose() { *m = m.Transposed() }
func (m Matrix) cMat() C.Matrix { return C.Matrix{m[0].cVec(), m[1].cVec(), m[2].cVec()} }

func AxisNormalOrientation(axis, normal Vector) Matrix {
	axis.Normalize()
	binormal := axis.Cross(normal); binormal.Normalize()
	normal = binormal.Cross(axis)
	orientation := Matrix{axis, normal, binormal}
	return orientation
}
func AxisOrientation(axis Vector) Matrix {
	axis.Normalize()
	normal := axis.Cross(Vector{0, 0, 1})
	if normal.Len() < .1 {
		normal = axis.Cross(Vector{0, 1, 0})
	}
	return AxisNormalOrientation(axis, normal)
}


type Transform struct {
	Position Vector
	Orientation Quaternion
}
func TransformFromSegment(p1, p2 Vector) Transform { return transformFromCTrans(C.TransformFromSegment(p1.cVec(), p2.cVec())) }
func (t Transform) Transform(u Transform) Transform { return Transform{t.TransformVector(u.Position), t.Orientation.Mul(u.Orientation)} }
func (t Transform) TransformVector(v Vector) Vector { return t.Orientation.Rotate(v).Add(t.Position) }
func (t Transform) TransformInv(u Transform) Transform {
	qinv := t.Orientation.Conjugate()
	return Transform{qinv.Rotate(u.Position.Sub(t.Position)), qinv.Mul(u.Orientation)}
}
func transformFromCTrans(t C.Transform) Transform { return Transform{vectorFromCVec(t.p), quaternionFromCQuat(t.o)} }
func (t Transform) cTrans() C.Transform { return C.Transform{t.Position.cVec(), t.Orientation.cQuat()} }

