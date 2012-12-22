package main

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type Input struct { *port }
type Output struct { *port }
type port struct {
	*ViewBase
	out bool
	node Node
	info *ValueInfo
	valueView *typeView
	connections []*Connection
	focused bool
	connectionsChanged func()
}

const portSize = 11

func newInput(n Node, info *ValueInfo) *Input {
	p := &Input{}
	p.port = newPort(p, false, n, info)
	p.valueView.Move(Pt(-p.valueView.Width() - 12, -p.valueView.Height() / 2))
	return p
}
func newOutput(n Node, info *ValueInfo) *Output {
	p := &Output{}
	p.port = newPort(p, true, n, info)
	p.valueView.Move(Pt(12, -p.valueView.Height() / 2))
	return p
}
func newPort(self View, out bool, n Node, info *ValueInfo) *port {
	p := &port{out:out, node:n, info:info}
	p.ViewBase = NewView(p)
	p.valueView = newValueView(info)
	p.valueView.Hide()
	p.connectionsChanged = func(){}
	p.AddChild(p.valueView)
	p.Resize(portSize, portSize)
	p.Pan(Pt(portSize, portSize).Div(-2))
	p.Self = self
	return p
}

func (p port) CanConnect(interface{}) bool { return true }
func (p *port) Connect(conn *Connection) {
	p.connections = append(p.connections, conn)
	if f := p.connectionsChanged; f != nil { f() }
}
func (p *port) Disconnect(conn *Connection) {
	for i, connection := range p.connections {
		if connection == conn {
			p.connections = append(p.connections[:i], p.connections[i+1:]...)
			if f := p.connectionsChanged; f != nil { f() }
			return
		}
	}
}

func (p *port) TookKeyboardFocus() { p.focused = true; p.Repaint(); p.valueView.Show() }
func (p *port) LostKeyboardFocus() { p.focused = false; p.Repaint(); p.valueView.Hide() }

func (p *port) KeyPressed(event KeyEvent) {
	if p.out && event.Key == glfw.KeyLeft || !p.out && event.Key == glfw.KeyRight {
		p.node.TakeKeyboardFocus()
		return
	}
	if p.out && event.Key == glfw.KeyRight && len(p.connections) > 0 || !p.out && event.Key == glfw.KeyLeft && len(p.connections) > 0 {
		p.connections[0].TakeKeyboardFocus()
		return
	}
	
	switch event.Key {
	case glfw.KeyEnter:
		conn := p.node.Block().NewConnection(p.Center())
		if p.out {
			conn.SetSource(p.Self.(*Output))
		} else {
			conn.SetDestination(p.Self.(*Input))
		}
		conn.StartEditing()
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		p.node.Block().Outermost().FocusNearestView(p, event.Key)
	case glfw.KeyEsc:
		p.node.TakeKeyboardFocus()
	default:
		p.ViewBase.KeyPressed(event)
	}
}

func (p *port) MousePressed(button int, pt Point) {
	p.TakeKeyboardFocus()
	conn := p.node.Block().NewConnection(p.MapTo(pt, p.node.Block()))
	if p.out {
		conn.SetSource(p.Self.(*Output))
		conn.dstHandle.SetMouseFocus(conn.dstHandle, button)
	} else {
		conn.SetDestination(p.Self.(*Input))
		conn.srcHandle.SetMouseFocus(conn.srcHandle, button)
	}
	conn.StartEditing()
}
func (p port) MouseDragged(button int, pt Point) {}
func (p port) MouseReleased(button int, pt Point) {}

func (p port) Paint() {
	SetColor(map[bool]Color{true:{.5, .5, 1, .5}, false:{1, 1, 1, .25}}[p.focused])
	for f := 1.0; f > .1; f /= 2 {
		SetPointSize(f * portSize)
		DrawPoint(ZP)
	}
}
