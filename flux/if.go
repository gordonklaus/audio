package main

import (
	."code.google.com/p/gordon-go/gui"
	"code.google.com/p/go.exp/go/types"
)

type ifNode struct {
	*ViewBase
	AggregateMouseHandler
	blk *block
	input *port
	seqIn, seqOut *port
	falseblk *block
	trueblk *block
	focused bool
}

func newIfNode() *ifNode {
	n := &ifNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.input = newInput(n, &types.Var{Type: types.Typ[types.Bool]})
	n.falseblk = newBlock(n)
	n.trueblk = newBlock(n)
	n.AddChild(n.input)
	n.AddChild(n.falseblk)
	n.AddChild(n.trueblk)
	
	n.seqIn = newInput(n, &types.Var{Name: "seq", Type: seqType})
	n.seqIn.MoveCenter(Pt(-portSize, 0))
	n.AddChild(n.seqIn)
	n.seqOut = newOutput(n, &types.Var{Name: "seq", Type: seqType})
	n.seqOut.MoveCenter(Pt(portSize, 0))
	n.AddChild(n.seqOut)

	n.input.MoveCenter(Pt(-2*portSize, 0))
	n.update()
	return n
}

func (n ifNode) block() *block { return n.blk }
func (n *ifNode) setBlock(b *block) { n.blk = b }
func (n ifNode) inputs() []*port { return []*port{n.seqIn, n.input} }
func (n ifNode) outputs() []*port { return []*port{n.seqOut} }

func (n ifNode) inConns() []*connection {
	return append(append(append(n.input.conns, n.seqIn.conns...), n.falseblk.inConns()...), n.trueblk.inConns()...)
}

func (n ifNode) outConns() []*connection {
	return append(append(n.seqOut.conns, n.falseblk.outConns()...), n.trueblk.outConns()...)
}

func (n *ifNode) update() bool {
	f, t := !n.falseblk.update(), !n.trueblk.update()
	if f && t {
		return false
	}
	n.falseblk.Move(Pt(-blockRadius, -4 - n.falseblk.Height()))
	n.trueblk.Move(Pt(-blockRadius, 4))
	ResizeToFit(n, 0)
	return true
}

func (n *ifNode) Move(p Point) {
	n.ViewBase.Move(p)
	for _, c := range append(n.inConns(), n.outConns()...) {
		c.reform()
	}
}

func (n *ifNode) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *ifNode) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *ifNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight:
		n.blk.outermost().focusNearestView(n, event.Key)
	case KeyUp:
		n.trueblk.TakeKeyboardFocus()
	case KeyDown:
		n.falseblk.TakeKeyboardFocus()
	case KeyEscape:
		n.blk.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n ifNode) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	DrawLine(n.input.MapToParent(n.input.Center()), n.seqOut.MapToParent(n.seqOut.Center()))
	DrawLine(Pt(0, -4), Pt(0, 4))
}
