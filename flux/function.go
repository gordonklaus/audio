package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."fmt"
)

type Function struct {
	*ViewBase
	info *FuncInfo
	pkgRefs map[*PackageInfo]int
	inputNode *InputNode
	block *Block
}

func NewFunction(info *FuncInfo) *Function {
	f := &Function{info:info}
	f.ViewBase = NewView(f)
	f.pkgRefs = map[*PackageInfo]int{}
	f.block = NewBlock(nil)
	f.block.function = f
	f.inputNode = NewInputNode(f.block)
	f.block.AddNode(f.inputNode)
	f.AddChild(f.block)
	
	if !loadFunction(f) { saveFunction(*f) }
	
	return f
}

func (f Function) pkg() *PackageInfo {
	parent := f.info.parent
	if typ, ok := parent.(*NamedType); ok {
		return typ.parent.(*PackageInfo)
	}
	return parent.(*PackageInfo)
}

func (f Function) imports() (x []*PackageInfo) {
	for p := range f.pkgRefs {
		x = append(x, p)
	}
	return
}

func (f *Function) AddPackageRef(x interface{}) {
	switch x := x.(type) {
	case Info:
		if p, ok := x.Parent().(*PackageInfo); ok && p != f.pkg() && p != builtinPkg {
			f.pkgRefs[p]++
		}
	case Type:
		f.walkType(x, (*Function).AddPackageRef)
	default:
		panic(Sprintf("can't AddPackageRef for %#v\n", x))
	}
}
func (f *Function) SubPackageRef(x interface{}) {
	switch x := x.(type) {
	case Info:
		if p, ok := x.Parent().(*PackageInfo); ok {
			f.pkgRefs[p]--
			if f.pkgRefs[p] <= 0 {
				delete(f.pkgRefs, p)
			}
		}
	case Type:
		f.walkType(x, (*Function).SubPackageRef)
	default:
		panic(Sprintf("can't SubPackageRef for %#v\n", x))
	}
}
func (f *Function) walkType(t Type, op func(*Function, interface{})) {
	switch t := t.(type) {
	case PointerType:
		f.walkType(t.element, op)
	case ArrayType:
		f.walkType(t.element, op)
	case SliceType:
		f.walkType(t.element, op)
	case MapType:
		f.walkType(t.key, op)
		f.walkType(t.value, op)
	case ChanType:
		f.walkType(t.element, op)
	case FuncType:
		for _, v := range append(t.parameters, t.results...) { f.walkType(v.typ, op) }
	case InterfaceType:
		for _, m := range t.methods { f.walkType(m.typ, op) }
	case StructType:
		for _, v := range t.fields { f.walkType(v.typ, op) }
	case *NamedType:
		op(f, t)
	default:
		panic(Sprintf("unexpected type %#v\n", t))
	}
}

func (f *Function) TookKeyboardFocus() { f.block.TakeKeyboardFocus() }

func (f *Function) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyF1:
		saveFunction(*f)
	default:
		f.ViewBase.KeyPressed(event)
	}
}
