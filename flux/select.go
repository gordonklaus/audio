// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
)

type selectNode struct {
	*ViewBase
	AggregateMouser
	name *Text

	blk           *block
	seqIn, seqOut *port

	cases   []*selectCase
	focused int

	arranged blockchan
}

func newSelectNode(arranged blockchan) *selectNode {
	n := &selectNode{focused: -2, arranged: arranged}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(n), NewMover(n)}
	n.name = NewText("select")
	n.name.SetBackgroundColor(noColor)
	n.name.SetTextColor(color(special{}, true, false))
	n.name.SetFrameSize(3)
	n.name.Move(Pt(-Width(n.name)-portSize/2, -portSize))
	n.Add(n.name)

	n.seqIn = newInput(n, newVar("seq", seqType))
	n.Add(n.seqIn)
	n.seqOut = newOutput(n, newVar("seq", seqType))
	n.Add(n.seqOut)

	return n
}

func (n *selectNode) newCase() *selectCase {
	c := &selectCase{n: n, send: true}
	c.ch = newInput(n, nil)
	c.ch.connsChanged = c.connsChanged
	n.Add(c.ch)
	c.elem = newInput(n, nil)
	n.Add(c.elem)
	c.blk = newBlock(n, n.arranged)
	n.cases = append(n.cases, c)
	rearrange(n.blk)
	return c
}

func (n *selectNode) connectable(t types.Type, dst *port) bool {
	for _, c := range n.cases {
		if dst == c.ch {
			t, ok := underlying(t).(*types.Chan)
			if !ok {
				return false
			}
			switch t.Dir {
			case types.SendRecv:
				return true
			case types.SendOnly:
				return c.send
			case types.RecvOnly:
				return !c.send
			}
		}
		if dst == c.elem {
			if inputType(c.ch) == nil {
				// A connection whose destination is being edited may currently be connected to n.ch.  It is temporarily disconnected during the call to connectable, but inputs (such as n.elem) with dependent types are not updated, so we have to specifically check for this case here.
				return false
			}
			return assignable(t, dst.obj.Type)
		}
	}
	panic("unreachable")
}

func (n selectNode) block() *block      { return n.blk }
func (n *selectNode) setBlock(b *block) { n.blk = b }
func (n selectNode) inputs() []*port {
	p := []*port{n.seqIn}
	for _, c := range n.cases {
		if c.ch != nil {
			p = append(p, c.ch)
			if c.elem != nil {
				p = append(p, c.elem)
			}
		}
	}
	return p
}
func (n selectNode) outputs() []*port { return []*port{n.seqOut} }

func (n selectNode) inConns() []*connection {
	conns := n.seqIn.conns
	for _, c := range n.cases {
		if c.ch != nil {
			conns = append(conns, c.ch.conns...)
			if c.elem != nil {
				conns = append(conns, c.elem.conns...)
			}
		}
		conns = append(conns, c.blk.inConns()...)
	}
	return conns
}

func (n selectNode) outConns() []*connection {
	conns := n.seqOut.conns
	for _, c := range n.cases {
		conns = append(conns, c.blk.outConns()...)
	}
	return conns
}

func (n *selectNode) focus(i int) {
	n.focused = i
	if i == -1 {
		n.name.SetFrameColor(focusColor)
		panTo(n.name, Center(n.name))
	} else {
		n.name.SetFrameColor(noColor)
		panTo(n, Pt(CenterInParent(n.cases[i].blk).X, 0))
	}
	SetKeyFocus(n)
	Repaint(n)
}

func (n *selectNode) focusFrom(v View) {
	for i, c := range n.cases {
		switch v {
		case c.ch, c.elem, c.blk, c.elemOk:
			n.focus(i)
			return
		}
	}
}

func (n *selectNode) removePort(p *port) {
	var c *selectCase
	var i int
	for i2, c2 := range n.cases {
		if c2.ch == nil { // only one default case is allowed
			return
		}
		if p == c2.ch {
			c, i = c2, i2
		}
	}
	if c != nil {
		c.setDefault()
		n.focus(i)
	}
}

func (n *selectNode) Move(p Point) {
	n.ViewBase.Move(p)
	nodeMoved(n)
}

func (n *selectNode) TookKeyFocus() {
	if n.focused < -1 {
		n.focused = -1
	}
	Repaint(n)
	if n.focused == -1 {
		n.name.SetFrameColor(focusColor)
		panTo(n.name, Center(n.name))
	} else {
		n.name.SetFrameColor(noColor)
		panTo(n, Pt(CenterInParent(n.cases[n.focused].blk).X, 0))
	}
}
func (n *selectNode) LostKeyFocus() {
	n.focused = -2
	Repaint(n)
	n.name.SetFrameColor(noColor)
}

