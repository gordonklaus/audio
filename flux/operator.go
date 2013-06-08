package main

import (
	"code.google.com/p/go.exp/go/types"
)

var operators []*types.Func

func init() {
	add := func(s string, params, results []*types.Var) {
		operators = append(operators, &types.Func{Name: s, Type: &types.Signature{Params: params, Results: results}})
	}
	genericVar := func() *types.Var { return &types.Var{Type: generic{}} }
	for _, s := range ([]string{"+", "-", "*", "/", "%", "&", "|", "^", "&^", "&&", "||", "!"}) {
		add(s, []*types.Var{genericVar(), genericVar()}, []*types.Var{genericVar()})
	}
	untypedBool := []*types.Var{{Type: types.Typ[types.UntypedBool]}}
	add("==", []*types.Var{genericVar(), genericVar()}, untypedBool)
	add("!=", []*types.Var{genericVar(), genericVar()}, untypedBool)
}

func findOp(s string) *types.Func {
	for _, op := range operators {
		if op.Name == s {
			return op
		}
	}
	panic("unknown operator: " + s)
}

type operatorNode struct {
	*nodeBase
	op string
}

func newOperatorNode(obj types.Object) *operatorNode {
	n := &operatorNode{op: obj.GetName()}
	n.nodeBase = newNodeBase(n)
	n.text.SetText(n.op)
	
	t := obj.GetType().(*types.Signature)
	for _, v := range t.Params { n.newInput(v) }
	for _, v := range t.Results { n.newOutput(v) }
	
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
	case "<", "<=", ">", ">=":
		
	}
	return n
}
