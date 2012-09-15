package physX

// #cgo LDFLAGS: CphysX.so
// #include "scene.h"
import "C"

import "unsafe"

type Scene struct { s unsafe.Pointer }

func NewScene() Scene { return Scene{C.newScene()} }
func (s Scene) NewDynamicActor(pose Transform) DynamicActor { return DynamicActor{C.Scene_newDynamicActor(s.s, pose.c())} }
func (s Scene) Simulate(dt float32) { C.Scene_simulate(s.s, C.float(dt)) }

