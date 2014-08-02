package gui

import (
	"code.google.com/p/gordon-go/glfw"
	gl "github.com/chsc/gogl/gl21"
	"runtime"
)

type Window struct {
	w *glfw.Window
	*ViewBase
	centralView View
	keyFocus    View
	mouseIn     MouserView
	mouser      map[int]MouserView
	close       bool
	paint       chan bool
	do          chan func()

	bufWidth, bufHeight gl.Sizei
}

func NewWindow(self View, title string, init func(w *Window)) {
	w := &Window{}
	if self == nil {
		self = w
	}
	doMain(func() {
		w.w = glfw.NewWindow(960, 520, title)
		windows = append([]*Window{w}, windows...)
	})
	w.ViewBase = NewView(self)
	w.mouser = make(map[int]MouserView)
	w.paint = make(chan bool, 1)
	w.do = make(chan func())
	go w.run(init)
	go doMain(w.registerCallbacks)
}

func (w *Window) run(init func(w *Window)) {
	runtime.LockOSThread()
	glfw.MakeContextCurrent(w.w)
	defer glfw.MakeContextCurrent(nil)

	init(w)

	// glfw should fire initial resize events to avoid this duplication (https://github.com/glfw/glfw/issues/62)
	w.resized(w.w.Size())
	w.framebufferResized(w.w.FramebufferSize())

	gl.Enable(gl.SCISSOR_TEST)
	gl.Enable(gl.BLEND)
	gl.Enable(gl.POINT_SMOOTH)
	gl.Enable(gl.LINE_SMOOTH)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	for !w.close {
		select {
		case f := <-w.do:
			f()
		case <-w.paint:
			gl.MatrixMode(gl.MODELVIEW)
			gl.LoadIdentity()

			gl.Scissor(0, 0, w.bufWidth, w.bufHeight)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			w.base().paint()
			w.w.SwapBuffers()
		}
	}
}

func (w *Window) registerCallbacks() {
	w.w.OnFocus(func(focused bool) {
		if focused {
			go doMain(func() {
				for i, w2 := range windows {
					if w2 == w {
						windows = append(append([]*Window{w}, windows[:i]...), windows[i+1:]...)
						break
					}
				}
			})
		}
	})
	w.w.OnClose(w.Close)
	w.w.OnResize(func(width, height int) { w.Do(func() { w.resized(width, height) }) })
	w.w.OnFramebufferResize(func(width, height int) { w.Do(func() { w.framebufferResized(width, height) }) })

	k := KeyEvent{}
	w.w.OnKey(func(key, scancode, action, mods int) {
		w.Do(func() {
			k.Key = key
			k.action = action
			k.Repeat = action == glfw.Repeat
			k.Shift = mods&glfw.ModShift != 0
			k.Ctrl = mods&glfw.ModControl != 0
			k.Alt = mods&glfw.ModAlt != 0
			k.Super = mods&glfw.ModSuper != 0
			k.Command = commandKey(k)
			if key >= KeyEscape || action == glfw.Release {
				k.Text = ""
				if w.keyFocus != nil {
					if k.action != glfw.Release {
						w.keyFocus.KeyPress(k)
					} else {
						w.keyFocus.KeyRelease(k)
					}
				}
			}
		})
	})
	w.w.OnChar(func(char rune) {
		w.Do(func() {
			if char < KeyEscape {
				k.Text = string(char)
				if w.keyFocus != nil {
					if k.action != glfw.Release {
						w.keyFocus.KeyPress(k)
					} else {
						w.keyFocus.KeyRelease(k)
					}
				}
			}
		})
	})

	m := MouseEvent{}
	w.w.OnMouseMove(func(x, y float64) {
		m.Pos = Pt(x, y)
		m.Move, m.Press, m.Release, m.Drag = true, false, false, false
		w.mouse(m)
	})
	w.w.OnMouseButton(func(button, action, mods int) {
		m.Button = button
		m.Move, m.Press, m.Release, m.Drag = false, action == glfw.Press, action == glfw.Release, false
		w.mouse(m)
	})
	w.w.OnScroll(func(dx, dy float64) {
		w.Do(func() {
			s := ScrollEvent{m.Pos, Pt(dx, -dy), k.Shift, k.Ctrl, k.Alt, k.Super, k.Command}
			s.Pos = w.mapToWindow(s.Pos)
			v, _ := viewAtFunc(w.Self, s.Pos, func(v View) View {
				v, _ = v.(ScrollerView)
				return v
			}).(ScrollerView)
			if v != nil {
				s.Pos = Map(s.Pos, w.Self, v)
				v.Scroll(s)
			}
		})
	})
}

