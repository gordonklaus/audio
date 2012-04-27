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
	
	currentInfo flux.Info
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
	
	n.currentInfo = flux.GetPackageInfo()
	n.activeIndices = []int{}
	for i := range n.currentInfo.Children() { n.activeIndices = append(n.activeIndices, i) }
	
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
	for info := n.currentInfo; info != nil; info = info.Parent() { pathStr = path.Join(info.Name(), pathStr) }
	if len(n.currentInfo.Children()) > 0 && len(pathStr) > 0 { pathStr += "/" }
	n.pathText.SetText(pathStr)
	xOffset := n.pathText.Width()
	
	currentIndex := n.activeIndices[n.currentActiveIndex]
	n.activeIndices = []int{}
	for i, child := range n.currentInfo.Children() {
		if HasPrefix(ToLower(child.Name()), ToLower(n.text.GetText())) {
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
	for i, activeIndex := range n.activeIndices {
		child := n.currentInfo.Children()[activeIndex]
		l := NewText(child.Name())
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
	gl.Rectd(0, lower, right, upper)
}

type nodeNameText struct {
	Text
	n *NodeCreator
}
func newNodeNameText(n *NodeCreator) *nodeNameText {
	t := &nodeNameText{}
	t.Text = *NewTextBase(t, "")
	t.n = n
	t.SetValidator(func(text *string) bool {
		for _, child := range n.currentInfo.Children() {
			if HasPrefix(ToLower(child.Name()), ToLower(*text)) {
				*text = child.Name()[:len(*text)]
				return true
			}
		}
		return false
	})
	return t
}
func (t *nodeNameText) LostKeyboardFocus() { t.n.Close() }
func (t *nodeNameText) KeyPressed(event KeyEvent) {
	n := t.n
	switch event.Key {
	case glfw.KeyUp:
		n.currentActiveIndex++
		n.indexChanged()
	case glfw.KeyDown:
		n.currentActiveIndex--
		n.indexChanged()
	case glfw.KeyBackspace:
		if len(t.GetText()) > 0 {
			t.Text.KeyPressed(event)
			break
		}
		fallthrough
	case glfw.KeyLeft:
		if parent := n.currentInfo.Parent(); parent != nil {
			previous := n.currentInfo
			n.currentInfo = parent
			n.activeIndices = []int{}
			for i, child := range parent.Children() {
				n.activeIndices = append(n.activeIndices, i)
				if child == previous { n.currentActiveIndex = i; break }
			}
			t.SetText("")
		}
	case glfw.KeyEnter:
		fallthrough
	case glfw.KeyRight:
		info := n.currentInfo.Children()[n.activeIndices[n.currentActiveIndex]]
		if packageInfo, ok := info.(*flux.PackageInfo); ok {
			packageInfo.Load()
		}
		if len(info.Children()) > 0 {
			n.currentInfo = info
			n.currentActiveIndex = 0
			t.SetText("")
		}
	case glfw.KeyEsc:
		n.Close()
	default:
		t.Text.KeyPressed(event)
	}
}
