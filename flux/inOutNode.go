package main

import (
	."code.google.com/p/gordon-go/gui"
)

type InOutNode struct {
	*NodeBase
	out bool
	editable bool
}

func NewInputNode(block *Block) *InOutNode {
	n := newInOutNode(block)
	n.out = false
	return n
}
func NewOutputNode(block *Block) *InOutNode {
	n := newInOutNode(block)
	n.out = true
	return n
}
func newInOutNode(block *Block) *InOutNode {
	n := &InOutNode{}
	n.NodeBase = NewNodeBase(n, block)
	n.reform()
	return n
}

func (n *InOutNode) KeyPressed(event KeyEvent) {
	if n.editable && event.Text == "," {
		var p *port
		if n.out {
			p = n.newInput(&ValueInfo{}).port
		} else {
			p = n.newOutput(&ValueInfo{}).port
		}
		p.valueView.Show()
		p.valueView.edit(func() {
			if p.info.typ != nil {
				f := n.block.Func()
				if n.out {
					f.info.typ.results = append(f.info.typ.results, p.info)
				} else {
					f.info.typ.parameters = append(f.info.typ.parameters, p.info)
				}
				f.AddPackageRef(p.info.typ)
				p.TakeKeyboardFocus()
			} else {
				n.RemoveChild(p.Self)
				n.TakeKeyboardFocus()
			}
		})
	} else {
		n.NodeBase.KeyPressed(event)
	}
}

func (n InOutNode) Paint() {
	SetColor(map[bool]Color{true:{.5, .5, 1, .5}, false:{1, 1, 1, .25}}[n.focused])
	// TODO:  draw half-circle instead
	for f := 1.0; f > .1; f /= 2 {
		SetPointSize(f * 12)
		DrawPoint(ZP)
	}
	n.NodeBase.Paint()
}
