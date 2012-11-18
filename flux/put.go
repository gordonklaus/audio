package main

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type put struct {
	*ViewBase
	spec putSpecializer
	node Node
	info *ValueInfo
	valueView *typeView
	connections []*Connection
	focused bool
}

const putSize = 11

type putSpecializer interface {
	View
	ConnectTo(conn *Connection)
	PassMouseFocusToFreeConnectionHandle(conn *Connection, button int)
}

func Newput(spec putSpecializer, n Node, info *ValueInfo) *put {
	p := &put{}
	p.ViewBase = NewView(p)
	p.spec = spec
	p.node = n
	p.info = info
	p.valueView = newValueView(info)
	p.valueView.Hide()
	p.AddChild(p.valueView)
	p.Resize(putSize, putSize)
	p.Pan(Pt(putSize, putSize).Div(-2))
	p.Self = spec
	return p
}

func (p put) CanConnect(interface{}) bool { return true }
func (p *put) Connect(conn *Connection) { p.connections = append(p.connections, conn) }
func (p *put) Disconnect(conn *Connection) {
	for i, connection := range p.connections {
		if connection == conn {
			p.connections = append(p.connections[:i], p.connections[i+1:]...)
			return
		}
	}
}

func (p *put) TookKeyboardFocus() { p.focused = true; p.Repaint(); p.valueView.Show() }
func (p *put) LostKeyboardFocus() { p.focused = false; p.Repaint(); p.valueView.Hide() }

func (p *put) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyEnter:
		conn := p.node.Block().NewConnection(p.Center())
		p.spec.ConnectTo(conn)
		conn.BeStraightLine()
		conn.StartEditing()
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		p.node.Block().Outermost().FocusNearestView(p.spec, event.Key)
	case glfw.KeyEsc:
		p.node.TakeKeyboardFocus()
	default:
		p.ViewBase.KeyPressed(event)
	}
}

func (p *put) MousePressed(button int, pt Point) {
	p.TakeKeyboardFocus()
	conn := p.node.Block().NewConnection(p.MapTo(pt, p.node.Block()))
	p.spec.ConnectTo(conn)
	p.spec.PassMouseFocusToFreeConnectionHandle(conn, button)
	conn.StartEditing()
}
func (p put) MouseDragged(button int, pt Point) {}
func (p put) MouseReleased(button int, pt Point) {}

func (p put) Paint() {
	color := map[bool]Color{true:{.5, .5, 1, .5}, false:{1, 1, 1, .25}}[p.focused]
	for f := 1.0; f > .1; f /= 2 {
		SetColor(color)
		SetPointSize(f * putSize)
		DrawPoint(ZP)
	}
}


type Input struct {
	*put
}
func NewInput(n Node, info *ValueInfo) *Input {
	p := &Input{}
	p.put = Newput(p, n, info)
	p.valueView.Move(Pt(-p.valueView.Width() - 12, -p.valueView.Height() / 2))
	return p
}

func (p *Input) ConnectTo(conn *Connection) { conn.SetDestination(p) }
func (p *Input) PassMouseFocusToFreeConnectionHandle(conn *Connection, button int) { conn.srcHandle.SetMouseFocus(conn.srcHandle, button) }

func (p *Input) KeyPressed(event KeyEvent) {
	if event.Key == glfw.KeyLeft && len(p.connections) > 0 {
		p.connections[0].TakeKeyboardFocus()
	} else if event.Key == glfw.KeyRight {
		p.node.TakeKeyboardFocus()
	} else {
		p.put.KeyPressed(event)
	}
}


type Output struct {
	*put
}
func NewOutput(n Node, info *ValueInfo) *Output {
	p := &Output{}
	p.put = Newput(p, n, info)
	p.valueView.Move(Pt(12, -p.valueView.Height() / 2))
	return p
}

func (p *Output) ConnectTo(conn *Connection) { conn.SetSource(p) }
func (p *Output) PassMouseFocusToFreeConnectionHandle(conn *Connection, button int) { conn.dstHandle.SetMouseFocus(conn.dstHandle, button) }

func (p *Output) KeyPressed(event KeyEvent) {
	if event.Key == glfw.KeyLeft {
		p.node.TakeKeyboardFocus()
	} else if event.Key == glfw.KeyRight && len(p.connections) > 0 {
		p.connections[0].TakeKeyboardFocus()
	} else {
		p.put.KeyPressed(event)
	}
}
