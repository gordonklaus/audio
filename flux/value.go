// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type valueNode struct {
	*nodeBase
	obj         types.Object // var or struct field, or nil if this is an assign (=) or indirect node
	set         bool
	x           *port // the target of the operation (struct or pointer)
	y           *port // the result of the read (output) or the argument to write (input)
	addressable bool
}

func newValueNode(obj types.Object, set bool) *valueNode {
	n := &valueNode{obj: obj, set: set}
	n.nodeBase = newNodeBase(n)
	dot := ""
	switch obj.(type) {
	case field, *types.Func, nil:
		if _, ok := obj.(*types.Func); !ok || isMethod(obj) {
			n.x = n.newInput(nil)
			if obj == nil {
				n.x.connsChanged = n.reform
			}
			dot = "."
		}
	default:
	}
	if obj != nil {
		n.text.SetText(dot + obj.GetName())
	}
	if set {
		n.y = n.newInput(nil)
	} else {
		n.y = n.newOutput(nil)
	}
	switch obj.(type) {
	case *types.Var, field, *localVar:
		n.addSeqPorts()
	default:
	}
	n.reform()
	return n
}

func (n *valueNode) reform() {
	if n.set {
		if n.y.out {
			n.removePortBase(n.y)
			n.y = n.newInput(nil)
		}
	} else {
		if !n.y.out {
			n.removePortBase(n.y)
			n.y = n.newOutput(nil)
		}
	}
	var xt, yt types.Type
	if n.obj != nil {
		yt = n.obj.GetType()
	}
	n.addressable = false
	switch obj := n.obj.(type) {
	case *types.Const:
	case *types.Var, *localVar:
		n.addressable = true
	case *types.Func:
		if isMethod(obj) {
			xt = obj.Type.(*types.Signature).Recv.Type
			// TODO: remove Recv? (from copy)
		}
	case field:
		xt = obj.recv
		n.addressable = obj.addressable
	case nil:
		if len(n.x.conns) > 0 {
			xt = n.x.conns[0].src.obj.Type
			yt, _ = indirect(xt)
		}
		if n.set {
			n.text.SetText("=")
		} else {
			n.text.SetText("indirect")
		}
	}
	if !n.set && n.addressable {
		yt = &types.Pointer{Elem: yt}
	}
	if n.x != nil {
		n.x.setType(xt)
	}
	n.y.setType(yt)
}

func (n *valueNode) KeyPress(event KeyEvent) {
	if event.Text == "=" && n.addressable {
		n.set = !n.set
		n.reform()
		SetKeyFocus(n)
	} else {
		n.nodeBase.KeyPress(event)
	}
}
