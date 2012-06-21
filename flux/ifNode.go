package main

import (
	."github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
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
	n.falseBlock = NewBlock(n)
	n.trueBlock = NewBlock(n)
	n.AddChild(n.input)
	n.AddChild(n.falseBlock)
	n.AddChild(n.trueBlock)
	n.Resize(n.falseBlock.Width() + n.trueBlock.Width(), n.input.Height() + n.falseBlock.Height())
	n.input.Move(Pt((n.Width() - n.input.Width()) / 2, 0))
	n.falseBlock.Move(Pt(0, n.input.Height()))
	n.trueBlock.Move(Pt(n.falseBlock.Width(), n.input.Height()))
	return n
}

func (n IfNode) Block() *Block { return n.block }
func (n IfNode) Inputs() []*Input { return []*Input{n.input} }
func (n IfNode) Outputs() []*Output { return nil }

func (n IfNode) GoCode(string) string { return "" }

func (n IfNode) Moved(Point) {
	for _, conn := range n.input.connections { conn.reform() }
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
	width, height := gl.Double(n.Width()), gl.Double(n.Height())
	if n.focused {
		gl.Color4d(.4, .4, 1, .4)
	} else {
		gl.Color4d(0, 0, 0, .5)
	}
	gl.Rectd(0, 0, width, height)
	gl.Color4d(1, 1, 1, 1)
	gl.Begin(gl.LINE_LOOP)
	gl.Vertex2d(0, 0)
	gl.Vertex2d(width, 0)
	gl.Vertex2d(width, height)
	gl.Vertex2d(0, height)
	gl.End()
}
