package main

import (
	."fmt"
	."io/ioutil"
	."strconv"
	."strings"
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
)

type Function struct {
	*ViewBase
	info *FunctionInfo
	pkgRefs map[*PackageInfo]int
	block *Block
}

func NewFunction(info *FunctionInfo) *Function {
	f := &Function{info:info}
	f.ViewBase = NewView(f)
	f.pkgRefs = map[*PackageInfo]int{}
	f.block = NewBlock(nil)
	f.block.function = f
	f.AddChild(f.block)
	
	if !f.load() { f.Save() }
	
	return f
}

var varNameIndex int
func newVarName() string { varNameIndex++; return "v" + Itoa(varNameIndex) }

func tabs(n int) string { return Repeat("\t", n) }

var nodeID int
func (f Function) Save() {
	nodeID = 0
	s := "func\n"
	for pkg := range f.pkgRefs {
		s += pkg.buildPackage.ImportPath + "\n"
	}
	s += f.block.Save(0, map[Node]int{})
	if err := WriteFile(f.info.FluxSourcePath(), []byte(s), 0644); err != nil { Println(err) }
	
	var pkgInfo *PackageInfo
	parent := f.info.Parent()
	typeInfo, isMethod := parent.(TypeInfo)
	if isMethod {
		pkgInfo = typeInfo.Parent().(*PackageInfo)
	} else {
		pkgInfo = parent.(*PackageInfo)
	}
	
	varNameIndex = 0
	s = Sprintf("package %v\n\nimport (\n", pkgInfo.Name())
	for pkg := range f.pkgRefs {
		s += Sprintf("\t\"%v\"\n", pkg.buildPackage.ImportPath)
	}
	s += Sprintf(")\n\nfunc %v() {\n", f.info.Name())
	s += f.block.Code(1, map[*Input]string{})
	s += "}"
	WriteFile(Sprintf("%v/%v.go", pkgInfo.FluxSourcePath(), f.info.Name()), []byte(s), 0644)
}

func (f *Function) load() bool {
	if s, err := ReadFile(f.info.FluxSourcePath()); err == nil {
		_, s := Split2(string(s), "\n")
		// TODO:  verify/update signature?
		pkgNames := map[string]*PackageInfo{}
		for line := ""; s[0] != '\\'; {
			line, s = Split2(string(s), "\n")
			pkg := FindPackageInfo(line)
			// TODO:  handle name collisions
			pkgNames[pkg.name] = pkg
		}
		f.block.Load(s, 0, map[int]Node{}, pkgNames)
		return true
	}
	return false
}

func (f *Function) Resize(w, h float64) {
	f.ViewBase.Resize(w, h)
	f.block.Resize(w, h)
}

func (f *Function) TookKeyboardFocus() { f.block.TakeKeyboardFocus() }

func (f *Function) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyF1:
		f.Save()
	default:
		f.ViewBase.KeyPressed(event)
	}
}
