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
	focused, bad bool
	connsChanged func()

	conntxt *Text
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
	p.valView = newValueView(v, nil) // TODO: pass currentPkg here so that named types are not package qualified in the current package and so that unexported fields in unnamed literals (rare) are hidden as appropriate
	Hide(p.valView)
	p.connsChanged = func() {}
	p.Add(p.valView)
	p.SetRect(ZR.Inset(-portSize / 2))
	p.setType(*p.valView.typ)

	p.conntxt = NewText("")
	p.conntxt.SetTextColor(lineColor)
	p.conntxt.SetBackgroundColor(noColor)
	p.conntxt.Validate = validateID
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
		p.valView.Move(Pt(-Width(p.valView)/2, -Height(p.valView)-12))
	} else {
		p.valView.Move(Pt(-Width(p.valView)/2, 12))
	}

	for i := 0; i < len(p.conns); {
		c := p.conns[i]
		if c.src != nil && c.dst != nil && !c.connectable(c.src, c.dst) {
			c.blk.removeConn(c)
		} else {
			if p.out && c.dst != nil {
				c.dst.connsChanged()
			}
			i++
		}
	}
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

// TODO: work with all of a node's in/out conns (including seq conns), not just a single port's conns.
func (p *port) focusNextConn(curƟ float64, dir int) {
	less := func(x, y float64) bool { return x < y }
	if p.out != (dir == KeyRight) {
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
		if p.out == (dir == KeyRight) {
			Ɵ = -Ɵ
		}
		p.focusNextConn(Ɵ, dir)
	}
}

func (p *port) next(dir int) *port {
	if p.obj.Type == seqType {
		// TODO: return nearest
		return nil
	}

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
	if dir == KeyLeft {
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

func (p *port) TookKeyFocus() {
	p.focused = true
	Repaint(p)
	Show(p.valView)
	panTo(p, ZP)
}

func (p *port) LostKeyFocus() {
	p.focused = false
	Repaint(p)
	Hide(p.valView)
}

func (p *port) KeyPress(event KeyEvent) {
	if event.Alt && !event.Shift {
		switch event.Key {
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			p.node.block().focusNearestView(p, event.Key)
			return
		}
	}

	switch k := event.Key; k {
	case KeyUp, KeyDown:
		if p.out == (k == KeyDown) {
			p.focusMiddle()
		} else {
			if f, ok := p.node.(focuserFrom); ok {
				f.focusFrom(p)
			} else {
				SetKeyFocus(p.node)
			}
		}
	case KeyLeft, KeyRight:
		if p := p.next(k); p != nil {
			SetKeyFocus(p)
		}
	case KeyPeriod:
		if event.Ctrl {
			p.ViewBase.KeyPress(event)
		}
		if !p.out {
			break
		}
		t := p.obj.Type
		if pt, ok := indirect(t); ok { // automatically dereference **t
			if _, ok := indirect(pt); ok {
				t = pt
			}
		}
		b := newSelectionBrowser(t, p)
		if len(b.objs) == 0 {
			break
		}
		p.Add(b)
		b.accepted = func(obj types.Object) {
			b.Close()
			n := p.node.block().newNode(obj, b.funcAsVal, "")
			c := newConnection()
			c.setSrc(p)
			c.setDst(ins(n)[0])
		}
		b.canceled = func() {
			b.Close()
			SetKeyFocus(p)
		}
		SetKeyFocus(b)
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
		if f, ok := p.node.(focuserFrom); ok {
			f.focusFrom(p)
		} else {
			SetKeyFocus(p.node)
		}
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
	if m.Press {
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
}

func (p *port) Paint() {
	if p.focused {
		SetColor(focusColor)
		SetPointSize(portSize)
		DrawPoint(ZP)
	}
	if p.bad {
		SetColor(Color{1, 0, 0, 1})
		SetLineWidth(3)
		r := Rect(p)
		DrawLine(r.Min, r.Max)
		DrawLine(Pt(r.Min.X, r.Max.Y), Pt(r.Max.X, r.Min.Y))
	}
}
