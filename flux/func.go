package main

import (
	."code.google.com/p/gordon-go/gui"
	"code.google.com/p/go.exp/go/types"
	"fmt"
	"time"
)

type funcNode struct {
	*ViewBase
	AggregateMouseHandler
	obj types.Object
	pkgRefs map[*types.Package]int
	inputsNode, outputsNode *portsNode
	funcblk *block
	awaken, stop chan bool
	done func()
}

func newFuncNode(obj types.Object) *funcNode {
	n := &funcNode{obj:obj}
	n.ViewBase = NewView(n)
	n.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(n), NewViewDragger(n)}
	n.pkgRefs = map[*types.Package]int{}
	n.funcblk = newBlock(n)
	n.inputsNode = newInputsNode()
	n.inputsNode.editable = true
	n.funcblk.addNode(n.inputsNode)
	n.outputsNode = newOutputsNode()
	n.outputsNode.editable = true
	n.funcblk.addNode(n.outputsNode)
	n.AddChild(n.funcblk)
	n.awaken = make(chan bool, 1)
	n.stop = make(chan bool, 1)
	
	if !loadFunc(n) {
		if m, ok := obj.(method); ok {
			n.inputsNode.newOutput(m.Type.Recv)
		}
		saveFunc(*n)
	}
	
	return n
}

func (n *funcNode) Close() {
	saveFunc(*n)
	n.stop <- true
	select {
	case n.awaken <- true:
	default:
	}
	n.ViewBase.Close()
	n.done()
}

func (n funcNode) pkg() *types.Package {
	return n.obj.GetPkg()
}

func (n funcNode) imports() (x []*types.Package) {
	for p := range n.pkgRefs {
		x = append(x, p)
	}
	return
}

func (n *funcNode) addPkgRef(x interface{}) {
	switch x := x.(type) {
	case types.Object:
		if p := x.GetPkg(); p != n.pkg() && p != nil {
			n.pkgRefs[p]++
		}
	case types.Type:
		walkType(x, func(t *types.NamedType) { n.addPkgRef(t.Obj) })
	default:
		panic(fmt.Sprintf("can't addPkgRef for %#v\n", x))
	}
}
func (n *funcNode) subPkgRef(x interface{}) {
	switch x := x.(type) {
	case types.Object:
		p := x.GetPkg()
		n.pkgRefs[p]--
		if n.pkgRefs[p] <= 0 {
			delete(n.pkgRefs, p)
		}
	case types.Type:
		walkType(x, func(t *types.NamedType) { n.subPkgRef(t.Obj) })
	default:
		panic(fmt.Sprintf("can't subPkgRef for %#v\n", x))
	}
}

func (n funcNode) block() *block { return nil }
func (n funcNode) setBlock(b *block) {}
func (n funcNode) inputs() []*port { return nil }
func (n funcNode) outputs() []*port { return nil }
func (n funcNode) inConns() []*connection { return nil }
func (n funcNode) outConns() []*connection { return nil }

const fps = 60

func (n *funcNode) animate() {
	updated := make(chan bool)
	// TODO:  end me
	for {
		next := time.After(time.Second / fps)
		n.Do(func() {
			updated <- n.update()
		})
		select {
		case ok := <-updated:
			if ok {
				<-next
			} else {
				<-n.awaken
			}
		case <-n.stop:
			return
		}
	}
}

func (n *funcNode) update() bool {
	if !n.funcblk.update() {
		return false
	}
	n.inputsNode.reposition()
	n.outputsNode.reposition()
	ResizeToFit(n, 0)
	return true
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

func (n *funcNode) Repaint() {
	n.ViewBase.Repaint()
	select {
	case n.awaken <- true:
	default:
	}
}
