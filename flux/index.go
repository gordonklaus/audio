// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
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
	if set {
		n.elem = n.newInput(nil)
	} else {
		n.elem = n.newOutput(nil)
	}
	n.text.SetText("[]")
	n.text.SetTextColor(color(&types.Func{}, true, false))
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
	if inputType(n.x) == nil {
		// A connection whose destination is being edited may currently be connected to n.x.  It is temporarily disconnected during the call to connectable, but inputs (such as n.elem) with dependent types are not updated, so we have to specifically check for this case here.
		return false
	}
	return assignable(t, dst.obj.Type)
}

func (n *indexNode) connsChanged() {
	if n.set == n.elem.out {
		n.removePortBase(n.elem)
		if n.set {
			n.elem = n.newInput(nil)
		} else {
			n.elem = n.newOutput(nil)
		}
	}

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
			}
			n.addressable = true
		case *types.Slice:
			elem = t.Elem
			if !n.set {
				elem = &types.Pointer{Elem: elem}
			}
			n.addressable = true
		case *types.Map:
			key, elem = t.Key, t.Elem
		}
	}
	n.x.setType(t)
	n.key.setType(key)
	n.elem.setType(elem)

	_, ok := underlying(t).(*types.Map)
	if !n.set && ok && n.ok == nil {
		n.ok = n.newOutput(newVar("ok", types.Typ[types.Bool]))
	}
	if (n.set || !ok) && n.ok != nil {
		n.removePortBase(n.ok)
		n.ok = nil
	}
}

func (n *indexNode) KeyPress(event KeyEvent) {
	if event.Text == "=" {
		if _, ok := underlying(inputType(n.x)).(*types.Map); ok || n.addressable {
			n.set = !n.set
			n.connsChanged()
			SetKeyFocus(n)
		}
	} else {
		n.nodeBase.KeyPress(event)
	}
}
