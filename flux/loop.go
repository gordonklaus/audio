package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type LoopNode struct {
	*ViewBase
	AggregateMouseHandler
	block *Block
	input *Input
	inputNode *InOutNode
	loopBlock *Block
	focused bool
	leftmost Point
}

func NewLoopNode(block *Block) *LoopNode {
	n := &LoopNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.block = block
	n.input = newInput(n, &ValueInfo{})
	n.input.connectionsChanged = func() {
		if conns := n.input.connections; len(conns) > 0 {
			if o := conns[0].src; o != nil { n.updateInputType(o.info.typ) }
		} else {
			n.updateInputType(nil)
		}
	}
	n.AddChild(n.input)
	n.loopBlock = NewBlock(n)
	n.inputNode = NewInputNode(n.loopBlock)
	n.inputNode.newOutput(&ValueInfo{})
	n.loopBlock.AddNode(n.inputNode)
	n.AddChild(n.loopBlock)
	go n.loopBlock.animate()
	
	n.input.MoveCenter(Pt(-2*portSize, 0))
	n.updateInputType(nil)
	return n
}

func (n LoopNode) Block() *Block { return n.block }
func (n LoopNode) Inputs() []*Input { return []*Input{n.input} }
func (n LoopNode) Outputs() []*Output { return nil }
func (n LoopNode) InputConnections() []*Connection {
	return append(n.input.connections, n.loopBlock.InputConnections()...)
}
func (n LoopNode) OutputConnections() []*Connection {
	return n.loopBlock.OutputConnections()
}

func (n *LoopNode) updateInputType(t Type) {
	in := n.inputNode
	switch t.(type) {
	case nil, *NamedType, *ChanType:
		if len(in.outputs) == 2 {
			in.RemoveChild(in.outputs[1])
			in.outputs = in.outputs[:1]
		}
	case *ArrayType, *SliceType, *MapType:
		if len(in.outputs) == 1 {
			in.newOutput(&ValueInfo{})
		}
	}
	switch t.(type) {
	case nil:
	case *NamedType:
	case *ArrayType:
	case *SliceType:
	case *MapType:
	case *ChanType:
	}
}

func (n *LoopNode) positionBlocks() {
	b := n.loopBlock
	n.leftmost = b.points[0]
	for _, p := range b.points { if p.X < n.leftmost.X { n.leftmost = p } }
	n.inputNode.MoveOrigin(n.leftmost)
	n.leftmost = b.MapToParent(n.leftmost)
	n.input.MoveCenter(n.leftmost.Sub(Pt(portSize, 0)))
	ResizeToFit(n, 0)
}

func (n *LoopNode) Move(p Point) {
	n.ViewBase.Move(p)
	for _, conn := range n.input.connections { conn.reform() }
}

func (n *LoopNode) Resize(width, height float64) {
	n.ViewBase.Resize(width, height)
}

func (n *LoopNode) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *LoopNode) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *LoopNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		n.block.Outermost().FocusNearestView(n, event.Key)
	case KeyEsc:
		n.block.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n LoopNode) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	DrawLine(n.leftmost, n.input.MapToParent(n.input.Center()))
}
