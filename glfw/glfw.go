package glfw

/*
#cgo pkg-config: glfw3
#include <GL/glfw3.h>
*/
import "C"

import (
	"errors"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

// This function must be called from the main thread (i.e. from either init() or main()).
func Init() error {
	if C.glfwInit() == C.GL_FALSE {
		return errors.New("Failed to initialize GLFW.")
	}
	return nil
}

func Terminate() {
	C.glfwTerminate()
}

func MakeContextCurrent(w *Window) {
	if w == nil {
		C.glfwMakeContextCurrent(nil)
	} else {
		C.glfwMakeContextCurrent(w.w)
	}
}

func WaitEvents() {
	C.glfwWaitEvents()
}
