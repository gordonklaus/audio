package gui

import (
	"code.google.com/p/gordon-go/glfw"
	gl "github.com/chsc/gogl/gl21"
	"log"
)

var do = make(chan func(), 1)
var windows []*Window

// must be called on the main goroutine, which will run on the main thread because glfw calls runtime.LockOSThread()
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
	go init()
	for len(windows) == 0 {
		(<-do)()
	}
	for len(windows) > 0 {
		glfw.WaitEvents()
		select {
		case f := <-do:
			f()
		default:
		}
	}
	return nil
}

func doMain(f func()) {
	done := make(chan bool)
	do <- func() {
		f()
		done <- true
	}
	glfw.PostEmptyEvent()
	<-done
}

func closeWindow(w *Window) {
	for i, w2 := range windows {
		if w2 == w {
			windows = append(windows[:i], windows[i+1:]...)
			break
		}
	}
	w.close <- true // wait for window before Destroying it to be sure it is done with the OpenGL context
	w.w.Destroy()
	if len(windows) > 0 {
		windows[0].w.Show()
	}
}

func Quit() {
	go doMain(func() {
		for len(windows) > 0 {
			closeWindow(windows[0])
		}
	})
}