func (n *selectNode) KeyPress(event KeyEvent) {
	switch event.Key {
	case KeyUp, KeyDown, KeyLeft, KeyRight:
		if event.Alt {
			n.ViewBase.KeyPress(event)
			return
		}
	}

	i := n.focused
	var c *selectCase
	if i >= 0 {
		c = n.cases[i]
	}
	switch event.Key {
	case KeyUp:
		if i >= 0 && c.ch != nil {
			c.ch.focusMiddle()
		}
	case KeyDown:
		if i >= 0 {
			if c.elemOk != nil {
				SetKeyFocus(c.elemOk)
			} else if len(c.blk.nodes) == 0 {
				SetKeyFocus(c.blk)
			} else {
				n.blk.focusNearestView(MapToParent(n, Pt(CenterInParent(c.blk).X, 0)), event.Key)
			}
		}
	case KeyLeft:
		if i > -1 {
			n.focus(i - 1)
		}
	case KeyRight:
		if i < len(n.cases)-1 {
			n.focus(i + 1)
		}
	case KeyEqual:
		if i >= 0 && c.ch != nil {
			if t := c.ch.obj.Type; t == nil || underlying(t).(*types.Chan).Dir == types.SendRecv {
				c.send = !c.send
				c.connsChanged()
			}
		}
	case KeyComma:
		n.newCase()
		n.focus(len(n.cases) - 1)
	case KeyBackspace, KeyDelete:
		if i == -1 {
			n.ViewBase.KeyPress(event)
			return
		}
		c := n.cases[i]
		c.blk.close()
		if c.ch != nil {
			for _, c := range c.ch.conns {
				c.blk.removeConn(c)
			}
			n.Remove(c.ch)
			if c.elem != nil {
				for _, c := range c.elem.conns {
					c.blk.removeConn(c)
				}
				n.Remove(c.elem)
			}
		}
		n.Remove(c.blk)
		n.cases = append(n.cases[:i], n.cases[i+1:]...)
		if event.Key == KeyBackspace || i == len(n.cases) {
			i--
		}
		n.focus(i)
		rearrange(n.blk)
	default:
		n.ViewBase.KeyPress(event)
	}
}

func (n selectNode) Paint() {
	for i, c := range n.cases {
		r := RectInParent(c.blk)
		left := r.Min.X - portSize/2
		right := r.Max.X + portSize/2
		center := r.Center().X
		if i == 0 {
			left = center
		}
		if i == len(n.cases)-1 {
			right = center
		}
		origin := Pt(center, 0)
		SetColor(lineColor)
		SetLineWidth(3)
		DrawLine(Pt(left, 0), Pt(right, 0))
		DrawLine(origin, Pt(center, -portSize))
		if c.ch != nil {
			p := CenterInParent(c.ch)
			DrawBezier(origin, Pt(center, p.Y/2), Pt(p.X, p.Y/2), p)
			if c.elem != nil {
				p := CenterInParent(c.elem)
				DrawBezier(origin, Pt(center, p.Y/2), Pt(p.X, p.Y/2), p)
			}
		}
		if i == n.focused {
			SetPointSize(2 * portSize)
			SetColor(focusColor)
			DrawPoint(origin)
		}
	}
}

type selectCase struct {
	n        *selectNode
	send     bool
	ch, elem *port
	blk      *block
	elemOk   *portsNode
}

func (c *selectCase) setDefault() {
	for _, c := range c.ch.conns {
		c.blk.removeConn(c)
	}
	c.n.Remove(c.ch)
	c.ch = nil
	if c.elem != nil {
		for _, c := range c.elem.conns {
			c.blk.removeConn(c)
		}
		c.n.Remove(c.elem)
		c.elem = nil
	} else {
		c.blk.removeNode(c.elemOk)
		c.elemOk = nil
	}
	rearrange(c.n.blk)
}

func (c *selectCase) connsChanged() {
	if c.send && c.elem == nil {
		c.elem = newInput(c.n, nil)
		c.n.Add(c.elem)
		c.blk.removeNode(c.elemOk)
		c.elemOk = nil
		rearrange(c.n.blk)
	}
	if !c.send && c.elem != nil {
		for _, c := range c.elem.conns {
			c.blk.removeConn(c)
		}
		c.n.Remove(c.elem)
		c.elem = nil
		c.elemOk = newInputsNode()
		c.blk.addNode(c.elemOk)
		c.elemOk.newOutput(nil)
		c.elemOk.newOutput(newVar("ok", nil))
		rearrange(c.n.blk)
	}

	t := inputType(c.ch)
	var elem, ok types.Type
	if t != nil {
		elem = underlying(t).(*types.Chan).Elem
		ok = types.Typ[types.Bool]
	}
	c.ch.setType(t)
	if c.send {
		c.elem.setType(elem)
	} else {
		c.elemOk.outs[0].setType(elem)
		c.elemOk.outs[1].setType(ok)
	}
}
