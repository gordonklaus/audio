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
	n.text.SetTextColor(color(&types.Func{}, true, false))
	n.addSeqPorts()
	in := n.newInput(newVar("", nil))
	out := n.newOutput(newVar("", nil))
	in.connsChanged = func() {
		t := inputType(in)
		in.setType(t)
		if t == nil {
			for _, p := range ins(n)[1:] {
				n.removePortBase(p)
			}
		} else if n.ellipsis() {
			p := ins(n)[1]
			p.valView.setEllipsis()
			p.setType(t)
		} else {
			for _, p := range ins(n)[1:] {
				p.setType(underlying(t).(*types.Slice).Elem)
			}
		}
		out.setType(t)
	}
	return n
}

func (n *appendNode) connectable(t types.Type, dst *port) bool {
	if dst == ins(n)[0] {
		_, ok := underlying(t).(*types.Slice)
		return ok
	}
	return assignable(t, dst.obj.Type)
}

func (n *appendNode) KeyPress(event KeyEvent) {
	ins := ins(n)
	v := ins[0].obj
	t, ok := v.Type.(*types.Slice)
	if ok && event.Text == "," {
		if n.ellipsis() {
			n.removePortBase(ins[1])
		}
		SetKeyFocus(n.newInput(newVar("", t.Elem)))
	} else if ok && event.Key == KeyPeriod && event.Ctrl {
		if n.ellipsis() {
			n.removePortBase(ins[1])
		} else {
			for _, p := range ins[1:] {
				n.removePortBase(p)
			}
			p := n.newInput(v)
			p.valView.setEllipsis()
			SetKeyFocus(p)
		}
	} else {
		n.ViewBase.KeyPress(event)
	}
}

func (n *appendNode) removePort(p *port) {
	for _, p2 := range ins(n)[1:] {
		if p2 == p {
			n.removePortBase(p)
			break
		}
	}
}

func (n *appendNode) ellipsis() bool {
	ins := ins(n)
	return len(ins) == 2 && ins[1].obj == ins[0].obj
}

type complexNode struct {
	*nodeBase
	re, im, out *port
}

func newComplexNode() *complexNode {
	n := &complexNode{}
	n.nodeBase = newNodeBase(n)
	n.text.SetText("complex")
	n.text.SetTextColor(color(&types.Func{}, true, false))
	n.re = n.newInput(newVar("real", nil))
	n.re.connsChanged = n.connsChanged
	n.im = n.newInput(newVar("imag", nil))
	n.im.connsChanged = n.connsChanged
	n.out = n.newOutput(newVar("", nil))
	return n
}

func (n *complexNode) connectable(t types.Type, dst *port) bool {
	b, ok := underlying(t).(*types.Basic)
	return ok && b.Info&types.IsFloat != 0 && assignableToAll(t, n.ins...)
}

func (n *complexNode) connsChanged() {
	t := untypedToTyped(inputType(n.ins...))
	n.ins[0].setType(t)
	n.ins[1].setType(t)
	if t != nil {
		if underlying(t).(*types.Basic).Kind == types.Float32 {
			t = types.Typ[types.Complex64]
		} else {
			t = types.Typ[types.Complex128]
		}
	}
	n.outs[0].setType(t)
}

type deleteNode struct {
	*nodeBase
}

func newDeleteNode(godefer string) *deleteNode {
	n := &deleteNode{}
	n.nodeBase = newGoDeferNodeBase(n, godefer)
	n.text.SetText("delete")
	n.text.SetTextColor(color(&types.Func{}, true, false))
	m := n.newInput(newVar("map", nil))
	key := n.newInput(newVar("key", nil))
	m.connsChanged = func() {
		t := inputType(m)
		m.setType(t)
		if t != nil {
			key.setType(underlying(t).(*types.Map).Key)
		} else {
			key.setType(nil)
		}
	}
	n.addSeqPorts()
	return n
}

func (n *deleteNode) connectable(t types.Type, dst *port) bool {
	if dst == ins(n)[0] {
		_, ok := underlying(t).(*types.Map)
		return ok
	}
	return assignable(t, dst.obj.Type)
}

