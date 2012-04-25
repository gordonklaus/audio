package main

import (
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/gui"
	."image"
)

func Max(x, y int) int {
	if x > y { return x }
	return y
}

type Node struct {
	ViewBase
	AggregateMouseHandler
	function *Function
	// inPorts []*InPort
	// outPorts []*OutPort
	name *Text
	focused bool
}

const (
	margin = 12
	
	nodeVerticalMargin = 7.0
	portDepth = 7.0
	nodeStringFontSize = 12
)

func NewNode(function *Function) *Node {
	n := &Node{}
	n.ViewBase = *NewView(n)
	n.function = function
	function.AddChild(n)
	// n.inPorts = make([]*InPort, 0)
	// n.outPorts = make([]*OutPort, 0)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}

	n.name = NewText("")
	n.AddChild(n.name)
	n.name.TextChanged.Connect(func(...interface{}) { n.Resize(n.name.Width() + 2*margin, n.name.Height() + 2*margin) })
	n.name.SetText("")
	n.name.Move(Pt(margin, margin))
	n.name.TakeKeyboardFocus()
	
	// numInPorts := len(node.InPorts())
	// numOutPorts := len(node.OutPorts())
	// maxPorts := Max(numInPorts, numOutPorts)
	// n.Resize(96, 2*nodeVerticalMargin + maxPorts * portSize)
	// inPortsVerticalOffset := portSize * float64(maxPorts - numInPorts) / 2
	// for i, inPort := range node.InPorts() {
	// 	p := NewInPort(n, inPort)
	// 	p.Move(image.Pt(0, int(nodeVerticalMargin + inPortsVerticalOffset + float64(i) * portSize)))
	// 	n.AddChild(p)
	// 	n.inPorts = append(n.inPorts, p)
	// }
	// outPortsVerticalOffset := portSize * float64(maxPorts - numOutPorts) / 2
	// for i, outPort := range node.OutPorts() {
	// 	p := NewOutPort(n, outPort)
	// 	p.Move(image.Pt(int(n.Width()) - portSize, int(nodeVerticalMargin + outPortsVerticalOffset + float64(i) * portSize)))
	// 	n.AddChild(p)
	// 	n.outPorts = append(n.outPorts, p)
	// }
	
	return n
}

// func (n Node) GetPorts() []View {
// 	ports := make([]View, 0, len(n.inPorts) + len(n.outPorts))
// 	for _, p := range n.inPorts { ports = append(ports, p) }
// 	for _, p := range n.outPorts { ports = append(ports, p) }
// 	return ports
// }
// 
// func (n Node) Moved(image.Point) {
// 	f := func(p *Port) { for _, ch := range p.connections { ch.reform() } }
// 	for _, p := range n.inPorts { f(p.Port) }
// 	for _, p := range n.outPorts { f(p.Port) }
// }
// 
// func (n *Node) TookKeyboardFocus() { n.focused = true; n.RepaintAll() }
// func (n *Node) LostKeyboardFocus() { n.focused = false; n.RepaintAll() }
// 
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
	// numInPorts, numOutPorts := len(n.inPorts), len(n.outPorts)
	// maxPorts := Max(numInPorts, numOutPorts)
	// 
	// const horizontalOffset = 3 * portSize / 5
	// 
	// edgeColor := map[bool]image.NRGBAColor{false:{31, 31, 31, 191}, true:{31, 31, 255, 191}}
	// painter.SetStrokeColor(edgeColor[n.focused])
	// painter.SetLineWidth(2)
	// painter.SetFillColor(image.NRGBAColor{95, 95, 95, 191})
	// inPortsVerticalOffset := portSize * float64(maxPorts - numInPorts) / 2
	// painter.MoveTo(horizontalOffset, nodeVerticalMargin + inPortsVerticalOffset)
	// for i := 0; i < numInPorts; i++ {
	// 	painter.RCubicCurveTo(0, portSize / 4, portDepth, portSize / 4, portDepth, portSize / 2)
	// 	painter.RCubicCurveTo(0, portSize / 4, -portDepth, portSize / 4, -portDepth, portSize / 2)
	// }
	// outPortsVerticalOffset := portSize * float64(maxPorts - numOutPorts) / 2
	// painter.CubicCurveTo(horizontalOffset, float64(n.Height()), float64(n.Width()) - horizontalOffset, float64(n.Height()), float64(n.Width()) - horizontalOffset, float64(n.Height()) - nodeVerticalMargin - outPortsVerticalOffset)
	// for i := 0; i < numOutPorts; i++ {
	// 	painter.RCubicCurveTo(0, -portSize / 4, -portDepth, -portSize / 4, -portDepth, -portSize / 2)
	// 	painter.RCubicCurveTo(0, -portSize / 4, portDepth, -portSize / 4, portDepth, -portSize / 2)
	// }
	// painter.CubicCurveTo(float64(n.Width()) - horizontalOffset, 0, horizontalOffset, 0, horizontalOffset, nodeVerticalMargin + inPortsVerticalOffset)
	// painter.FillStroke()
	// 
	// painter.SetFontSize(nodeStringFontSize)
	// painter.SetStrokeColor(image.NRGBAColor{31, 31, 31, 255})
	// painter.MoveTo(32, float64(n.Height() + nodeStringFontSize) / 2)
	// painter.FillString(n.node.String())
	
	gl.Color4d(1, 1, 1, .25)
	gl.Rectd(0, 0, gl.Double(n.Width()), gl.Double(n.Height()))
}
