package main

import (
	"code.google.com/p/go.exp/go/types"
)

type indexNode struct {
	*nodeBase
	set bool
	x, key, inVal *input
	outVal, ok *output
}
func newIndexNode(b *block, set bool) *indexNode {
	n := &indexNode{set:set}
	n.nodeBase = newNodeBase(n, b)
	n.x = n.newInput(&types.Var{})
	n.x.connsChanged = func() {
		if conns := n.x.conns; len(conns) > 0 {
			if o := conns[0].src; o != nil { n.updateInputType(o.obj.GetType()) }
		} else {
			n.updateInputType(nil)
		}
	}
	n.key = n.newInput(&types.Var{})
	if set {
		n.inVal = n.newInput(&types.Var{})
		n.text.SetText("[]=")
	} else {
		n.outVal = n.newOutput(&types.Var{})
		n.text.SetText("[]")
	}
	return n
}

func (n *indexNode) updateInputType(t types.Type) {
	if !n.set {
		switch t.(type) {
		case nil, *types.NamedType, *types.Array, *types.Slice:
			if n.ok != nil {
				n.RemoveChild(n.ok)
				n.outs = n.outs[:1]
				n.ok = nil
			}
		case *types.Map:
			if n.ok == nil {
				n.ok = n.newOutput(&types.Var{Name:"ok"})
			}
		}
	}
	switch t.(type) {
	case nil:
	case *types.NamedType:
	case *types.Array:
	case *types.Slice:
	case *types.Map:
	}
}
