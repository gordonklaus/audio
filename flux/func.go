package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."fmt"
)

type funcNode struct {
	*ViewBase
	AggregateMouseHandler
	info *FuncInfo
	pkgRefs map[*PackageInfo]int
	inputsNode, outputsNode *portsNode
	funcblk *block
}

func newFuncNode(info *FuncInfo) *funcNode {
	f := &funcNode{info:info}
	f.ViewBase = NewView(f)
	f.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(f), NewViewDragger(f)}
	f.pkgRefs = map[*PackageInfo]int{}
	f.funcblk = newBlock(f)
	f.inputsNode = newInputsNode(f.funcblk)
	f.inputsNode.editable = true
	f.funcblk.addNode(f.inputsNode)
	f.outputsNode = newOutputsNode(f.funcblk)
	f.outputsNode.editable = true
	f.funcblk.addNode(f.outputsNode)
	f.AddChild(f.funcblk)
	go f.funcblk.animate()
	
	if info.receiver != nil {
		f.inputsNode.newOutput(info.typeWithReceiver().parameters[0])
	}
	
	if !loadFunc(f) { saveFunc(*f) }
	
	return f
}

func (f funcNode) pkg() *PackageInfo {
	parent := f.info.parent
	if t, ok := parent.(*NamedType); ok {
		return t.parent.(*PackageInfo)
	}
	return parent.(*PackageInfo)
}

func (f funcNode) imports() (x []*PackageInfo) {
	for p := range f.pkgRefs {
		x = append(x, p)
	}
	return
}

func (f *funcNode) addPackageRef(x interface{}) {
	switch x := x.(type) {
	case Info:
		if p, ok := x.Parent().(*PackageInfo); ok && p != f.pkg() && p != builtinPkg {
			f.pkgRefs[p]++
		}
	case Type:
		walkType(x, func(t *NamedType) { f.addPackageRef(t) })
	default:
		panic(Sprintf("can't addPackageRef for %#v\n", x))
	}
}
func (f *funcNode) subPackageRef(x interface{}) {
	switch x := x.(type) {
	case Info:
		if p, ok := x.Parent().(*PackageInfo); ok {
			f.pkgRefs[p]--
			if f.pkgRefs[p] <= 0 {
				delete(f.pkgRefs, p)
			}
		}
	case Type:
		walkType(x, func(t *NamedType) { f.subPackageRef(t) })
	default:
		panic(Sprintf("can't subPackageRef for %#v\n", x))
	}
}

func (n funcNode) block() *block { return nil }
func (n funcNode) inputs() []*input { return nil }
func (n funcNode) outputs() []*output { return nil }
func (n funcNode) inConns() []*connection { return nil }
func (n funcNode) outConns() []*connection { return nil }

func (n *funcNode) positionblocks() {
	b := n.funcblk
	leftmost, rightmost := b.points[0], b.points[0]
	for _, p := range b.points {
		if p.X < leftmost.X { leftmost = p }
		if p.X > rightmost.X { rightmost = p }
	}
	n.inputsNode.MoveOrigin(leftmost)
	n.outputsNode.MoveOrigin(rightmost)
	ResizeToFit(n, 0)
}

func (f *funcNode) TookKeyboardFocus() { f.funcblk.TakeKeyboardFocus() }

func (f *funcNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyF1:
		saveFunc(*f)
	default:
		f.ViewBase.KeyPressed(event)
	}
}
