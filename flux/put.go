package main

import (
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/gui"
)

type put struct {
	ViewBase
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
	// switconn key {
	// case Key_Enter:
	// 	conn := p.node.function.NewConnection(p.Center())
	// 	p.spec.ConnectToConnection(conn)
	// 	conn.BeStraightLine()
	// 	conn.StartEditing()
	// case Key_Left, Key_Right, Key_Up, Key_Down:
	// 	p.node.function.FocusNearestView(p.spec, key)
	// case Key_Escape:
	// 	p.node.TakeKeyboardFocus()
	// default:
	// 	p.ViewBase.KeyPressed(key)
	// }
}

func (p put) Paint() {
	width, height := gl.Double(p.Width()), gl.Double(p.Height())
	gl.Color4d(0, 0, 0, .5)
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
	// if key == Key_Left && len(p.connections) > 0 {
	// 	p.connections[0].dstHandle.TakeKeyboardFocus()
	// } else if key == Key_Right {
	// 	p.node.TakeKeyboardFocus()
	// } else {
	// 	p.put.KeyPressed(key)
	// }
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
	// if key == Key_Left {
	// 	p.node.TakeKeyboardFocus()
	// } else if key == Key_Right && len(p.connections) > 0 {
	// 	p.connections[0].srcHandle.TakeKeyboardFocus()
	// } else {
	// 	p.put.KeyPressed(key)
	// }
}
