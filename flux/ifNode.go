package main

import (
	."fmt"
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
}

func NewIfNode(block *Block) *IfNode {
	n := &IfNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.block = block
	n.input = NewInput(n, ValueInfo{})
	n.falseBlock = NewBlock(nil)
	n.trueBlock = NewBlock(nil)
	n.falseBlock.node = n
	n.trueBlock.node = n
	n.AddChild(n.input)
	n.AddChild(n.falseBlock)
	n.AddChild(n.trueBlock)

	n.input.MoveCenter(Pt(-2*putSize, 0))
	n.trueBlock.Move(ZP)
	n.positionBlocks()
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

func (n IfNode) Code(indent int, vars map[*Input]string, _ string) (s string) {
	name := "false"
	if len(n.input.connections) > 0 {
		name = vars[n.input]
	}
	s += Sprintf("%vif %v {\n", tabs(indent), name)
	s += n.trueBlock.Code(indent + 1, vars)
	if s2 := n.falseBlock.Code(indent + 1, vars); len(s2) > 0 {
		s += Sprintf("%v} else {\n%v", tabs(indent), s2)
	}
	s += tabs(indent) + "}\n"
	return
}

func (n *IfNode) positionBlocks() {
	n.falseBlock.Move(Pt(0, -n.falseBlock.Height()))
	rect := n.input.MapRectToParent(n.input.Rect()).Union(n.falseBlock.MapRectToParent(n.falseBlock.Rect())).Union(n.trueBlock.MapRectToParent(n.trueBlock.Rect()))
	n.Pan(rect.Min)
	n.Resize(rect.Dx(), rect.Dy())
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
}
