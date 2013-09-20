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


var (
	windows = map[*Window]bool{}
	
	// Currently, messages on this channel are not processed until
	// WaitEvents returns, which is an indefinite amount of time (if
	// usually small).
	closeWindow = make(chan *Window)
)

func Run() {
	defer glfw.Terminate()
	for len(windows) > 0 {
		glfw.WaitEvents()
F:		for {
			select {
			case w := <-closeWindow:
				delete(windows, w)
			default:
				break F
			}
		}
	}
}
