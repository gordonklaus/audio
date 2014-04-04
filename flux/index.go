// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
)

type indexNode struct {
	*nodeBase
	set, addressable bool
	x, key, elem, ok *port
}

func newIndexNode(set bool) *indexNode {
	n := &indexNode{set: set}
	n.nodeBase = newNodeBase(n)
	n.x = n.newInput(nil)
	n.x.connsChanged = n.connsChanged
	n.key = n.newInput(nil)
	n.key.connsChanged = n.connsChanged
	if set {
		n.elem = n.newInput(nil)
		n.text.SetText("[]=")
	} else {
		n.elem = n.newOutput(nil)
		n.text.SetText("[]")
	}
	n.elem.connsChanged = n.connsChanged
	n.addSeqPorts()
	n.connsChanged()
	return n
}

func (n *indexNode) connectable(t types.Type, dst *port) bool {
	if dst == n.x {
		ok := false
		switch t := underlying(t).(type) {
		case *types.Array, *types.Slice, *types.Map:
			ok = true
		case *types.Pointer:
			_, ok = underlying(t.Elem).(*types.Array)
		}
		return ok
	}
	return assignable(t, dst.obj.Type)
}

func (n *indexNode) connsChanged() {
	t := inputType(n.x)
	var key, elem types.Type
	n.addressable = false
	if t != nil {
		key = types.Typ[types.Int]
		switch t := underlying(t).(type) {
		case *types.Array:
			elem = t.Elem
		case *types.Pointer:
			elem = underlying(t.Elem).(*types.Array).Elem
			if !n.set {
				elem = &types.Pointer{Elem: elem}
				n.addressable = true
			}
		case *types.Slice:
			elem = t.Elem
			if !n.set {
				elem = &types.Pointer{Elem: elem}
				n.addressable = true
			}
		case *types.Map:
			key, elem = t.Key, t.Elem
		}
	}
	n.x.setType(t)
	n.key.setType(key)
	n.elem.setType(elem)

	if !n.set {
		_, isMap := underlying(t).(*types.Map)
		if isMap && n.ok == nil {
			n.ok = n.newOutput(newVar("ok", types.Typ[types.Bool]))
		}
		if !isMap && n.ok != nil {
			n.removePortBase(n.ok)
			n.ok = nil
		}
	}
}
