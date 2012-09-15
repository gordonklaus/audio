package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/util"
	."code.google.com/p/gordon-go/gui"
	."fmt"
	."io/ioutil"
	."strings"
	"os"
)

type Browser struct {
	*ViewBase
	mode browserMode
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

type browserMode int
const (
	browse = iota
	fluxSourceOnly
	typesOnly
)

func NewBrowser(mode browserMode) *Browser {
	b := &Browser{created:NewSignal(), canceled:NewSignal()}
	b.ViewBase = NewView(b)
	
	b.mode = mode
	b.currentInfo = rootPackageInfo
	rootPackageInfo.types = builtinPkg.types
	rootPackageInfo.functions = builtinPkg.functions
	rootPackageInfo.constants = builtinPkg.constants
	b.activeIndices = []int{}
	
	b.text = newNodeNameText(b)
	b.text.SetBackgroundColor(Color{0, 0, 0, 0})
	b.AddChild(b.text)
	b.text.TextChanged.Connect(func(...interface{}) { b.update() })
	b.text.SetText("")
	
	return b
}

func (b *Browser) Cancel() {
	if !b.finished {
		b.finished = true
		b.Close()
		b.canceled.Emit()
	}
}

func getTextColor(info Info, alpha float64) Color {
	switch info.(type) {
	case *PackageInfo:
		return Color{1, 1, 1, alpha}
	case *NamedType:
		return Color{.6, 1, .6, alpha}
	case *FuncInfo:
		return Color{1, .6, .6, alpha}
	case *ValueInfo:
		return Color{.6, .6, 1, alpha}
	}
	return Color{}
}

func (b Browser) lastPathText() (*Text, bool) {
	if np := len(b.pathTexts); np > 0 {
		return b.pathTexts[np - 1], true
	}
	return nil, false
}

func (b Browser) filteredInfos() (infos []Info) {
	switch b.mode {
	case browse:
		infos = b.currentInfo.Children()
	case fluxSourceOnly:
		for _, child := range b.currentInfo.Children() {
			if _, err := os.Stat(child.FluxSourcePath()); err == nil { infos = append(infos, child) }
		}
	case typesOnly:
		for _, child := range b.currentInfo.Children() {
			switch child.(type) { case *PackageInfo, *NamedType: infos = append(infos, child) }
		}
	}
	return
}

func (b Browser) currentActiveInfo() Info {
	if len(b.activeIndices) == 0 || len(b.filteredInfos()) == 0 { return nil }
	return b.filteredInfos()[b.activeIndices[b.currentActiveIndex]]
}

func (b *Browser) update() {
	currentIndex := 0
	if len(b.activeIndices) > 0 {
		b.currentActiveIndex %= len(b.activeIndices)
		if b.currentActiveIndex < 0 { b.currentActiveIndex += len(b.activeIndices) }
		currentIndex = b.activeIndices[b.currentActiveIndex]
	}
	
	infos := b.filteredInfos()
	if b.newInfo != nil {
		b.newInfo.SetName(b.text.GetText())
		newIndex := 0
		for i, child := range infos {
			if child.Name() >= b.newInfo.Name() {
				switch child.(type) {
				case *PackageInfo: if _, ok := b.newInfo.(*PackageInfo); !ok { continue }
				case *FuncInfo: if _, ok := b.newInfo.(*FuncInfo); !ok { continue }
				default: continue
				}
				newIndex = i
				break
			}
		}
		infos = append(infos[:newIndex], append([]Info{b.newInfo}, infos[newIndex:]...)...)
		currentIndex = newIndex
	}
	
	b.activeIndices = []int{}
	for i, child := range infos {
		if HasPrefix(ToLower(child.Name()), ToLower(b.text.GetText())) {
			b.activeIndices = append(b.activeIndices, i)
		}
	}
	for i, index := range b.activeIndices {
		if index >= currentIndex {
			b.currentActiveIndex = i
			break
		}
	}
	if b.currentActiveIndex >= len(b.activeIndices) { b.currentActiveIndex = len(b.activeIndices) - 1 }
	
	if t, ok := b.lastPathText(); ok && len(b.activeIndices) > 0 {
		sep := ""; if _, ok := infos[b.activeIndices[b.currentActiveIndex]].(*PackageInfo); ok { sep = "/" } else { sep = "." }
		text := t.GetText()
		t.SetText(text[:len(text) - 1] + sep)
	}
	xOffset := 0.0; if t, ok := b.lastPathText(); ok { xOffset = t.Position().X + t.Width() }
	
	for _, l := range b.nameTexts { l.Close() }
	b.nameTexts = []*Text{}
	width := 0.0
	for i, activeIndex := range b.activeIndices {
		child := infos[activeIndex]
		l := NewText(child.Name())
		l.SetTextColor(getTextColor(child, .7))
		l.SetBackgroundColor(Color{0, 0, 0, .7})
		b.AddChild(l)
		b.nameTexts = append(b.nameTexts, l)
		l.Move(Pt(xOffset, float64(len(b.activeIndices) - i - 1)*l.Height()))
		if l.Width() > width { width = l.Width() }
	}
	b.text.Raise()
	b.Resize(xOffset + width, float64(len(b.nameTexts))*b.text.Height())
	
	yOffset := float64(len(b.activeIndices) - b.currentActiveIndex - 1)*b.text.Height()
	b.text.Move(Pt(xOffset, yOffset))
	if len(b.activeIndices) > 0 {
		b.text.SetTextColor(getTextColor(infos[b.activeIndices[b.currentActiveIndex]], 1))
	}
	for _, p := range b.pathTexts { p.Move(Pt(p.Position().X, yOffset)) }
	
	b.Pan(Pt(0, yOffset))
}

func (b *Browser) Paint() {
	rect := ZR
	if b.newInfo == nil && len(b.nameTexts) > 0 {
		cur := b.nameTexts[b.currentActiveIndex]
		rect = Rect(0, cur.Position().Y, cur.Position().X + cur.Width(), cur.Position().Y + cur.Height())
	} else {
		rect = b.text.MapRectToParent(b.text.Rect())
		rect.Min.X = 0
	}
	SetColor(Color{1, 1, 1, .7})
	FillRect(rect)
}

type nodeNameText struct {
	Text
	b *Browser
}
func newNodeNameText(b *Browser) *nodeNameText {
	t := &nodeNameText{}
	t.Text = *NewTextBase(t, "")
	t.b = b
	t.SetValidator(func(text *string) bool {
		if b.newInfo != nil { return true }
		for _, info := range b.filteredInfos() {
			if HasPrefix(ToLower(info.Name()), ToLower(*text)) {
				*text = info.Name()[:len(*text)]
				return true
			}
		}
		return false
	})
	return t
}
func (t *nodeNameText) LostKeyboardFocus() { t.b.Cancel() }
func (t *nodeNameText) KeyPressed(event KeyEvent) {
	b := t.b
	switch event.Key {
	case KeyUp:
		if b.newInfo == nil {
			b.currentActiveIndex--
			b.update()
		}
	case KeyDown:
		if b.newInfo == nil {
			b.currentActiveIndex++
			b.update()
		}
	case KeyBackspace:
		if len(t.GetText()) > 0 {
			t.Text.KeyPressed(event)
			break
		}
		fallthrough
	case KeyLeft:
		if parent := b.currentInfo.Parent(); b.newInfo == nil && parent != nil {
			previous := b.currentInfo
			b.currentInfo = parent
			b.currentActiveIndex = 0
			for i, info := range b.filteredInfos() {
				if info == previous { b.activeIndices = []int{i}; break }
			}
			
			length := len(b.pathTexts)
			b.pathTexts[length - 1].Close()
			b.pathTexts = b.pathTexts[:length - 1]
			
			t.SetText("")
		}
	case KeyEnter:
		info := b.newInfo
		existing := false
		if info == nil {
			info = b.currentActiveInfo()
		} else if b.currentActiveInfo() != nil && info.Name() == b.currentActiveInfo().Name() {
			info = b.currentActiveInfo()
			existing = true
		}
		if b.newInfo != nil && !existing {
			b.currentInfo.AddChild(info)
			switch info := info.(type) {
			case *PackageInfo:
				*info = *newPackageInfo(info.parent.(*PackageInfo), info.name)
				if err := os.Mkdir(info.FluxSourcePath(), 0755); err != nil { Println(err) }
			case *NamedType:
				if err := WriteFile(info.FluxSourcePath(), []byte("type"), 0644); err != nil { Println(err) }
			case *FuncInfo:
				NewFunction(info)
			}
			
			b.currentActiveIndex = 0
			for i, child := range b.filteredInfos() {
				if child == info { b.activeIndices = []int{i}; break }
			}
		}
		b.newInfo = nil
		if _, ok := info.(*PackageInfo); !ok {
			b.Close()
			b.finished = true
			b.created.Emit(info)
			return
		}
		fallthrough
	case KeyRight:
		if b.newInfo == nil {
			switch info := b.currentActiveInfo().(type) {
			case *PackageInfo, *NamedType:
				b.currentInfo = info
				b.activeIndices[0], b.currentActiveIndex = 0, 0
			
				pathText := NewText(info.Name() + "/")
				pathText.SetTextColor(getTextColor(info, 1))
				pathText.SetBackgroundColor(Color{0, 0, 0, .7})
				b.AddChild(pathText)
				x := 0.0; if t, ok := b.lastPathText(); ok { x = t.Position().X + t.Width() }
				pathText.Move(Pt(x, 0))
				b.pathTexts = append(b.pathTexts, pathText)
			
				t.SetText("")
			}
		}
	case KeyEsc:
		if b.newInfo == nil {
			b.Cancel()
		} else {
			b.newInfo = nil
			t.SetText("")
		}
	default:
		_, inPkg := b.currentInfo.(*PackageInfo)
		_, inType := b.currentInfo.(*NamedType)
		if event.Ctrl && b.mode != typesOnly && b.newInfo == nil && (inPkg || inType && event.Text == "3") {
			switch event.Text {
			case "1": b.newInfo = &PackageInfo{}
			case "2": b.newInfo = &NamedType{}
			case "3": b.newInfo = &FuncInfo{}
			case "4": b.newInfo = &ValueInfo{}
			case "5": b.newInfo = &ValueInfo{constant:true}
			default: t.Text.KeyPressed(event); return
			}
			t.SetText("")
		} else {
			t.Text.KeyPressed(event)
		}
	}
}
