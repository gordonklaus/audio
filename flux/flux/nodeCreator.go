package main

import (
	// ."github.com/jteeuwen/glfw"
	// gl "github.com/chsc/gogl/gl21"
	// ."code.google.com/p/gordon-go/ftgl"
	// ."code.google.com/p/gordon-go/util"
	."code.google.com/p/gordon-go/gui"
	."image"
)

type NodeCreator struct {
	ViewBase
	function *Function
	
	names []string
	
	nameLabels []*Label
	nameBox *TextBox
}

func NewNodeCreator(function *Function) *NodeCreator {
	n := &NodeCreator{}
	n.ViewBase = *NewView(n)
	n.function = function
	function.AddChild(n)
	
	n.names = []string{"audio", "aux", "auto", "bufio", "builtin", "compress", "container", "flag", "fmt", "font", "hash", "html", "index", "io", "reflect", "regexp", "strconv", "strings", "sync", "syscall", "unicode", "unsafe"}
	
	n.nameLabels = []*Label{}
	
	n.nameBox = NewTextBox()
	n.AddChild(n.nameBox)
	// n.nameBox.TextChanged.Connect(func(...interface{}) { n.Resize(n.nameBox.Width(), n.nameBox.Height()) })
	n.nameBox.SetText("")
	n.nameBox.TakeKeyboardFocus()
	
	n.update()
	return n
}

func (n *NodeCreator) update() {
	for _, l := range n.nameLabels {
		n.RemoveChild(l)
	}
	n.nameLabels = []*Label{}
	width := 0
	for i, name := range n.names {
		l := NewLabel(name)
		if l.Width() > width { width = l.Width() }
		n.AddChild(l)
		n.nameLabels = append(n.nameLabels, l)
		l.Move(Pt(0, i*l.Height()))
	}
	height := len(n.nameLabels)
	if height > 0 {
		height *= n.nameLabels[0].Height()
	} else {
		width, height = n.nameBox.Width(), n.nameBox.Height()
	}
	
	center := n.Center()
	n.Resize(width, height)
	n.MoveCenter(center)
}