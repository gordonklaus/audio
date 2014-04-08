// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type chanNode struct {
	*nodeBase
	send         bool
	ch, elem, ok *port
}

func newChanNode(send bool) *chanNode {
	n := &chanNode{send: send}
	n.nodeBase = newNodeBase(n)
	n.ch = n.newInput(nil)
	n.ch.connsChanged = n.connsChanged
	if send {
		n.elem = n.newInput(nil)
	} else {
		n.elem = n.newOutput(nil)
		n.ok = n.newOutput(newVar("ok", nil))
	}
	n.text.SetText("<-")
	n.text.SetTextColor(color(&types.Func{}, true, false))
	n.addSeqPorts()
	n.connsChanged()
	return n
}

func (n *chanNode) connectable(t types.Type, dst *port) bool {
	if dst == n.ch {
		t, ok := underlying(t).(*types.Chan)
		if !ok {
			return false
		}
		switch t.Dir {
		case types.SendRecv:
			return true
		case types.SendOnly:
			return n.send
		case types.RecvOnly:
			return !n.send
		}
	}
	if inputType(n.ch) == nil {
		// A connection whose destination is being edited may currently be connected to n.ch.  It is temporarily disconnected during the call to connectable, but inputs (such as n.elem) with dependent types are not updated, so we have to specifically check for this case here.
		return false
	}
	return assignable(t, dst.obj.Type)
}

func (n *chanNode) connsChanged() {
	if n.send == n.elem.out {
		n.removePortBase(n.elem)
		if n.send {
			n.elem = n.newInput(nil)
			n.removePortBase(n.ok)
		} else {
			n.elem = n.newOutput(nil)
			n.ok = n.newOutput(newVar("ok", nil))
		}
	}

	t := inputType(n.ch)
	var elem, ok types.Type
	if t != nil {
		elem = underlying(t).(*types.Chan).Elem
		ok = types.Typ[types.Bool]
	}
	n.ch.setType(t)
	n.elem.setType(elem)
	if !n.send {
		n.ok.setType(ok)
	}
}

func (n *chanNode) KeyPress(event KeyEvent) {
	if event.Text == "=" {
		t, _ := underlying(inputType(n.ch)).(*types.Chan)
		if t == nil || t.Dir == types.SendRecv {
			n.send = !n.send
			n.connsChanged()
			SetKeyFocus(n)
		}
	} else {
		n.nodeBase.KeyPress(event)
	}
}

type closeNode struct {
	*nodeBase
}

func newCloseNode() *closeNode {
	n := &closeNode{}
	n.nodeBase = newNodeBase(n)
	n.text.SetText("close")
	n.text.SetTextColor(color(&types.Func{}, true, false))
	in := n.newInput(newVar("", nil))
	in.connsChanged = func() {
		in.setType(inputType(in))
	}
	n.addSeqPorts()
	return n
}

func (n *closeNode) connectable(t types.Type, dst *port) bool {
	ch, ok := underlying(t).(*types.Chan)
	return ok && ch.Dir != types.RecvOnly
}
