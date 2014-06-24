package glfw

//#include <GLFW/glfw3.h>
import "C"

import (
	"sync"
	"unsafe"
)

type Window struct {
	w                   *C.GLFWwindow
	closeCB             CloseFunc
	focusCB             FocusFunc
	resizeCB            ResizeFunc
	framebufferResizeCB ResizeFunc
	keyCB               KeyFunc
	charCB              CharFunc
	mouseMoveCB         MouseMoveFunc
	mouseButtonCB       MouseButtonFunc
	scrollCB            ScrollFunc
}

var windows = struct {
	sync.Mutex
	m map[*C.GLFWwindow]*Window
}{m: map[*C.GLFWwindow]*Window{}}

func NewWindow(width, height int, title string) *Window {
	win := C.glfwCreateWindow(C.int(width), C.int(height), C.CString(title), nil, nil)
	if win == nil {
		panic("GLFW:  Failed to create window.")
	}
	w := &Window{w: win}
	C.glfwSetWindowUserPointer(win, unsafe.Pointer(w))
	windows.Lock()
	windows.m[win] = w
	windows.Unlock()
	return w
}

func (w *Window) Destroy() {
	C.glfwDestroyWindow(w.w)
	windows.Lock()
	delete(windows.m, w.w)
	windows.Unlock()
}

func (w *Window) Hide() { C.glfwHideWindow(w.w) }
func (w *Window) Show() { C.glfwShowWindow(w.w) }

func (w *Window) SetTitle(s string) {
	C.glfwSetWindowTitle(w.w, C.CString(s))
}

func (w *Window) Size() (width, height int) {
	var wid, hei C.int
	C.glfwGetWindowSize(w.w, &wid, &hei)
	return int(wid), int(hei)
}

func (w *Window) Move(x, y int) {
	C.glfwSetWindowPos(w.w, C.int(x), C.int(y))
}

func (w *Window) FramebufferSize() (width, height int) {
	var wid, hei C.int
	C.glfwGetFramebufferSize(w.w, &wid, &hei)
	return int(wid), int(hei)
}

func (w *Window) ShouldClose() bool {
	return C.glfwWindowShouldClose(w.w) == C.GL_TRUE
}

func (w *Window) SetShouldClose(b bool) {
	v := C.GL_FALSE
	if b {
		v = C.GL_TRUE
	}
	C.glfwSetWindowShouldClose(w.w, C.int(v))
}

func (w *Window) MousePosition() (float64, float64) {
	var x, y C.double
	C.glfwGetCursorPos(w.w, &x, &y)
	return float64(x), float64(y)
}

func (w *Window) SwapBuffers() {
	C.glfwSwapBuffers(w.w)
}
