package gui

import (
	"code.google.com/p/gordon-go/glfw"
	gl "github.com/chsc/gogl/gl21"
	"runtime"
	"sync"
	"time"
)

type Window struct {
	w *glfw.Window
	*ViewBase
	centralView View
	keyboardFocus View
	mouseFocus map[int]MouseHandlerView
	control chan interface{}
	key chan KeyEvent
	mouse chan interface{}
	paint chan Paint
}

func NewWindow(self View) *Window {
	w := &Window{w:glfw.NewWindow(800, 600, "Flux")}
	
	// Somehow, this seems to work:
	// Have the context current in both threads, as this one needs to be able to call
	// Font.Advance and Font.LineHeight (from Text.SetText), and the other thread renders.
	// But is this portable?  Or even a good idea?
	glfw.MakeContextCurrent(w.w)
	
	if self == nil { self = w }
	w.ViewBase = NewView(w)
	w.mouseFocus = make(map[int]MouseHandlerView)
	w.control = make(chan interface{})
	w.key = make(chan KeyEvent, 1)
	w.mouse = make(chan interface{})
	w.paint = make(chan Paint, 1)
	
	windows[w] = true
	w.run()
	width, height := w.w.Size()
	w.Resize(float64(width), float64(height))
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

type Paint struct {}
type Close struct {
	w *Window
}
type Resize struct {
	w, h float64
}
type MouseMove struct {
	Pos Point
}
type MouseButton struct {
	Pos Point
	Button int
	Action int
}

func (w *Window) run() {
	w.w.OnClose(func() bool {
		w.control <- Close{w}
		return false
	})
	w.w.OnResize(func(width, height int) {
		w.control <- Resize{float64(width), float64(height)}
	})
	
	keyEvent := KeyEvent{}
	w.w.OnKey(func(key, action int) {
		keyEvent.Key = key
		keyEvent.Action = action
		if key >= KeyEscape {
			keyEvent.Text = ""
			switch key {
			case KeyLeftShift, KeyRightShift: keyEvent.Shift = action == Press
			case KeyLeftControl, KeyRightControl: keyEvent.Ctrl = action == Press
			case KeyLeftAlt, KeyRightAlt: keyEvent.Alt = action == Press
			case KeyLeftSuper, KeyRightSuper: keyEvent.Super = action == Press
			}
			w.key <- keyEvent
		} else if action == Release {
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
	
	var pos Point
	w.w.OnMouseMove(func(x, y int) {
		pos = Pt(float64(x), w.Height() - float64(y))
		w.mouse <- MouseMove{pos}
	})
	w.w.OnMouseButton(func(button, action int) {
		w.mouse <- MouseButton{pos, button, action}
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
		
		ticker := time.NewTicker(33 * time.Millisecond)
		for {
			select {
			case c := <-w.control:
				switch c := c.(type) {
				case Close:
					windowControl <- c
					return
				case Resize:
					w.Self.Resize(c.w, c.h)
					if w.centralView != nil { w.centralView.Resize(c.w, c.h) }
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
				switch m := m.(type) {
				case MouseMove:
					for button, v := range w.mouseFocus {
						pt := v.MapFrom(m.Pos, w.Self)
						v.MouseDragged(button, pt)
					}
				case MouseButton:
					if m.Action == Press {
						v := w.Self.GetMouseFocus(m.Button, m.Pos)
						if v != nil {
							w.mouseFocus[m.Button] = v
							pt := v.MapFrom(m.Pos, w.Self)
							v.MousePressed(m.Button, pt)
						}
					} else if m.Action == Release {
						if v, ok := w.mouseFocus[m.Button]; ok {
							pt := v.MapFrom(m.Pos, w.Self)
							v.MouseReleased(m.Button, pt)
							delete(w.mouseFocus, m.Button)
						}
					}
				}
			case <-ticker.C:
				w.repaint()
			}
		}
	} ()
	wg.Wait()
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
func (w Window) GetKeyboardFocus() View { return w.keyboardFocus }

func (w *Window) SetMouseFocus(focus MouseHandlerView, button int) { w.mouseFocus[button] = focus }

func (w *Window) Repaint() {
	select {
	case w.paint <- Paint{}:
	default:
	}
}

func (w Window) repaint() {
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
	w.w.SwapBuffers()
}
