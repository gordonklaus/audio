package main

import (
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	."math"
)

type node interface {
	View
	MouseHandler
	block() *block
	inputs() []*input
	outputs() []*output
	inConns() []*connection
	outConns() []*connection
}

type nodeText struct {
	*TextBase
	node *nodeBase
}
func newNodeText(node *nodeBase) *nodeText {
	t := &nodeText{node:node}
	t.TextBase = NewTextBase(t, "")
	return t
}
func (t *nodeText) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyEnter:
		t.node.TakeKeyboardFocus()
	default:
		t.TextBase.KeyPressed(event)
	}
	t.node.reform()
}

type nodeBase struct {
	*ViewBase
	self node
	AggregateMouseHandler
	blk *block
	text *nodeText
	ins []*input
	outs []*output
	focused bool
}

const (
	nodeMargin = 3
)

func newNodeBase(self node, b *block) *nodeBase {
	n := &nodeBase{self:self}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(self), NewViewDragger(self)}
	n.blk = b
	n.text = newNodeText(n)
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
	n.AddChild(n.text)
	n.ViewBase.Self = self
	return n
}

func (n *nodeBase) newInput(v *Value) *input {
	p := newInput(n.self, v)
	n.AddChild(p)
	n.ins = append(n.ins, p)
	n.reform()
	return p
}

func (n *nodeBase) newOutput(v *Value) *output {
	p := newOutput(n.self, v)
	n.AddChild(p)
	n.outs = append(n.outs, p)
	n.reform()
	return p
}

func (n *nodeBase) RemoveChild(v View) {
	n.ViewBase.RemoveChild(v)
	switch v := v.(type) {
	case *input:
		SliceRemove(&n.ins, v)
	case *output:
		SliceRemove(&n.outs, v)
	}
	n.reform()
}

func (n *nodeBase) reform() {
	numIn := float64(len(n.ins))
	numOut := float64(len(n.outs))
	rx, ry := 2.0 * portSize, (Max(numIn, numOut) + 1) * portSize / 2
	
	rect := ZR
	for i, p := range n.ins {
		y := -portSize * (float64(i) - (numIn - 1) / 2)
		p.MoveCenter(Pt(-rx * Sqrt(ry * ry - y * y) / ry, y))
		rect = rect.Union(p.MapRectToParent(p.Rect()))
	}
	for i, p := range n.outs {
		y := -portSize * (float64(i) - (numOut - 1) / 2)
		p.MoveCenter(Pt(rx * Sqrt(ry * ry - y * y) / ry, y))
		rect = rect.Union(p.MapRectToParent(p.Rect()))
	}

	n.text.MoveCenter(Pt(0, rect.Max.Y + n.text.Height() / 2))
	n.Pan(rect.Min)
	n.Resize(rect.Dx(), rect.Dy())
}

func (n nodeBase) block() *block { return n.blk }
func (n nodeBase) inputs() []*input { return n.ins }
func (n nodeBase) outputs() []*output { return n.outs }

func (n nodeBase) inConns() (conns []*connection) {
	for _, p := range n.inputs() {
		for _, c := range p.conns {
			conns = append(conns, c)
		}
	}
	return
}

func (n nodeBase) outConns() (conns []*connection) {
	for _, p := range n.outputs() {
		for _, c := range p.conns {
			conns = append(conns, c)
		}
	}
	return
}

func (n *nodeBase) Move(p Point) {
	n.ViewBase.Move(p)
	f := func(p *port) { for _, c := range p.conns { c.reform() } }
	for _, p := range n.ins { f(p.port) }
	for _, p := range n.outs { f(p.port) }
}

func (n nodeBase) Center() Point { return ZP }

func (n *nodeBase) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *nodeBase) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *nodeBase) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		n.blk.outermost().focusNearestView(n, event.Key)
	case KeyEscape:
		n.blk.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n nodeBase) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	for _, p := range n.ins {
		pt := p.MapToParent(p.Center())
		DrawCubic([4]Point{ZP, Pt(pt.X / 2, 0), Pt(pt.X / 2, pt.Y), pt}, int(pt.Len() / 2))
	}
	for _, p := range n.outs {
		pt := p.MapToParent(p.Center())
		DrawCubic([4]Point{ZP, Pt(pt.X / 2, 0), Pt(pt.X / 2, pt.Y), pt}, int(pt.Len() / 2))
	}
	if len(n.ins) == 0 || len(n.outs) == 0 {
		SetPointSize(7)
		DrawPoint(ZP)
	}
}

func newNode(i Info, b *block) node {
	switch i := i.(type) {
	case Special:
		switch i.Name() {
		case "[]":
			return newIndexNode(b, false)
		case "[]=":
			return newIndexNode(b, true)
		case "if":
			return newIfNode(b)
		case "loop":
			return newLoopNode(b)
		}
	// case StringType:
	// 	return NewBasicLiteralNode(i)
	case *Func:
		return newCallNode(i, b)
	}
	return nil
}

type callNode struct {
	*nodeBase
	info *Func
}
func newCallNode(i *Func, b *block) *callNode {
	n := &callNode{info:i}
	n.nodeBase = newNodeBase(n, b)
	n.text.SetText(i.name)
	for _, v := range i.typ.parameters { n.newInput(v) }
	for _, v := range i.typ.results { n.newOutput(v) }
	
	return n
}

type constantNode struct { *nodeBase }
func newStringConstantNode(b *block) *constantNode {
	n := &constantNode{}
	n.nodeBase = newNodeBase(n, b)
	n.newOutput(&Value{})
	return n
}
