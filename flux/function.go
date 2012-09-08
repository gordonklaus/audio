package main

import (
	."io/ioutil"
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
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
	
	if !f.load() { saveFunction(*f) }
	
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

func (f *Function) load() bool {
	if s, err := ReadFile(f.info.FluxSourcePath()); err == nil {
		line, s := Split2(string(s), "\n")
		pkgNames := map[string]*PackageInfo{}
		for s[0] != '\\' {
			line, s = Split2(s, "\n")
			pkg := FindPackageInfo(line)
			// TODO:  handle name collisions
			pkgNames[pkg.name] = pkg
		}
		for _, parameter := range f.info.typ.parameters {
			// TODO:  increment pkgRef for this parameter's type
			p := NewOutput(f.inputNode, parameter)
			f.inputNode.AddChild(p)
			f.inputNode.outputs = append(f.inputNode.outputs, p)
		}
		f.inputNode.reform()
		f.block.Load(s, 0, map[int]Node{}, pkgNames)
		return true
	}
	return false
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
