package main

import (
	."code.google.com/p/gordon-go/gui"
)

type ifNode struct {
	*ViewBase
	AggregateMouseHandler
	blk *block
	input *input
	falseblk *block
	trueblk *block
	focused bool
}

func newIfNode(b *block) *ifNode {
	n := &ifNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.blk = b
	n.input = newInput(n, &Value{})
	n.falseblk = newBlock(n)
	n.trueblk = newBlock(n)
	n.AddChild(n.input)
	n.AddChild(n.falseblk)
	n.AddChild(n.trueblk)
	go n.falseblk.animate()
	go n.trueblk.animate()

	n.input.MoveCenter(Pt(-2*portSize, 0))
	n.trueblk.Move(Pt(portSize, 4))
	return n
}

func (n ifNode) block() *block { return n.blk }
func (n ifNode) inputs() []*input { return []*input{n.input} }
func (n ifNode) outputs() []*output { return nil }

func (n ifNode) inConns() []*connection {
	return append(append(n.input.conns, n.falseblk.inConns()...), n.trueblk.inConns()...)
}

func (n ifNode) outConns() []*connection {
	return append(n.falseblk.outConns(), n.trueblk.outConns()...)
}

func (n *ifNode) positionBlocks() {
	n.falseblk.Move(Pt(portSize, -4 - n.falseblk.Height()))
	ResizeToFit(n, 0)
}

func (n *ifNode) Move(p Point) {
	n.ViewBase.Move(p)
	for _, c := range n.input.conns { c.reform() }
}

func (n *ifNode) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *ifNode) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *ifNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		n.blk.outermost().focusNearestView(n, event.Key)
	case KeyEscape:
		n.blk.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n ifNode) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	DrawLine(ZP, n.input.MapToParent(n.input.Center()))
	top, bottom := 4 + n.trueblk.Size().Y / 2, -4 - n.falseblk.Size().Y / 2
	DrawLine(Pt(0, top), Pt(0, bottom))
	DrawLine(Pt(0, top), Pt(portSize, top))
	DrawLine(Pt(0, bottom), Pt(portSize, bottom))
}
