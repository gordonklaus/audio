// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type appendNode struct {
	*nodeBase
}

func newAppendNode() *appendNode {
	n := &appendNode{}
	n.nodeBase = newNodeBase(n)
	n.text.SetText("append")
	slice := n.newInput(newVar("", generic{}))
	val := n.newInput(newVar("", generic{}))
	out := n.newOutput(newVar("", generic{}))
	slice.connsChanged = func() {
		if len(slice.conns) > 0 {
			t, _ := indirect(slice.conns[0].src.obj.Type)
			if t, ok := t.(*types.Slice); ok {
				slice.setType(t)
				val.setType(t.Elem)
				out.setType(t)
			}
		} else {
			slice.setType(generic{})
			val.setType(generic{})
			out.setType(generic{})
		}
	}
	n.addSeqPorts()
	return n
}

type deleteNode struct {
	*nodeBase
}

func newDeleteNode() *deleteNode {
	n := &deleteNode{}
	n.nodeBase = newNodeBase(n)
	n.text.SetText("delete")
	m := n.newInput(newVar("map", generic{}))
	key := n.newInput(newVar("key", generic{}))
	m.connsChanged = func() {
		if len(m.conns) > 0 {
			t, _ := indirect(m.conns[0].src.obj.Type)
			if t, ok := t.(*types.Map); ok {
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
	in := n.newInput(newVar("", generic{}))
	n.newOutput(newVar("", types.Typ[types.Int]))
	in.connsChanged = func() {
		if len(in.conns) > 0 {
			t := in.conns[0].src.obj.Type
			if tt, ok := indirect(t); ok {
				if _, ok := tt.(*types.Array); !ok {
					t = tt
				}
			}
			in.setType(t)
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
		if nt, ok := t.(*types.Named); ok {
			t = nt.UnderlyingT
		}
		n.newInput(newVar("len", types.Typ[types.Int]))
		if _, ok := t.(*types.Slice); ok {
			n.newInput(newVar("cap", types.Typ[types.Int]))
		}
		MoveCenter(n.typ, Pt(0, Rect(n).Max.Y+Height(n.typ)/2))
		SetKeyFocus(n)
	}
}

func newVar(name string, typ types.Type) *types.Var {
	return types.NewVar(0, nil, name, typ)
}
