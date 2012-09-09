package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
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

func (f *Function) AddPackageRef(p *PackageInfo) { f.pkgRefs[p]++ }
func (f *Function) SubPackageRef(p *PackageInfo) { f.pkgRefs[p]--; if f.pkgRefs[p] == 0 { delete(f.pkgRefs, p) } }

func (f *Function) TookKeyboardFocus() { f.block.TakeKeyboardFocus() }

func (f *Function) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyF1:
		saveFunction(*f)
	default:
		f.ViewBase.KeyPressed(event)
	}
}
