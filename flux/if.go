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
	seqIn, seqOut *port

	cond    []*port
	blocks  []*block
	focused int

	arranged blockchan
}

func newIfNode(arranged blockchan) *ifNode {
	n := &ifNode{focused: -1, arranged: arranged}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(n), NewMover(n)}

	n.seqIn = newInput(n, newVar("seq", seqType))
	MoveCenter(n.seqIn, Pt(0, -portSize))
	n.Add(n.seqIn)
	n.seqOut = newOutput(n, newVar("seq", seqType))
	n.Add(n.seqOut)

	return n
}

func (n *ifNode) newBlock() (b *block, cond *port) {
	b = newBlock(n, n.arranged)
	n.blocks = append(n.blocks, b)

	cond = newInput(n, newVar("", nil))
	cond.connsChanged = func() {
		cond.setType(untypedToTyped(inputType(cond)))
	}
	n.cond = append(n.cond, cond)
	n.Add(cond)

	rearrange(n.blk)
	return
}

func (n *ifNode) connectable(t types.Type, dst *port) bool {
	b, ok := underlying(t).(*types.Basic)
	return ok && b.Info&types.IsBoolean != 0
}

func (n ifNode) block() *block      { return n.blk }
func (n *ifNode) setBlock(b *block) { n.blk = b }
func (n ifNode) inputs() []*port    { return append([]*port{n.seqIn}, n.cond...) }
func (n ifNode) outputs() []*port   { return []*port{n.seqOut} }

func (n ifNode) inConns() []*connection {
	c := n.seqIn.conns
	for _, p := range n.cond {
		c = append(c, p.conns...)
	}
	for _, b := range n.blocks {
		c = append(c, b.inConns()...)
	}
	return c
}

func (n ifNode) outConns() []*connection {
	c := n.seqOut.conns
	for _, b := range n.blocks {
		c = append(c, b.outConns()...)
	}
	return c
}

func (n *ifNode) focus(i int) {
	n.focused = i
	SetKeyFocus(n)
	Repaint(n)
}

func (n *ifNode) focusFrom(p *port) {
	for i, p2 := range n.cond {
		if p2 == p {
			n.focus(i)
			break
		}
	}
}

func (n *ifNode) Move(p Point) {
	n.ViewBase.Move(p)
	nodeMoved(n)
}

func (n *ifNode) TookKeyFocus() {
	if n.focused < 0 {
		n.focused = 0
	}
	Repaint(n)
}
func (n *ifNode) LostKeyFocus() { n.focused = -1; Repaint(n) }

func (n *ifNode) KeyPress(event KeyEvent) {
	if b, ok := KeyFocus(n).(*block); ok {
		for i, b2 := range n.blocks {
			if b == b2 && event.Key == KeyUp {
				n.focus(i)
			}
		}
		return
	}

	switch event.Key {
	case KeyUp:
		SetKeyFocus(n.cond[n.focused])
	case KeyDown:
		b := n.blocks[n.focused]
		if len(b.nodes) == 0 {
			SetKeyFocus(b)
		} else {
			n.blk.focusNearestView(MapToParent(n, Pt(CenterInParent(b).X, 0)), event.Key)
		}
	case KeyLeft:
		if n.focused > 0 {
			n.focus(n.focused - 1)
		}
	case KeyRight:
		if n.focused < len(n.blocks)-1 {
			n.focus(n.focused + 1)
		}
	case KeyComma:
		n.newBlock()
		n.focus(len(n.blocks) - 1)
	case KeyBackspace, KeyDelete:
		if len(n.blocks) == 1 {
			n.ViewBase.KeyPress(event)
			return
		}
		i := n.focused
		n.blocks[i].close()
		for _, c := range n.cond[i].conns {
			c.blk.removeConn(c)
		}
		n.Remove(n.cond[i])
		n.Remove(n.blocks[i])
		n.cond = append(n.cond[:i], n.cond[i+1:]...)
		n.blocks = append(n.blocks[:i], n.blocks[i+1:]...)
		if i > 0 && (event.Key == KeyBackspace || i == len(n.blocks)) {
			i--
		}
		n.focus(i)
		rearrange(n.blk)
	default:
		n.ViewBase.KeyPress(event)
	}
}

func (n ifNode) Paint() {
	for i, b := range n.blocks {
		SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[i == n.focused])
		r := RectInParent(b)
		left := r.Min.X - portSize/2
		right := r.Max.X + portSize/2
		center := r.Center().X
		if i == 0 {
			left = center
		}
		if i == len(n.blocks)-1 {
			right = center
		}
		DrawLine(Pt(left, portSize), Pt(right, portSize))
		DrawLine(Pt(center, 2*portSize), Pt(center, 0))
	}
}
