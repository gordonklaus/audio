package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/util"
	."code.google.com/p/gordon-go/gui"
	."strings"
)

type NodeCreator struct {
	*ViewBase
	mode creatorMode
	finished bool
	created *Signal
	canceled *Signal
	
	currentInfo Info
	activeIndices []int
	currentActiveIndex int
	
	pathTexts []*Text
	nameTexts []*Text
	text *nodeNameText
}

type creatorMode int
const (
	browse = iota
	newFunction
)

func NewNodeCreator() *NodeCreator {
	n := &NodeCreator{mode:browse, finished:false, created:NewSignal(), canceled:NewSignal()}
	n.ViewBase = NewView(n)
	
	n.currentInfo = GetPackageInfo()
	n.activeIndices = []int{}
	for i := range n.currentInfo.Children() { n.activeIndices = append(n.activeIndices, i) }
	
	n.text = newNodeNameText(n)
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
	n.AddChild(n.text)
	n.text.TextChanged.Connect(func(...interface{}) { n.update() })
	n.text.SetText("")
	
	return n
}

func (n *NodeCreator) Cancel() {
	if !n.finished {
		n.finished = true
		n.Close()
		n.canceled.Emit()
	}
}

func getTextColor(info Info, alpha float64) Color {
	switch info.(type) {
	case *PackageInfo:
		return Color{1, 1, 1, alpha}
	case TypeInfo:
		return Color{.6, 1, .6, alpha}
	case *FunctionInfo:
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
	currentIndex := 0
	if len(n.activeIndices) > 0 {
		n.currentActiveIndex %= len(n.activeIndices)
		if n.currentActiveIndex < 0 { n.currentActiveIndex += len(n.activeIndices) }
		currentIndex = n.activeIndices[n.currentActiveIndex]
	}
	
	infos := n.currentInfo.Children()
	if n.mode != browse {
		var newInfo Info
		switch n.mode {
		case newFunction: newInfo = &FunctionInfo{InfoBase:InfoBase{name:n.text.GetText()}}
		}
		newIndex := 0
		for i, child := range infos {
			if child.Name() >= n.text.GetText() {
				switch child.(type) {
				case *FunctionInfo: if n.mode != newFunction { continue }
				default: continue
				}
				newIndex = i
				break
			}
		}
		infos = append(infos[:newIndex], append([]Info{newInfo}, infos[newIndex:]...)...)
		currentIndex = newIndex
	}
	
	n.activeIndices = []int{}
	for i, child := range infos {
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
		sep := ""; if _, ok := infos[n.activeIndices[n.currentActiveIndex]].(*PackageInfo); ok { sep = "/" } else { sep = "." }
		text := t.GetText()
		t.SetText(text[:len(text) - 1] + sep)
	}
	xOffset := 0.0; if t, ok := n.lastPathText(); ok { xOffset = t.Position().X + t.Width() }
	
	for _, l := range n.nameTexts { l.Close() }
	n.nameTexts = []*Text{}
	width := 0.0
	for i, activeIndex := range n.activeIndices {
		child := infos[activeIndex]
		l := NewText(child.Name())
		l.SetTextColor(getTextColor(child, .7))
		l.SetBackgroundColor(Color{0, 0, 0, .7})
		n.AddChild(l)
		n.nameTexts = append(n.nameTexts, l)
		l.Move(Pt(xOffset, float64(len(n.activeIndices) - i - 1)*l.Height()))
		if l.Width() > width { width = l.Width() }
	}
	n.text.Raise()
	n.Resize(xOffset + width, float64(len(n.nameTexts))*n.text.Height())
	
	yOffset := float64(len(n.activeIndices) - n.currentActiveIndex - 1)*n.text.Height()
	n.text.Move(Pt(xOffset, yOffset))
	n.text.SetTextColor(getTextColor(infos[n.activeIndices[n.currentActiveIndex]], 1))
	for _, p := range n.pathTexts { p.Move(Pt(p.Position().X, yOffset)) }
	
	n.Pan(Pt(0, yOffset))
}

func (n *NodeCreator) Paint() {
	rect := ZR
	if n.mode == browse {
		cur := n.nameTexts[n.currentActiveIndex]
		rect = Rect(0, cur.Position().Y, cur.Position().X + cur.Width(), cur.Position().Y + cur.Height())
	} else {
		rect = n.text.MapRectToParent(n.text.Rect())
		rect.Min.X = 0
	}
	SetColor(Color{1, 1, 1, .7})
	FillRect(rect)
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
		if n.mode != browse { return true }
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
func (t *nodeNameText) LostKeyboardFocus() { t.n.Cancel() }
func (t *nodeNameText) KeyPressed(event KeyEvent) {
	n := t.n
	switch event.Key {
	case KeyUp:
		if n.mode == browse {
			n.currentActiveIndex--
			n.update()
		}
	case KeyDown:
		if n.mode == browse {
			n.currentActiveIndex++
			n.update()
		}
	case KeyBackspace:
		if len(t.GetText()) > 0 {
			t.Text.KeyPressed(event)
			break
		}
		fallthrough
	case KeyLeft:
		if parent := n.currentInfo.Parent(); n.mode == browse && parent != nil {
			previous := n.currentInfo
			n.currentInfo = parent
			n.currentActiveIndex = 0
			for i, child := range parent.Children() {
				if child == previous { n.activeIndices = []int{i}; break }
			}
			
			length := len(n.pathTexts)
			n.pathTexts[length - 1].Close()
			n.pathTexts = n.pathTexts[:length - 1]
			
			t.SetText("")
		}
	case KeyEnter:
		if info, ok := n.currentActiveInfo().(*FunctionInfo); ok {
			n.Close()
			n.finished = true
			n.created.Emit(info)
			return
		}
		fallthrough
	case KeyRight:
		if info := n.currentActiveInfo(); n.mode == browse && len(info.Children()) > 0 {
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
	case KeyEsc:
		if n.mode == browse {
			n.Cancel()
		} else {
			n.mode = browse
			t.SetText("")
		}
	default:
		if n.mode == browse && event.Text == "\\" {
			n.mode = newFunction
			t.SetText("")
		} else {
			t.Text.KeyPressed(event)
		}
	}
}
