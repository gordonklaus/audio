package flux

import (
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/gui"
)

func Max(x, y int) int {
	if x > y { return x }
	return y
}

type Component struct {
	ViewBase
	AggregateMouseHandler
	compound *Compound
	// inPorts []*InPort
	// outPorts []*OutPort
	focused bool
}

const (
	componentVerticalMargin = 7.0
	portDepth = 7.0
	componentStringFontSize = 12
)

func NewComponent(compound *Compound) *Component {
	c := &Component{}
	c.ViewBase = *NewView(c)
	c.compound = compound
	// c.inPorts = make([]*InPort, 0)
	// c.outPorts = make([]*OutPort, 0)
	c.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(c), NewViewDragger(c)}
	
	c.Resize(96, 64)
	
	// numInPorts := len(component.InPorts())
	// numOutPorts := len(component.OutPorts())
	// maxPorts := Max(numInPorts, numOutPorts)
	// c.Resize(96, 2*componentVerticalMargin + maxPorts * portSize)
	// inPortsVerticalOffset := portSize * float64(maxPorts - numInPorts) / 2
	// for i, inPort := range component.InPorts() {
	// 	p := NewInPort(c, inPort)
	// 	p.Move(image.Pt(0, int(componentVerticalMargin + inPortsVerticalOffset + float64(i) * portSize)))
	// 	c.AddChild(p)
	// 	c.inPorts = append(c.inPorts, p)
	// }
	// outPortsVerticalOffset := portSize * float64(maxPorts - numOutPorts) / 2
	// for i, outPort := range component.OutPorts() {
	// 	p := NewOutPort(c, outPort)
	// 	p.Move(image.Pt(int(c.Width()) - portSize, int(componentVerticalMargin + outPortsVerticalOffset + float64(i) * portSize)))
	// 	c.AddChild(p)
	// 	c.outPorts = append(c.outPorts, p)
	// }
	
	return c
}

// func (c Component) GetPorts() []View {
// 	ports := make([]View, 0, len(c.inPorts) + len(c.outPorts))
// 	for _, p := range c.inPorts { ports = append(ports, p) }
// 	for _, p := range c.outPorts { ports = append(ports, p) }
// 	return ports
// }
// 
// func (c Component) Moved(image.Point) {
// 	f := func(p *Port) { for _, ch := range p.connections { ch.reform() } }
// 	for _, p := range c.inPorts { f(p.Port) }
// 	for _, p := range c.outPorts { f(p.Port) }
// }
// 
// func (c *Component) TookKeyboardFocus() { c.focused = true; c.RepaintAll() }
// func (c *Component) LostKeyboardFocus() { c.focused = false; c.RepaintAll() }
// 
// func (c *Component) KeyPressed(key int) {
// 	switch key {
// 	case Key_Left, Key_Right, Key_Up, Key_Down:
// 		c.compound.FocusNearestView(c, key)
// 	case Key_Escape:
// 		c.compound.TakeKeyboardFocus()
// 	default:
// 		c.ViewBase.KeyPressed(key)
// 	}
// }

func (c Component) Paint() {
	// numInPorts, numOutPorts := len(c.inPorts), len(c.outPorts)
	// maxPorts := Max(numInPorts, numOutPorts)
	// 
	// const horizontalOffset = 3 * portSize / 5
	// 
	// edgeColor := map[bool]image.NRGBAColor{false:{31, 31, 31, 191}, true:{31, 31, 255, 191}}
	// painter.SetStrokeColor(edgeColor[c.focused])
	// painter.SetLineWidth(2)
	// painter.SetFillColor(image.NRGBAColor{95, 95, 95, 191})
	// inPortsVerticalOffset := portSize * float64(maxPorts - numInPorts) / 2
	// painter.MoveTo(horizontalOffset, componentVerticalMargin + inPortsVerticalOffset)
	// for i := 0; i < numInPorts; i++ {
	// 	painter.RCubicCurveTo(0, portSize / 4, portDepth, portSize / 4, portDepth, portSize / 2)
	// 	painter.RCubicCurveTo(0, portSize / 4, -portDepth, portSize / 4, -portDepth, portSize / 2)
	// }
	// outPortsVerticalOffset := portSize * float64(maxPorts - numOutPorts) / 2
	// painter.CubicCurveTo(horizontalOffset, float64(c.Height()), float64(c.Width()) - horizontalOffset, float64(c.Height()), float64(c.Width()) - horizontalOffset, float64(c.Height()) - componentVerticalMargin - outPortsVerticalOffset)
	// for i := 0; i < numOutPorts; i++ {
	// 	painter.RCubicCurveTo(0, -portSize / 4, -portDepth, -portSize / 4, -portDepth, -portSize / 2)
	// 	painter.RCubicCurveTo(0, -portSize / 4, portDepth, -portSize / 4, portDepth, -portSize / 2)
	// }
	// painter.CubicCurveTo(float64(c.Width()) - horizontalOffset, 0, horizontalOffset, 0, horizontalOffset, componentVerticalMargin + inPortsVerticalOffset)
	// painter.FillStroke()
	// 
	// painter.SetFontSize(componentStringFontSize)
	// painter.SetStrokeColor(image.NRGBAColor{31, 31, 31, 255})
	// painter.MoveTo(32, float64(c.Height() + componentStringFontSize) / 2)
	// painter.FillString(c.component.String())
	
	gl.Color4d(1, 1, 1, 1)
	gl.Rectd(0, 0, gl.Double(c.Width()), gl.Double(c.Height()))
}
