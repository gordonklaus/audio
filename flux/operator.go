// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	"unicode"
)

type operatorNode struct {
	*nodeBase
	op string
}

func newOperatorNode(obj types.Object) *operatorNode {
	n := &operatorNode{op: obj.GetName()}
	n.nodeBase = newNodeBase(n)
	n.text.SetText(n.op)

	n.newInput(nil)
	n.newInput(nil)
	n.newOutput(nil)

	switch n.op {
	case "!":
		n.removePortBase(n.ins[1])
	case "==", "!=", "<", "<=", ">", ">=":
		n.outs[0].setType(types.Typ[types.UntypedBool])
	}

	for _, p := range ins(n) {
		p.connsChanged = n.connsChanged
	}
	n.connsChanged()
	return n
}

func (n *operatorNode) connectable(t types.Type, dst *port) bool {
	if !n.connectable1(t, dst) {
		return false
	}
	if n.op != "<<" && n.op != ">>" {
		return assignableToAll(t, ins(n)...)
	}
	return true
}

func (n *operatorNode) connectable1(t types.Type, dst *port) bool {
	if n.op == "==" || n.op == "!=" {
		switch underlying(t).(type) {
		case *types.Slice, *types.Map, *types.Signature:
			// these types are comparable only with nil
			other := ins(n)[0]
			if other == dst {
				other = ins(n)[1]
			}
			return len(other.conns) == 0
		}
		return types.Comparable(t)
	}
	b, ok := underlying(t).(*types.Basic)
	if !ok {
		return false
	}
	i := b.Info
	switch n.op {
	case "!", "&&", "||":
		return i&types.IsBoolean != 0
	case "+":
		return i&(types.IsString|types.IsNumeric) != 0
	case "-", "*", "/":
		return i&types.IsNumeric != 0
	case "%", "&", "|", "^", "&^", "<<", ">>":
		if (n.op == "<<" || n.op == ">>") && dst == ins(n)[1] {
			return i&types.IsUnsigned != 0
		}
		return i&types.IsInteger != 0
	case "<", "<=", ">", ">=":
		return i&types.IsOrdered != 0
	}
	panic(n.op)
}

func (n *operatorNode) connsChanged() {
	switch n.op {
	case "!", "&&", "||", "+", "-", "*", "/", "%", "&", "|", "^", "&^":
		t := untypedToTyped(inputType(ins(n)...))
		for _, p := range ins(n) {
			p.setType(t)
		}
		n.outs[0].setType(t)
	case "<<", ">>":
		t := untypedToTyped(inputType(n.ins[0]))
		u := untypedToTyped(inputType(n.ins[1]))
		n.ins[0].setType(t)
		n.ins[1].setType(u)
		n.outs[0].setType(t)
	case "==", "!=", "<", "<=", ">", ">=":
		t := untypedToTyped(inputType(ins(n)...))
		for _, p := range ins(n) {
			p.setType(t)
		}
	}
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

func isOperator(obj types.Object) bool {
	name := obj.GetName()
	return len(name) > 0 && !unicode.IsLetter([]rune(name)[0])
}
