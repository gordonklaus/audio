package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type port struct {
	*ViewBase
	out          bool
	node         node
	obj          *types.Var
	valView      *typeView
	conns        []*connection
	focused      bool
	connsChanged func()
}

const portSize = 11.0

func newInput(n node, v *types.Var) *port  { return newPort(false, n, v) }
func newOutput(n node, v *types.Var) *port { return newPort(true, n, v) }
func newPort(out bool, n node, v *types.Var) *port {
	p := &port{out: out, node: n, obj: v}
	p.ViewBase = NewView(p)
	p.valView = newValueView(v)
	Hide(p.valView)
	p.connsChanged = func() {}
	p.Add(p.valView)
	p.SetRect(ZR.Inset(-portSize / 2))
	p.setType(*p.valView.typ)
	return p
}

func (p *port) setType(t types.Type) {
	p.valView.setType(t)
	if p.out {
		p.valView.Move(Pt(12, -Height(p.valView)/2))
	} else {
		p.valView.Move(Pt(-Width(p.valView)-12, -Height(p.valView)/2))
	}
}

func canConnect(src, dst *port, feedback bool) bool {
	if src.out == dst.out {
		return false
	}
	for _, c := range src.conns {
		if c.dst == dst {
			return false
		}
	}
	// TODO: types.IsAssignableTo(src.obj.Type, dst.obj.Type) (handles seqType, too?)
	if (src.obj.Type == seqType) != (dst.obj.Type == seqType) {
		return false
	}
	if src.obj.Type == seqType && (src.node.block() != dst.node.block() || precedes(src.node, dst.node)) {
		return false
	}

	// TODO: recursive func literals
	// if f, ok := src.node.(*funcNode); ok && parentNodeInBlock(dst.node, f.funcblk) != nil {
	// 	return true
	// }

	n1, n2 := src.node, dst.node
	for b := n1.block(); ; n1, b = b.node, b.outer() {
		if n := parentNodeInBlock(n2, b); n != nil {
			n2 = n
			break
		}
	}
	if feedback {
		for b := n1.block(); ; b = b.outer() {
			if b == nil {
				return false
			}
			if _, ok := b.node.(*loopNode); ok {
				break
			}
		}
		n1, n2 = n2, n1
	} else if n1 == n2 {
		return false
	}
	return !precedes(n2, n1)
}

func precedes(n1, n2 node) bool {
	for _, dst := range dstsInBlock(n1) {
		if dst == n2 || precedes(dst, n2) {
			return true
		}
	}
	return false
}

func (p *port) connect(c *connection) {
	p.conns = append(p.conns, c)
}

func (p *port) disconnect(c *connection) {
	for i, c2 := range p.conns {
		if c2 == c {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			return
		}
	}
}

func (p *port) TookKeyFocus() { p.focused = true; Repaint(p); Show(p.valView) }
func (p *port) LostKeyFocus() { p.focused = false; Repaint(p); Hide(p.valView) }

func (p *port) KeyPress(event KeyEvent) {
	if p.obj.Type != seqType &&
		(p.out && event.Key == KeyRight && len(p.conns) > 0 ||
			!p.out && event.Key == KeyLeft && len(p.conns) > 0) {
		SetKeyFocus(p.conns[len(p.conns)-1])
		return
	}
	if p.obj.Type == seqType && event.Key == KeyDown && len(p.conns) > 0 {
		SetKeyFocus(p.conns[len(p.conns)-1])
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
	case KeyBackspace, KeyDelete:
		if p.out && event.Key == KeyBackspace || !p.out && event.Key == KeyDelete {
			if n, ok := p.node.(interface {
				removePort(*port)
			}); ok {
				n.removePort(p)
			} else {
				SetKeyFocus(p.node)
			}
		} else if len(p.conns) > 0 {
			SetKeyFocus(p.conns[len(p.conns)-1])
		}
	case KeyEscape:
		SetKeyFocus(p.node)
	default:
		p.ViewBase.KeyPress(event)
	}
}

func (p *port) Mouse(m MouseEvent) {
	SetKeyFocus(p)
	c := newConnection()
	if p.out {
		c.setSrc(p)
		SetMouser(c.dstHandle, m.Button)
	} else {
		c.setDst(p)
		SetMouser(c.srcHandle, m.Button)
	}
	c.startEditing()
}

func (p port) Paint() {
	if p.focused {
		SetColor(Color{.3, .3, .7, .5})
		SetPointSize(portSize)
		DrawPoint(ZP)
	}
}
