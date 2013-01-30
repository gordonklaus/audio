package main

import (
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
	n.text.SetText("[]")
	n.x = n.newInput(&Value{})
	n.x.connsChanged = func() {
		if conns := n.x.conns; len(conns) > 0 {
			if o := conns[0].src; o != nil { n.updateInputType(o.val.typ) }
		} else {
			n.updateInputType(nil)
		}
	}
	n.key = n.newInput(&Value{})
	if set {
		n.inVal = n.newInput(&Value{})
	} else {
		n.outVal = n.newOutput(&Value{})
	}
	return n
}

func (n *indexNode) updateInputType(t Type) {
	if !n.set {
		switch t.(type) {
		case nil, *NamedType, *ArrayType, *SliceType:
			if n.ok != nil {
				n.RemoveChild(n.ok)
				n.outs = n.outs[:1]
				n.ok = nil
			}
		case *MapType:
			if n.ok == nil {
				n.ok = n.newOutput(&Value{})
				n.ok.val.name = "ok"
			}
		}
	}
	switch t.(type) {
	case nil:
	case *NamedType:
	case *ArrayType:
	case *SliceType:
	case *MapType:
	}
}
