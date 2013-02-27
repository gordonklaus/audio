#include <GL/glfw3.h>

#include "callback.h"
#include "_cgo_export.h"

void setErrorCallback() {
	glfwSetErrorCallback((GLFWerrorfun)&errorCB);
}
void clearErrorCallback() {
	glfwSetErrorCallback(0);
}
void setCloseCallback(GLFWwindow* w) {
	glfwSetWindowCloseCallback(w, (GLFWwindowclosefun)&closeCB);
}
void clearCloseCallback(GLFWwindow* w) {
	glfwSetWindowCloseCallback(w, 0);
}
void setResizeCallback(GLFWwindow* w) {
	glfwSetWindowSizeCallback(w, (GLFWwindowsizefun)&resizeCB);
}
void clearResizeCallback(GLFWwindow* w) {
	glfwSetWindowSizeCallback(w, 0);
}
void setKeyCallback(GLFWwindow* w) {
	glfwSetKeyCallback(w, (GLFWkeyfun)&keyCB);
}
void clearKeyCallback(GLFWwindow* w) {
	glfwSetKeyCallback(w, 0);
}
void setCharCallback(GLFWwindow* w) {
	glfwSetCharCallback(w, (GLFWcharfun)&charCB);
}
void clearCharCallback(GLFWwindow* w) {
	glfwSetCharCallback(w, 0);
}
void setMouseMoveCallback(GLFWwindow* w) {
	glfwSetCursorPosCallback(w, (GLFWcursorposfun)&mouseMoveCB);
}
void clearMouseMoveCallback(GLFWwindow* w) {
	glfwSetCursorPosCallback(w, 0);
}
void setMouseButtonCallback(GLFWwindow* w) {
	glfwSetMouseButtonCallback(w, (GLFWmousebuttonfun)&mouseButtonCB);
}
void clearMouseButtonCallback(GLFWwindow* w) {
	glfwSetMouseButtonCallback(w, 0);
}
