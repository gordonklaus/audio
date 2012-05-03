package main

import (
	gl "github.com/chsc/gogl/gl21"
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/util"
	."code.google.com/p/gordon-go/gui"
	."strings"
)

type NodeCreator struct {
	ViewBase
	function *Function
	Created *Signal
	
	currentInfo Info
	activeIndices []int
	currentActiveIndex int
	
	pathTexts []*Text
	nameTexts []*Text
	text *nodeNameText
}

func NewNodeCreator(function *Function) *NodeCreator {
	n := &NodeCreator{}
	n.ViewBase = *NewView(n)
	n.function = function
	function.AddChild(n)
	n.Created = NewSignal()
	
	n.currentInfo = GetPackageInfo()
	n.activeIndices = []int{}
	for i := range n.currentInfo.Children() { n.activeIndices = append(n.activeIndices, i) }
	
	n.text = newNodeNameText(n)
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
	n.AddChild(n.text)
	n.text.TextChanged.Connect(func(...interface{}) { n.update() })
	n.text.SetText("")
	n.text.TakeKeyboardFocus()
	
	return n
}

func getTextColor(info Info, alpha float32) Color {
	switch info.(type) {
	case *PackageInfo:
		return Color{1, 1, 1, alpha}
	case *TypeInfo:
		return Color{.6, 1, .6, alpha}
	case FunctionInfo:
		return Color{1, .6, .6, alpha}
	case ValueInfo:
		return Color{.6, .6, 1, alpha}
	}
	return Color{}
}

func (n NodeCreator) lastPathText() (*Text, bool) {
	if np := len(n.pathTexts); np > 0 {
		return n.pathTexts[np - 1], true
	}
	return nil, false
}

func (n NodeCreator) currentActiveInfo() Info { return n.currentInfo.Children()[n.activeIndices[n.currentActiveIndex]] }

func (n *NodeCreator) update() {
	n.currentActiveIndex %= len(n.activeIndices)
	if n.currentActiveIndex < 0 { n.currentActiveIndex += len(n.activeIndices) }
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
	
	if t, ok := n.lastPathText(); ok {
		sep := ""; if _, ok := n.currentActiveInfo().(*PackageInfo); ok { sep = "/" } else { sep = "." }
		text := t.GetText()
		t.SetText(text[:len(text) - 1] + sep)
	}
	xOffset := 0.0; if t, ok := n.lastPathText(); ok { xOffset = t.Position().X + t.Width() }
	
	for _, l := range n.nameTexts { l.Close() }
	n.nameTexts = []*Text{}
	width := 0.0
	for i, activeIndex := range n.activeIndices {
		child := n.currentInfo.Children()[activeIndex]
		l := NewText(child.Name())
		l.SetTextColor(getTextColor(child, .7))
		l.SetBackgroundColor(Color{0, 0, 0, .7})
		n.AddChild(l)
		n.nameTexts = append(n.nameTexts, l)
		l.Move(Pt(xOffset, float64(len(n.activeIndices) - i - 1)*l.Height()))
		if l.Width() > width { width = l.Width() }
	}
	n.text.Raise()
	n.Resize(xOffset + width, float64(len(n.nameTexts))*n.nameTexts[0].Height())
	
	yOffset := float64(len(n.activeIndices) - n.currentActiveIndex - 1)*n.text.Height()
	n.text.Move(Pt(xOffset, yOffset))
	n.text.SetTextColor(getTextColor(n.currentActiveInfo(), 1))
	for _, p := range n.pathTexts { p.Move(Pt(p.Position().X, yOffset)) }
	
	n.Pan(Pt(0, yOffset))
}

func (n *NodeCreator) Paint() {
	cur := n.nameTexts[n.currentActiveIndex]
	right := gl.Double(cur.Position().X + cur.Width())
	lower := gl.Double(cur.Position().Y)
	upper := gl.Double(cur.Position().Y + cur.Height())
	gl.Color4d(1, 1, 1, .7)
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
		n.currentActiveIndex--
		n.update()
	case glfw.KeyDown:
		n.currentActiveIndex++
		n.update()
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
			
			length := len(n.pathTexts)
			n.pathTexts[length - 1].Close()
			n.pathTexts = n.pathTexts[:length - 1]
			
			t.SetText("")
		}
	case glfw.KeyEnter:
		if creator, ok := n.currentActiveInfo().(interface{NewNode()*Node}); ok {
			n.Close()
			n.Created.Emit(creator.NewNode())
			return
		}
		fallthrough
	case glfw.KeyRight:
		if info := n.currentActiveInfo(); len(info.Children()) > 0 {
			n.currentInfo = info
			n.activeIndices[0], n.currentActiveIndex = 0, 0
			
			pathText := NewText(info.Name() + "/")
			pathText.SetTextColor(getTextColor(info, 1))
			pathText.SetBackgroundColor(Color{0, 0, 0, .7})
			n.AddChild(pathText)
			x := 0.0; if t, ok := n.lastPathText(); ok { x = t.Position().X + t.Width() }
			pathText.Move(Pt(x, 0))
			n.pathTexts = append(n.pathTexts, pathText)
			
			t.SetText("")
		}
	case glfw.KeyEsc:
		n.Close()
	default:
		t.Text.KeyPressed(event)
	}
}
