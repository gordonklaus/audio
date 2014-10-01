// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/flux/go/types"
	. "code.google.com/p/gordon-go/flux/gui"
)

type typeAssertNode struct {
	*nodeBase
}

func newTypeAssertNode(currentPkg *types.Package) *typeAssertNode {
	n := &typeAssertNode{}
	n.nodeBase = newNodeBase(n)
	in := n.newInput(nil)
	out := n.newOutput(nil)
	ok := n.newOutput(newVar("ok", nil))
	in.connsChanged = func() {
		t := inputType(in)
		var u, b types.Type
		if t != nil {
			u = *n.typ.typ
			b = types.Typ[types.Bool]
		}
		in.setType(t)
		out.setType(u)
		ok.setType(b)
	}
	n.typ = newTypeView(new(types.Type), currentPkg)
	n.typ.mode = anyType
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
	if t != nil {
		n.blk.func_().addPkgRef(t)
		n.reform()
		SetKeyFocus(n)
	}
}

func (n *typeAssertNode) connectable(t types.Type, dst *port) bool {
	i, ok := underlying(t).(*types.Interface)
	return ok && types.AssertableTo(i, *n.typ.typ)
}
