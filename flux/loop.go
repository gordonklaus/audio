package main

import (
	."code.google.com/p/gordon-go/gui"
	"code.google.com/p/go.exp/go/types"
)

type loopNode struct {
	*ViewBase
	AggregateMouseHandler
	blk *block
	input *port
	inputsNode *portsNode
	loopblk *block
	focused bool
}

func newLoopNode() *loopNode {
	n := &loopNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.input = newInput(n, &types.Var{})
	n.input.connsChanged = func() { n.updateInputType() }
	n.AddChild(n.input)
	n.loopblk = newBlock(n)
	n.inputsNode = newInputsNode()
	n.inputsNode.newOutput(&types.Var{})
	n.loopblk.addNode(n.inputsNode)
	n.AddChild(n.loopblk)
	
	n.input.MoveCenter(Pt(-2*portSize, 0))
	n.updateInputType()
	return n
}

func (n loopNode) block() *block { return n.blk }
func (n *loopNode) setBlock(b *block) { n.blk = b }
func (n loopNode) inputs() []*port { return []*port{n.input} }
func (n loopNode) outputs() []*port { return nil }
func (n loopNode) inConns() []*connection {
	return append(n.input.conns, n.loopblk.inConns()...)
}
func (n loopNode) outConns() []*connection {
	return n.loopblk.outConns()
}

func (n *loopNode) updateInputType() {
	var t, key, elt types.Type
	key = types.Typ[types.Int]
	if len(n.input.conns) > 0 {
		if p := n.input.conns[0].src; p != nil {
			t = p.obj.Type
			switch t := t.(type) {
			case *types.Basic, *types.NamedType:
				u := t
				if n, ok := u.(*types.NamedType); ok {
					u = n.Underlying
				}
				b := u.(*types.Basic)
				if b.Info & types.IsString != 0 {
					elt = types.Typ[types.Rune]
				} else if b.Kind != types.UntypedInt {
					key = b
				}
			case *types.Array: elt = t.Elt
			case *types.Slice: elt = t.Elt
			case *types.Map:   key, elt = t.Key, t.Elt
			case *types.Chan:  key = t.Elt
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
	
	n.input.valView.setType(t)
	in.outs[0].valView.setType(key)
	if len(in.outs) == 2 {
		in.outs[1].valView.setType(elt)
	}
}

func (n *loopNode) update() bool {
	b := n.loopblk
	if !b.update() {
		return false
	}
	y2 := b.Size().Y / 2
	b.Move(Pt(0, -y2))
	n.inputsNode.MoveOrigin(b.Rect().Min.Add(Pt(0, y2)))
	ResizeToFit(n, 0)
	return true
}

func (n *loopNode) Move(p Point) {
	n.ViewBase.Move(p)
	for _, c := range n.input.conns { c.reform() }
}

func (n *loopNode) TookKeyboardFocus() { n.focused = true; n.Repaint() }
func (n *loopNode) LostKeyboardFocus() { n.focused = false; n.Repaint() }

func (n *loopNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		n.blk.outermost().focusNearestView(n, event.Key)
	case KeyEscape:
		n.blk.TakeKeyboardFocus()
	default:
		n.ViewBase.KeyPressed(event)
	}
}

func (n loopNode) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[n.focused])
	DrawLine(ZP, Pt(-2*portSize, 0))
}
