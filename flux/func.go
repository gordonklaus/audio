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
	blk                     *block
	output                  *port
	inputsNode, outputsNode *portsNode
	funcblk                 *block
	focused                 bool

	obj          types.Object
	pkgRefs      map[*types.Package]int
	awaken, stop chan struct{}
	done         func()
}

func newFuncLiteralNode() *funcNode {
	n := newFuncNode()
	n.output = newOutput(n, &types.Var{Type: &types.Signature{}})
	n.Add(n.output)
	return n
}

// helper constructor -- does not produce a complete funcNode
func newFuncNode() *funcNode {
	n := &funcNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(n), NewMover(n)}
	n.funcblk = newBlock(n)
	n.inputsNode = newInputsNode()
	n.inputsNode.editable = true
	n.funcblk.addNode(n.inputsNode)
	n.outputsNode = newOutputsNode()
	n.outputsNode.editable = true
	n.funcblk.addNode(n.outputsNode)
	n.Add(n.funcblk)
	return n
}

func (n funcNode) lit() bool { return n.obj == nil }

func (n *funcNode) Close() {
	if !n.lit() {
		saveFunc(n)
		n.stop <- struct{}{}
		n.wakeUp()
		n.done()
	}
	n.ViewBase.Close()
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

func (n funcNode) block() *block      { return n.blk }
func (n *funcNode) setBlock(b *block) { n.blk = b }
func (n funcNode) inputs() []*port    { return nil }
func (n funcNode) outputs() (p []*port) {
	if n.lit() {
		p = append(p, n.output)
	}
	return
}
func (n funcNode) inConns() []*connection { return n.funcblk.inConns() }
func (n funcNode) outConns() []*connection {
	c := n.funcblk.outConns()
	if n.lit() {
		c = append(c, n.output.conns...)
	}
	return c
}

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
	b := n.funcblk
	if !b.update() {
		return false
	}
	b.Move(Pt(0, -Height(b)/2))
	if n.lit() {
		MoveCenter(n.output, Pt(Width(b)+portSize, 0))
	}
	n.inputsNode.reposition()
	n.outputsNode.reposition()
	c := CenterInParent(n)
	ResizeToFit(n, 0)
	MoveCenter(n, c)
	return true
}

func (n *funcNode) Move(p Point) {
	n.ViewBase.Move(p)
	nodeMoved(n)
}

func (n *funcNode) TookKeyFocus() {
	if n.lit() {
		n.focused = true
		Repaint(n)
	} else {
		SetKeyFocus(n.funcblk)
	}
}
func (n *funcNode) LostKeyFocus() { n.focused = false; Repaint(n) }

func (n *funcNode) KeyPress(event KeyEvent) {
	switch event.Key {
	case KeyF1:
		if !n.lit() {
			saveFunc(n)
		} else {
			n.ViewBase.KeyPress(event)
		}
	default:
		n.ViewBase.KeyPress(event)
	}
}

func (n funcNode) Paint() {
	SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[n.focused])
	if n.lit() {
		x := Width(n.funcblk)
		DrawLine(Pt(x, 0), Pt(x+portSize, 0))
	}
}
