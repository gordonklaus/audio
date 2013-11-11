package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type valueNode struct {
	*nodeBase
	obj types.Object // package var or struct field, or nil if this is an assign (=) or indirect node
	set bool
	x   *port // the target of the operation (struct or pointer)
	y   *port // the result of the read (output) or the argument to write (input)
}

func newValueNode(obj types.Object, set bool) *valueNode {
	n := &valueNode{obj: obj, set: set}
	n.nodeBase = newNodeBase(n)
	text := ""
	switch obj.(type) {
	case field, method, nil:
		n.x = n.newInput(&types.Var{})
		n.x.connsChanged = n.reform
		text = "."
	default:
	}
	if obj != nil {
		text += obj.GetName()
		n.text.SetText(text)
	}
	if set {
		n.y = n.newInput(&types.Var{})
	} else {
		n.y = n.newOutput(&types.Var{})
	}
	switch obj.(type) {
	case *types.Var, field:
		n.addSeqPorts()
	default:
	}
	n.reform()
	return n
}

func (n *valueNode) reform() {
	if n.set {
		if n.y.out {
			n.removePortBase(n.y)
			n.y = n.newInput(&types.Var{})
		}
	} else {
		if !n.y.out {
			n.removePortBase(n.y)
			n.y = n.newOutput(&types.Var{})
		}
	}
	var xt, yt types.Type
	if n.obj != nil {
		yt = n.obj.GetType()
	}
	switch obj := n.obj.(type) {
	case *types.Const:
	case *types.Var:
		if !n.set {
			yt = &types.Pointer{yt}
		}
	case *types.Func:
	case field:
		xt = obj.recv
		// TODO: use indirect result of types.LookupFieldOrMethod, or types.Selection.Indirect()
		if n.set {
			xt = &types.Pointer{xt}
		} else {
			if len(n.x.conns) > 0 {
				xt = n.x.conns[0].src.obj.Type
			}
			if _, ok := xt.(*types.Pointer); ok {
				yt = &types.Pointer{yt}
			}
		}
	case method:
		xt = obj.Type.Recv.Type
		// TODO: remove Recv? (from copy)
	case nil:
		if len(n.x.conns) > 0 {
			xt = n.x.conns[0].src.obj.Type
			yt, _ = indirect(xt)
		}
		if n.set {
			n.text.SetText("=")
		} else {
			n.text.SetText("indirect")
		}
	}
	if n.x != nil {
		n.x.setType(xt)
	}
	n.y.setType(yt)
}

func (n *valueNode) KeyPress(event KeyEvent) {
	if _, ok := n.obj.(*types.Const); ok || n.obj == nil {
		n.nodeBase.KeyPress(event)
	} else {
		switch event.Text {
		case "=":
			if n.x != nil {
				// TODO: use indirect result of types.LookupFieldOrMethod, or types.Selection.Indirect()
				if _, ok := n.x.obj.Type.(*types.Pointer); !ok {
					break
				}
			}
			n.set = !n.set
			n.reform()
			SetKeyFocus(n)
		default:
			n.nodeBase.KeyPress(event)
		}
	}
}
