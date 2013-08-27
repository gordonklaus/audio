package main

import (
	"code.google.com/p/go.exp/go/types"
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
	case "+", "-", "*", "/", "%", "&", "|", "^", "&^", "&&", "||", "!":
		f := func() {
			t := types.Type(generic{})
			for _, p := range ins(n) {
				if len(p.conns) > 0 {
					t = p.conns[0].src.obj.Type
					break
				}
			}
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
		
	case "==", "!=":
		f := func() {
			t := types.Type(generic{})
			for _, p := range ins(n) {
				if len(p.conns) > 0 {
					t = p.conns[0].src.obj.Type
					break
				}
			}
			for _, p := range ins(n) {
				p.setType(t)
			}
		}
		for _, p := range ins(n) {
			p.connsChanged = f
		}
		f()
		n.outs[0].setType(types.Typ[types.UntypedBool])
	case "<", "<=", ">", ">=":
		
	}
	return n
}
