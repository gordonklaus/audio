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
		browser := NewBrowser(typesOnly)
		n.AddChild(browser)
		browser.Move(n.Center())
		browser.created.Connect(func(info ...interface{}) {
			typ := info[0].(Type)
			output := NewOutput(n, ValueInfo{typ:typ})
			n.AddChild(output)
			n.outputs = append(n.outputs, output)
			if t, ok := typ.(*NamedType); ok {
				n.block.Outermost().function.AddPackageRef(t.Parent().(*PackageInfo))
			}
			n.reform()
			output.TakeKeyboardFocus()
		})
		browser.canceled.Connect(func(...interface{}) { n.TakeKeyboardFocus() })
		browser.text.TakeKeyboardFocus()
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
