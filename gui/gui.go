package gui

import (
	"code.google.com/p/gordon-go/glfw"
	gl "github.com/chsc/gogl/gl21"
	"log"
)

func init() {
	glfw.OnError(func(err error) {
		log.Println("GLFW: ", err)
	})
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	if err := gl.Init(); err != nil {
		panic(err)
	}
}

var windows = map[*Window]struct{}{}

func Run() {
	defer glfw.Terminate()
	for len(windows) > 0 {
		glfw.WaitEvents()
		for w := range windows {
			if w.w.ShouldClose() {
				delete(windows, w)
				w.close <- struct{}{} // wait for window to stop before Destroying it to be sure it is done with the OpenGL context
				w.w.Destroy()
			}
		}
	}
}
