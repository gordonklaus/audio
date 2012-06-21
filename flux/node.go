package main

import (
	"github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/gui"
	."math"
	."fmt"
)

type Node interface {
	View
	MouseHandler
	Block() *Block
	Inputs() []*Input
	Outputs() []*Output
	GoCode(inputs string) string
}

type nodeText struct {
	Text
	node *NodeBase
}
func newNodeText(node *NodeBase) *nodeText {
	t := &nodeText{}
	t.Text = *NewTextBase(t, "")
	t.node = node
	return t
}
func (t *nodeText) LostKeyboardFocus() { t.SetEditable(false) }
func (t *nodeText) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyEnter:
		t.SetEditable(false)
		t.node.TakeKeyboardFocus()
	default:
		t.Text.KeyPressed(event)
	}
	t.node.reform()
}

type NodeBase struct {
	*ViewBase
	AggregateMouseHandler
	block *Block
	text *nodeText
	inputs []*Input
	outputs []*Output
	focused bool
}

const (
	nodeMargin = 3
)

func NewNodeBase(self Node, block *Block) *NodeBase {
	n := &NodeBase{}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(self), NewViewDragger(self)}
	n.block = block
	n.text = newNodeText(n)
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
	n.AddChild(n.text)
	n.Self = self
	return n
}

func (n *NodeBase) reform() {
	numInputs := float64(len(n.inputs))
	numOutputs := float64(len(n.outputs))
	maxputs := Max(numInputs, numOutputs)
	width := 2 * nodeMargin + Max(n.text.Width(), maxputs * putSize)
	height := n.text.Height() + 2*putSize
	n.Resize(width, height)
	n.text.Move(Pt((width - n.text.Width()) / 2, putSize))
	for i, input := range n.inputs {
		input.Move(Pt((float64(i) + .5) * width / numInputs - putSize / 2, 0))
	}
	for i, output := range n.outputs {
		output.Move(Pt((float64(i) + .5) * width / numOutputs - putSize / 2, n.Height() - putSize))
	}
}

func (n NodeBase) Block() *Block { return n.block }
func (n NodeBase) Inputs() []*Input { return n.inputs }
func (n NodeBase) Outputs() []*Output { return n.outputs }

func (n NodeBase) Moved(Point) {
	f := func(p *put) { for _, conn := range p.connections { conn.reform() } }
	for _, p := range n.inputs { f(p.put) }
	for _, p := range n.outputs { f(p.put) }
}

func (n *NodeBase) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *NodeBase) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *NodeBase) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		n.block.Outermost().FocusNearestView(n, event.Key)
	case glfw.KeyEsc:
		n.block.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n NodeBase) Paint() {
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

func NewNode(info Info, block *Block) Node {
	switch info := info.(type) {
	// case StringTypeInfo:
	// 	return NewBasicLiteralNode(info)
	case FunctionInfo:
		return NewFunctionNode(info, block)
	}
	return nil
}

type FunctionNode struct { *NodeBase }
func NewFunctionNode(info FunctionInfo, block *Block) *FunctionNode {
	n := &FunctionNode{}
	n.NodeBase = NewNodeBase(n, block)
	n.text.SetText(info.name)
	for _, parameter := range info.parameters {
		p := NewInput(n, parameter)
		n.AddChild(p)
		n.inputs = append(n.inputs, p)
	}
	for _, result := range info.results {
		p := NewOutput(n, result)
		n.AddChild(p)
		n.outputs = append(n.outputs, p)
	}
	n.reform()
	
	return n
}
func (n FunctionNode) GoCode(inputs string) string {
	return Sprintf("%v(%v)", n.text.GetText(), inputs)
}

type ConstantNode struct { *NodeBase }
func NewStringConstantNode(block *Block) *ConstantNode {
	n := &ConstantNode{}
	n.NodeBase = NewNodeBase(n, block)
	n.text.SetEditable(true)
	p := NewOutput(n, ValueInfo{})
	n.AddChild(p)
	n.outputs = []*Output{p}
	n.reform()
	return n
}

func (n ConstantNode) GoCode(string) string {
	return Sprintf(`"%v"`, n.text.GetText())
}
