package gui

import (
	."github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
)

func init() {
	if err := Init(); err != nil { panic(err) }
	if err := gl.Init(); err != nil { panic(err) }
	if err := OpenWindow(800, 600, 8, 8, 8, 8, 0, 0, Windowed); err != nil { panic(err) }
	gl.Enable(gl.BLEND)
	gl.Enable(gl.POINT_SMOOTH)
	gl.Enable(gl.LINE_SMOOTH)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	Disable(AutoPollEvents)
	SetSwapInterval(1)
}

type Window struct {
	*ViewBase
	centralView View
	keyboardFocus View
	mouseFocus map[int]MouseHandlerView
	repaintMe bool
}

func NewWindow(self View) *Window {
	w := &Window{nil, nil, nil, make(map[int]MouseHandlerView), false}
	if self == nil { self = w }
	w.ViewBase = NewView(w)
	w.Self = self
	return w
}

func (w *Window) SetCentralView(v View) {
	if w.centralView != nil { w.RemoveChild(w.centralView) }
	w.centralView = v
	if v != nil {
		if v.Parent() != w { w.AddChild(v) }
		v.Resize(w.Size().XY())
		w.SetKeyboardFocus(v)
	}
}

func (w *Window) Close() {
	CloseWindow()
	Terminate()
}

func (w *Window) HandleEvents() {
	SetWindowSizeCallback(func(width, height int) {
		wid, hei := float64(width), float64(height)
		w.Self.Resize(wid, hei)
		if w.centralView != nil { w.centralView.Resize(wid, hei) }
	})
	
	keyEvent := KeyEvent{}
	SetKeyCallback(func(key, state int) {
		keyEvent.Key = key
		if key > KeySpecial {
			keyEvent.Text = ""
			switch key {
			case KeyLshift, KeyRshift: keyEvent.Shift = state == KeyPress
			case KeyLctrl, KeyRctrl: keyEvent.Ctrl = state == KeyPress
			case KeyLalt, KeyRalt: keyEvent.Alt = state == KeyPress
			case KeyLsuper, KeyRsuper: keyEvent.Super = state == KeyPress
			}
			if state == KeyPress {
				if w.keyboardFocus != nil { w.keyboardFocus.KeyPressed(keyEvent) }
			} else if state == KeyRelease {
				if w.keyboardFocus != nil { w.keyboardFocus.KeyReleased(keyEvent) }
			}
		}
	})
	SetCharCallback(func(char, state int) {
		if char < KeySpecial {
			keyEvent.Text = string(char)
			if state == KeyPress {
				if w.keyboardFocus != nil { w.keyboardFocus.KeyPressed(keyEvent) }
			} else if state == KeyRelease {
				if w.keyboardFocus != nil { w.keyboardFocus.KeyReleased(keyEvent) }
			}
		}
	})
	
	var mousePos Point
	SetMousePosCallback(func(x, y int) {
		mousePos = Pt(float64(x), w.Height() - float64(y))
		for button, v := range w.mouseFocus {
			pt := v.MapFrom(mousePos, w.Self)
			v.MouseDragged(button, pt)
		}
	})
	SetMouseButtonCallback(func(button, state int) {
		if state == KeyPress {
			v := w.Self.GetMouseFocus(button, mousePos)
			if v != nil {
				w.mouseFocus[button] = v
				pt := v.MapFrom(mousePos, w.Self)
				v.MousePressed(button, pt)
			}
		} else if state == KeyRelease {
			if v, ok := w.mouseFocus[button]; ok {
				pt := v.MapFrom(mousePos, w.Self)
				v.MouseReleased(button, pt)
				delete(w.mouseFocus, button)
			}
		}
	})

	for WindowParam(Opened) == 1 {
		PollEvents()
		w.repaint()
	}
}

func (w *Window) SetKeyboardFocus(view View) {
	if w.keyboardFocus != view {
		// change w.keyboardFocus first to avoid possible infinite recursion
		oldFocus := w.keyboardFocus
		w.keyboardFocus = view
		if oldFocus != nil { oldFocus.LostKeyboardFocus() }
		if w.keyboardFocus != nil { w.keyboardFocus.TookKeyboardFocus() }
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
	SwapBuffers()
}
