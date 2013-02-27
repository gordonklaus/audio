#ifndef GLFWGO_CALLBACK_H
#define GLFWGO_CALLBACK_H

void setErrorCallback();
void clearErrorCallback();
void setCloseCallback(GLFWwindow* w);
void clearCloseCallback(GLFWwindow* w);
void setResizeCallback(GLFWwindow* w);
void clearResizeCallback(GLFWwindow* w);
void setKeyCallback(GLFWwindow* w);
void clearKeyCallback(GLFWwindow* w);
void setCharCallback(GLFWwindow* w);
void clearCharCallback(GLFWwindow* w);
void setMouseMoveCallback(GLFWwindow* w);
void clearMouseMoveCallback(GLFWwindow* w);
void setMouseButtonCallback(GLFWwindow* w);
void clearMouseButtonCallback(GLFWwindow* w);

#endif
