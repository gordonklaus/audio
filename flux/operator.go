// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
)

type operatorNode struct {
	*nodeBase
	op string
}

func newOperatorNode(obj types.Object) *operatorNode {
	n := &operatorNode{op: obj.GetName()}
	n.nodeBase = newNodeBase(n)
	n.text.SetText(n.op)

	n.newInput(&types.Var{})
	n.newInput(&types.Var{})
	n.newOutput(&types.Var{})

	switch n.op {
	case "!":
		n.removePortBase(n.ins[1])
		fallthrough
	case "+", "-", "*", "/", "%", "&", "|", "^", "&^", "&&", "||":
		f := func() {
			t := combineInputTypes(ins(n))
			t = untypedToTyped(t)
			for _, p := range ins(n) {
				p.setType(t)
			}
			n.outs[0].setType(t)
		}
		for _, p := range ins(n) {
			p.connsChanged = f
		}
		f()
	case "<<", ">>":

	case "==", "!=", "<", "<=", ">", ">=":
		f := func() {
			t := combineInputTypes(ins(n))
			for _, p := range ins(n) {
				p.setType(t)
			}
		}
		for _, p := range ins(n) {
			p.connsChanged = f
		}
		f()
		n.outs[0].setType(types.Typ[types.UntypedBool])
	}
	return n
}

func combineInputTypes(p []*port) (t types.Type) {
	for _, p := range p {
		if len(p.conns) > 0 {
			t2, _ := indirect(p.conns[0].src.obj.Type)
			switch {
			case t == nil:
				t = t2
			case isUntyped(t) && isUntyped(t2):
				// TODO: combine untypeds
			case isUntyped(t):
				t = t2
			case isUntyped(t2):
			default:
			}
		}
	}
	return
}

func isUntyped(t types.Type) bool {
	b, ok := t.(*types.Basic)
	return ok && b.Info&types.IsUntyped != 0
}

func untypedToTyped(t types.Type) types.Type {
	b, ok := t.(*types.Basic)
	if !ok {
		return t
	}
	switch b.Kind {
	case types.UntypedBool:
		return types.Typ[types.Bool]
	case types.UntypedInt:
		return types.Typ[types.Int]
	case types.UntypedRune:
		return types.Typ[types.Rune]
	case types.UntypedFloat:
		return types.Typ[types.Float64]
	case types.UntypedComplex:
		return types.Typ[types.Complex128]
	case types.UntypedString:
		return types.Typ[types.String]
	default:
		return t
	}
}
