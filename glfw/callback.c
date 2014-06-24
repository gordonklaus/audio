#include "callback.h"
#include "_cgo_export.h"

void errorCallback(int error, const char* description) {
	errorCB(error, (char*)description);
}
void setErrorCallback() {
	glfwSetErrorCallback(&errorCallback);
}
void clearErrorCallback() {
	glfwSetErrorCallback(0);
}

void focusCallback(GLFWwindow* w, int focused) {
	focusCB(w, focused);
}
void setFocusCallback(GLFWwindow* w) {
	glfwSetWindowFocusCallback(w, &focusCallback);
}
void clearFocusCallback(GLFWwindow* w) {
	glfwSetWindowFocusCallback(w, 0);
}

void closeCallback(GLFWwindow* w) {
	closeCB(w);
}
void setCloseCallback(GLFWwindow* w) {
	glfwSetWindowCloseCallback(w, &closeCallback);
}
void clearCloseCallback(GLFWwindow* w) {
	glfwSetWindowCloseCallback(w, 0);
}

void resizeCallback(GLFWwindow* w, int width, int height) {
	resizeCB(w, width, height);
}
void setResizeCallback(GLFWwindow* w) {
	glfwSetWindowSizeCallback(w, &resizeCallback);
}
void clearResizeCallback(GLFWwindow* w) {
	glfwSetWindowSizeCallback(w, 0);
}

void framebufferResizeCallback(GLFWwindow* w, int width, int height) {
	framebufferResizeCB(w, width, height);
}
void setFramebufferResizeCallback(GLFWwindow* w) {
	glfwSetFramebufferSizeCallback(w, &framebufferResizeCallback);
}
void clearFramebufferResizeCallback(GLFWwindow* w) {
	glfwSetFramebufferSizeCallback(w, 0);
}

void keyCallback(GLFWwindow* w, int key, int scancode, int action, int mods) {
	keyCB(w, key, scancode, action, mods);
}
void setKeyCallback(GLFWwindow* w) {
	glfwSetKeyCallback(w, &keyCallback);
}
void clearKeyCallback(GLFWwindow* w) {
	glfwSetKeyCallback(w, 0);
}

void charCallback(GLFWwindow* w, unsigned int character) {
	charCB(w, character);
}
void setCharCallback(GLFWwindow* w) {
	glfwSetCharCallback(w, &charCallback);
}
void clearCharCallback(GLFWwindow* w) {
	glfwSetCharCallback(w, 0);
}

void mouseMoveCallback(GLFWwindow* w, double x, double y) {
	mouseMoveCB(w, x, y);
}
void setMouseMoveCallback(GLFWwindow* w) {
	glfwSetCursorPosCallback(w, &mouseMoveCallback);
}
void clearMouseMoveCallback(GLFWwindow* w) {
	glfwSetCursorPosCallback(w, 0);
}

void mouseButtonCallback(GLFWwindow* w, int button, int action, int mods) {
	mouseButtonCB(w, button, action, mods);
}
void setMouseButtonCallback(GLFWwindow* w) {
	glfwSetMouseButtonCallback(w, &mouseButtonCallback);
}
void clearMouseButtonCallback(GLFWwindow* w) {
	glfwSetMouseButtonCallback(w, 0);
}

void scrollCallback(GLFWwindow* w, double dx, double dy) {
	scrollCB(w, dx, dy);
}
void setScrollCallback(GLFWwindow* w) {
	glfwSetScrollCallback(w, &scrollCallback);
}
void clearScrollCallback(GLFWwindow* w) {
	glfwSetScrollCallback(w, 0);
}
