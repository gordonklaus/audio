package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
	"fmt"
	"time"
)

type funcNode struct {
	*ViewBase
	AggregateMouser
	obj                     types.Object
	pkgRefs                 map[*types.Package]int
	inputsNode, outputsNode *portsNode
	funcblk                 *block
	awaken, stop            chan struct{}
	done                    func()
}

func newFuncNode(obj types.Object) *funcNode {
	n := &funcNode{obj: obj}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(n), NewMover(n)}
	n.pkgRefs = map[*types.Package]int{}
	n.funcblk = newBlock(n)
	n.inputsNode = newInputsNode()
	n.inputsNode.editable = true
	n.funcblk.addNode(n.inputsNode)
	n.outputsNode = newOutputsNode()
	n.outputsNode.editable = true
	n.funcblk.addNode(n.outputsNode)
	n.Add(n.funcblk)
	n.awaken = make(chan struct{}, 1)
	n.stop = make(chan struct{}, 1)

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
	n.stop <- struct{}{}
	n.wakeUp()
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

func (n funcNode) block() *block           { return nil }
func (n funcNode) setBlock(b *block)       {}
func (n funcNode) inputs() []*port         { return nil }
func (n funcNode) outputs() []*port        { return nil }
func (n funcNode) inConns() []*connection  { return nil }
func (n funcNode) outConns() []*connection { return nil }

const fps = 60

func (n *funcNode) animate() {
	updated := make(chan bool)
	for {
		next := time.After(time.Second / fps)
		Do(n, func() {
			updated <- n.update()
		})

		if <-updated {
			<-next
		} else {
			<-n.awaken
		}
		select {
		case <-n.stop:
			return
		default:
		}
	}
}

func (n *funcNode) wakeUp() {
	select {
	case n.awaken <- struct{}{}:
	default:
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

func (n *funcNode) TookKeyFocus() { SetKeyFocus(n.funcblk) }

func (n *funcNode) KeyPress(event KeyEvent) {
	switch event.Key {
	case KeyF1:
		saveFunc(*n)
	default:
		n.ViewBase.KeyPress(event)
	}
}
