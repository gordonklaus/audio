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

func newCallNode(obj types.Object, currentPkg *types.Package, godefer string) node {
	if obj == nil {
		n := &callNode{}
		n.nodeBase = newGoDeferNodeBase(n, godefer)
		if godefer == "" {
			n.text.SetText("call")
		} else {
			n.text.SetText("") //trigger TextChanged
		}
		n.text.SetTextColor(color(special{}, true, false))
		n.addSeqPorts()
		in := n.newInput(nil)
		in.connsChanged = func() {
			t := inputType(in)
			in.setType(t)
			// TODO: add/remove/modify only the affected ports.  beware: t == in.obj.Type because portsNode mutates the signature in place.
			for _, p := range append(ins(n)[1:], outs(n)...) {
				n.removePortBase(p)
			}
			if t != nil {
				n.addPorts(underlying(t).(*types.Signature))
			}
		}
		return n
	}

	if sig, ok := obj.GetType().(*types.Signature); ok {
		n := &callNode{obj: obj}
		n.nodeBase = newGoDeferNodeBase(n, godefer)
		name := obj.GetName()
		if sig.Recv != nil {
			name = "." + name
		} else if p := obj.GetPkg(); p != currentPkg && p != nil {
			n.pkg.setPkg(obj.GetPkg())
		}
		n.text.SetText(name)
		n.text.SetTextColor(color(&types.Func{}, true, false))
		n.addSeqPorts()
		n.addPorts(sig)
		return n
	}

	switch name := obj.GetName(); name {
	case "append":
		return newAppendNode()
	case "close":
		return newCloseNode(godefer)
	case "complex":
		return newComplexNode()
	case "copy":
		return newCopyNode(godefer)
	case "delete":
		return newDeleteNode(godefer)
	case "len", "cap":
		return newLenCapNode(name)
	case "make":
		return newMakeNode(currentPkg)
	case "new":
		return newNewNode(currentPkg)
	case "real", "imag":
		return newRealImagNode(name)
	default:
		panic("unknown builtin: " + name)
	}
}

func (n *callNode) connectable(t types.Type, dst *port) bool {
	f := ins(n)[0]
	if n.obj == nil && dst == f {
		_, ok := underlying(t).(*types.Signature)
		return ok
	}
	if n.obj == nil && inputType(f) == nil {
		// A connection whose destination is being edited may currently be connected to f.  It is temporarily disconnected during the call to connectable, but inputs with dependent types are not updated, so we have to specifically check for this case here.
		return false
	}
	return assignable(t, dst.obj.Type)
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
		if event.Text == "," {
			if n.ellipsis() {
				n.removePortBase(ins[i])
			}
			SetKeyFocus(n.newInput(newVar(v.Name, v.Type.(*types.Slice).Elem)))
		} else if event.Key == KeyPeriod && event.Ctrl {
			if n.ellipsis() {
				n.removePortBase(ins[i])
			} else {
				for _, in := range ins[i:] {
					n.removePortBase(in)
				}
				in := n.newInput(v)
				in.valView.setEllipsis()
				SetKeyFocus(in)
			}
		} else {
			n.ViewBase.KeyPress(event)
		}
	} else {
		n.ViewBase.KeyPress(event)
	}
}

func (n *callNode) removePort(p *port) {
	if p.bad {
		n.removePortBase(p)
		return
	}
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
	} else if t := inputType(ins(n)[0]); t != nil {
		sig = underlying(t).(*types.Signature)
	}
	if sig == nil || !sig.IsVariadic {
		return -1, nil
	}
	i := len(sig.Params) - 1
	v := sig.Params[i]
	if n.obj == nil {
		i++ // for func sig input
	}
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

func (n *callNode) Paint() {
	n.nodeBase.Paint()
	if n.obj != nil && unknown(n.obj) {
		SetColor(Color{1, 0, 0, 1})
		SetLineWidth(3)
		r := RectInParent(n.text)
		DrawLine(r.Min, r.Max)
		DrawLine(Pt(r.Min.X, r.Max.Y), Pt(r.Max.X, r.Min.Y))
	}
}
