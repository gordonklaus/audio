package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type IfNode struct {
	*ViewBase
	AggregateMouseHandler
	block *Block
	input *Input
	falseBlock *Block
	trueBlock *Block
	focused bool
	blockEnds [2]Point
}

func NewIfNode(block *Block) *IfNode {
	n := &IfNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.block = block
	n.input = NewInput(n, &ValueInfo{})
	n.falseBlock = NewBlock(n)
	n.trueBlock = NewBlock(n)
	n.AddChild(n.input)
	n.AddChild(n.falseBlock)
	n.AddChild(n.trueBlock)
	go n.falseBlock.animate()
	go n.trueBlock.animate()

	n.input.MoveCenter(Pt(-2*putSize, 0))
	n.trueBlock.Move(Pt(putSize, 4))
	return n
}

func (n IfNode) Block() *Block { return n.block }
func (n IfNode) Inputs() []*Input { return []*Input{n.input} }
func (n IfNode) Outputs() []*Output { return nil }

func (n IfNode) InputConnections() []*Connection {
	return append(append(n.input.connections, n.falseBlock.InputConnections()...), n.trueBlock.InputConnections()...)
}

func (n IfNode) OutputConnections() []*Connection {
	return append(n.falseBlock.OutputConnections(), n.trueBlock.OutputConnections()...)
}

func (n *IfNode) positionBlocks() {
	n.falseBlock.Move(Pt(putSize, -4 - n.falseBlock.Height()))
	for i, b := range []*Block{n.falseBlock, n.trueBlock} {
		leftmost := b.points[0]
		for _, p := range b.points { if p.X < leftmost.X { leftmost = p } }
		n.blockEnds[i] = b.MapToParent(leftmost)
	}
	ResizeToFit(n, 0)
}

func (n *IfNode) Move(p Point) {
	n.ViewBase.Move(p)
	for _, conn := range n.input.connections { conn.reform() }
}

func (n *IfNode) Resize(width, height float64) {
	n.ViewBase.Resize(width, height)
}

func (n *IfNode) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *IfNode) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *IfNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		n.block.Outermost().FocusNearestView(n, event.Key)
	case KeyEsc:
		n.block.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n IfNode) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	DrawLine(ZP, n.input.MapToParent(n.input.Center()))
	top, bottom := n.blockEnds[0], n.blockEnds[1]; top.X, bottom.X = 0, 0
	DrawLine(top, bottom)
	DrawLine(top, n.blockEnds[0])
	DrawLine(bottom, n.blockEnds[1])
}
