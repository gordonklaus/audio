package main

import (
	"code.google.com/p/go.exp/go/types"
)

type indexNode struct {
	*nodeBase
	set bool
	x, key, inVal *port
	outVal, ok *port
}
func newIndexNode(set bool) *indexNode {
	n := &indexNode{set:set}
	n.nodeBase = newNodeBase(n)
	n.x = n.newInput(&types.Var{})
	f := func() { n.updateInputType() }
	n.x.connsChanged = f
	n.key = n.newInput(&types.Var{})
	n.key.connsChanged = f
	if set {
		n.inVal = n.newInput(&types.Var{})
		n.inVal.connsChanged = f
		n.text.SetText("[]=")
	} else {
		n.outVal = n.newOutput(&types.Var{})
		n.text.SetText("[]")
	}
	return n
}

func (n *indexNode) updateInputType() {
	var t, key, elt types.Type
	if len(n.x.conns) > 0 {
		if p := n.x.conns[0].src; p != nil {
			t = p.obj.Type
			if n, ok := t.(*types.NamedType); ok {
				t = n.Underlying
			}
			key = types.Typ[types.Int]
			switch t := t.(type) {
			case *types.Array: elt = t.Elt
			case *types.Slice: elt = t.Elt
			case *types.Map:   key, elt = t.Key, t.Elt
			}
		}
	} else {
		if len(n.key.conns) > 0 {
			if o := n.key.conns[0].src; o != nil {
				key = o.obj.Type
			}
		}
		if n.set && len(n.inVal.conns) > 0 {
			if o := n.inVal.conns[0].src; o != nil {
				elt = o.obj.Type
			}
		}
	}
	if t == nil && key != nil && elt != nil {
		t = &types.Map{Key: key, Elt: elt}
	}
	
	if !n.set {
		switch t.(type) {
		default:
			if n.ok != nil {
				for _, c := range n.ok.conns {
					c.blk.removeConnection(c)
				}
				n.RemoveChild(n.ok)
				n.outs = n.outs[:1]
				n.ok = nil
			}
		case *types.Map:
			if n.ok == nil {
				n.ok = n.newOutput(&types.Var{Name: "ok", Type: types.Typ[types.Bool]})
			}
		}
	}
	
	n.x.valView.setType(t)
	n.key.valView.setType(key)
	if n.set {
		n.inVal.valView.setType(elt)
	} else {
		n.outVal.valView.setType(elt)
	}
}
