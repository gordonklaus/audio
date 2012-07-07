package main

import (
	."io/ioutil"
	."strconv"
	."strings"
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type Function struct {
	*ViewBase
	block *Block
}

func NewFunction() *Function {
	f := &Function{}
	f.ViewBase = NewView(f)
	f.block = NewBlock(nil)
	f.AddChild(f.block)
	return f
}

var varNameIndex int
func newVarName() string { varNameIndex++; return "v" + Itoa(varNameIndex) }

func tabs(n int) string { return Repeat("\t", n) }

func (f Function) Save() {
	varNameIndex = 0
	s := "package main\n\nimport (\n"
	s += "\t.\"fmt\"\n"
	s += ")\n\nfunc main() {\n"
	s += f.block.Code(1, map[*Input]string{})
	s += "}"
	WriteFile("../main.go", []byte(s), 0644)
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

func (Function) Paint() {}
