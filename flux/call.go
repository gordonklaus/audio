package main

import (
	"code.google.com/p/go.exp/go/types"
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
			n.addPorts(sig)
			n.addSeqPorts()
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
		in := n.newInput(&types.Var{})
		in.connsChanged = func() {
			for _, p := range append(ins(n), outs(n)...) {
				if p != in {
					n.removePortBase(p)
				}
			}
			if len(in.conns) > 0 {
				t, _ := indirect(in.conns[0].src.obj.Type)
				if sig, ok := t.(*types.Signature); ok {
					in.setType(sig)
					n.addPorts(sig)
				}
			} else {
				in.setType(nil)
			}
		}
		n.addSeqPorts()
		return n
	}
}

func (n *callNode) addPorts(sig *types.Signature) {
	if sig.Recv != nil {
		n.newInput(sig.Recv)
	}
	for _, v := range sig.Params {
		n.newInput(v)
	}
	for _, v := range sig.Results {
		n.newOutput(v)
	}
}