type lenCapNode struct {
	*nodeBase
	name string
}

func newLenCapNode(name string) *lenCapNode {
	n := &lenCapNode{name: name}
	n.nodeBase = newNodeBase(n)
	n.text.SetText(name)
	n.text.SetTextColor(color(&types.Func{}, true, false))
	in := n.newInput(newVar("", nil))
	len := n.newOutput(newVar("", nil))
	in.connsChanged = func() {
		t := inputType(in)
		in.setType(t)
		if t != nil {
			len.setType(types.Typ[types.Int])
		} else {
			len.setType(nil)
		}
	}
	n.addSeqPorts()
	return n
}

func (n *lenCapNode) connectable(t types.Type, dst *port) bool {
	ok := false
	switch t := underlying(t).(type) {
	case *types.Basic:
		ok = n.name == "len" && t.Info&types.IsString != 0
	case *types.Map:
		ok = n.name == "len"
	case *types.Array, *types.Slice, *types.Chan:
		ok = true
	case *types.Pointer:
		_, ok = underlying(t.Elem).(*types.Array)
	}
	return ok
}

type makeNode struct {
	*nodeBase
}

func newMakeNode(currentPkg *types.Package) *makeNode {
	n := &makeNode{}
	n.nodeBase = newNodeBase(n)
	n.text.SetText("make")
	n.text.SetTextColor(color(&types.Func{}, true, false))
	out := n.newOutput(nil)
	n.typ = newTypeView(&out.obj.Type, currentPkg)
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
		n.reform()
		SetKeyFocus(n)
	}
}

type newNode struct {
	*nodeBase
}

func newNewNode(currentPkg *types.Package) *newNode {
	n := &newNode{}
	n.nodeBase = newNodeBase(n)
	n.text.SetText("new")
	n.text.SetTextColor(color(&types.Func{}, true, false))
	t := &types.Pointer{}
	n.newOutput(newVar("", t))
	n.typ = newTypeView(&t.Elem, currentPkg)
	n.Add(n.typ)
	return n
}

func (n *newNode) editType() {
	n.typ.editType(func() {
		if t := *n.typ.typ; t != nil {
			n.setType(t)
		} else {
			n.blk.removeNode(n)
			SetKeyFocus(n.blk)
		}
	})
}

func (n *newNode) setType(t types.Type) {
	n.typ.setType(t)
	n.outs[0].setType(n.outs[0].obj.Type)
	if t != nil {
		n.blk.func_().addPkgRef(t)
		n.reform()
		SetKeyFocus(n)
	}
}

type panicRecoverNode struct {
	*nodeBase
	name string
}

func newPanicRecoverNode(name string, godefer string) *panicRecoverNode {
	n := &panicRecoverNode{name: name}
	n.nodeBase = newGoDeferNodeBase(n, godefer)
	n.text.SetText(name)
	n.text.SetTextColor(color(&types.Func{}, true, false))
	t := &types.Interface{}
	if name == "panic" {
		n.newInput(newVar("", t))
	} else {
		n.newOutput(newVar("", t))
	}
	n.addSeqPorts()
	return n
}

type realImagNode struct {
	*nodeBase
	name string
}

func newRealImagNode(name string) *realImagNode {
	n := &realImagNode{name: name}
	n.nodeBase = newNodeBase(n)
	n.text.SetText(name)
	n.text.SetTextColor(color(&types.Func{}, true, false))
	in := n.newInput(newVar("", nil))
	out := n.newOutput(newVar("", nil))
	in.connsChanged = func() {
		t := untypedToTyped(inputType(in))
		in.setType(t)
		if t != nil {
			if underlying(t).(*types.Basic).Kind == types.Complex64 {
				t = types.Typ[types.Float32]
			} else {
				t = types.Typ[types.Float64]
			}
		}
		out.setType(t)
	}
	return n
}

func (n *realImagNode) connectable(t types.Type, dst *port) bool {
	b, ok := underlying(t).(*types.Basic)
	return ok && b.Info&types.IsComplex != 0
}

func newVar(name string, typ types.Type) *types.Var {
	return types.NewVar(0, nil, name, typ)
}
