package main

import (
	."code.google.com/p/gordon-go/gui"
	"code.google.com/p/go.exp/go/types"
)

type port struct {
	*ViewBase
	out bool
	node node
	obj *types.Var
	valView *typeView
	conns []*connection
	focused bool
	connsChanged func()
}

const portSize = 11

func newInput(n node, v *types.Var) *port { return newPort(false, n, v) }
func newOutput(n node, v *types.Var) *port { return newPort(true, n, v) }
func newPort(out bool, n node, v *types.Var) *port {
	p := &port{out:out, node:n, obj:v}
	p.ViewBase = NewView(p)
	p.valView = newValueView(v)
	p.valView.Hide()
	p.connsChanged = func(){}
	p.AddChild(p.valView)
	p.Resize(portSize, portSize)
	p.Pan(Pt(portSize, portSize).Div(-2))
	p.setType(*p.valView.typ)
	return p
}

func (p *port) setType(t types.Type) {
	p.valView.setType(t)
	if p.out {
		p.valView.Move(Pt(12, -p.valView.Height() / 2))
	} else {
		p.valView.Move(Pt(-p.valView.Width() - 12, -p.valView.Height() / 2))
	}
}

func (p port) canConnect(q *port) bool {
	return p.out != q.out
}
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
	if p.obj.Type != seqType &&
	  (p.out && event.Key == KeyRight && len(p.conns) > 0 ||
	   !p.out && event.Key == KeyLeft && len(p.conns) > 0) {
		p.conns[len(p.conns) - 1].TakeKeyboardFocus()
		return
	}
	if p.obj.Type == seqType && event.Key == KeyDown && len(p.conns) > 0 {
		p.conns[len(p.conns) - 1].TakeKeyboardFocus()
		return
	}
	
	switch event.Key {
	case KeyEnter:
		c := newConnection()
		if p.out {
			c.setSrc(p)
		} else {
			c.setDst(p)
		}
		c.startEditing()
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		p.node.block().outermost().focusNearestView(p, event.Key)
	case KeyBackspace, KeyDelete:
		if p.out && event.Key == KeyBackspace || !p.out && event.Key == KeyDelete {
			if n, ok := p.node.(interface{removePort(*port)}); ok {
				n.removePort(p)
			} else {
				p.node.TakeKeyboardFocus()
			}
		} else if len(p.conns) > 0 {
			p.conns[len(p.conns) - 1].TakeKeyboardFocus()
		}
	case KeyEscape:
		p.node.TakeKeyboardFocus()
	default:
		p.ViewBase.KeyPressed(event)
	}
}

func (p *port) MousePressed(button int, pt Point) {
	p.TakeKeyboardFocus()
	c := newConnection()
	if p.out {
		c.setSrc(p)
		c.dstHandle.SetMouseFocus(c.dstHandle, button)
	} else {
		c.setDst(p)
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
