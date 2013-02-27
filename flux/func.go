package main

import (
	."code.google.com/p/gordon-go/gui"
	."fmt"
)

type funcNode struct {
	*ViewBase
	AggregateMouseHandler
	info *Func
	pkgRefs map[*Package]int
	inputsNode, outputsNode *portsNode
	funcblk *block
}

func newFuncNode(info *Func) *funcNode {
	n := &funcNode{info:info}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.pkgRefs = map[*Package]int{}
	n.funcblk = newBlock(n)
	n.inputsNode = newInputsNode(n.funcblk)
	n.inputsNode.editable = true
	n.funcblk.addNode(n.inputsNode)
	n.outputsNode = newOutputsNode(n.funcblk)
	n.outputsNode.editable = true
	n.funcblk.addNode(n.outputsNode)
	n.AddChild(n.funcblk)
	go n.funcblk.animate()
	
	if info.receiver != nil {
		n.inputsNode.newOutput(info.typeWithReceiver().parameters[0])
	}
	
	if !loadFunc(n) { saveFunc(*n) }
	
	return n
}

func (n funcNode) pkg() *Package {
	parent := n.info.parent
	if t, ok := parent.(*NamedType); ok {
		return t.parent.(*Package)
	}
	return parent.(*Package)
}

func (n funcNode) imports() (x []*Package) {
	for p := range n.pkgRefs {
		x = append(x, p)
	}
	return
}

func (n *funcNode) addPkgRef(x interface{}) {
	switch x := x.(type) {
	case Info:
		if p, ok := x.Parent().(*Package); ok && p != n.pkg() && p != builtinPkg {
			n.pkgRefs[p]++
		}
	case Type:
		walkType(x, func(t *NamedType) { n.addPkgRef(t) })
	default:
		panic(Sprintf("can't addPkgRef for %#v\n", x))
	}
}
func (n *funcNode) subPkgRef(x interface{}) {
	switch x := x.(type) {
	case Info:
		if p, ok := x.Parent().(*Package); ok {
			n.pkgRefs[p]--
			if n.pkgRefs[p] <= 0 {
				delete(n.pkgRefs, p)
			}
		}
	case Type:
		walkType(x, func(t *NamedType) { n.subPkgRef(t) })
	default:
		panic(Sprintf("can't subPkgRef for %#v\n", x))
	}
}

func (n funcNode) block() *block { return nil }
func (n funcNode) inputs() []*input { return nil }
func (n funcNode) outputs() []*output { return nil }
func (n funcNode) inConns() []*connection { return nil }
func (n funcNode) outConns() []*connection { return nil }

func (n *funcNode) positionBlocks() {
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

func (n *funcNode) TookKeyboardFocus() { n.funcblk.TakeKeyboardFocus() }

func (n *funcNode) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyF1:
		saveFunc(*n)
	default:
		n.ViewBase.KeyPressed(event)
	}
}
