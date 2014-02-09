// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type callNode struct {
	*nodeBase
	obj types.Object
}

func newCallNode(obj types.Object) node {
	if obj != nil {
		if sig, ok := obj.GetType().(*types.Signature); ok {
			n := &callNode{obj: obj}
			n.nodeBase = newNodeBase(n)
			name := obj.GetName()
			if sig.Recv != nil {
				name = "." + name
			}
			n.text.SetText(name)
			n.addSeqPorts()
			n.addPorts(sig)
			return n
		}

		switch obj.GetName() {
		case "append":
			return newAppendNode()
		case "delete":
			return newDeleteNode()
		case "len":
			return newLenNode()
		case "make":
			return newMakeNode()
		default:
			panic("unknown builtin: " + obj.GetName())
		}
	} else {
		n := &callNode{}
		n.nodeBase = newNodeBase(n)
		n.text.SetText("call")
		n.addSeqPorts()
		in := n.newInput(nil)
		in.connsChanged = func() {
			for _, p := range append(ins(n), outs(n)...) {
				if p != in {
					n.removePortBase(p)
				}
			}
			if len(in.conns) > 0 {
				t, _ := indirect(in.conns[0].src.obj.Type)
				if sig, ok := t.(*types.Signature); ok { // TODO: remove ok (always true) after canConnect checks types
					in.setType(sig)
					n.addPorts(sig)
				}
			} else {
				in.setType(nil)
			}
		}
		return n
	}
}

func (n *callNode) addPorts(sig *types.Signature) {
	if sig.Recv != nil {
		n.newInput(sig.Recv)
	}
	params := sig.Params
	if sig.IsVariadic {
		params = params[:len(params)-1]
	}
	for _, v := range params {
		n.newInput(v)
	}
	for _, v := range sig.Results {
		n.newOutput(v)
	}
}

func (n *callNode) KeyPress(event KeyEvent) {
	if i, v := n.variadic(); v != nil {
		ins := ins(n)
		if event.Key == KeyComma {
			if n.ellipsis() {
				n.removePortBase(ins[i])
			}
			SetKeyFocus(n.newInput(newVar(v.Name, v.Type.(*types.Slice).Elem)))
			rearrange(n.blk)
		} else if event.Key == KeyPeriod && event.Ctrl {
			if n.ellipsis() {
				n.removePortBase(ins[i])
			} else {
				for _, in := range ins[i:] {
					n.removePortBase(in)
				}
				in := n.newInput(v)
				in.valView.setVariadic()
				SetKeyFocus(in)
				rearrange(n.blk)
			}
		} else {
			n.ViewBase.KeyPress(event)
		}
	} else {
		n.ViewBase.KeyPress(event)
	}
}

func (n *callNode) removePort(p *port) {
	if i, v := n.variadic(); v != nil {
		for _, p2 := range ins(n)[i:] {
			if p2 == p {
				n.removePortBase(p)
				break
			}
		}
	}
}

// returns index of first variadic port and its var
func (n *callNode) variadic() (int, *types.Var) {
	var sig *types.Signature
	if n.obj != nil {
		sig = n.obj.GetType().(*types.Signature)
	} else if in := ins(n)[0]; len(in.conns) > 0 {
		t, _ := indirect(in.conns[0].src.obj.Type)
		sig = t.(*types.Signature)
	}
	if sig == nil || !sig.IsVariadic {
		return -1, nil
	}
	i := len(sig.Params) - 1
	v := sig.Params[i]
	if sig.Recv != nil {
		i++
	}
	return i, v
}

func (n *callNode) ellipsis() bool {
	i, v := n.variadic()
	ins := ins(n)
	return v != nil && i == len(ins)-1 && ins[i].obj == v
}
