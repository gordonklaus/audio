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
	centralView View
	keyFocus    View
	mouser      map[int]MouserView
	control     chan interface{}
	key         chan KeyEvent
	mouse       chan MouseEvent
	scroll      chan ScrollEvent
	paint       chan struct{}
	do          chan func()
}

func NewWindow(self View) *Window {
	w := &Window{w: glfw.NewWindow(960, 520, "Flux")}

	// Somehow, this seems to work:
	// Have the context current in both threads, as this one needs to be able to call
	// Font.Advance and Font.LineHeight (from Text.SetText), and the other thread renders.
	// But is this portable?  Or even a good idea?
	glfw.MakeContextCurrent(w.w)

	if self == nil {
		self = w
	}
	w.ViewBase = NewView(w)
	w.mouser = make(map[int]MouserView)
	w.control = make(chan interface{})
	w.key = make(chan KeyEvent)
	w.mouse = make(chan MouseEvent)
	w.scroll = make(chan ScrollEvent)
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

func (w *Window) win() *Window { return w }

func (w *Window) SetCentralView(v View) {
	if w.centralView != nil {
		w.Remove(w.centralView)
	}
	w.centralView = v
	if v != nil {
		if Parent(v) != w {
			w.Add(v)
		}
		Resize(v, Size(w))
		SetKeyFocus(v)
	}
}

func (w *Window) run() {
	type (
		close  *Window
		resize Point
	)

	w.w.OnClose(func() {
		// callbacks may still be called after OnClose so they must be unregistered to avoid deadlock
		w.w.OnResize(nil)
		w.w.OnFramebufferResize(nil)
		w.w.OnKey(nil)
		w.w.OnChar(nil)
		w.w.OnMouseMove(nil)
		w.w.OnMouseButton(nil)
		w.w.OnScroll(nil)
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
		keyEvent.action = action
		keyEvent.Repeat = action == glfw.Repeat
		keyEvent.Shift = mods&glfw.ModShift != 0
		keyEvent.Ctrl = mods&glfw.ModControl != 0
		keyEvent.Alt = mods&glfw.ModAlt != 0
		keyEvent.Super = mods&glfw.ModSuper != 0
		keyEvent.Command = commandKey(keyEvent)
		if key >= KeyEscape || action == glfw.Release {
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
		m.Pos = Pt(x, y)
		m.Move, m.Press, m.Release, m.Drag = true, false, false, false
		w.mouse <- m
	})
	w.w.OnMouseButton(func(button, action, mods int) {
		m.Button = button
		m.Move, m.Press, m.Release, m.Drag = false, action == glfw.Press, action == glfw.Release, false
		w.mouse <- m
	})
	w.w.OnScroll(func(dx, dy float64) {
		w.scroll <- ScrollEvent{m.Pos, Pt(dx, -dy)}
	})

	mapToWindow := func(p Point) Point {
		p.Y = Height(w) - p.Y
		return p.Add(Rect(w).Min)
	}

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
				if w.keyFocus != nil {
					if k.action != glfw.Release {
						w.keyFocus.KeyPress(k)
					} else {
						w.keyFocus.KeyRelease(k)
					}
				}
			case m := <-w.mouse:
				switch {
				case m.Press:
					m.Pos = mapToWindow(m.Pos)
					v, _ := viewAtFunc(w.Self, m.Pos, func(v View) View {
						v, _ = v.(MouserView)
						return v
					}).(MouserView)
					if v != nil {
						w.mouser[m.Button] = v
						m.Pos = MapFrom(v, m.Pos, w.Self)
						v.Mouse(m)
					}
				case m.Move:
					m.Pos = mapToWindow(m.Pos)
					for button, v := range w.mouser {
						m := m
						m.Pos = MapFrom(v, m.Pos, w.Self)
						m.Move, m.Drag = false, true
						m.Button = button
						v.Mouse(m)
					}
				case m.Release:
					m.Pos = mapToWindow(m.Pos)
					if v, ok := w.mouser[m.Button]; ok {
						m.Pos = MapFrom(v, m.Pos, w.Self)
						v.Mouse(m)
						delete(w.mouser, m.Button)
					}
				}
			case s := <-w.scroll:
				s.Pos = mapToWindow(s.Pos)
				v, _ := viewAtFunc(w.Self, s.Pos, func(v View) View {
					v, _ = v.(ScrollerView)
					return v
				}).(ScrollerView)
				if v != nil {
					s.Pos = MapFrom(v, s.Pos, w.Self)
					v.Scroll(s)
				}
			case <-w.paint:
				gl.MatrixMode(gl.MODELVIEW)
				gl.LoadIdentity()

				gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
				w.base().paint()
				w.w.SwapBuffers()
			case f := <-w.do:
				f()
			}
		}
	}()
	wg.Wait()
}

func (w *Window) setKeyFocus(view View) {
	if w.keyFocus != view {
		// change w.keyFocus first to avoid possible infinite recursion
		oldFocus := w.keyFocus
		w.keyFocus = view
		if oldFocus != nil {
			oldFocus.LostKeyFocus()
		}
		if w.keyFocus != nil {
			w.keyFocus.TookKeyFocus()
		}
	}
}

func (w *Window) setMouser(m MouserView, button int) { w.mouser[button] = m }

func (w *Window) repaint() {
	select {
	case w.paint <- struct{}{}:
	default:
	}
}