func (w *Window) resized(width, height int) {
	wid, hei := float64(width), float64(height)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, gl.Double(wid), 0, gl.Double(hei), -1, 1)
	w.Self.Resize(wid, hei)
	if w.centralView != nil {
		w.centralView.Resize(wid, hei)
	}
}

func (w *Window) framebufferResized(width, height int) {
	w.bufWidth, w.bufHeight = gl.Sizei(width), gl.Sizei(height)
	gl.Viewport(0, 0, w.bufWidth, w.bufHeight)
}

func (w *Window) mouse(m MouseEvent) {
	w.Do(func() {
		switch {
		case m.Press:
			m.Pos = w.mapToWindow(m.Pos)
			v, _ := viewAtFunc(w.Self, m.Pos, func(v View) View {
				v, _ = v.(MouserView)
				return v
			}).(MouserView)
			if v != nil {
				w.mouser[m.Button] = v
				m.Pos = Map(m.Pos, w.Self, v)
				v.Mouse(m)
			}
		case m.Move:
			m.Pos = w.mapToWindow(m.Pos)
			m.Move = false
			v, _ := viewAtFunc(w.Self, m.Pos, func(v View) View {
				v, _ = v.(MouserView)
				return v
			}).(MouserView)
			if w.mouseIn != v {
				p := commonParent(w.mouseIn, v)
				for v := View(w.mouseIn); v != p && v != nil; v = Parent(v) {
					if v, ok := v.(MouserView); ok {
						m := m
						m.Pos = Map(m.Pos, w.Self, v)
						m.Leave = true
						v.Mouse(m)
					}
				}
				for v := View(v); v != p && v != nil; v = Parent(v) {
					if v, ok := v.(MouserView); ok {
						m := m
						m.Pos = Map(m.Pos, w.Self, v)
						m.Enter = true
						v.Mouse(m)
					}
				}
				w.mouseIn = v
			}
			for button, v := range w.mouser {
				m := m
				m.Pos = Map(m.Pos, w.Self, v)
				m.Drag = true
				m.Button = button
				v.Mouse(m)
			}
		case m.Release:
			m.Pos = w.mapToWindow(m.Pos)
			if v, ok := w.mouser[m.Button]; ok {
				m.Pos = Map(m.Pos, w.Self, v)
				v.Mouse(m)
				delete(w.mouser, m.Button)
			}
		}
	})
}

func (w *Window) mapToWindow(p Point) Point {
	p.Y = Height(w) - p.Y
	return InnerRect(w).Min.Add(p)
}

func (w *Window) Do(f func()) {
	done := make(chan bool)
	w.do <- func() {
		f()
		done <- true
	}
	<-done
}

func (w *Window) Close() {
	go doMain(func() {
		closeWindow(w)
	})
}

func (w *Window) SetTitle(s string) { w.w.SetTitle(s) }

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
		v.Resize(Size(w))
		SetKeyFocus(v)
	}
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

func (w *Window) KeyPress(k KeyEvent) {
	if k.Command {
		switch k.Key {
		case KeyW:
			w.Close()
		case KeyQ:
			Quit()
		}
	}
}

func (w *Window) repaint() {
	select {
	case w.paint <- true:
	default:
	}
}
