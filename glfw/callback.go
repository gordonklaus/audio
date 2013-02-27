package glfw

//#include <GL/glfw3.h>
//#include "callback.h"
import "C"

import (
	"errors"
	"unsafe"
)

type ErrorFunc func(err error)
type CloseFunc func() bool
type ResizeFunc func(width, height int)
type KeyFunc func(key, action int)
type CharFunc func(char rune)
type MouseMoveFunc func(x, y int)
type MouseButtonFunc func(button, action int)

var errorCallback ErrorFunc

func OnError(f ErrorFunc) {
	errorCallback = f
	if f != nil {
		C.setErrorCallback()
	} else {
		C.clearErrorCallback()
	}
}
func (w *Window) OnClose(f CloseFunc) {
	w.onClose = f
	if f != nil {
		C.setCloseCallback(w.w)
	} else {
		C.clearCloseCallback(w.w)
	}
}
func (w *Window) OnResize(f ResizeFunc) {
	w.onResize = f
	if f != nil {
		C.setResizeCallback(w.w)
	} else {
		C.clearResizeCallback(w.w)
	}
}
func (w *Window) OnKey(f KeyFunc) {
	w.onKey = f
	if f != nil {
		C.setKeyCallback(w.w)
	} else {
		C.clearKeyCallback(w.w)
	}
}
func (w *Window) OnChar(f CharFunc) {
	w.onChar = f
	if f != nil {
		C.setCharCallback(w.w)
	} else {
		C.clearCharCallback(w.w)
	}
}
func (w *Window) OnMouseMove(f MouseMoveFunc) {
	w.onMouseMove = f
	if f != nil {
		C.setMouseMoveCallback(w.w)
	} else {
		C.clearMouseMoveCallback(w.w)
	}
}
func (w *Window) OnMouseButton(f MouseButtonFunc) {
	w.onMouseButton = f
	if f != nil {
		C.setMouseButtonCallback(w.w)
	} else {
		C.clearMouseButtonCallback(w.w)
	}
}

//export errorCB
func errorCB(err C.int, desc *C.char) {
	errorCallback(errors.New(C.GoString(desc)))
}
//export closeCB
func closeCB(win unsafe.Pointer) C.int {
	w := (*Window)(C.glfwGetWindowUserPointer((*C.GLFWwindow)(win)))
	if w.onClose() {
		return C.GL_TRUE
	}
	return C.GL_FALSE
}
//export resizeCB
func resizeCB(win unsafe.Pointer, width, height C.int) {
	w := (*Window)(C.glfwGetWindowUserPointer((*C.GLFWwindow)(win)))
	w.onResize(int(width), int(height))
}
//export keyCB
func keyCB(win unsafe.Pointer, key, action C.int) {
	w := (*Window)(C.glfwGetWindowUserPointer((*C.GLFWwindow)(win)))
	w.onKey(int(key), int(action))
}
//export charCB
func charCB(win unsafe.Pointer, char C.int) {
	w := (*Window)(C.glfwGetWindowUserPointer((*C.GLFWwindow)(win)))
	w.onChar(rune(char))
}
//export mouseMoveCB
func mouseMoveCB(win unsafe.Pointer, x, y C.int) {
	w := (*Window)(C.glfwGetWindowUserPointer((*C.GLFWwindow)(win)))
	w.onMouseMove(int(x), int(y))
}
//export mouseButtonCB
func mouseButtonCB(win unsafe.Pointer, button, action C.int) {
	w := (*Window)(C.glfwGetWindowUserPointer((*C.GLFWwindow)(win)))
	w.onMouseButton(int(button), int(action))
}
