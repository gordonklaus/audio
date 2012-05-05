package main

import (
	"github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/gui"
	."math"
)

type Node struct {
	ViewBase
	AggregateMouseHandler
	function *Function
	inputs []*Input
	outputs []*Output
	name *Text
	focused bool
}

const (
	nodeMargin = 3
)

func newNode(name string) *Node {
	n := &Node{}
	n.ViewBase = *NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}

	n.name = NewText(name)
	n.name.SetBackgroundColor(Color{0, 0, 0, 0})
	n.AddChild(n.name)
	return n
}

func NewFunctionNode(info FunctionInfo) *Node {
	n := newNode(info.name)
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

func NewStringLiteralNode(text string) *Node {
	n := newNode(text)
	width := 2 * nodeMargin + n.name.Width()
	height := n.name.Height() + putSize
	n.Resize(width, height)
	n.name.Move(Pt((width - n.name.Width()) / 2, 0))
	n.name.SetTextColor(Color{1, 1, 0, 1})
	p := NewOutput(n)
	p.Move(Pt((width - putSize) / 2, n.Height() - putSize))
	n.AddChild(p)
	n.outputs = []*Output{p}
	return n
}

func (n Node) Getputs() []View {
	puts := make([]View, 0, len(n.inputs) + len(n.outputs))
	for _, p := range n.inputs { puts = append(puts, p) }
	for _, p := range n.outputs { puts = append(puts, p) }
	return puts
}

func (n Node) Moved(Point) {
	f := func(p put) { for _, conn := range p.connections { conn.reform() } }
	for _, p := range n.inputs { f(p.put) }
	for _, p := range n.outputs { f(p.put) }
}

func (n *Node) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *Node) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *Node) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		n.function.FocusNearestView(n, event.Key)
	case glfw.KeyEsc:
		n.function.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n Node) Paint() {
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
