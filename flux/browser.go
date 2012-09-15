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
	accepted, canceled *Signal
	
	path []Info
	indices []int
	i int
	newInfo Info
	
	pathTexts, nameTexts []*Text
	text *nodeNameText
}

type browserMode int
const (
	browse = iota
	fluxSourceOnly
	typesOnly
)

func NewBrowser(mode browserMode) *Browser {
	b := &Browser{accepted:NewSignal(), canceled:NewSignal()}
	b.ViewBase = NewView(b)
	
	b.mode = mode
	b.path = []Info{rootPackageInfo}
	rootPackageInfo.types = builtinPkg.types
	rootPackageInfo.functions = builtinPkg.functions
	rootPackageInfo.constants = builtinPkg.constants
	
	b.text = newNodeNameText(b)
	b.text.SetBackgroundColor(Color{0, 0, 0, 0})
	b.AddChild(b.text)
	b.text.TextChanged.Connect(func(...interface{}) { b.update() })
	b.text.SetText("")
	
	return b
}

func (b *Browser) cancel() {
	if !b.finished {
		b.finished = true
		b.Close()
		b.canceled.Emit()
	}
}

func (b *Browser) update() {
	currentIndex := 0
	n := len(b.indices)
	if n > 0 {
		b.i %= n
		if b.i < 0 { b.i += n }
		currentIndex = b.indices[b.i]
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
	
	b.indices = nil
	for i, child := range infos {
		if HasPrefix(ToLower(child.Name()), ToLower(b.text.GetText())) {
			b.indices = append(b.indices, i)
		}
	}
	n = len(b.indices)
	for i, index := range b.indices {
		if index >= currentIndex {
			b.i = i
			break
		}
	}
	if b.i >= n { b.i = n - 1 }
	
	if t, ok := b.lastPathText(); ok && n > 0 {
		sep := ""; if _, ok := infos[b.indices[b.i]].(*PackageInfo); ok { sep = "/" } else { sep = "." }
		text := t.GetText()
		t.SetText(text[:len(text) - 1] + sep)
	}
	xOffset := 0.0; if t, ok := b.lastPathText(); ok { xOffset = t.Position().X + t.Width() }
	
	for _, l := range b.nameTexts { l.Close() }
	b.nameTexts = []*Text{}
	width := 0.0
	for i, activeIndex := range b.indices {
		child := infos[activeIndex]
		l := NewText(child.Name())
		l.SetTextColor(getTextColor(child, .7))
		l.SetBackgroundColor(Color{0, 0, 0, .7})
		b.AddChild(l)
		b.nameTexts = append(b.nameTexts, l)
		l.Move(Pt(xOffset, float64(n - i - 1)*l.Height()))
		if l.Width() > width { width = l.Width() }
	}
	b.text.Raise()
	b.Resize(xOffset + width, float64(len(b.nameTexts))*b.text.Height())
	
	yOffset := float64(n - b.i - 1)*b.text.Height()
	b.text.Move(Pt(xOffset, yOffset))
	if n > 0 {
		b.text.SetTextColor(getTextColor(infos[b.indices[b.i]], 1))
	}
	for _, p := range b.pathTexts { p.Move(Pt(p.Position().X, yOffset)) }
	
	b.Pan(Pt(0, yOffset))
}

func (b Browser) filteredInfos() (infos []Info) {
	switch b.mode {
	case browse:
		infos = b.path[0].Children()
	case fluxSourceOnly:
		for _, child := range b.path[0].Children() {
			if _, err := os.Stat(child.FluxSourcePath()); err == nil { infos = append(infos, child) }
		}
	case typesOnly:
		for _, child := range b.path[0].Children() {
			switch child.(type) { case *PackageInfo, *NamedType: infos = append(infos, child) }
		}
	}
	return
}

func (b Browser) currentInfo() Info {
	infos := b.filteredInfos()
	if len(b.indices) == 0 || len(infos) == 0 { return nil }
	return infos[b.indices[b.i]]
}

func (b Browser) lastPathText() (*Text, bool) {
	if np := len(b.pathTexts); np > 0 {
		return b.pathTexts[np - 1], true
	}
	return nil, false
}

func (b *Browser) Paint() {
	rect := ZR
	if b.newInfo == nil && len(b.nameTexts) > 0 {
		cur := b.nameTexts[b.i]
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
func (t *nodeNameText) LostKeyboardFocus() { t.b.cancel() }
func (t *nodeNameText) KeyPressed(event KeyEvent) {
	b := t.b
	switch event.Key {
	case KeyUp:
		if b.newInfo == nil {
			b.i--
			b.update()
		}
	case KeyDown:
		if b.newInfo == nil {
			b.i++
			b.update()
		}
	case KeyBackspace:
		if len(t.GetText()) > 0 {
			t.Text.KeyPressed(event)
			break
		}
		fallthrough
	case KeyLeft:
		if len(b.path) > 1 && b.newInfo == nil {
			previous := b.path[0]
			b.path = b.path[1:]
			b.i = 0
			for i, info := range b.filteredInfos() {
				if info == previous { b.indices = []int{i}; break }
			}
			
			length := len(b.pathTexts)
			b.pathTexts[length - 1].Close()
			b.pathTexts = b.pathTexts[:length - 1]
			
			t.SetText("")
		}
	case KeyEnter:
		info := b.newInfo
		existing := false
		if cur := b.currentInfo(); info == nil {
			info = cur
		} else if cur != nil && info.Name() == cur.Name() {
			info = cur
			existing = true
		}
		if b.newInfo != nil && !existing {
			b.path[0].AddChild(info)
			switch info := info.(type) {
			case *PackageInfo:
				*info = *newPackageInfo(info.parent.(*PackageInfo), info.name)
				if err := os.Mkdir(info.FluxSourcePath(), 0755); err != nil { Println(err) }
			case *NamedType:
				if err := WriteFile(info.FluxSourcePath(), []byte("type"), 0644); err != nil { Println(err) }
			case *FuncInfo:
				NewFunction(info)
			}
			
			b.i = 0
			for i, child := range b.filteredInfos() {
				if child == info { b.indices = []int{i}; break }
			}
		}
		b.newInfo = nil
		if _, ok := info.(*PackageInfo); !ok {
			b.Close()
			b.finished = true
			b.accepted.Emit(info)
			return
		}
		fallthrough
	case KeyRight:
		if b.newInfo == nil {
			switch info := b.currentInfo().(type) {
			case *PackageInfo, *NamedType:
				b.path = append([]Info{info}, b.path...)
				b.indices = nil
				
				sep := ""; if _, ok := info.(*PackageInfo); ok { sep = "/" } else { sep = "." }
				pathText := NewText(info.Name() + sep)
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
			b.cancel()
		} else {
			b.newInfo = nil
			t.SetText("")
		}
	default:
		_, inPkg := b.path[0].(*PackageInfo)
		_, inType := b.path[0].(*NamedType)
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
