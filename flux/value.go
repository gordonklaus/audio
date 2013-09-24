package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type valueNode struct {
	*nodeBase
	obj                 types.Object
	addr, indirect, set bool
	x                   *port // the value to be read from or written to (input)
	y                   *port // the result of the read (output) or the argument to write (input)
}

func newValueNode(obj types.Object, addr, indirect, set bool) *valueNode {
	n := &valueNode{obj: obj, addr: addr, indirect: indirect, set: set}
	n.nodeBase = newNodeBase(n)
	if f, ok := obj.(field); ok {
		n.x = n.newInput(&types.Var{Name: "x", Type: f.recv})
	} else if obj == nil {
		n.x = n.newInput(&types.Var{Name: "x"})
		n.x.connsChanged = n.reform
	}
	if set {
		n.y = n.newInput(&types.Var{})
	} else {
		n.y = n.newOutput(&types.Var{})
	}
	if _, ok := n.obj.(*types.Const); !ok {
		n.addSeqPorts()
	}
	n.reform()
	return n
}

func (n *valueNode) reform() {
	text := ""
	if n.addr {
		text = "&"
	} else if n.indirect {
		text = "*"
	}
	if _, ok := n.obj.(field); ok {
		text += "."
	}
	if n.obj != nil {
		text += n.obj.GetName()
	} else {
		text += "x"
	}
	n.text.SetText(text)

	if n.set {
		if n.y.out {
			n.removePortBase(n.y)
			n.y = n.newInput(&types.Var{})
			if n.x != nil {
				n.y.connsChanged = n.reform
			}
		}
	} else {
		if !n.y.out {
			n.y.connsChanged = func() {}
			n.removePortBase(n.y)
			n.y = n.newOutput(&types.Var{})
		}
	}
	var yt types.Type
	if n.obj != nil {
		yt = n.obj.GetType()
		if n.addr {
			yt = &types.Pointer{Base: yt}
		} else if n.indirect {
			yt, _ = indirect(yt)
		}
	} else {
		var xt types.Type
		if len(n.x.conns) > 0 {
			xt = n.x.conns[0].src.obj.Type
			yt = xt
			if n.addr {
				yt = &types.Pointer{Base: yt}
			} else if n.indirect {
				yt, _ = indirect(yt)
			}
		} else if n.set && len(n.y.conns) > 0 {
			yt = n.y.conns[0].src.obj.Type
			xt = yt
			if n.addr {
				xt, _ = indirect(xt)
			} else if n.indirect {
				xt = &types.Pointer{Base: xt}
			}
		}
		n.x.setType(xt)
	}
	n.y.setType(yt)
}

func (n *valueNode) KeyPress(event KeyEvent) {
	if _, ok := n.obj.(*types.Const); ok {
		n.nodeBase.KeyPress(event)
	} else {
		switch event.Text {
		case "&":
			if !n.set {
				n.addr = !n.addr
				n.indirect = false
				n.reform()
			}
		case "*":
			var t types.Type
			if n.obj != nil {
				t = n.obj.GetType()
			} else {
				t = n.x.obj.Type
			}
			if _, ok := indirect(t); ok || t == nil {
				n.addr = false
				n.indirect = !n.indirect
				n.reform()
			}
		case "=":
			if !n.addr {
				n.set = !n.set
				n.reform()
				SetKeyFocus(n)
			}
		default:
			n.nodeBase.KeyPress(event)
		}
	}
}
