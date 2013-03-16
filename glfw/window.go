package glfw

//#include <GL/glfw3.h>
import "C"

import (
	"unsafe"
)

type Window struct {
	w *C.GLFWwindow
	onClose CloseFunc
	onResize ResizeFunc
	onKey KeyFunc
	onChar CharFunc
	onMouseMove MouseMoveFunc
	onMouseButton MouseButtonFunc
}

func NewWindow(width, height int, title string) *Window {
	win := C.glfwCreateWindow(C.int(width), C.int(height), C.CString(title), nil, nil)
	if win == nil {
		panic("GLFW:  Failed to create window.")
	}
	w := &Window{w:win}
	C.glfwSetWindowUserPointer(win, unsafe.Pointer(w))
	return w
}

func (w *Window) Destroy() {
	C.glfwDestroyWindow(w.w)
}

func (w *Window) Size() (width, height int) {
	var wid, hei C.int
	C.glfwGetWindowSize(w.w, &wid, &hei)
	return int(wid), int(hei)
}

func (w *Window) ShouldClose() bool {
	return C.glfwWindowShouldClose(w.w) == C.GL_TRUE
}

func (w *Window) SwapBuffers() {
	C.glfwSwapBuffers(w.w)
}
