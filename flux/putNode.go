package main

import (
	."code.google.com/p/gordon-go/gui"
)

type InputNode struct {
	*NodeBase
}

func NewInputNode(block *Block) *InputNode {
	n := &InputNode{}
	n.NodeBase = NewNodeBase(n, block)
	n.reform()
	return n
}

func (n *InputNode) KeyPressed(event KeyEvent) {
	if event.Text == "," {
		output := NewOutput(n, new(ValueInfo))
		n.AddChild(output)
		n.outputs = append(n.outputs, output)
		n.reform()
		output.valueView.Show()
		output.valueView.edit(func() {
			if output.info.typ != nil {
				f := n.block.Outermost().function
				f.info.typ.parameters = append(f.info.typ.parameters, output.info)
				f.AddPackageRef(output.info.typ)
				output.TakeKeyboardFocus()
			} else {
				n.RemoveChild(output)
				n.outputs = n.outputs[:len(n.outputs) - 1]
				n.reform()
				n.TakeKeyboardFocus()
			}
		})
	} else {
		n.NodeBase.KeyPressed(event)
	}
}

func (n InputNode) Paint() {
	color := map[bool]Color{true:{.5, .5, 1, .5}, false:{1, 1, 1, .25}}[n.focused]
	for f := 1.0; f > .1; f /= 2 {
		SetColor(color)
		SetPointSize(f * 25)
		DrawPoint(ZP)
	}
}
