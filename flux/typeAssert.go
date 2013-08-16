package main

import (
	."code.google.com/p/gordon-go/gui"
	"code.google.com/p/go.exp/go/types"
)

type typeAssertNode struct {
	*nodeBase
	typ *typeView
}

func newTypeAssertNode() *typeAssertNode {
	n := &typeAssertNode{}
	n.nodeBase = newNodeBase(n)
	n.newInput(&types.Var{})
	v := &types.Var{}
	n.newOutput(v)
	n.newOutput(&types.Var{Name: "ok", Type: types.Typ[types.Bool]})
	n.typ = newTypeView(&v.Type)
	n.typ.mode = typesOnly
	n.AddChild(n.typ)
	return n
}

func (n *typeAssertNode) editType() {
	n.typ.editType(func() {
		if t := *n.typ.typ; t != nil {
			n.setType(t)
		} else {
			n.blk.removeNode(n)
			n.blk.TakeKeyboardFocus()
		}
	})
}

func (n *typeAssertNode) setType(t types.Type) {
	n.typ.setType(t)
	n.outs[0].setType(t)
	if t != nil {
		n.blk.func_().addPkgRef(t)
		n.typ.MoveCenter(Pt(0, n.Rect().Max.Y + n.typ.Height() / 2))
		n.TakeKeyboardFocus()
	}
}
