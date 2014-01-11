// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	. "code.google.com/p/gordon-go/gui"
	. "github.com/chsc/gogl/gl21"
	"math"
	"sort"
	"strings"
)

type connection struct {
	*ViewBase
	AggregateMouser
	blk      *block
	src      *port
	dst      *port
	feedback bool

	focused, focusSrc bool
	srcPt             Point
	dstPt             Point
	hidden            bool

	savedPort *port
	editing   bool
}

func newConnection() *connection {
	c := &connection{}
	c.ViewBase = NewView(c)
	c.AggregateMouser = AggregateMouser{NewClickFocuser(c)}
	return c
}

func (c connection) connected() bool { return c.src != nil && c.dst != nil }
func (c *connection) disconnect() {
	c.setSrc(nil)
	c.setDst(nil)
}

func (c *connection) setSrc(src *port) {
	txt := ""
	if c.src != nil {
		c.src.disconnect(c)
		c.src.connsChanged()
		txt = c.src.conntxt.GetText()
		c.updateSrcTxt("")
	}
	c.src = src
	if src != nil {
		src.connect(c)
		if c.dst != nil {
			c.src.connsChanged()
			c.dst.connsChanged()
		}
		c.updateSrcTxt(txt)
	}
	c.reblock()
	c.reform()
}

func (c *connection) setDst(dst *port) {
	if c.dst != nil {
		c.dst.disconnect(c)
		c.dst.connsChanged()
		c.updateDstTxt()
	}
	c.dst = dst
	if dst != nil {
		dst.connect(c)
		if c.src != nil {
			c.src.connsChanged()
			c.dst.connsChanged()
		}
		c.updateDstTxt()
	}
	c.reblock()
	c.reform()
}

func (c *connection) reblock() {
	var newblk *block
	switch {
	case c.src == nil && c.dst == nil:
		return
	case c.src == nil:
		newblk = c.dst.node.block()
	case c.dst == nil:
		newblk = c.src.node.block()
	default:
		for b := c.src.node.block(); ; b = b.outer() {
			if b.find(c.dst.node) != nil {
				newblk = b
				break
			}
		}
		if c.feedback {
			for {
				n := newblk.node
				newblk = n.block()
				if _, ok := n.(*loopNode); ok {
					break
				}
			}
		}
	}
	newblk.addConn(c)
}

func (c *connection) reform() {
	unconnectedOffset := Pt(32, 0)
	if c.feedback {
		unconnectedOffset.X = -96
	}
	if c.src != nil {
		c.srcPt = MapTo(c.src, Center(c.src), c.blk)
	}
	if c.dst != nil {
		c.dstPt = MapTo(c.dst, Center(c.dst), c.blk)
	} else {
		c.dstPt = c.srcPt.Add(unconnectedOffset)
	}
	if c.src == nil {
		c.srcPt = c.dstPt.Sub(unconnectedOffset)
	}

	var rect Rectangle
	if c.src != nil && c.src.obj.Type == seqType || c.dst != nil && c.dst.obj.Type == seqType {
		rect = Rectangle{Pt(c.srcPt.X, math.Min(c.srcPt.Y, c.dstPt.Y)-25), Pt(c.dstPt.X, math.Max(c.srcPt.Y, c.dstPt.Y))}
	} else {
		rect = Rectangle{c.srcPt, c.dstPt}.Canon()
	}

	pos := rect.Min

	// center the origin so that keyboard navigation works
	center := rect.Center()
	c.srcPt = c.srcPt.Sub(center)
	c.dstPt = c.dstPt.Sub(center)
	rect = rect.Sub(center)

	c.Move(pos)
	c.SetRect(rect)
}

func (c *connection) focus(focusSrc bool) {
	c.focusSrc = focusSrc
	c.updateTextColors()
	SetKeyFocus(c)
	Repaint(c)
}

func (c *connection) startEditing() {
	if c.focusSrc {
		if c.hidden {
			return // TODO: alternately, edit all hidden connections simultaneously?
		}
		c.savedPort = c.src
	} else {
		c.savedPort = c.dst
	}
	c.editing = true
	c.updateTextColors()
	c.reform()
}

func (c *connection) cancelEditing() {
	if c.editing {
		if c.focusSrc {
			c.setSrc(c.savedPort)
		} else {
			c.setDst(c.savedPort)
		}
		c.stopEditing()
	}
}

func (c *connection) stopEditing() {
	if c.editing {
		c.editing = false
		if c.connected() {
			c.reform()
		} else {
			p := c.src
			if c.focusSrc {
				p = c.dst
			}
			c.blk.removeConn(c)
			SetKeyFocus(p)
		}
		c.updateTextColors()
	}
}

