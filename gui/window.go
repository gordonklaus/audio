package gui

import (
	"github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
	"image"
)

func init() {
	if err := glfw.Init(); err != nil { panic(err) }
	if err := gl.Init(); err != nil { panic(err) }
	if err := glfw.OpenWindow(800, 600, 8, 8, 8, 8, 0, 0, glfw.Windowed); err != nil { panic(err) }
	gl.Enable(gl.BLEND)
	gl.Enable(gl.LINE_SMOOTH)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	glfw.Disable(glfw.AutoPollEvents)
	glfw.SetSwapInterval(1)
}

type Window struct {
	ViewBase
	ClickHandler
	centralView View
	keyboardFocus View
	mouseFocus map[int]MouseHandlerView
	repaintMe bool
}

func NewWindow(self View, centralView View) *Window {
	w := &Window{ViewBase{}, *NewClickKeyboardFocuser(centralView), centralView, centralView, make(map[int]MouseHandlerView), false}
	if self == nil { self = w }
	w.ViewBase = *NewView(self)
	w.AddChild(centralView)
	
	return w
}

func (w *Window) Close() {
	glfw.CloseWindow()
	glfw.Terminate()
}

func (w *Window) HandleEvents() {
	glfw.SetWindowSizeCallback(func(width, height int) {
		w.Resize(width, height)
		w.centralView.Resize(width, height)
	})
	
	keyEvent := KeyEvent{}
	glfw.SetKeyCallback(func(key, state int) {
		keyEvent.Key = key
		if key > glfw.KeySpecial {
			keyEvent.Text = ""
			if state == glfw.KeyPress {
				w.keyboardFocus.KeyPressed(keyEvent)
			} else if state == glfw.KeyRelease {
				w.keyboardFocus.KeyReleased(keyEvent)
			}
		}
	})
	glfw.SetCharCallback(func(char, state int) {
		if char < glfw.KeySpecial {
			keyEvent.Text = string(char)
			if state == glfw.KeyPress {
				w.keyboardFocus.KeyPressed(keyEvent)
			} else if state == glfw.KeyRelease {
				w.keyboardFocus.KeyReleased(keyEvent)
			}
		}
	})
	
	var mousePos image.Point
	glfw.SetMousePosCallback(func(x, y int) {
		mousePos = image.Pt(x, w.Height() - y)
		for button, v := range w.mouseFocus {
			pt := v.MapFrom(mousePos, w.self)
			v.MouseDragged(button, pt)
		}
	})
	glfw.SetMouseButtonCallback(func(button, state int) {
		if state == glfw.KeyPress {
			v := w.GetMouseFocus(button, mousePos)
			if v != nil {
				w.mouseFocus[button] = v
				pt := v.MapFrom(mousePos, w.self)
				v.MousePressed(button, pt)
			}
		} else if state == glfw.KeyRelease {
			if v, ok := w.mouseFocus[button]; ok {
				pt := v.MapFrom(mousePos, w.self)
				v.MouseReleased(button, pt)
				delete(w.mouseFocus, button)
			}
		}
	})

	for glfw.WindowParam(glfw.Opened) == 1 {
		glfw.WaitEvents()
		w.repaint()
	}
}

func (w *Window) SetKeyboardFocus(view View) {
	if w.keyboardFocus != view {
		w.keyboardFocus.LostKeyboardFocus()
		w.keyboardFocus = view
		w.keyboardFocus.TookKeyboardFocus()
	}
}

func (w *Window) SetMouseFocus(focus MouseHandlerView, button int) { w.mouseFocus[button] = focus }

func (w *Window) Repaint() { w.repaintMe = true }
func (w Window) repaint() {
	if !w.repaintMe { return }
	w.repaintMe = false
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	width, height := w.Width(), w.Height()
	gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))
	wid, hei := gl.Double(width)/2, gl.Double(height)/2
	gl.Ortho(-wid, wid, -hei, hei, -1, 1)
	gl.Translated(-wid, -hei, 0)

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	w.GetViewBase().paintBase()
	glfw.SwapBuffers()
}
