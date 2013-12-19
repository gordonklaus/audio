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

	obj     types.Object
	pkgRefs map[*types.Package]int
	done    func()

	newArrange chan *nodeArrange
	stop       chan struct{}
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

	n.newArrange = make(chan *nodeArrange, 1)
	n.stop = make(chan struct{})

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

func (n funcNode) lit() bool { return n.obj == nil } // TODO: make field, as this is constant

func (n *funcNode) Close() {
	if !n.lit() {
		saveFunc(n)
		close(n.stop)
		n.rearrange()
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
	if n.output != nil {
		p = append(p, n.output)
	}
	return
}
func (n funcNode) inConns() []*connection { return n.funcblk.inConns() }
func (n funcNode) outConns() []*connection {
	c := n.funcblk.outConns()
	if n.output != nil {
		c = append(c, n.output.conns...)
	}
	return c
}

const fps = 60.0

func (n *funcNode) animate() {
	animate := make(chan *nodeArrange, 1)
	go n.arrange(animate)
	a := newNodeArrange(n, nil, map[*port]*portArrange{})
	done := make(chan bool, 1)
	for {
		next := time.After(time.Second / fps)
		Do(n, func() {
			c := CenterInParent(n)
			done <- a.animate()
			MoveCenter(n, c)
		})
		if <-done {
			next = nil
		}
		select {
		case <-next:
		case a = <-animate:
		case <-n.stop:
			return
		}
	}
}

func (n *funcNode) arrange(animate chan *nodeArrange) {
	a := newNodeArrange(n, nil, map[*port]*portArrange{})
	for {
		if a.arrange() {
			animate <- a
			a = <-n.newArrange
		}
		select {
		case <-n.stop:
			return
		default:
		}
	}
}

func (n *funcNode) rearrange() {
	select {
	case <-n.newArrange:
	default:
	}
	n.newArrange <- newNodeArrange(n, nil, map[*port]*portArrange{})
}

func (n *funcNode) Move(p Point) {
	n.ViewBase.Move(p)
	nodeMoved(n)
}

func (n *funcNode) TookKeyFocus() { n.focused = true; Repaint(n) }
func (n *funcNode) LostKeyFocus() { n.focused = false; Repaint(n) }

func (n *funcNode) KeyPress(event KeyEvent) {
	if event.Command && event.Key == KeyS && !n.lit() {
		saveFunc(n)
	} else if event.Key == KeyLeft && n.lit() {
		SetKeyFocus(n.outputsNode)
	} else {
		n.ViewBase.KeyPress(event)
	}
}

func (n funcNode) Paint() {
	SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[n.focused])
	if n.lit() {
		DrawLine(Pt(-portSize/2, 0), Pt(portSize/2, 0))
	}
}
