// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type ifNode struct {
	*ViewBase
	AggregateMouser
	blk           *block
	input         *port
	seqIn, seqOut *port
	falseblk      *block
	trueblk       *block
	focused       bool
}

func newIfNode(arranged blockchan) *ifNode {
	n := &ifNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(n), NewMover(n)}
	n.input = newInput(n, newVar("", types.Typ[types.Bool]))
	n.Add(n.input)
	n.falseblk = newBlock(n, arranged)
	n.trueblk = newBlock(n, arranged)

	n.seqIn = newInput(n, newVar("seq", seqType))
	MoveCenter(n.seqIn, Pt(-portSize, 0))
	n.Add(n.seqIn)
	n.seqOut = newOutput(n, newVar("seq", seqType))
	MoveCenter(n.seqOut, Pt(portSize, 0))
	n.Add(n.seqOut)

	MoveCenter(n.input, Pt(-2*portSize, 0))
	return n
}

func (n ifNode) block() *block      { return n.blk }
func (n *ifNode) setBlock(b *block) { n.blk = b }
func (n ifNode) inputs() []*port    { return []*port{n.seqIn, n.input} }
func (n ifNode) outputs() []*port   { return []*port{n.seqOut} }

func (n ifNode) inConns() []*connection {
	return append(append(append(n.input.conns, n.seqIn.conns...), n.falseblk.inConns()...), n.trueblk.inConns()...)
}

func (n ifNode) outConns() []*connection {
	return append(append(n.seqOut.conns, n.falseblk.outConns()...), n.trueblk.outConns()...)
}

func (n *ifNode) Move(p Point) {
	n.ViewBase.Move(p)
	nodeMoved(n)
}

func (n *ifNode) TookKeyFocus() { n.focused = true; Repaint(n) }
func (n *ifNode) LostKeyFocus() { n.focused = false; Repaint(n) }

func (n *ifNode) KeyPress(event KeyEvent) {
	if b, ok := KeyFocus(n).(*block); ok {
		if b == n.trueblk && event.Key == KeyDown || b == n.falseblk && event.Key == KeyUp {
			SetKeyFocus(n)
		}
		return
	}

	switch k := event.Key; k {
	case KeyDown, KeyUp:
		if k == KeyUp && len(n.trueblk.nodes) == 0 {
			SetKeyFocus(n.trueblk)
		} else if k == KeyDown && len(n.falseblk.nodes) == 0 {
			SetKeyFocus(n.falseblk)
		} else {
			n.blk.focusNearestView(n, k)
		}
	default:
		n.ViewBase.KeyPress(event)
	}
}

func (n ifNode) Paint() {
	SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[n.focused])
	DrawLine(CenterInParent(n.input), CenterInParent(n.seqOut))
	DrawLine(Pt(0, -4), Pt(0, 4))
}
