package main

import (
	gl "github.com/chsc/gogl/gl21"
	"github.com/jteeuwen/glfw"
	// ."code.google.com/p/gordon-go/util"
	."code.google.com/p/gordon-go/gui"
	"code.google.com/p/gordon-go/flux"
	."image"
	."strings"
	"path"
)

type NodeCreator struct {
	ViewBase
	function *Function
	
	currentPackageInfo *flux.PackageInfo
	activeIndices []int
	currentActiveIndex int
	
	pathText *Text
	nameTexts []*Text
	text *nodeNameText
	offset Point
}

func NewNodeCreator(function *Function) *NodeCreator {
	n := &NodeCreator{}
	n.ViewBase = *NewView(n)
	n.function = function
	function.AddChild(n)
	
	n.currentPackageInfo = flux.GetPackageInfo()
	n.activeIndices = []int{}
	for i := range n.currentPackageInfo.SubPackages { n.activeIndices = append(n.activeIndices, i) }
	
	n.pathText = NewText("")
	n.pathText.SetBackgroundColor(Color{0, 0, 0, .7})
	n.AddChild(n.pathText)
	n.nameTexts = []*Text{}
	
	n.text = newNodeNameText(n)
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
	n.AddChild(n.text)
	n.text.TextChanged.Connect(func(...interface{}) { n.textChanged() })
	n.text.SetText("")
	n.text.TakeKeyboardFocus()
	
	return n
}

func (n *NodeCreator) textChanged() {
	pathStr := ""
	for packageInfo := n.currentPackageInfo; packageInfo != nil; packageInfo = packageInfo.Parent { pathStr = path.Join(packageInfo.Name, pathStr) }
	if len(n.currentPackageInfo.SubPackages) > 0 && len(pathStr) > 0 { pathStr += "/" }
	n.pathText.SetText(pathStr)
	xOffset := n.pathText.Width()
	
	currentIndex := n.activeIndices[n.currentActiveIndex]
	n.activeIndices = []int{}
	for i, info := range n.currentPackageInfo.SubPackages {
		if HasPrefix(ToLower(info.Name), ToLower(n.text.GetText())) {
			n.activeIndices = append(n.activeIndices, i)
		}
	}
	for i, index := range n.activeIndices {
		if index >= currentIndex {
			n.currentActiveIndex = i
			break
		}
	}
	if n.currentActiveIndex >= len(n.activeIndices) { n.currentActiveIndex = len(n.activeIndices) - 1 }
	
	for _, l := range n.nameTexts {
		n.RemoveChild(l)
	}
	n.nameTexts = []*Text{}
	width := 0
	for i, infoIndex := range n.activeIndices {
		subPackageInfo := n.currentPackageInfo.SubPackages[infoIndex]
		text := subPackageInfo.Name
		if len(subPackageInfo.SubPackages) > 0 { text += "->" }
		l := NewText(text)
		l.SetTextColor(Color{.7, .7, .7, 1})
		l.SetBackgroundColor(Color{0, 0, 0, .7})
		n.AddChild(l)
		n.nameTexts = append(n.nameTexts, l)
		l.Move(Pt(xOffset, i*l.Height()))
		if l.Width() > width { width = l.Width() }
	}
	n.text.Raise()
	height := len(n.nameTexts)
	if height > 0 {
		height *= n.nameTexts[0].Height()
	} else {
		width, height = n.text.Width(), n.text.Height()
	}
	n.Resize(xOffset + width, height)
	
	n.indexChanged()
}

func (n *NodeCreator) indexChanged() {
	n.currentActiveIndex %= len(n.activeIndices)
	if n.currentActiveIndex < 0 { n.currentActiveIndex += len(n.activeIndices) }
	xOffset := n.pathText.Width()
	n.text.Move(Pt(xOffset, n.currentActiveIndex*n.text.Height()))
	n.pathText.Move(Pt(0, n.currentActiveIndex*n.text.Height()))
	
	offset := Pt(0, -n.currentActiveIndex*n.nameTexts[0].Height())
	n.Move(n.Position().Sub(n.offset).Add(offset))
	n.offset = offset
}

func (n *NodeCreator) Paint() {
	cur := n.nameTexts[n.currentActiveIndex]
	right := gl.Double(cur.Position().X + cur.Width())
	lower := gl.Double(cur.Position().Y)
	upper := gl.Double(cur.Position().Y + cur.Height())
	gl.Color4d(.5, .75, 1, .9)
	gl.Begin(gl.LINE_LOOP)
	gl.Vertex2d(0, lower)
	gl.Vertex2d(right, lower)
	gl.Vertex2d(right, upper)
	gl.Vertex2d(0, upper)
	gl.End()
}

type nodeNameText struct {
	Text
	n *NodeCreator
}
func newNodeNameText(n *NodeCreator) *nodeNameText {
	t := &nodeNameText{}
	t.Text = *NewTextBase(t, "")
	t.n = n
	return t
}
func (t *nodeNameText) KeyPressed(event KeyEvent) {
	n := t.n
	switch event.Key {
	case glfw.KeyUp:
		n.currentActiveIndex++
		n.indexChanged()
	case glfw.KeyDown:
		n.currentActiveIndex--
		n.indexChanged()
	case glfw.KeyLeft:
		if parent := n.currentPackageInfo.Parent; parent != nil {
			previous := n.currentPackageInfo
			n.currentPackageInfo = parent
			n.activeIndices = []int{}
			for i, subPackage := range parent.SubPackages {
				n.activeIndices = append(n.activeIndices, i)
				if subPackage == previous { n.currentActiveIndex = i; break }
			}
			n.text.SetText("")
		}
	case glfw.KeyRight:
		if subPackage := n.currentPackageInfo.SubPackages[n.activeIndices[n.currentActiveIndex]]; len(subPackage.SubPackages) > 0 {
			n.currentPackageInfo = subPackage
			n.currentActiveIndex = 0
			n.text.SetText("")
		}
	default:
		t.Text.KeyPressed(event)
		return
	}
}
