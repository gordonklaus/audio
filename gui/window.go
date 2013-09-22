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
	keyboardFocus View
	mouseFocus map[int]MouseHandlerView
	control chan interface{}
	key chan KeyEvent
	mouse chan interface{}
	paint chan paint
	do chan func()
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
	w.key = make(chan KeyEvent, 10)
	w.mouse = make(chan interface{}, 100)
	w.paint = make(chan paint, 1)
	w.do = make(chan func())
	
	// glfw should fire initial resize events to avoid this duplication (https://github.com/glfw/glfw/issues/62)
	width, height := w.w.Size()
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, gl.Double(width), 0, gl.Double(height), -1, 1)
	w.Resize(float64(width), float64(height))
	width, height = w.w.FramebufferSize()
	gl.Viewport(0, 0, gl.Sizei(width), gl.Sizei(height))
	
	windows[w] = true
	w.run()
	
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

type paint struct {}
type close struct {
	w *Window
}
type resize struct {
	w, h float64
}
type mouseMove struct {
	Pos Point
}
type mouseButton struct {
	Pos Point
	Button int
	Action int
}

func (w *Window) run() {
	w.w.OnClose(func() {
		w.control <- close{w}
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
		keyEvent.Shift = mods & glfw.ModShift != 0
		keyEvent.Ctrl = mods & glfw.ModControl != 0
		keyEvent.Alt = mods & glfw.ModAlt != 0
		keyEvent.Super = mods & glfw.ModSuper != 0
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
	
	var pos Point
	w.w.OnMouseMove(func(x, y float64) {
		pos = Pt(x, w.Height() - y)
		w.mouse <- mouseMove{pos}
	})
	w.w.OnMouseButton(func(button, action, mods int) {
		w.mouse <- mouseButton{pos, button, action}
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
					closeWindow <- c.w
					return
				case resize:
					gl.MatrixMode(gl.PROJECTION)
					gl.LoadIdentity()
					gl.Ortho(0, gl.Double(c.w), 0, gl.Double(c.h), -1, 1)
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
				case mouseMove:
					for button, v := range w.mouseFocus {
						pt := v.MapFrom(m.Pos, w.Self)
						v.MouseDragged(button, pt)
					}
				case mouseButton:
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
			case <-w.paint:
				w.repaint()
			case f := <-w.do:
				f()
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
	case w.paint <- paint{}:
	default:
	}
}

func (w Window) repaint() {
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	w.GetViewBase().paintBase()
	w.w.SwapBuffers()
}

func (w Window) Do(f func()) {
	w.do <- f
}
