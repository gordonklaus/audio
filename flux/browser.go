package main

import (
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	."fmt"
	."github.com/jteeuwen/glfw"
	"go/ast"
	"os"
	."strings"
)

type Browser struct {
	*ViewBase
	mode browserMode
	currentPkg *PackageInfo
	finished bool
	accepted, canceled *Signal
	
	path []Info
	indices []int
	i int
	newInfo Info
	
	pathTexts, nameTexts []Text
	text *nodeNameText
	typeView *typeView
}

type browserMode int
const (
	browse = iota
	fluxSourceOnly
	typesOnly
)

func NewBrowser(mode browserMode, currentPkg *PackageInfo, imports []*PackageInfo) *Browser {
	b := &Browser{currentPkg:currentPkg, accepted:NewSignal(), canceled:NewSignal()}
	b.ViewBase = NewView(b)
	
	b.mode = mode
	b.path = []Info{rootPackageInfo}
	rootPackageInfo.types = append([]*NamedType{}, builtinPkg.types...)
	rootPackageInfo.functions = append([]*FuncInfo{}, builtinPkg.functions...)
	rootPackageInfo.variables = []*ValueInfo{}
	rootPackageInfo.constants = append([]*ValueInfo{}, builtinPkg.constants...)
	if currentPkg != nil { imports = append(imports, currentPkg) }
	for _, p := range imports {
		rootPackageInfo.types = append(rootPackageInfo.types, p.types...)
		rootPackageInfo.functions = append(rootPackageInfo.functions, p.functions...)
		rootPackageInfo.variables = append(rootPackageInfo.variables, p.variables...)
		rootPackageInfo.constants = append(rootPackageInfo.constants, p.constants...)
	}
	Sort(rootPackageInfo.types, "Name")
	rootPackageInfo.types = append([]*NamedType{protoPointer, protoArray, protoSlice, protoMap, protoChan, protoFunc, protoInterface, protoStruct}, rootPackageInfo.types...)
	Sort(rootPackageInfo.functions, "Name")
	Sort(rootPackageInfo.variables, "Name")
	Sort(rootPackageInfo.constants, "Name")
	
	b.text = newNodeNameText(b)
	b.text.SetBackgroundColor(Color{0, 0, 0, 0})
	b.AddChild(b.text)
	
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

func (b Browser) filteredInfos() (infos []Info) {
	for _, i := range b.path[0].Children() {
		switch b.mode {
		case fluxSourceOnly:
			if p := i.Parent(); p == nil || p == builtinPkg { continue }
			if _, err := os.Stat(i.FluxSourcePath()); err != nil { continue }
		case typesOnly:
			switch i.(type) {
				default: continue
				case *PackageInfo, *NamedType:
			}
		}
		if b.currentPkg != nil {
			switch i.(type) {
			default:
				if p := i.Parent(); p != nil && p != builtinPkg && p != cPkg && p != b.currentPkg && !ast.IsExported(i.Name()) { continue }
			case *PackageInfo:
			}
		}
		infos = append(infos, i)
	}
	return
}

func (b Browser) currentInfo() Info {
	infos := b.filteredInfos()
	if len(b.indices) == 0 || len(infos) == 0 { return nil }
	return infos[b.indices[b.i]]
}

func (b Browser) lastPathText() (Text, bool) {
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
	*TextBase
	b *Browser
}
func newNodeNameText(b *Browser) *nodeNameText {
	t := &nodeNameText{}
	t.TextBase = NewTextBase(t, "")
	t.b = b
	return t
}
func (t *nodeNameText) SetText(text string) {
	b := t.b
	if b.newInfo == nil {
		if infos := b.filteredInfos(); len(infos) > 0 {
			for _, info := range infos {
				if HasPrefix(ToLower(info.Name()), ToLower(text)) {
					goto ok
				}
			}
			return
		}
	}
ok:
	currentIndex := 0
	n := len(b.indices)
	if n > 0 {
		b.i %= n
		if b.i < 0 { b.i += n }
		currentIndex = b.indices[b.i]
	}
	
	infos := b.filteredInfos()
	if b.newInfo != nil {
		b.newInfo.SetName(text)
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
		if HasPrefix(ToLower(child.Name()), ToLower(text)) {
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
	
	var cur Info; if n > 0 { cur = infos[b.indices[b.i]] }
	if cur != nil {
		text = cur.Name()[:len(text)]
	} else {
		text = ""
	}
	t.TextBase.SetText(text)
	
	if t, ok := b.lastPathText(); ok && cur != nil {
		sep := ""; if _, ok := cur.(*PackageInfo); ok { sep = "/" } else { sep = "." }
		text := t.GetText()
		t.SetText(text[:len(text) - 1] + sep)
	}
	xOffset := 0.0; if t, ok := b.lastPathText(); ok { xOffset = t.Position().X + t.Width() }

	for _, l := range b.nameTexts { l.Close() }
	b.nameTexts = []Text{}
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
	if b.typeView != nil { b.RemoveChild(b.typeView) }
	if cur != nil {
		b.text.SetTextColor(getTextColor(cur, 1))
		switch i := cur.(type) {
		case *NamedType:
			if p := i.parent; p != builtinPkg && p != cPkg {
				b.typeView = newTypeView(&i.underlying)
				b.AddChild(b.typeView)
			}
		case *FuncInfo:
			t := Type(i.typ)
			b.typeView = newTypeView(&t)
			b.AddChild(b.typeView)
		case *ValueInfo:
			b.typeView = newTypeView(&i.typ)
			b.AddChild(b.typeView)
		}
		if b.typeView != nil {
			b.typeView.Move(Pt(xOffset + width + 16, yOffset - (b.typeView.Height() - b.text.Height()) / 2))
		}
	}
	for _, p := range b.pathTexts { p.Move(Pt(p.Position().X, yOffset)) }

	b.Pan(Pt(0, yOffset))
}
func (t *nodeNameText) LostKeyboardFocus() { t.b.cancel() }
func (t *nodeNameText) KeyPressed(event KeyEvent) {
	b := t.b
	switch event.Key {
	case KeyUp:
		if b.newInfo == nil {
			b.i--
			t.SetText(t.GetText())
		}
	case KeyDown:
		if b.newInfo == nil {
			b.i++
			t.SetText(t.GetText())
		}
	case KeyBackspace:
		if len(t.GetText()) > 0 {
			t.TextBase.KeyPressed(event)
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
				if err := os.Mkdir(info.FluxSourcePath(), 0777); err != nil { Println(err) }
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
		recv, inType := b.path[0].(*NamedType)
		if event.Ctrl && b.mode != typesOnly && b.newInfo == nil && (inPkg || inType && event.Text == "3") {
			switch event.Text {
			case "1": b.newInfo = &PackageInfo{}
			case "2": b.newInfo = &NamedType{}
			case "3":
				f := &FuncInfo{typ:&FuncType{}}
				if inType {
					f.receiver = &PointerType{element:recv}
				}
				b.newInfo = f
			case "4": b.newInfo = &ValueInfo{}
			case "5": b.newInfo = &ValueInfo{constant:true}
			default: t.TextBase.KeyPressed(event); return
			}
			t.SetText("")
		} else {
			t.TextBase.KeyPressed(event)
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

var (
	protoPointer = &NamedType{InfoBase:InfoBase{"pointer", nil}}
	protoArray = &NamedType{InfoBase:InfoBase{"array", nil}}
	protoSlice = &NamedType{InfoBase:InfoBase{"slice", nil}}
	protoMap = &NamedType{InfoBase:InfoBase{"map", nil}}
	protoChan = &NamedType{InfoBase:InfoBase{"chan", nil}}
	protoFunc = &NamedType{InfoBase:InfoBase{"func", nil}}
	protoInterface = &NamedType{InfoBase:InfoBase{"interface", nil}}
	protoStruct = &NamedType{InfoBase:InfoBase{"struct", nil}}
	
	protoType = map[*NamedType]bool{protoPointer:true, protoArray:true, protoSlice:true, protoMap:true, protoChan:true, protoFunc:true, protoInterface:true, protoStruct:true}
)

func newProtoType(t *NamedType) (p Type) {
	switch t {
	case protoPointer: p = &PointerType{}
	case protoArray: p = &ArrayType{}
	case protoSlice: p = &SliceType{}
	case protoMap: p = &MapType{}
	case protoChan: p = &ChanType{send:true, recv:true}
	case protoFunc: p = &FuncType{}
	case protoInterface: p = &InterfaceType{}
	case protoStruct: p = &StructType{}
	default: panic(Sprintf("not a proto type %#v", t))
	}
	return
}
