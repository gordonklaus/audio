package main

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	."math"
	."fmt"
	."strconv"
	."strings"
)

type Node interface {
	View
	MouseHandler
	Block() *Block
	Inputs() []*Input
	Outputs() []*Output
	InputConnections() []*Connection
	OutputConnections() []*Connection
	Save(indent int, nodeIDs map[Node]int) string
	Code(indent int, vars map[*Input]string, inputs string) string
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
	rx, ry := 2.0 * putSize, (maxputs + 1) * putSize / 2
	
	rect := ZR
	for i, input := range n.inputs {
		y := putSize * (float64(i) - (numInputs - 1) / 2)
		input.MoveCenter(Pt(-rx * Sqrt(ry * ry - y * y) / ry, y))
		rect = rect.Union(input.MapRectToParent(input.Rect()))
	}
	for i, output := range n.outputs {
		y := putSize * (float64(i) - (numOutputs - 1) / 2)
		output.MoveCenter(Pt(rx * Sqrt(ry * ry - y * y) / ry, y))
		rect = rect.Union(output.MapRectToParent(output.Rect()))
	}

	n.text.MoveCenter(Pt(0, rect.Max.Y + n.text.Height() / 2))
	rect = rect.Union(n.text.MapRectToParent(n.text.Rect()))
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
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	for _, p := range n.inputs { DrawLine(ZP, p.MapToParent(p.Center())) }
	for _, p := range n.outputs { DrawLine(ZP, p.MapToParent(p.Center())) }
}

func NewNode(info Info, block *Block) Node {
	switch info := info.(type) {
	// case StringTypeInfo:
	// 	return NewBasicLiteralNode(info)
	case *FunctionInfo:
		return NewFunctionNode(info, block)
	}
	return nil
}

func LoadNode(b *Block, s string, indent int, nodes map[int]Node, pkgNames map[string]*PackageInfo) (node Node, rest string) {
	line, rest := Split2(s, "\n")
	fields := Fields(line)
	if fields[1][0] == '"' {
		strNode := NewStringConstantNode(b)
		text := fields[1]
		strNode.text.SetText(text[1:len(text) - 1])
		node = strNode
	} else if fields[1] == "\\in" {
		for n := range b.nodes {
			if _, ok := n.(*InputNode); ok {
				node = n
			}
		}
	} else if fields[1] == "if" {
		node, rest = LoadIfNode(s, indent, b, nodes, pkgNames)
	} else {
		pkgName, name := Split2(fields[1], ".")
		for _, info := range pkgNames[pkgName].Children() {
			if info.Name() != name { continue }
			switch info := info.(type) {
			case *FunctionInfo:
				node = NewFunctionNode(info, b)
			default:
				panic("not yet implemented")
			}
		}
	}
	nodeID, _ := Atoi(fields[0])
	nodes[nodeID] = node
	return node, rest
}

type FunctionNode struct {
	*NodeBase
	info *FunctionInfo
}
func NewFunctionNode(info *FunctionInfo, block *Block) *FunctionNode {
	n := &FunctionNode{info:info}
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
func (n FunctionNode) Package() *PackageInfo { return n.info.Parent().(*PackageInfo) }
func (n FunctionNode) Save(int, map[Node]int) string {
	pkgName := n.Package().Name() // TODO:  handle name collisions
	return Sprintf("%v.%v", pkgName, n.info.Name())
}
func (n FunctionNode) Code(_ int, _ map[*Input]string, inputs string) string {
	pkgName := n.Package().Name() // TODO:  handle name collisions
	return Sprintf("%v.%v(%v)", pkgName, n.info.Name(), inputs)
}

type ConstantNode struct { *NodeBase }
func NewStringConstantNode(block *Block) *ConstantNode {
	n := &ConstantNode{}
	n.NodeBase = NewNodeBase(n, block)
	p := NewOutput(n, ValueInfo{})
	n.AddChild(p)
	n.outputs = []*Output{p}
	n.reform()
	return n
}

func (n ConstantNode) Save(int, map[Node]int) string {
	return Sprintf(`"%v"`, n.text.GetText())
}
func (n ConstantNode) Code(int, map[*Input]string, string) string {
	return Sprintf(`"%v"`, n.text.GetText())
}
