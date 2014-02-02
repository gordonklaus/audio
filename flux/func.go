// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	"fmt"
)

type funcNode struct {
	*ViewBase
	AggregateMouser
	blk                     *block
	output                  *port
	funcblk                 *block
	inputsNode, outputsNode *portsNode
	focused                 bool

	obj     types.Object
	literal bool
	pkgRefs map[*types.Package]int
	done    func()

	animate blockchan
	stop    stopchan
}

func newFuncNode(obj types.Object, arranged blockchan) *funcNode {
	n := &funcNode{obj: obj, literal: obj == nil}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(n), NewMover(n)}
	if n.literal {
		n.output = newOutput(n, newVar("", &types.Signature{}))
		n.Add(n.output)
	} else {
		n.pkgRefs = map[*types.Package]int{}
		n.animate = make(blockchan)
		n.stop = make(stopchan)
		arranged = n.animate
	}
	n.funcblk = newBlock(n, arranged)
	n.inputsNode = newInputsNode()
	n.inputsNode.editable = true
	n.funcblk.addNode(n.inputsNode)
	n.outputsNode = newOutputsNode()
	n.outputsNode.editable = true
	n.funcblk.addNode(n.outputsNode)
	return n
}

func (n *funcNode) Close() {
	if !n.literal {
		n.funcblk.close()
		n.stop.stop()
		saveFunc(n)
		n.done()
	}
	n.ViewBase.Close()
}

func (n *funcNode) sig() *types.Signature {
	obj := n.obj
	if obj == nil {
		obj = n.output.obj
	}
	return obj.GetType().(*types.Signature)
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
		walkType(x, func(t *types.Named) { n.addPkgRef(t.Obj) })
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
		walkType(x, func(t *types.Named) { n.subPkgRef(t.Obj) })
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

func (n *funcNode) Move(p Point) {
	n.ViewBase.Move(p)
	nodeMoved(n)
}

func (n *funcNode) TookKeyFocus() { n.focused = true; Repaint(n) }
func (n *funcNode) LostKeyFocus() { n.focused = false; Repaint(n) }

func (n *funcNode) KeyPress(event KeyEvent) {
	if event.Command && event.Key == KeyS && !n.literal {
		saveFunc(n)
	} else if event.Key == KeyLeft && n.literal {
		SetKeyFocus(n.outputsNode)
	} else {
		n.ViewBase.KeyPress(event)
	}
}

func (n funcNode) Paint() {
	SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[n.focused])
	if n.literal {
		DrawLine(Pt(-portSize/2, 0), Pt(portSize/2, 0))
	}
}
