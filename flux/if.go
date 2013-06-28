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

	n.input.MoveCenter(Pt(-2*portSize, 0))
	n.update()
	return n
}

func (n ifNode) block() *block { return n.blk }
func (n *ifNode) setBlock(b *block) { n.blk = b }
func (n ifNode) inputs() []*port { return []*port{n.input} }
func (n ifNode) outputs() []*port { return nil }

func (n ifNode) inConns() []*connection {
	return append(append(n.input.conns, n.falseblk.inConns()...), n.trueblk.inConns()...)
}

func (n ifNode) outConns() []*connection {
	return append(n.falseblk.outConns(), n.trueblk.outConns()...)
}

func (n *ifNode) update() bool {
	if !n.falseblk.update() && !n.trueblk.update() {
		return false
	}
	n.falseblk.Move(Pt(-blockRadius, -4 - n.falseblk.Height()))
	n.trueblk.Move(Pt(-blockRadius, 4))
	ResizeToFit(n, 0)
	return true
}

func (n *ifNode) Move(p Point) {
	n.ViewBase.Move(p)
	for _, c := range n.input.conns { c.reform() }
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
	DrawLine(ZP, n.input.MapToParent(n.input.Center()))
	DrawLine(Pt(0, -4), Pt(0, 4))
}
