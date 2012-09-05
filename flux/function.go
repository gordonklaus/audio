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
	
	if !f.load() { f.Save() }
	
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

var varNameIndex int
func newVarName() string { varNameIndex++; return "v" + Itoa(varNameIndex) }

func tabs(n int) string { return Repeat("\t", n) }

var nodeID int
func (f Function) Save() {
	pkg := f.pkg()
	
	nodeID = 0
	var params []string
	for _, output := range f.inputNode.Outputs() {
		s := ""
		if name := output.info.Name(); name != "" { s = name + " " }
		s += qualifiedName(output.info.typ, pkg)
		params = append(params, s)
	}
	s := Sprintf("func(%v)\n", Join(params, ", "))
	for p := range f.pkgRefs {
		if p != pkg {
			s += p.buildPackage.ImportPath + "\n"
		}
	}
	s += f.block.Save(0, map[Node]int{})
	if err := WriteFile(f.info.FluxSourcePath(), []byte(s), 0644); err != nil { Println(err) }
	
	varNameIndex = 0
	vars := map[*Input]string{}
	s = Sprintf("package %v\n\nimport (\n", pkg.Name())
	for p := range f.pkgRefs {
		if p != pkg {
			s += Sprintf("\t\"%v\"\n", p.buildPackage.ImportPath)
		}
	}
	params = nil
	for _, output := range f.inputNode.Outputs() {
		// TODO: use output.info.Name(), handle name collisions?
		name := ""
		if len(output.connections) > 0 {
			name = newVarName()
			vars[output.connections[0].dst] = name
		} else {
			name = "_"
		}
		params = append(params, name + " " + qualifiedName(output.info.typ, pkg))
	}
	s += Sprintf(")\n\nfunc %v(%v) {\n", f.info.Name(), Join(params, ", "))
	s += f.block.Code(1, vars)
	s += "}"
	WriteFile(Sprintf("%v/%v.go", f.info.parent.FluxSourcePath(), f.info.Name()), []byte(s), 0644)
}

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
		f.Save()
	default:
		f.ViewBase.KeyPressed(event)
	}
}
