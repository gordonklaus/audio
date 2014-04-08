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
	n.text.SetTextColor(color(&types.Func{}, true, false))

	n.newInput(nil).connsChanged = n.connsChanged
	if n.op != "!" {
		n.newInput(nil).connsChanged = n.connsChanged
	}
	n.newOutput(nil)

	n.connsChanged()
	return n
}

func (n *operatorNode) connectable(t types.Type, dst *port) bool {
	if !n.connectable1(t, dst) {
		return false
	}
	if n.op != "<<" && n.op != ">>" {
		return assignableToAll(t, n.ins...)
	}
	return true
}

func (n *operatorNode) connectable1(t types.Type, dst *port) bool {
	if n.op == "==" || n.op == "!=" {
		switch underlying(t).(type) {
		case *types.Slice, *types.Map, *types.Signature:
			// these types are comparable only with nil
			other := n.ins[0]
			if other == dst {
				other = n.ins[1]
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
		if (n.op == "<<" || n.op == ">>") && dst == n.ins[1] {
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
		t := untypedToTyped(inputType(n.ins...))
		n.ins[0].setType(t)
		n.ins[1].setType(t)
		n.outs[0].setType(t)
	case "<<", ">>":
		t := untypedToTyped(inputType(n.ins[0]))
		u := untypedToTyped(inputType(n.ins[1]))
		n.ins[0].setType(t)
		n.ins[1].setType(u)
		n.outs[0].setType(t)
	case "==", "!=", "<", "<=", ">", ">=":
		t := untypedToTyped(inputType(n.ins...))
		n.ins[0].setType(t)
		n.ins[1].setType(t)
		if t != nil {
			n.outs[0].setType(types.Typ[types.UntypedBool])
		} else {
			n.outs[0].setType(nil)
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
