package main

import (
	"code.google.com/p/go.exp/go/types"
	."code.google.com/p/gordon-go/gui"
)

type makeNode struct {
	*nodeBase
	typ *typeView
}

func newMakeNode() *makeNode {
	n := &makeNode{}
	n.nodeBase = newNodeBase(n)
	v := &types.Var{}
	n.newOutput(v)
	n.typ = newTypeView(&v.Type)
	n.typ.mode = makeableType
	n.AddChild(n.typ)
	return n
}

func (n *makeNode) editType() {
	n.typ.editType(func() {
		if t := *n.typ.typ; t != nil {
			n.setType(t)
		} else {
			n.blk.removeNode(n)
			n.blk.TakeKeyboardFocus()
		}
	})
}

func (n *makeNode) setType(t types.Type) {
	n.typ.setType(t)
	n.outs[0].setType(t)
	if t != nil {
		n.blk.func_().addPkgRef(t)
		if nt, ok := t.(*types.NamedType); ok {
			t = nt.Underlying
		}
		n.newInput(&types.Var{Name: "len", Type: types.Typ[types.Int]})
		if _, ok := t.(*types.Slice); ok {
			n.newInput(&types.Var{Name: "cap", Type: types.Typ[types.Int]})
		}
		n.typ.MoveCenter(Pt(0, n.Rect().Max.Y + n.typ.Height() / 2))
		n.TakeKeyboardFocus()
	}
}
