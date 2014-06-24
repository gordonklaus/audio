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
}

func NewWindow(self View, title string, init func(w *Window)) {
	w := &Window{}
	if self == nil {
		self = w
	}
	w.ViewBase = NewView(self)
	w.mouser = make(map[int]MouserView)
	w.paint = make(chan bool, 1)
	w.do = make(chan func())

	doMain(func() {
		w.w = glfw.NewWindow(960, 520, title)

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
		w.w.OnResize(func(width, height int) {
			w.Do(func() {
				s := Pt(float64(width), float64(height))
				gl.MatrixMode(gl.PROJECTION)
				gl.LoadIdentity()
				gl.Ortho(0, gl.Double(s.X), 0, gl.Double(s.Y), -1, 1)
				Resize(w.Self, s)
				if w.centralView != nil {
					Resize(w.centralView, s)
				}
			})
		})
		w.w.OnFramebufferResize(func(width, height int) {
			w.Do(func() {
				gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))
			})
		})

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
				s := ScrollEvent{m.Pos, Pt(dx, -dy)}
				s.Pos = w.mapToWindow(s.Pos)
				v, _ := viewAtFunc(w.Self, s.Pos, func(v View) View {
					v, _ = v.(ScrollerView)
					return v
				}).(ScrollerView)
				if v != nil {
					s.Pos = MapFrom(v, s.Pos, w.Self)
					v.Scroll(s)
				}
			})
		})

		windows = append([]*Window{w}, windows...)
	})

	go func() {
		runtime.LockOSThread()
		glfw.MakeContextCurrent(w.w)
		defer glfw.MakeContextCurrent(nil)

		// glfw should fire initial resize events to avoid this duplication (https://github.com/glfw/glfw/issues/62)
		width, height := w.w.Size()
		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()
		gl.Ortho(0, gl.Double(width), 0, gl.Double(height), -1, 1)
		Resize(w, Pt(float64(width), float64(height)))
		width, height = w.w.FramebufferSize()
		gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))

		init(w)

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

				gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
				w.base().paint()
				w.w.SwapBuffers()
			}
		}
	}()
}

func (w *Window) mapToWindow(p Point) Point {
	p.Y = Height(w) - p.Y
	return p.Add(Rect(w).Min)
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
				m.Pos = MapFrom(v, m.Pos, w.Self)
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
						m.Pos = MapFrom(v, m.Pos, w.Self)
						m.Leave = true
						v.Mouse(m)
					}
				}
				for v := View(v); v != p && v != nil; v = Parent(v) {
					if v, ok := v.(MouserView); ok {
						m := m
						m.Pos = MapFrom(v, m.Pos, w.Self)
						m.Enter = true
						v.Mouse(m)
					}
				}
				w.mouseIn = v
			}
			for button, v := range w.mouser {
				m := m
				m.Pos = MapFrom(v, m.Pos, w.Self)
				m.Drag = true
				m.Button = button
				v.Mouse(m)
			}
		case m.Release:
			m.Pos = w.mapToWindow(m.Pos)
			if v, ok := w.mouser[m.Button]; ok {
				m.Pos = MapFrom(v, m.Pos, w.Self)
				v.Mouse(m)
				delete(w.mouser, m.Button)
			}
		}
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
		Resize(v, Size(w))
		SetKeyFocus(v)
	}
}

func commonParent(v1, v2 View) (p View) {
	for ; v1 != nil; v1 = Parent(v1) {
		for v2 := v2; v2 != nil; v2 = Parent(v2) {
			if v1 == v2 {
				return v1
			}
		}
	}
	return nil
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
