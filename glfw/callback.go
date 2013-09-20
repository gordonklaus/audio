package glfw

//#include "callback.h"
import "C"

import (
	"errors"
	"unsafe"
)

type ErrorFunc func(err error)
type CloseFunc func()
type ResizeFunc func(width, height int)
type KeyFunc func(key, scancode, action, mods int)
type CharFunc func(char rune)
type MouseMoveFunc func(x, y float64)
type MouseButtonFunc func(button, action, mods int)

const (
	ModShift = 1
	ModControl = 2
	ModAlt = 4
	ModSuper = 8
)

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
	w.closeCB = f
	if f != nil {
		C.setCloseCallback(w.w)
	} else {
		C.clearCloseCallback(w.w)
	}
}

func (w *Window) OnResize(f ResizeFunc) {
	w.resizeCB = f
	if f != nil {
		C.setResizeCallback(w.w)
	} else {
		C.clearResizeCallback(w.w)
	}
}

func (w *Window) OnFramebufferResize(f ResizeFunc) {
	w.framebufferResizeCB = f
	if f != nil {
		C.setFramebufferResizeCallback(w.w)
	} else {
		C.clearFramebufferResizeCallback(w.w)
	}
}

func (w *Window) OnKey(f KeyFunc) {
	w.keyCB = f
	if f != nil {
		C.setKeyCallback(w.w)
	} else {
		C.clearKeyCallback(w.w)
	}
}

func (w *Window) OnChar(f CharFunc) {
	w.charCB = f
	if f != nil {
		C.setCharCallback(w.w)
	} else {
		C.clearCharCallback(w.w)
	}
}

func (w *Window) OnMouseMove(f MouseMoveFunc) {
	w.mouseMoveCB = f
	if f != nil {
		C.setMouseMoveCallback(w.w)
	} else {
		C.clearMouseMoveCallback(w.w)
	}
}

func (w *Window) OnMouseButton(f MouseButtonFunc) {
	w.mouseButtonCB = f
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
func closeCB(w unsafe.Pointer) {
	win(w).closeCB()
}

//export resizeCB
func resizeCB(w unsafe.Pointer, width, height C.int) {
	win(w).resizeCB(int(width), int(height))
}

//export framebufferResizeCB
func framebufferResizeCB(w unsafe.Pointer, width, height C.int) {
	win(w).framebufferResizeCB(int(width), int(height))
}

//export keyCB
func keyCB(w unsafe.Pointer, key, scancode, action, mods C.int) {
	win(w).keyCB(int(key), int(scancode), int(action), int(mods))
}

//export charCB
func charCB(w unsafe.Pointer, char C.uint) {
	win(w).charCB(rune(char))
}

//export mouseMoveCB
func mouseMoveCB(w unsafe.Pointer, x, y C.double) {
	win(w).mouseMoveCB(float64(x), float64(y))
}

//export mouseButtonCB
func mouseButtonCB(w unsafe.Pointer, button, action, mods C.int) {
	win(w).mouseButtonCB(int(button), int(action), int(mods))
}

func win(w unsafe.Pointer) *Window {
	return (*Window)(C.glfwGetWindowUserPointer((*C.GLFWwindow)(w)))
}
