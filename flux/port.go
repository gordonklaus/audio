// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	"go/ast"
	"go/parser"
	"math"
)

type port struct {
	*ViewBase
	out          bool
	node         node
	obj          *types.Var
	valView      *typeView
	conns        []*connection
	focused      bool
	connsChanged func()

	conntxt *TextBase
}

const portSize = 11.0

func newInput(n node, v *types.Var) *port  { return newPort(false, n, v) }
func newOutput(n node, v *types.Var) *port { return newPort(true, n, v) }
func newPort(out bool, n node, v *types.Var) *port {
	if v == nil {
		v = &types.Var{}
	}

	p := &port{out: out, node: n, obj: v}
	p.ViewBase = NewView(p)
	p.valView = newValueView(v)
	Hide(p.valView)
	p.connsChanged = func() {}
	p.Add(p.valView)
	p.SetRect(ZR.Inset(-portSize / 2))
	p.setType(*p.valView.typ)

	p.conntxt = NewText("")
	p.conntxt.Move(Pt(0, -Height(p.conntxt)/2))
	p.conntxt.SetBackgroundColor(Color{0, 0, 0, 0})
	p.conntxt.SetValidator(validateID)
	p.Add(p.conntxt)
	return p
}

func validateID(text *string) bool {
	x, err := parser.ParseExpr(*text)
	_, ok := x.(*ast.Ident)
	return *text == "" || err == nil && ok
}

func (p *port) setType(t types.Type) {
	p.valView.setType(t)
	if p.out {
		p.valView.Move(Pt(12, -Height(p.valView)/2))
	} else {
		p.valView.Move(Pt(-Width(p.valView)-12, -Height(p.valView)/2))
	}
}

func canConnect(src, dst *port, c *connection) bool {
	if src.out == dst.out {
		return false
	}
	for _, c := range src.conns {
		if c.dst == dst {
			return false
		}
	}
	// TODO: types.IsAssignableTo(src.obj.Type, dst.obj.Type) (handles seqType, too?)
	if (src.obj.Type == seqType) != (dst.obj.Type == seqType) {
		return false
	}
	if src.obj.Type == seqType && (src.node.block() != dst.node.block() || precedes(src.node, dst.node)) {
		return false
	}

	// TODO: recursive func literals
	// if f, ok := src.node.(*funcNode); ok && parentOrSelfInBlock(dst.node, f.funcblk) != nil {
	// 	return true
	// }

	n1, n2 := src.node, dst.node
	for b := n1.block(); ; n1, b = b.node, b.outer() {
		if n := b.find(n2); n != nil {
			n2 = n
			break
		}
	}
	if c.feedback {
		for b := n1.block(); ; b = b.outer() {
			if b == nil {
				return false
			}
			if _, ok := b.node.(*loopNode); ok {
				break
			}
		}
		n1, n2 = n2, n1
	} else if n1 == n2 {
		return false
	}
	return !precedes(n2, n1)
}

func precedes(n1, n2 node) bool {
	for _, dst := range dstsInBlock(n1) {
		if dst == n2 || precedes(dst, n2) {
			return true
		}
	}
	return false
}

func (p *port) connect(c *connection) {
	p.conns = append(p.conns, c)
}

func (p *port) disconnect(c *connection) {
	for i, c2 := range p.conns {
		if c2 == c {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			return
		}
	}
}

func (p *port) focusMiddle() {
	var conn *connection
	dist := 0.0
	for _, c := range p.conns {
		d := math.Abs(c.dstPt.Sub(c.srcPt).Angle())
		if conn == nil || d < dist {
			conn, dist = c, d
		}
	}
	if conn != nil {
		conn.focus(p.out)
	} else {
		SetKeyFocus(p)
	}
}

func (p *port) focusNextConn(curƟ float64, dir int) {
	less := func(x, y float64) bool { return x < y }
	if p.out != (dir == KeyUp) {
		less = func(x, y float64) bool { return x > y }
	}
	var conn *connection
	nearestƟ := 0.0
	for _, c := range p.conns {
		Ɵ := c.dstPt.Sub(c.srcPt).Angle()
		if Ɵ != curƟ && less(curƟ, Ɵ) && (conn == nil || less(Ɵ, nearestƟ)) {
			conn, nearestƟ = c, Ɵ
		}
	}
	if conn != nil {
		conn.focus(p.out)
		return
	}

	if p := p.next(dir); p != nil {
		Ɵ := math.Pi
		if p.out == (dir == KeyUp) {
			Ɵ = -Ɵ
		}
		p.focusNextConn(Ɵ, dir)
	}
}

func (p *port) next(dir int) *port {
	ports := ins(p.node)
	if p.out {
		ports = outs(p.node)
	}
	i := 0
	for j, p2 := range ports {
		if p2 == p {
			i = j
			break
		}
	}
	if dir == KeyUp {
		i--
	} else {
		i++
	}
	if i >= 0 && i < len(ports) {
		return ports[i]
	}
	return nil
}

func (p *port) Move(pt Point) {
	p.ViewBase.Move(pt)
	for _, c := range p.conns {
		c.reform()
	}
}

func (p *port) TookKeyFocus() { p.focused = true; Repaint(p); Show(p.valView) }
func (p *port) LostKeyFocus() { p.focused = false; Repaint(p); Hide(p.valView) }

func (p *port) KeyPress(event KeyEvent) {
	if event.Alt {
		switch event.Key {
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			p.node.block().focusNearestView(p, event.Key)
			return
		}
	}

	switch k := event.Key; k {
	case KeyLeft, KeyRight:
		if p.out == (k == KeyRight) {
			if p.obj.Type == seqType {
				ports := ins(p.node)
				if p.out {
					ports = outs(p.node)
				}
				if len(ports) > 0 {
					ports[len(ports)/2].focusMiddle()
				}
			} else {
				p.focusMiddle()
			}
		} else {
			SetKeyFocus(p.node)
		}
	case KeyDown:
		if p.obj.Type == seqType && len(p.conns) > 0 {
			p.conns[len(p.conns)-1].focus(p.out)
			return
		}
		fallthrough
	case KeyUp:
		if p := p.next(k); p != nil {
			SetKeyFocus(p)
		}
	case KeyEnter:
		c := newConnection()
		if p.out {
			c.setSrc(p)
			c.focus(false)
		} else {
			c.setDst(p)
			c.focus(true)
		}
		c.startEditing()
	case KeyBackspace, KeyDelete:
		if n, ok := p.node.(interface {
			removePort(*port)
		}); ok {
			n.removePort(p)
		} else {
			SetKeyFocus(p.node)
		}
	case KeyEscape:
		SetKeyFocus(p.node)
	default:
		if pn, ok := p.node.(*portsNode); ok && p.out && pn.outs[0] == p && event.Text == "*" {
			if t, ok := p.obj.Type.(*types.Pointer); ok {
				p.setType(t.Elem)
			} else {
				p.setType(&types.Pointer{Elem: p.obj.Type})
			}
		} else {
			p.ViewBase.KeyPress(event)
		}
	}
}

func (p *port) Mouse(m MouseEvent) {
	SetKeyFocus(p)
	c := newConnection()
	if p.out {
		c.setSrc(p)
		c.focus(false)
	} else {
		c.setDst(p)
		c.focus(true)
	}
	SetMouser(c, m.Button)
	c.startEditing()
}

func (p port) Paint() {
	if p.focused {
		SetColor(Color{.3, .3, .7, .5})
		SetPointSize(portSize)
		DrawPoint(ZP)
	}
}
