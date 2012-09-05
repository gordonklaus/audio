package physX

// #cgo LDFLAGS: CphysX.so
// #include "debugRenderer.h"
import "C"

type DebugRenderer struct { r *C.DebugRenderer }

func NewDebugRenderer(s Scene) DebugRenderer { return DebugRenderer{C.newDebugRenderer(s.s)} }
func (r DebugRenderer) Resize(w, h int) { C.DebugRenderer_resize(r.r, C.int(w), C.int(h)) }
func (r DebugRenderer) Zoom(factor float64) { C.DebugRenderer_zoom(r.r, C.double(factor)) }
func (r DebugRenderer) Rotate(dAngle, dElevation float64) { C.DebugRenderer_rotate(r.r, C.double(dAngle), C.double(dElevation)) }
func (r DebugRenderer) Pan(right, forward float64) { C.DebugRenderer_pan(r.r, C.double(right), C.double(forward)) }
func (r DebugRenderer) Render() { C.DebugRenderer_render(r.r) }

