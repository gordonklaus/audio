package main

import (
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/gui"
	."math"
)

type Node struct {
	ViewBase
	AggregateMouseHandler
	Function *Function
	inputs []*Input
	outputs []*Output
	name *Text
	focused bool
}

const (
	nodeMargin = 3
)

func NewFunctionNode(info FunctionInfo) *Node {
	n := &Node{}
	n.ViewBase = *NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}

	n.name = NewText(info.name)
	n.name.SetBackgroundColor(Color{0, 0, 0, 0})
	n.AddChild(n.name)
	
	numInputs := float64(len(info.parameters))
	numOutputs := float64(len(info.results))
	maxputs := Max(numInputs, numOutputs)
	width := 2 * nodeMargin + Max(n.name.Width(), maxputs * putSize)
	height := n.name.Height() + 2*putSize
	n.Resize(width, height)
	n.name.Move(Pt((width - n.name.Width()) / 2, putSize))
	for i := range info.parameters {
		p := NewInput(n)
		p.Move(Pt((float64(i) + .5) * width / numInputs - putSize / 2, 0))
		n.AddChild(p)
		n.inputs = append(n.inputs, p)
	}
	for i := range info.results {
		p := NewOutput(n)
		p.Move(Pt((float64(i) + .5) * width / numOutputs - putSize / 2, n.Height() - putSize))
		n.AddChild(p)
		n.outputs = append(n.outputs, p)
	}
	
	return n
}

func (n Node) Getputs() []View {
	puts := make([]View, 0, len(n.inputs) + len(n.outputs))
	for _, p := range n.inputs { puts = append(puts, p) }
	for _, p := range n.outputs { puts = append(puts, p) }
	return puts
}

// func (n Node) Moved(Point) {
// 	f := func(p *put) { for _, conn := range p.connections { conn.reform() } }
// 	for _, p := range n.inputs { f(p.put) }
// 	for _, p := range n.outputs { f(p.put) }
// }

func (n *Node) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *Node) LostKeyboardFocus() { n.focused = false; n.Repaint() }

// func (n *Node) KeyPressed(key int) {
// 	switch key {
// 	case Key_Left, Key_Right, Key_Up, Key_Down:
// 		n.function.FocusNearestView(n, key)
// 	case Key_Escape:
// 		n.function.TakeKeyboardFocus()
// 	default:
// 		n.ViewBase.KeyPressed(key)
// 	}
// }

func (n Node) Paint() {
	width, height := gl.Double(n.Width()), gl.Double(n.Height())
	gl.Color4d(0, 0, 0, .5)
	gl.Rectd(0, 0, width, height)
	gl.Color4d(1, 1, 1, 1)
	gl.Begin(gl.LINE_LOOP)
	gl.Vertex2d(0, 0)
	gl.Vertex2d(width, 0)
	gl.Vertex2d(width, height)
	gl.Vertex2d(0, height)
	gl.End()
}
