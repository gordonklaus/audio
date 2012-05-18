package main

import (
	"github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/gui"
)

type put struct {
	ViewBase
	spec putSpecializer
	node Node
	info ValueInfo
	connections []*Connection
	focused bool
}

const putSize = 11

type putSpecializer interface {
	View
	ConnectTo(conn *Connection)
	PassMouseFocusToFreeConnectionHandle(conn *Connection, button int)
}

func Newput(spec putSpecializer, n Node, info ValueInfo) *put {
	p := &put{}
	p.ViewBase = *NewView(spec)
	p.spec = spec
	p.node = n
	p.info = info
	p.Resize(putSize, putSize)
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

func (p *put) TookKeyboardFocus() { p.focused = true; p.Repaint() }
func (p *put) LostKeyboardFocus() { p.focused = false; p.Repaint() }

func (p *put) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyEnter:
		conn := p.node.Function().NewConnection(p.Center())
		p.spec.ConnectTo(conn)
		conn.BeStraightLine()
		conn.StartEditing()
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		p.node.Function().FocusNearestView(p.spec, event.Key)
	case glfw.KeyEsc:
		p.node.TakeKeyboardFocus()
	default:
		p.ViewBase.KeyPressed(event)
	}
}

func (p *put) MousePressed(button int, pt Point) {
	p.TakeKeyboardFocus()
	conn := p.node.Function().NewConnection(p.MapTo(pt, p.node.Function()))
	p.spec.ConnectTo(conn)
	p.spec.PassMouseFocusToFreeConnectionHandle(conn, button)
	conn.StartEditing()
}
func (p put) MouseDragged(button int, pt Point) {}
func (p put) MouseReleased(button int, pt Point) {}

func (p put) Paint() {
	width, height := gl.Double(p.Width()), gl.Double(p.Height())
	if p.focused {
		gl.Color4d(.6, .6, 1, .6)
	} else {
		gl.Color4d(0, 0, 0, .5)
	}
	gl.Rectd(0, 0, width, height)
	gl.Color4d(1, 1, 1, 1)
	gl.Begin(gl.LINE_LOOP)
	gl.Vertex2d(0, 0)
	gl.Vertex2d(width, 0)
	gl.Vertex2d(width, height)
	gl.Vertex2d(0, height)
	gl.End()
}


type Input struct {
	put
}
func NewInput(n Node, info ValueInfo) *Input {
	p := &Input{}
	p.put = *Newput(p, n, info)
	return p
}

func (p *Input) ConnectTo(conn *Connection) { conn.SetDestination(p) }
func (p *Input) PassMouseFocusToFreeConnectionHandle(conn *Connection, button int) { conn.srcHandle.SetMouseFocus(conn.srcHandle, button) }

func (p *Input) KeyPressed(event KeyEvent) {
	if event.Key == glfw.KeyDown && len(p.connections) > 0 {
		p.connections[0].TakeKeyboardFocus()
	} else if event.Key == glfw.KeyUp {
		p.node.TakeKeyboardFocus()
	} else {
		p.put.KeyPressed(event)
	}
}


type Output struct {
	put
}
func NewOutput(n Node, info ValueInfo) *Output {
	p := &Output{}
	p.put = *Newput(p, n, info)
	return p
}

func (p *Output) ConnectTo(conn *Connection) { conn.SetSource(p) }
func (p *Output) PassMouseFocusToFreeConnectionHandle(conn *Connection, button int) { conn.dstHandle.SetMouseFocus(conn.dstHandle, button) }

func (p *Output) KeyPressed(event KeyEvent) {
	if event.Key == glfw.KeyDown {
		p.node.TakeKeyboardFocus()
	} else if event.Key == glfw.KeyUp && len(p.connections) > 0 {
		p.connections[0].TakeKeyboardFocus()
	} else {
		p.put.KeyPressed(event)
	}
}
