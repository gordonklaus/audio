// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type loopNode struct {
	*ViewBase
	AggregateMouser
	blk           *block
	input         *port
	seqIn, seqOut *port
	loopblk       *block
	inputsNode    *portsNode
	focused       bool
}

func newLoopNode(arranged blockchan) *loopNode {
	n := &loopNode{}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(n), NewMover(n)}
	n.input = newInput(n, nil)
	n.input.connsChanged = n.connsChanged
	MoveCenter(n.input, Pt(0, portSize))
	n.Add(n.input)

	n.seqIn = newInput(n, newVar("seq", seqType))
	MoveCenter(n.seqIn, Pt(0, -portSize))
	n.Add(n.seqIn)
	n.seqOut = newOutput(n, newVar("seq", seqType))
	n.Add(n.seqOut)

	n.loopblk = newBlock(n, arranged)
	n.inputsNode = newInputsNode()
	n.inputsNode.newOutput(nil)
	n.loopblk.addNode(n.inputsNode)
	n.connsChanged()
	return n
}

func (n loopNode) block() *block      { return n.blk }
func (n *loopNode) setBlock(b *block) { n.blk = b }
func (n loopNode) inputs() []*port    { return []*port{n.seqIn, n.input} }
func (n loopNode) outputs() []*port   { return []*port{n.seqOut} }
func (n loopNode) inConns() []*connection {
	return append(n.seqIn.conns, append(n.input.conns, n.loopblk.inConns()...)...)
}
func (n loopNode) outConns() []*connection {
	return append(n.seqOut.conns, n.loopblk.outConns()...)
}

func (n *loopNode) connectable(t types.Type, dst *port) bool {
	ok := false
	switch t := underlying(t).(type) {
	case *types.Basic:
		ok = t.Info&(types.IsInteger|types.IsString) != 0
	case *types.Array, *types.Slice, *types.Map, *types.Chan:
		ok = true
	case *types.Pointer:
		_, ok = underlying(t.Elem).(*types.Array)
	}
	return ok
}

func (n *loopNode) connsChanged() {
	t := inputType(n.input)
	var key, elem types.Type
	key = types.Typ[types.Int]
	elemPort := true
	switch t := underlying(t).(type) {
	case *types.Basic:
		if t.Info&types.IsString != 0 {
			elem = types.Typ[types.Rune]
		} else {
			key = t
			elemPort = false
		}
	case *types.Array:
		elem = t.Elem
	case *types.Pointer:
		elem = &types.Pointer{Elem: underlying(t.Elem).(*types.Array).Elem}
	case *types.Slice:
		elem = &types.Pointer{Elem: t.Elem}
	case *types.Map:
		key, elem = t.Key, t.Elem
	case *types.Chan:
		key = t.Elem
		elemPort = false
	case nil:
		elemPort = false
	}

	in := n.inputsNode
	if elemPort && len(in.outs) == 1 {
		in.newOutput(nil)
	}
	if !elemPort && len(in.outs) == 2 {
		in.removePortBase(in.outs[1])
	}

	n.input.setType(t)
	in.outs[0].setType(key)
	if elemPort {
		in.outs[1].setType(elem)
	}
}

func (n *loopNode) Move(p Point) {
	n.ViewBase.Move(p)
	nodeMoved(n)
}

func (n *loopNode) TookKeyFocus() {
	n.focused = true
	Repaint(n)
	panTo(n, ZP)
}

func (n *loopNode) LostKeyFocus() {
	n.focused = false
	Repaint(n)
}

func (n *loopNode) KeyPress(event KeyEvent) {
	switch event.Key {
	case KeyUp:
		if event.Alt && event.Shift {
			SetKeyFocus(n.seqIn)
		} else {
			SetKeyFocus(n.input)
		}
	case KeyDown:
		if event.Alt && event.Shift {
			SetKeyFocus(n.seqOut)
		} else {
			SetKeyFocus(n.inputsNode)
		}
	default:
		n.ViewBase.KeyPress(event)
	}
}

func (n loopNode) Paint() {
	SetColor(lineColor)
	SetLineWidth(3)
	DrawLine(Pt(0, -portSize), Pt(0, portSize))
	if n.focused {
		SetPointSize(2 * portSize)
		SetColor(focusColor)
		DrawPoint(ZP)
	}
}
