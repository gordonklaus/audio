package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type loopNode struct {
	*ViewBase
	AggregateMouseHandler
	blk           *block
	input         *port
	seqIn, seqOut *port
	inputsNode    *portsNode
	loopblk       *block
	focused       bool
}

func newLoopNode() *loopNode {
	n := &loopNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.input = newInput(n, &types.Var{})
	n.input.connsChanged = n.updateInputType
	MoveCenter(n.input, Pt(-2*portSize, 0))
	n.AddChild(n.input)
	n.loopblk = newBlock(n)
	n.inputsNode = newInputsNode()
	n.inputsNode.newOutput(&types.Var{})
	n.loopblk.addNode(n.inputsNode)
	n.AddChild(n.loopblk)

	n.seqIn = newInput(n, &types.Var{Name: "seq", Type: seqType})
	MoveCenter(n.seqIn, ZP)
	n.AddChild(n.seqIn)
	n.seqOut = newOutput(n, &types.Var{Name: "seq", Type: seqType})
	n.AddChild(n.seqOut)

	n.updateInputType()
	return n
}

func (n loopNode) block() *block      { return n.blk }
func (n *loopNode) setBlock(b *block) { n.blk = b }
func (n loopNode) inputs() []*port    { return []*port{n.seqIn, n.input} }
func (n loopNode) outputs() []*port   { return []*port{n.seqOut} }
func (n loopNode) inConns() []*connection {
	return append(n.seqIn.conns, append(n.input.conns, n.loopblk.inConns()...)...)
}
func (n loopNode) outConns() []*connection {
	return append(n.seqOut.conns, n.loopblk.outConns()...)
}

func (n *loopNode) updateInputType() {
	var t, key, elt types.Type
	key = types.Typ[types.Int]
	if len(n.input.conns) > 0 {
		if p := n.input.conns[0].src; p != nil {
			t = p.obj.Type
			if n, ok := t.(*types.NamedType); ok {
				t = n.Underlying
			}
			switch t := t.(type) {
			case *types.Basic:
				if t.Info&types.IsString != 0 {
					elt = types.Typ[types.Rune]
				} else if t.Kind != types.UntypedInt {
					key = t
				}
			case *types.Array:
				elt = t.Elt
			case *types.Slice:
				elt = t.Elt
			case *types.Map:
				key, elt = t.Key, t.Elt
			case *types.Chan:
				key = t.Elt
			}
		}
	}

	in := n.inputsNode
	switch t.(type) {
	default:
		if len(in.outs) == 2 {
			for _, c := range in.outs[1].conns {
				c.blk.removeConnection(c)
			}
			in.RemoveChild(in.outs[1])
			in.outs = in.outs[:1]
		}
	case *types.Array, *types.Slice, *types.Map:
		if len(in.outs) == 1 {
			in.newOutput(&types.Var{})
		}
	}

	n.input.setType(t)
	in.outs[0].setType(key)
	if len(in.outs) == 2 {
		in.outs[1].setType(elt)
	}
	in.reform()
	in.reposition()
}

func (n *loopNode) update() bool {
	b := n.loopblk
	if !b.update() {
		return false
	}
	b.Move(Pt(0, -Height(b)/2))
	n.inputsNode.reposition()
	MoveCenter(n.seqOut, Pt(Width(b), 0))
	ResizeToFit(n, 0)
	return true
}

func (n *loopNode) Move(p Point) {
	n.ViewBase.Move(p)
	for _, c := range append(n.inConns(), n.outConns()...) {
		c.reform()
	}
}

func (n *loopNode) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *loopNode) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *loopNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		n.blk.outermost().focusNearestView(n, event.Key)
	case KeyEscape:
		SetKeyboardFocus(n.blk)
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n loopNode) Paint() {
	SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[n.focused])
	DrawLine(ZP, Pt(-2*portSize, 0))
}
