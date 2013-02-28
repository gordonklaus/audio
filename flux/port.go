package main

import (
	."code.google.com/p/gordon-go/gui"
)

type input struct { *port }
type output struct { *port }
type port struct {
	*ViewBase
	out bool
	node node
	val *Value
	valView *typeView
	conns []*connection
	focused bool
	connsChanged func()
}

const portSize = 11

func newInput(n node, v *Value) *input {
	p := &input{}
	p.port = newPort(p, false, n, v)
	p.valView.Move(Pt(-p.valView.Width() - 12, -p.valView.Height() / 2))
	return p
}
func newOutput(n node, v *Value) *output {
	p := &output{}
	p.port = newPort(p, true, n, v)
	p.valView.Move(Pt(12, -p.valView.Height() / 2))
	return p
}
func newPort(self View, out bool, n node, v *Value) *port {
	p := &port{out:out, node:n, val:v}
	p.ViewBase = NewView(p)
	p.valView = newValueView(v)
	p.valView.Hide()
	p.connsChanged = func(){}
	p.AddChild(p.valView)
	p.Resize(portSize, portSize)
	p.Pan(Pt(portSize, portSize).Div(-2))
	p.Self = self
	return p
}

func (p port) canConnect(interface{}) bool { return true }
func (p *port) connect(c *connection) {
	p.conns = append(p.conns, c)
	p.connsChanged()
}
func (p *port) disconnect(c *connection) {
	for i, c2 := range p.conns {
		if c2 == c {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			p.connsChanged()
			return
		}
	}
}

func (p *port) TookKeyboardFocus() { p.focused = true; p.Repaint(); p.valView.Show() }
func (p *port) LostKeyboardFocus() { p.focused = false; p.Repaint(); p.valView.Hide() }

func (p *port) KeyPressed(event KeyEvent) {
	if p.out && event.Key == KeyLeft || !p.out && event.Key == KeyRight {
		p.node.TakeKeyboardFocus()
		return
	}
	if p.out && event.Key == KeyRight && len(p.conns) > 0 || !p.out && event.Key == KeyLeft && len(p.conns) > 0 {
		p.conns[0].TakeKeyboardFocus()
		return
	}
	
	switch event.Key {
	case KeyEnter:
		c := newConnection(p.node.block(), p.Center())
		if p.out {
			c.setSrc(p.Self.(*output))
		} else {
			c.setDst(p.Self.(*input))
		}
		c.startEditing()
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		p.node.block().outermost().focusNearestView(p, event.Key)
	case KeyEscape:
		p.node.TakeKeyboardFocus()
	default:
		p.ViewBase.KeyPressed(event)
	}
}

func (p *port) MousePressed(button int, pt Point) {
	p.TakeKeyboardFocus()
	c := newConnection(p.node.block(), p.MapTo(pt, p.node.block()))
	if p.out {
		c.setSrc(p.Self.(*output))
		c.dstHandle.SetMouseFocus(c.dstHandle, button)
	} else {
		c.setDst(p.Self.(*input))
		c.srcHandle.SetMouseFocus(c.srcHandle, button)
	}
	c.startEditing()
}
func (p port) MouseDragged(button int, pt Point) {}
func (p port) MouseReleased(button int, pt Point) {}

func (p port) Paint() {
	if p.focused {
		SetColor(Color{.3, .3, .7, .5})
		SetPointSize(portSize)
		DrawPoint(ZP)
	}
}
