// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/flux/go/types"
	. "code.google.com/p/gordon-go/flux/gui"
)

type convertNode struct {
	*nodeBase
}

func newConvertNode(currentPkg *types.Package) *convertNode {
	n := &convertNode{}
	n.nodeBase = newNodeBase(n)
	in := n.newInput(nil)
	out := n.newOutput(nil)
	in.connsChanged = func() {
		t := untypedToTyped(inputType(in))
		in.setType(t)
		if t != nil {
			out.setType(*n.typ.typ)
		} else {
			out.setType(nil)
		}
	}
	n.typ = newTypeView(new(types.Type), currentPkg)
	n.typ.mode = anyType
	n.Add(n.typ)
	return n
}

func (n *convertNode) editType() {
	n.typ.editType(func() {
		if t := *n.typ.typ; t != nil {
			n.setType(t)
		} else {
			n.blk.removeNode(n)
			SetKeyFocus(n.blk)
		}
	})
}

func (n *convertNode) setType(t types.Type) {
	n.typ.setType(t)
	if t != nil {
		n.blk.func_().addPkgRef(t)
		n.reform()
		SetKeyFocus(n)
	}
}

func (n *convertNode) connectable(t types.Type, dst *port) bool {
	return types.ConvertibleTo(t, *n.typ.typ)
}
