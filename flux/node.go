package main

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	."math"
)

type Node interface {
	View
	MouseHandler
	Block() *Block
	Inputs() []*Input
	Outputs() []*Output
	InputConnections() []*Connection
	OutputConnections() []*Connection
}

type nodeText struct {
	*TextBase
	node *NodeBase
}
func newNodeText(node *NodeBase) *nodeText {
	t := &nodeText{node:node}
	t.TextBase = NewTextBase(t, "")
	return t
}
func (t *nodeText) LostKeyboardFocus() { t.SetEditable(false) }
func (t *nodeText) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyEnter:
		t.SetEditable(false)
		t.node.TakeKeyboardFocus()
	default:
		t.TextBase.KeyPressed(event)
	}
	t.node.reform()
}

type NodeBase struct {
	*ViewBase
	Self Node
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
	n := &NodeBase{Self:self}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(self), NewViewDragger(self)}
	n.block = block
	n.text = newNodeText(n)
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
	n.AddChild(n.text)
	n.ViewBase.Self = self
	return n
}

func (n *NodeBase) newInput(i *ValueInfo) *Input {
	p := newInput(n.Self, i)
	n.AddChild(p)
	n.inputs = append(n.inputs, p)
	n.reform()
	return p
}

func (n *NodeBase) newOutput(i *ValueInfo) *Output {
	p := newOutput(n.Self, i)
	n.AddChild(p)
	n.outputs = append(n.outputs, p)
	n.reform()
	return p
}

func (n *NodeBase) RemoveChild(v View) {
	n.ViewBase.RemoveChild(v)
	switch v := v.(type) {
	case *Input:
		SliceRemove(&n.inputs, v)
	case *Output:
		SliceRemove(&n.outputs, v)
	}
	n.reform()
}

func (n *NodeBase) reform() {
	numInputs := float64(len(n.inputs))
	numOutputs := float64(len(n.outputs))
	maxputs := Max(numInputs, numOutputs)
	rx, ry := 2.0 * portSize, (maxputs + 1) * portSize / 2
	
	rect := ZR
	for i, input := range n.inputs {
		y := -portSize * (float64(i) - (numInputs - 1) / 2)
		input.MoveCenter(Pt(-rx * Sqrt(ry * ry - y * y) / ry, y))
		rect = rect.Union(input.MapRectToParent(input.Rect()))
	}
	for i, output := range n.outputs {
		y := -portSize * (float64(i) - (numOutputs - 1) / 2)
		output.MoveCenter(Pt(rx * Sqrt(ry * ry - y * y) / ry, y))
		rect = rect.Union(output.MapRectToParent(output.Rect()))
	}

	n.text.MoveCenter(Pt(0, rect.Max.Y + n.text.Height() / 2))
	n.Pan(rect.Min)
	n.Resize(rect.Dx(), rect.Dy())
}

func (n NodeBase) Block() *Block { return n.block }
func (n NodeBase) Inputs() []*Input { return n.inputs }
func (n NodeBase) Outputs() []*Output { return n.outputs }

func (n NodeBase) InputConnections() (connections []*Connection) {
	for _, input := range n.Inputs() {
		for _, conn := range input.connections {
			connections = append(connections, conn)
		}
	}
	return
}

func (n NodeBase) OutputConnections() (connections []*Connection) {
	for _, output := range n.Outputs() {
		for _, conn := range output.connections {
			connections = append(connections, conn)
		}
	}
	return
}

func (n *NodeBase) Move(p Point) {
	n.ViewBase.Move(p)
	f := func(p *port) { for _, conn := range p.connections { conn.reform() } }
	for _, p := range n.inputs { f(p.port) }
	for _, p := range n.outputs { f(p.port) }
}

func (n NodeBase) Center() Point { return ZP }

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
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	for _, p := range n.inputs { DrawLine(ZP, p.MapToParent(p.Center())) }
	for _, p := range n.outputs { DrawLine(ZP, p.MapToParent(p.Center())) }
}

func NewNode(info Info, block *Block) Node {
	switch info := info.(type) {
	case SpecialInfo:
		switch info.Name() {
		case "if":
			return NewIfNode(block)
		case "loop":
			return NewLoopNode(block)
		}
	// case StringType:
	// 	return NewBasicLiteralNode(info)
	case *FuncInfo:
		return NewCallNode(info, block)
	}
	return nil
}

type CallNode struct {
	*NodeBase
	info *FuncInfo
}
func NewCallNode(info *FuncInfo, block *Block) *CallNode {
	n := &CallNode{info:info}
	n.NodeBase = NewNodeBase(n, block)
	n.text.SetText(info.name)
	for _, parameter := range info.typ.parameters { n.newInput(parameter) }
	for _, result := range info.typ.results { n.newOutput(result) }
	
	return n
}
func (n CallNode) Package() *PackageInfo { return n.info.Parent().(*PackageInfo) }

type ConstantNode struct { *NodeBase }
func NewStringConstantNode(block *Block) *ConstantNode {
	n := &ConstantNode{}
	n.NodeBase = NewNodeBase(n, block)
	n.newOutput(&ValueInfo{})
	return n
}
