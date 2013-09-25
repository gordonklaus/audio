package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type deleteNode struct {
	*nodeBase
}

func newDeleteNode() *deleteNode {
	n := &deleteNode{}
	n.nodeBase = newNodeBase(n)
	n.text.SetText("delete")
	m := n.newInput(&types.Var{Name: "map", Type: generic{}})
	key := n.newInput(&types.Var{Name: "key", Type: generic{}})
	m.connsChanged = func() {
		if len(m.conns) > 0 {
			if t, ok := m.conns[0].src.obj.Type.(*types.Map); ok {
				m.setType(t)
				key.setType(t.Key)
			}
		} else {
			m.setType(generic{})
			key.setType(generic{})
		}
	}
	n.addSeqPorts()
	return n
}

type lenNode struct {
	*nodeBase
}

func newLenNode() *lenNode {
	n := &lenNode{}
	n.nodeBase = newNodeBase(n)
	n.text.SetText("len")
	in := n.newInput(&types.Var{Type: generic{}})
	n.newOutput(&types.Var{Type: types.Typ[types.Int]})
	in.connsChanged = func() {
		if len(in.conns) > 0 {
			in.setType(in.conns[0].src.obj.Type)
		} else {
			in.setType(generic{})
		}
	}
	n.addSeqPorts()
	return n
}

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
	n.Add(n.typ)
	return n
}

func (n *makeNode) editType() {
	n.typ.editType(func() {
		if t := *n.typ.typ; t != nil {
			n.setType(t)
		} else {
			n.blk.removeNode(n)
			SetKeyFocus(n.blk)
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
		MoveCenter(n.typ, Pt(0, Rect(n).Max.Y+Height(n.typ)/2))
		SetKeyFocus(n)
	}
}
