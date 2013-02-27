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
	// usually small).  A solution might be found in this feature
	// request:  http://sourceforge.net/p/glfw/feature-requests/27/
	windowControl = make(chan interface{})
)

func Run() {
	defer glfw.Terminate()
	for len(windows) > 0 {
		glfw.WaitEvents()
F:		for {
			select {
			case c := <-windowControl:
				switch c := c.(type) {
				case Close:
					c.w.w.Destroy()
					delete(windows, c.w)
				}
			default:
				break F
			}
		}
	}
}
