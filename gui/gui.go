package gui

import (
	"code.google.com/p/gordon-go/glfw"
	gl "github.com/chsc/gogl/gl21"
	"log"
)

var (
	windows     = map[*Window]bool{}
	newWindow   = make(chan *Window, 1)
	closeWindow = make(chan *Window, 1)
	quit        = make(chan bool, 1)
)

// must be called from the main goroutine, which will run on the main thread because glfw calls runtime.LockOSThread()
func Run(init func()) error {
	glfw.OnError(func(err error) {
		log.Println("GLFW: ", err)
	})
	if err := glfw.Init(); err != nil {
		return err
	}
	defer glfw.Terminate()
	if err := gl.Init(); err != nil {
		return err
	}
	init()
	windows[<-newWindow] = true
	for len(windows) > 0 {
		glfw.WaitEvents()
		select {
		case w := <-newWindow:
			windows[w] = true
		case w := <-closeWindow:
			closeWin(w)
		case <-quit:
			for w := range windows {
				closeWin(w)
			}
		default:
		}
	}
	return nil
}

func closeWin(w *Window) {
	delete(windows, w)
	w.close <- true // wait for window before Destroying it to be sure it is done with the OpenGL context
	w.w.Destroy()
}

func Quit() {
	select {
	case quit <- true:
	default:
	}
	glfw.PostEmptyEvent()
}
