package main

import (
	"github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/gui"
)

type put struct {
	ViewBase
	AggregateMouseHandler
	spec putSpecializer
	node *Node
	connections []*Connection
	focused bool
}

const putSize = 11

type putSpecializer interface {
	View
	ConnectToConnection(conn *Connection)
}

func Newput(spec putSpecializer, n *Node) *put {
	p := &put{}
	p.ViewBase = *NewView(spec)
	p.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(p)}
	p.spec = spec
	p.node = n
	p.Resize(putSize, putSize)
	return p
}

func (p *put) ConnectConnection(conn *Connection) { p.connections = append(p.connections, conn) }
func (p *put) DisconnectConnection(conn *Connection) {
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
	// case glfw.KeyEnter:
	// 	conn := p.node.function.NewConnection(p.Center())
	// 	p.spec.ConnectToConnection(conn)
	// 	conn.BeStraightLine()
	// 	conn.StartEditing()
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		p.node.function.FocusNearestView(p.spec, event.Key)
	case glfw.KeyEsc:
		p.node.TakeKeyboardFocus()
	default:
		p.ViewBase.KeyPressed(event)
	}
}

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
func NewInput(n *Node) *Input {
	p := &Input{}
	p.put = *Newput(p, n)
	return p
}

func (p *Input) ConnectToConnection(conn *Connection) { conn.SetDestination(p) }

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
	put
}
func NewOutput(n *Node) *Output {
	p := &Output{}
	p.put = *Newput(p, n)
	return p
}

func (p *Output) ConnectToConnection(conn *Connection) { conn.SetSource(p) }

func (p *Output) KeyPressed(event KeyEvent) {
	if event.Key == glfw.KeyLeft {
		p.node.TakeKeyboardFocus()
	} else if event.Key == glfw.KeyRight && len(p.connections) > 0 {
		p.connections[0].TakeKeyboardFocus()
	} else {
		p.put.KeyPressed(event)
	}
}
