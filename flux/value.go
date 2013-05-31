package main

import (
	."code.google.com/p/gordon-go/gui"
	"code.google.com/p/go.exp/go/types"
)

type valueNode struct {
	*ViewBase
	AggregateMouseHandler
	obj types.Object
	set bool
	blk *block
	text Text
	in, out *port
	focused bool
}

func newValueNode(obj types.Object, set bool) *valueNode {
	n := &valueNode{obj: obj, set: set}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.text = NewText(obj.GetName())
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
	n.text.MoveCenter(Pt(0, n.text.Height() / 2))
	n.AddChild(n.text)
	n.reform()
	return n
}

func (n *valueNode) reform() {
	if n.in != nil {
		n.RemoveChild(n.in)
		n.in = nil
	}
	if n.out != nil {
		n.RemoveChild(n.out)
		n.out = nil
	}
	if n.set {
		n.in = newInput(n, &types.Var{Type: n.obj.GetType()})
		n.AddChild(n.in)
		n.in.MoveCenter(Pt(-8 - 2*portSize, 0))
	} else {
		n.out = newOutput(n, &types.Var{Type: n.obj.GetType()})
		n.AddChild(n.out)
		n.out.MoveCenter(Pt(8 + 2*portSize, 0))
	}
	ResizeToFit(n, 0)
	n.TakeKeyboardFocus()
}

func (n valueNode) block() *block { return n.blk }
func (n *valueNode) setBlock(b *block) { n.blk = b }
func (n valueNode) inputs() []*port {
	if n.set {
		return []*port{n.in}
	}
	return nil
}
func (n valueNode) outputs() []*port {
	if n.set {
		return nil
	}
	return []*port{n.out}
}
func (n valueNode) inConns() []*connection {
	if n.set {
		return n.in.conns
	}
	return nil
}
func (n valueNode) outConns() []*connection {
	if n.set {
		return nil
	}
	return n.out.conns
}

func (n *valueNode) Move(p Point) {
	n.ViewBase.Move(p)
	for _, c := range append(n.inConns(), n.outConns()...) {
		c.reform()
	}
}

func (n *valueNode) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *valueNode) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *valueNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyBackslash:
		if _, ok := n.obj.(*types.Const); !ok {
			n.set = !n.set
			n.reform()
		}
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		n.blk.outermost().focusNearestView(n, event.Key)
	case KeyEscape:
		n.blk.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n valueNode) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	if n.set {
		DrawLine(n.in.MapToParent(n.in.Center()), ZP)
	} else {
		DrawLine(ZP, n.out.MapToParent(n.out.Center()))
	}
}
