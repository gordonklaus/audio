package main

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type input struct { *port }
type output struct { *port }
type port struct {
	*ViewBase
	out bool
	node node
	info *ValueInfo
	valueView *typeView
	conns []*connection
	focused bool
	connsChanged func()
}

const portSize = 11

func newInput(n node, info *ValueInfo) *input {
	p := &input{}
	p.port = newPort(p, false, n, info)
	p.valueView.Move(Pt(-p.valueView.Width() - 12, -p.valueView.Height() / 2))
	return p
}
func newOutput(n node, info *ValueInfo) *output {
	p := &output{}
	p.port = newPort(p, true, n, info)
	p.valueView.Move(Pt(12, -p.valueView.Height() / 2))
	return p
}
func newPort(self View, out bool, n node, info *ValueInfo) *port {
	p := &port{out:out, node:n, info:info}
	p.ViewBase = NewView(p)
	p.valueView = newValueView(info)
	p.valueView.Hide()
	p.connsChanged = func(){}
	p.AddChild(p.valueView)
	p.Resize(portSize, portSize)
	p.Pan(Pt(portSize, portSize).Div(-2))
	p.Self = self
	return p
}

func (p port) canConnect(interface{}) bool { return true }
func (p *port) connect(c *connection) {
	p.conns = append(p.conns, c)
	if f := p.connsChanged; f != nil { f() }
}
func (p *port) disconnect(c *connection) {
	for i, c2 := range p.conns {
		if c2 == c {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			if f := p.connsChanged; f != nil { f() }
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
	if p.out && event.Key == glfw.KeyRight && len(p.conns) > 0 || !p.out && event.Key == glfw.KeyLeft && len(p.conns) > 0 {
		p.conns[0].TakeKeyboardFocus()
		return
	}
	
	switch event.Key {
	case glfw.KeyEnter:
		c := newConnection(p.node.block(), p.Center())
		if p.out {
			c.setSource(p.Self.(*output))
		} else {
			c.setDestination(p.Self.(*input))
		}
		c.startEditing()
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		p.node.block().outermost().focusNearestView(p, event.Key)
	case glfw.KeyEsc:
		p.node.TakeKeyboardFocus()
	default:
		p.ViewBase.KeyPressed(event)
	}
}

func (p *port) MousePressed(button int, pt Point) {
	p.TakeKeyboardFocus()
	c := newConnection(p.node.block(), p.MapTo(pt, p.node.block()))
	if p.out {
		c.setSource(p.Self.(*output))
		c.dstHandle.SetMouseFocus(c.dstHandle, button)
	} else {
		c.setDestination(p.Self.(*input))
		c.srcHandle.SetMouseFocus(c.srcHandle, button)
	}
	c.startEditing()
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