func (c *connection) toggleHidden() {
	c.hidden = !c.hidden
	rearrange(c.blk)
	if c.hidden {
		Hide(c)
		if srctxt := c.src.conntxt; srctxt.GetText() == "" {
			srctxt.TextChanged = func(string) {
				for _, c := range c.src.conns {
					if c.hidden {
						c.updateDstTxt()
					}
				}
			}
			srctxt.Accept = func(name string) {
				names := map[string]bool{"": true}
				c.blk.outermost().walk(nil, nil, func(conn *connection) {
					if conn != c {
						names[conn.src.conntxt.GetText()] = true
					}
				})
				if names[name] {
					SetKeyFocus(srctxt)
				} else {
					c.focus(c.focusSrc)
				}
			}
			srctxt.Reject = func() {
				c.toggleHidden()
				c.focus(c.focusSrc)
			}
			SetKeyFocus(srctxt)
		}
	} else {
		Show(c)
	}
	c.updateSrcTxt("")
	c.updateDstTxt()
}

func (c *connection) updateSrcTxt(txt string) {
	anyHidden := false
	for _, c := range c.src.conns {
		if c.hidden {
			anyHidden = true
		}
	}
	if !anyHidden {
		c.src.conntxt.SetText("")
	} else if txt != "" {
		c.src.conntxt.SetText(txt)
	}
	c.src.updateTextColor()
}

func (c *connection) updateDstTxt() {
	if c.dst == nil {
		return
	}
	nameset := map[string]bool{}
	for _, c := range c.dst.conns {
		if c.hidden {
			nameset[c.src.conntxt.GetText()] = true
		}
	}
	names := []string{}
	for n := range nameset {
		names = append(names, n)
	}
	sort.StringSlice(names).Sort()
	dsttxt := c.dst.conntxt
	dsttxt.SetText(strings.Join(names, ","))
	dsttxt.Move(Pt(-Width(dsttxt), -Height(dsttxt)/2))
	c.dst.updateTextColor()
}

func (p *port) updateTextColor() {
	if p == nil {
		return
	}
	var focused, focusSrc, editing bool
	for _, c := range p.conns {
		if c.focused && c.hidden {
			focused = true
			focusSrc = c.focusSrc
			editing = c.editing
		}
	}
	c := Color{.6, .6, .6, 1}
	if focused {
		c = Color{.5, .5, .8, 1}
		if p.out == focusSrc {
			c = Color{.3, .3, .7, 1}
			if editing {
				c = Color{1, .5, 0, .5}
			}
		}
	}
	p.conntxt.SetTextColor(c)
}

func (c *connection) updateTextColors() {
	c.src.updateTextColor()
	c.dst.updateTextColor()
}

func (c *connection) TookKeyFocus() {
	c.focused = true
	c.updateTextColors()
	Repaint(c)
}

func (c *connection) LostKeyFocus() {
	c.cancelEditing()
	c.focused = false
	c.updateTextColors()
	Repaint(c)
}

func (c *connection) KeyPress(event KeyEvent) {
	if c.editing {
		switch event.Key {
		case KeyBackslash:
			if c.src == nil || c.dst == nil {
				c.feedback = !c.feedback
				c.reform()
			}
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			b := c.blk.outermost()
			ports := []View{}
			p1 := c.src
			if c.focusSrc {
				p1 = c.dst
			}
			for _, n := range b.allNodes() {
				p := n.inputs()
				if c.focusSrc {
					p = n.outputs()
				}
				for _, p2 := range p {
					src, dst := p1, p2
					if c.focusSrc {
						src, dst = dst, src
					}
					if canConnect(src, dst, c) {
						ports = append(ports, p2)
					}
				}
			}

			pt := c.dstPt
			if c.focusSrc {
				pt = c.srcPt
			}
			v := nearestView(b, ports, MapTo(c, pt, b), event.Key)
			if p, ok := v.(*port); ok {
				if c.focusSrc {
					c.setSrc(p)
				} else {
					c.setDst(p)
				}
			}
		case KeyEnter:
			c.stopEditing()
		case KeyEscape:
			c.cancelEditing()
		}
		return
	}

	if event.Alt {
		switch event.Key {
		case KeyLeft, KeyRight, KeyDown, KeyUp:
			p := c.dst
			if c.focusSrc {
				p = c.src
			}
			c.blk.focusNearestView(p, event.Key)
			return
		}
	}

	switch event.Key {
	case KeyLeft:
		if c.focusSrc {
			ins := ins(c.src.node)
			if len(ins) > 0 {
				ins[len(ins)/2].focusMiddle()
			} else {
				SetKeyFocus(c.src.node)
			}
		} else {
			c.focus(true)
		}
	case KeyRight:
		if c.focusSrc {
			c.focus(false)
		} else {
			outs := outs(c.dst.node)
			if len(outs) > 0 {
				outs[len(outs)/2].focusMiddle()
			} else {
				SetKeyFocus(c.dst.node)
			}
		}
	case KeyDown, KeyUp:
		p := c.dst
		if c.focusSrc {
			p = c.src
		}
		p.focusNextConn(c.dstPt.Sub(c.srcPt).Angle(), event.Key)
	case KeyBackspace:
		SetKeyFocus(c.src)
		c.blk.removeConn(c)
	case KeyDelete:
		SetKeyFocus(c.dst)
		c.blk.removeConn(c)
	case KeyEnter:
		c.startEditing()
	case KeyEscape:
		if c.focusSrc {
			SetKeyFocus(c.src)
		} else {
			SetKeyFocus(c.dst)
		}
	default:
		if event.Text == "_" {
			if c.src.obj.Type != seqType {
				c.toggleHidden()
			}
		} else {
			c.ViewBase.KeyPress(event)
		}
	}
}

