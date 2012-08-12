package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/util"
	."code.google.com/p/gordon-go/gui"
	."fmt"
	"go/build"
	."io/ioutil"
	."strings"
	"os"
)

type NodeCreator struct {
	*ViewBase
	fluxSourceOnly bool
	finished bool
	created *Signal
	canceled *Signal
	
	currentInfo Info
	activeIndices []int
	currentActiveIndex int
	newInfo Info
	
	pathTexts []*Text
	nameTexts []*Text
	text *nodeNameText
}

type creatorMode int
const (
	browse = iota
	newFunction
)

func NewNodeCreator(fluxSourceOnly bool) *NodeCreator {
	n := &NodeCreator{created:NewSignal(), canceled:NewSignal()}
	n.ViewBase = NewView(n)
	
	n.fluxSourceOnly = fluxSourceOnly
	n.currentInfo = GetPackageInfo()
	n.activeIndices = []int{}
	
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
	case *ValueInfo:
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

func (n NodeCreator) filteredInfos() (infos []Info) {
	if !n.fluxSourceOnly { return n.currentInfo.Children() }
	for _, child := range n.currentInfo.Children() {
		if _, err := os.Stat(child.FluxSourcePath()); err != nil { continue }
		infos = append(infos, child)
	}
	return
}

func (n NodeCreator) currentActiveInfo() Info {
	if len(n.activeIndices) == 0 || len(n.filteredInfos()) == 0 { return nil }
	return n.filteredInfos()[n.activeIndices[n.currentActiveIndex]]
}

func (n *NodeCreator) update() {
	currentIndex := 0
	if len(n.activeIndices) > 0 {
		n.currentActiveIndex %= len(n.activeIndices)
		if n.currentActiveIndex < 0 { n.currentActiveIndex += len(n.activeIndices) }
		currentIndex = n.activeIndices[n.currentActiveIndex]
	}
	
	infos := n.filteredInfos()
	if n.newInfo != nil {
		n.newInfo.SetName(n.text.GetText())
		newIndex := 0
		for i, child := range infos {
			if child.Name() >= n.newInfo.Name() {
				switch child.(type) {
				case *PackageInfo: if _, ok := n.newInfo.(*PackageInfo); !ok { continue }
				case *FunctionInfo: if _, ok := n.newInfo.(*FunctionInfo); !ok { continue }
				default: continue
				}
				newIndex = i
				break
			}
		}
		infos = append(infos[:newIndex], append([]Info{n.newInfo}, infos[newIndex:]...)...)
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
	
	if t, ok := n.lastPathText(); ok && len(n.activeIndices) > 0 {
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
	if len(n.activeIndices) > 0 {
		n.text.SetTextColor(getTextColor(infos[n.activeIndices[n.currentActiveIndex]], 1))
	}
	for _, p := range n.pathTexts { p.Move(Pt(p.Position().X, yOffset)) }
	
	n.Pan(Pt(0, yOffset))
}

func (n *NodeCreator) Paint() {
	rect := ZR
	if n.newInfo == nil && len(n.nameTexts) > 0 {
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
		if n.newInfo != nil { return true }
		for _, info := range n.filteredInfos() {
			if HasPrefix(ToLower(info.Name()), ToLower(*text)) {
				*text = info.Name()[:len(*text)]
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
		if n.newInfo == nil {
			n.currentActiveIndex--
			n.update()
		}
	case KeyDown:
		if n.newInfo == nil {
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
		if parent := n.currentInfo.Parent(); n.newInfo == nil && parent != nil {
			previous := n.currentInfo
			n.currentInfo = parent
			n.currentActiveIndex = 0
			for i, info := range n.filteredInfos() {
				if info == previous { n.activeIndices = []int{i}; break }
			}
			
			length := len(n.pathTexts)
			n.pathTexts[length - 1].Close()
			n.pathTexts = n.pathTexts[:length - 1]
			
			t.SetText("")
		}
	case KeyEnter:
		info := n.newInfo
		existing := false
		if info == nil {
			info = n.currentActiveInfo()
		} else if n.currentActiveInfo() != nil && info.Name() == n.currentActiveInfo().Name() {
			info = n.currentActiveInfo()
			existing = true
		}
		if n.newInfo != nil && !existing {
			n.currentInfo.AddChild(info)
			switch info := info.(type) {
			case *PackageInfo:
				srcDirs := build.Default.SrcDirs()
				info.buildPackage.Dir = Sprintf("%v/%v", srcDirs[len(srcDirs) - 1], info.name)
				if err := os.Mkdir(info.FluxSourcePath(), 0755); err != nil { Println(err) }
			case TypeInfo:
				if err := WriteFile(info.FluxSourcePath(), []byte("type"), 0644); err != nil { Println(err) }
			case *FunctionInfo:
				NewFunction(info)
			}
			
			n.currentActiveIndex = 0
			for i, child := range n.filteredInfos() {
				if child == info { n.activeIndices = []int{i}; break }
			}
		}
		n.newInfo = nil
		if _, ok := info.(*PackageInfo); !ok {
			n.Close()
			n.finished = true
			n.created.Emit(info)
			return
		}
		fallthrough
	case KeyRight:
		if n.newInfo == nil {
			switch info := n.currentActiveInfo().(type) {
			case *PackageInfo, TypeInfo:
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
		}
	case KeyEsc:
		if n.newInfo == nil {
			n.Cancel()
		} else {
			n.newInfo = nil
			t.SetText("")
		}
	default:
		if n.newInfo == nil && event.Ctrl {
			if _, ok := n.currentInfo.(*PackageInfo); !ok {
				if _, ok := n.currentInfo.(TypeInfo); !(ok && event.Text == "3") {
					t.Text.KeyPressed(event)
					return
				}
			}
			switch event.Text {
			case "1": n.newInfo = &PackageInfo{}
			case "2": n.newInfo = &TypeInfoBase{}
			case "3": n.newInfo = &FunctionInfo{}
			case "4": n.newInfo = &ValueInfo{}
			case "5": n.newInfo = &ValueInfo{constant:true}
			default: t.Text.KeyPressed(event); return
			}
			t.SetText("")
		} else {
			t.Text.KeyPressed(event)
		}
	}
}
