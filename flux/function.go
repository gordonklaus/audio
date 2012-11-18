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
	
	if info.receiver != nil {
		r := NewOutput(f.inputNode, info.typeWithReceiver().parameters[0])
		f.inputNode.AddChild(r)
		f.inputNode.outputs = append(f.inputNode.outputs, r)
	}
	
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
		walkType(x, func(t *NamedType) { f.AddPackageRef(t) })
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
		walkType(x, func(t *NamedType) { f.SubPackageRef(t) })
	default:
		panic(Sprintf("can't SubPackageRef for %#v\n", x))
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
