package physX

// #cgo LDFLAGS: CphysX.so
// #include "geometry.h"
import "C"

import . "math"

type Vector struct{ X, Y, Z float64 }

func (v Vector) Len() float64         { return Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z) }
func (v Vector) Add(w Vector) Vector  { return Vector{v.X + w.X, v.Y + w.Y, v.Z + w.Z} }
func (v Vector) Sub(w Vector) Vector  { return Vector{v.X - w.X, v.Y - w.Y, v.Z - w.Z} }
func (v Vector) Mul(x float64) Vector { return Vector{v.X * x, v.Y * x, v.Z * x} }
func (v Vector) Div(x float64) Vector { return Vector{v.X / x, v.Y / x, v.Z / x} }
func (v Vector) Dot(w Vector) float64 { return v.X*w.X + v.Y*w.Y + v.Z*w.Z }
func (v Vector) Cross(w Vector) Vector {
	return Vector{v.Y*w.Z - v.Z*w.Y, v.Z*w.X - v.X*w.Z, v.X*w.Y - v.Y*w.X}
}
func (v Vector) Normalized() Vector  { return v.Div(v.Len()) }
func (v *Vector) Normalize() float64 { d := v.Len(); *v = v.Div(d); return d }
func (v Vector) c() C.Vector         { return C.Vector{C.float(v.X), C.float(v.Y), C.float(v.Z)} }
func C2v(v C.Vector) Vector          { return Vector{float64(v.x), float64(v.y), float64(v.z)} }

type Quaternion struct{ X, Y, Z, W float64 }

var IQ = Quaternion{0, 0, 0, 1}

func AxisAngleQuaternion(axis Vector, angle float64) Quaternion {
	s := Sin(angle / 2)
	return Quaternion{axis.X * s, axis.Y * s, axis.Z * s, Cos(angle / 2)}
}
func (q Quaternion) XAxis() Vector { return q.Rotate(Vector{1, 0, 0}) }
func (q Quaternion) YAxis() Vector { return q.Rotate(Vector{0, 1, 0}) }
func (q Quaternion) ZAxis() Vector { return q.Rotate(Vector{0, 0, 1}) }
func (q Quaternion) Mul(r Quaternion) Quaternion {
	return Quaternion{q.W*r.X + q.X*r.W + q.Y*r.Z - q.Z*r.Y,
		q.W*r.Y + q.Y*r.W + q.Z*r.X - q.X*r.Z,
		q.W*r.Z + q.Z*r.W + q.X*r.Y - q.Y*r.X,
		q.W*r.W - q.X*r.X - q.Y*r.Y - q.Z*r.Z}
}
func (q Quaternion) Conjugate() Quaternion { return Quaternion{-q.X, -q.Y, -q.Z, q.W} }
func (q Quaternion) Rotate(v Vector) Vector {
	qv := Vector{q.X, q.Y, q.Z}
	return v.Mul(q.W*q.W - .5).Add(qv.Cross(v).Mul(q.W)).Add(qv.Mul(qv.Dot(v))).Mul(2)
}
func (q Quaternion) c() C.Quaternion {
	return C.Quaternion{C.float(q.X), C.float(q.Y), C.float(q.Z), C.float(q.W)}
}
func C2q(q C.Quaternion) Quaternion {
	return Quaternion{float64(q.x), float64(q.y), float64(q.z), float64(q.w)}
}

type Matrix [3]Vector

var IM = Matrix{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}

func (m Matrix) Transposed() Matrix {
	return Matrix{{m[0].X, m[1].X, m[2].X}, {m[0].Y, m[1].Y, m[2].Y}, {m[0].Z, m[1].Z, m[2].Z}}
}
func (m *Matrix) Transpose()        { *m = m.Transposed() }
func (m Matrix) ToQuat() Quaternion { return C2q(C.Matrix_toQuat(m.c())) }
func C2m(m C.Matrix) Matrix         { return Matrix{C2v(m.c1), C2v(m.c2), C2v(m.c3)} }
func (m Matrix) c() C.Matrix        { return C.Matrix{m[0].c(), m[1].c(), m[2].c()} }

func AxisNormalOrientation(axis, normal Vector) Quaternion {
	axis.Normalize()
	binormal := axis.Cross(normal)
	binormal.Normalize()
	normal = binormal.Cross(axis)
	return Matrix{axis, normal, binormal}.ToQuat()
}
func AxisOrientation(axis Vector) Quaternion {
	normal := axis.Cross(Vector{0, 0, 1})
	if normal.Len() < .1 {
		normal = axis.Cross(Vector{0, 1, 0})
	}
	return AxisNormalOrientation(axis, normal)
}

type Transform struct {
	Position    Vector
	Orientation Quaternion
}

func (t Transform) Transform(u Transform) Transform {
	return Transform{t.TransformVector(u.Position), t.Orientation.Mul(u.Orientation)}
}
func (t Transform) TransformVector(v Vector) Vector { return t.Orientation.Rotate(v).Add(t.Position) }
func (t Transform) TransformInv(u Transform) Transform {
	return Transform{t.TransformVectorInv(u.Position), t.Orientation.Conjugate().Mul(u.Orientation)}
}
func (t Transform) TransformVectorInv(v Vector) Vector { return t.Orientation.Conjugate().Rotate(v.Sub(t.Position)) }
func C2t(t C.Transform) Transform  { return Transform{C2v(t.p), C2q(t.o)} }
func (t Transform) c() C.Transform { return C.Transform{t.Position.c(), t.Orientation.c()} }

