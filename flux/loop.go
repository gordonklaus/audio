package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type loopNode struct {
	*ViewBase
	AggregateMouser
	blk           *block
	input         *port
	seqIn, seqOut *port
	loopblk       *block
	inputsNode    *portsNode
	focused       bool
}

func newLoopNode(arranged blockchan) *loopNode {
	n := &loopNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(n), NewMover(n)}
	n.input = newInput(n, &types.Var{})
	n.input.connsChanged = n.updateInputType
	MoveCenter(n.input, Pt(-portSize/2, 0))
	n.Add(n.input)

	n.seqIn = newInput(n, &types.Var{Name: "seq", Type: seqType})
	MoveCenter(n.seqIn, Pt(portSize/2, 0))
	n.Add(n.seqIn)
	n.seqOut = newOutput(n, &types.Var{Name: "seq", Type: seqType})
	n.Add(n.seqOut)

	n.loopblk = newBlock(n, arranged)
	n.inputsNode = newInputsNode()
	n.inputsNode.newOutput(&types.Var{})
	n.loopblk.addNode(n.inputsNode)
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
	var t, u, key, elt types.Type
	key = types.Typ[types.Int]
	// TODO: loop over conns until a non-nil src is found, then break after getting the type (and do the same in other nodes)
	if len(n.input.conns) > 0 {
		if p := n.input.conns[0].src; p != nil {
			ptr := false
			t, ptr = indirect(p.obj.Type)
			u = t
			if n, ok := t.(*types.NamedType); ok {
				u = n.Underlying
			}
			switch u := u.(type) {
			case *types.Basic:
				if u.Info&types.IsString != 0 {
					elt = types.Typ[types.Rune]
				} else if u.Kind != types.UntypedInt {
					key = u
				}
			case *types.Array:
				elt = u.Elt
				if ptr {
					t = p.obj.Type
					elt = &types.Pointer{elt}
				}
			case *types.Slice:
				elt = &types.Pointer{u.Elt}
			case *types.Map:
				key, elt = u.Key, u.Elt
			case *types.Chan:
				key = u.Elt
			}
		}
	}

	in := n.inputsNode
	switch u.(type) {
	default:
		if len(in.outs) == 2 {
			for _, c := range in.outs[1].conns {
				c.blk.removeConn(c)
			}
			in.Remove(in.outs[1])
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
	rearrange(n.loopblk)
}

func (n *loopNode) Move(p Point) {
	n.ViewBase.Move(p)
	nodeMoved(n)
}

func (n *loopNode) TookKeyFocus() { n.focused = true; Repaint(n) }
func (n *loopNode) LostKeyFocus() { n.focused = false; Repaint(n) }

func (n *loopNode) KeyPress(event KeyEvent) {
	switch event.Key {
	case KeyRight:
		SetKeyFocus(n.inputsNode)
	default:
		n.ViewBase.KeyPress(event)
	}
}

func (n loopNode) Paint() {
	SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[n.focused])
	DrawLine(Pt(-portSize/2, 0), Pt(portSize/2, 0))
}