func (c *connection) Mouse(m MouseEvent) {
	if m.Press {
		if m.Pos.Sub(c.srcPt).Len() < 2*portSize {
			c.focus(true)
			c.startEditing()
		} else if m.Pos.Sub(c.dstPt).Len() < 2*portSize {
			c.focus(false)
			c.startEditing()
		}
	}
	if !c.editing {
		return
	}

	p1 := c.src
	if c.focusSrc {
		p1 = c.dst
	}
	var p *port
	b := c.blk.outermost()
	if p2, ok := ViewAt(b, MapTo(c, m.Pos, b)).(*port); ok {
		src, dst := p1, p2
		if c.focusSrc {
			src, dst = dst, src
		}
		if src == c.src && dst == c.dst || canConnect(src, dst, c) {
			p = p2
		}
	}
	if c.focusSrc {
		c.srcPt = m.Pos
		c.setSrc(p)
	} else {
		c.dstPt = m.Pos
		c.setDst(p)
	}

	if m.Release {
		c.stopEditing()
	}
}

func (c *connection) Paint() {
	start, end := c.srcPt, c.dstPt
	var pts []Point
	if c.src != nil && c.src.obj.Type == seqType || c.dst != nil && c.dst.obj.Type == seqType {
		pts = []Point{start, Pt(start.X, start.Y-40), Pt(end.X, end.Y-40), end}
	} else {
		d := end.Sub(start)
		mid := start.Add(d.Div(2))
		if c.feedback {
			mid.Y = math.Max(start.Y, end.Y) + 128
		}
		dx := math.Abs(d.X / 3)
		p1 := start.Add(Pt(dx, 0))
		p2 := mid
		p3 := end.Sub(Pt(dx, 0))
		pts = []Point{start, p1, p2, p3, end}
	}
	steps := 0.0
	for i := range pts {
		if i > 0 {
			steps += pts[i].Sub(pts[i-1]).Len()
		}
	}

	c1 := Color{.5, .5, .5, 1}
	if c.focused {
		c2 := Color{.3, .3, .7, 1}
		if c.editing {
			c2 = Color{1, .5, 0, .5}
		}
		if c.focusSrc {
			c1, c2 = c2, c1
		}
		ctrlPts := []Double{Double(c1.R), Double(c1.G), Double(c1.B), Double(c1.A), Double(c2.R), Double(c2.G), Double(c2.B), Double(c2.A)}
		Map1d(MAP1_COLOR_4, 0, 1, 4, 2, &ctrlPts[0])
		Enable(MAP1_COLOR_4)
		defer Disable(MAP1_COLOR_4)
	} else {
		SetColor(c1)
	}

	ctrlPts := []Double{}
	for _, p := range pts {
		ctrlPts = append(ctrlPts, Double(p.X), Double(p.Y), 0)
	}
	Map1d(MAP1_VERTEX_3, 0, 1, 3, Int(len(pts)), &ctrlPts[0])
	Enable(MAP1_VERTEX_3)
	defer Disable(MAP1_VERTEX_3)
	MapGrid1d(Int(steps), 0, 1)
	EvalMesh1(LINE, 0, Int(steps))
}
