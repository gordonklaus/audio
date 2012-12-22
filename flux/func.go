package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."fmt"
)

type FuncNode struct {
	*ViewBase
	AggregateMouseHandler
	info *FuncInfo
	pkgRefs map[*PackageInfo]int
	inputNode, outputNode *InOutNode
	funcBlock *Block
}

func NewFuncNode(info *FuncInfo) *FuncNode {
	f := &FuncNode{info:info}
	f.ViewBase = NewView(f)
	f.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(f), NewViewDragger(f)}
	f.pkgRefs = map[*PackageInfo]int{}
	f.funcBlock = NewBlock(f)
	f.inputNode = NewInputNode(f.funcBlock)
	f.inputNode.editable = true
	f.funcBlock.AddNode(f.inputNode)
	f.outputNode = NewOutputNode(f.funcBlock)
	f.outputNode.editable = true
	f.funcBlock.AddNode(f.outputNode)
	f.AddChild(f.funcBlock)
	go f.funcBlock.animate()
	
	if info.receiver != nil {
		f.inputNode.newOutput(info.typeWithReceiver().parameters[0])
	}
	
	if !loadFunc(f) { saveFunc(*f) }
	
	return f
}

func (f FuncNode) pkg() *PackageInfo {
	parent := f.info.parent
	if typ, ok := parent.(*NamedType); ok {
		return typ.parent.(*PackageInfo)
	}
	return parent.(*PackageInfo)
}

func (f FuncNode) imports() (x []*PackageInfo) {
	for p := range f.pkgRefs {
		x = append(x, p)
	}
	return
}

func (f *FuncNode) AddPackageRef(x interface{}) {
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
func (f *FuncNode) SubPackageRef(x interface{}) {
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

func (n FuncNode) Block() *Block { return nil }
func (n FuncNode) Inputs() []*Input { return nil }
func (n FuncNode) Outputs() []*Output { return nil }
func (n FuncNode) InputConnections() []*Connection { return nil }
func (n FuncNode) OutputConnections() []*Connection { return nil }

func (n *FuncNode) positionBlocks() {
	b := n.funcBlock
	leftmost, rightmost := b.points[0], b.points[0]
	for _, p := range b.points {
		if p.X < leftmost.X { leftmost = p }
		if p.X > rightmost.X { rightmost = p }
	}
	n.inputNode.MoveOrigin(leftmost)
	n.outputNode.MoveOrigin(rightmost)
	ResizeToFit(n, 0)
}

func (f *FuncNode) TookKeyboardFocus() { f.funcBlock.TakeKeyboardFocus() }

func (f *FuncNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyF1:
		saveFunc(*f)
	default:
		f.ViewBase.KeyPressed(event)
	}
}
