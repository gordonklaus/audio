package gui

import (
	"code.google.com/p/gordon-go/glfw"
	gl "github.com/chsc/gogl/gl21"
	"runtime"
	"sync"
)

type Window struct {
	w *glfw.Window
	*ViewBase
	centralView   View
	keyboardFocus View
	mouseFocus    map[int]MouserView
	control       chan interface{}
	key           chan KeyEvent
	mouse         chan MouseEvent
	paint         chan struct{}
	do            chan func()
}

func NewWindow(self View) *Window {
	w := &Window{w: glfw.NewWindow(800, 600, "Flux")}

	// Somehow, this seems to work:
	// Have the context current in both threads, as this one needs to be able to call
	// Font.Advance and Font.LineHeight (from Text.SetText), and the other thread renders.
	// But is this portable?  Or even a good idea?
	glfw.MakeContextCurrent(w.w)

	if self == nil {
		self = w
	}
	w.ViewBase = NewView(w)
	w.mouseFocus = make(map[int]MouserView)
	w.control = make(chan interface{})
	w.key = make(chan KeyEvent, 10)
	w.mouse = make(chan MouseEvent, 100)
	w.paint = make(chan struct{}, 1)
	w.do = make(chan func())

	// glfw should fire initial resize events to avoid this duplication (https://github.com/glfw/glfw/issues/62)
	width, height := w.w.Size()
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, gl.Double(width), 0, gl.Double(height), -1, 1)
	Resize(w, Pt(float64(width), float64(height)))
	width, height = w.w.FramebufferSize()
	gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))

	windows[w] = true
	w.run()

	w.Self = self
	return w
}

func (w *Window) SetCentralView(v View) {
	if w.centralView != nil {
		w.RemoveChild(w.centralView)
	}
	w.centralView = v
	if v != nil {
		if v.Parent() != w {
			w.AddChild(v)
		}
		Resize(v, Size(w))
		SetKeyboardFocus(v)
	}
}

func (w *Window) run() {
	type (
		close  *Window
		resize Point
	)

	w.w.OnClose(func() {
		w.control <- close(w)
	})
	w.w.OnResize(func(width, height int) {
		w.control <- resize{float64(width), float64(height)}
	})
	w.w.OnFramebufferResize(func(width, height int) {
		gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))
	})

	keyEvent := KeyEvent{}
	w.w.OnKey(func(key, scancode, action, mods int) {
		keyEvent.Key = key
		keyEvent.Action = action
		keyEvent.Shift = mods&glfw.ModShift != 0
		keyEvent.Ctrl = mods&glfw.ModControl != 0
		keyEvent.Alt = mods&glfw.ModAlt != 0
		keyEvent.Super = mods&glfw.ModSuper != 0
		if key >= KeyEscape || action == Release {
			keyEvent.Text = ""
			w.key <- keyEvent
		}
	})
	w.w.OnChar(func(char rune) {
		if char < KeyEscape {
			keyEvent.Text = string(char)
			w.key <- keyEvent
		}
	})

	m := MouseEvent{}
	w.w.OnMouseMove(func(x, y float64) {
		m.Pos = Pt(x, Height(w)-y)
		m.Action = Move
		w.mouse <- m
	})
	w.w.OnMouseButton(func(button, action, mods int) {
		m.Button = button
		m.Action = action
		w.mouse <- m
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		runtime.LockOSThread()
		glfw.MakeContextCurrent(w.w)
		defer glfw.MakeContextCurrent(nil)

		initFont()
		wg.Done()

		gl.Enable(gl.BLEND)
		gl.Enable(gl.POINT_SMOOTH)
		gl.Enable(gl.LINE_SMOOTH)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

		for {
			select {
			case c := <-w.control:
				switch c := c.(type) {
				case close:
					closeWindow <- c
					return
				case resize:
					gl.MatrixMode(gl.PROJECTION)
					gl.LoadIdentity()
					gl.Ortho(0, gl.Double(c.X), 0, gl.Double(c.Y), -1, 1)
					Resize(w.Self, Point(c))
					if w.centralView != nil {
						Resize(w.centralView, Point(c))
					}
				}
			case k := <-w.key:
				if w.keyboardFocus != nil {
					if k.Action != Release {
						w.keyboardFocus.KeyPressed(k)
					} else {
						w.keyboardFocus.KeyReleased(k)
					}
				}
			case m := <-w.mouse:
				switch m.Action {
				case Press:
					v, _ := viewAtFunc(w, m.Pos, func(v View) View {
						v, _ = v.(MouserView)
						return v
					}).(MouserView)
					if v != nil {
						w.mouseFocus[m.Button] = v
						m.Pos = MapFrom(v, m.Pos, w.Self)
						v.Mouse(m)
					}
				case Move:
					for button, v := range w.mouseFocus {
						v.Mouse(MouseEvent{MapFrom(v, m.Pos, w.Self), Drag, button})
					}
				case Release:
					if v, ok := w.mouseFocus[m.Button]; ok {
						m.Pos = MapFrom(v, m.Pos, w.Self)
						v.Mouse(m)
						delete(w.mouseFocus, m.Button)
					}
				}
			case <-w.paint:
				w.repaint()
			case f := <-w.do:
				f()
			}
		}
	}()
	wg.Wait()
}

func (w *Window) setKeyboardFocus(view View) {
	if w.keyboardFocus != view {
		// change w.keyboardFocus first to avoid possible infinite recursion
		oldFocus := w.keyboardFocus
		w.keyboardFocus = view
		if oldFocus != nil {
			oldFocus.LostKeyboardFocus()
		}
		if w.keyboardFocus != nil {
			w.keyboardFocus.TookKeyboardFocus()
		}
	}
}
func (w Window) GetKeyboardFocus() View { return w.keyboardFocus }

func (w *Window) setMouseFocus(focus MouserView, button int) { w.mouseFocus[button] = focus }

func (w *Window) Repaint() {
	select {
	case w.paint <- struct{}{}:
	default:
	}
}

func (w Window) repaint() {
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	w.base().paint()
	w.w.SwapBuffers()
}

func (w Window) Do(f func()) {
	w.do <- f
}
