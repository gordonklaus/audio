// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type typeAssertNode struct {
	*nodeBase
	typ *typeView
}

func newTypeAssertNode() *typeAssertNode {
	n := &typeAssertNode{}
	n.nodeBase = newNodeBase(n)
	in := n.newInput(&types.Var{})
	in.connsChanged = func() {
		if len(in.conns) > 0 {
			in.setType(in.conns[0].src.obj.Type)
		}
	}
	v := &types.Var{}
	n.newOutput(v)
	n.newOutput(&types.Var{Name: "ok", Type: types.Typ[types.Bool]})
	n.typ = newTypeView(&v.Type)
	n.typ.mode = typesOnly
	n.Add(n.typ)
	return n
}

func (n *typeAssertNode) editType() {
	n.typ.editType(func() {
		if t := *n.typ.typ; t != nil {
			n.setType(t)
		} else {
			n.blk.removeNode(n)
			SetKeyFocus(n.blk)
		}
	})
}

func (n *typeAssertNode) setType(t types.Type) {
	n.typ.setType(t)
	n.outs[0].setType(t)
	if t != nil {
		n.blk.func_().addPkgRef(t)
		MoveCenter(n.typ, Pt(0, Rect(n).Max.Y+Height(n.typ)/2))
		SetKeyFocus(n)
	}
}
