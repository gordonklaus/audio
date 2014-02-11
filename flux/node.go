// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	"go/token"
	"math"
	"strconv"
	"strings"
)

type node interface {
	View
	Mouser
	block() *block
	setBlock(b *block)
	inputs() []*port
	outputs() []*port
	inConns() []*connection
	outConns() []*connection
}

type nodeBase struct {
	*ViewBase
	self node
	AggregateMouser

	blk  *block
	text *TextBase
	ins  []*port
	outs []*port

	focused bool
	gap     float64
}

func newNodeBase(self node) *nodeBase {
	n := &nodeBase{self: self}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(self), NewMover(self)}
	n.text = NewText("")
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
	n.text.TextChanged = func(string) {
		MoveCenter(n.text, ZP)
		n.gap = Height(n.text) / 2
	}
	n.Add(n.text)
	n.ViewBase.Self = self
	return n
}

func (n *nodeBase) newInput(v *types.Var) *port {
	p := newInput(n.self, v)
	n.Add(p)
	n.ins = append(n.ins, p)
	n.reform()
	return p
}

func (n *nodeBase) newOutput(v *types.Var) *port {
	p := newOutput(n.self, v)
	n.Add(p)
	n.outs = append(n.outs, p)
	n.reform()
	return p
}

func (n *nodeBase) addSeqPorts() {
	n.newInput(newVar("seq", seqType))
	n.newOutput(newVar("seq", seqType))
	n.reform()
}

func (n *nodeBase) removePortBase(p *port) { // intentionally named to not implement interface{removePort(*port)}
	for _, c := range p.conns {
		c.blk.removeConn(c)
	}
	ports := &n.ins
	if p.out {
		ports = &n.outs
	}
	for i, p2 := range *ports {
		if p2 == p {
			*ports = append((*ports)[:i], (*ports)[i+1:]...)
			n.Remove(p)
			n.reform()
			rearrange(n.blk)

			if i > 0 && (*ports)[i-1].obj.Type != seqType { // assumes sequencing port, if present, is at index 0
				i--
			}
			if i < len(*ports) {
				SetKeyFocus((*ports)[i])
			} else {
				SetKeyFocus(n.self)
			}
			break
		}
	}
}

func (n *nodeBase) reform() {
	ins, outs := ins(n), outs(n)

	numIn := float64(len(ins))
	numOut := float64(len(outs))
	rx, ry := (math.Max(numIn, numOut)+1)*portSize/2, 1.0*portSize

	rect := ZR
	for i, p := range ins {
		x := portSize * (float64(i) - (numIn-1)/2)
		y := n.gap + ry*math.Sqrt(rx*rx-x*x)/rx
		if numIn > 1 {
			y += 8
		}
		MoveCenter(p, Pt(x, y))
		rect = rect.Union(RectInParent(p))
	}
	for i, p := range outs {
		x := portSize * (float64(i) - (numOut-1)/2)
		y := -n.gap - ry*math.Sqrt(rx*rx-x*x)/rx
		if numOut > 1 {
			y -= 8
		}
		MoveCenter(p, Pt(x, y))
		rect = rect.Union(RectInParent(p))
	}
	if p := seqIn(n); p != nil {
		MoveCenter(p, Pt(0, n.gap))
	}
	if p := seqOut(n); p != nil {
		MoveCenter(p, Pt(0, -n.gap))
	}
	n.SetRect(rect)
}

func (n nodeBase) block() *block      { return n.blk }
func (n *nodeBase) setBlock(b *block) { n.blk = b }
func (n nodeBase) inputs() []*port    { return n.ins }
func (n nodeBase) outputs() []*port   { return n.outs }

func (n nodeBase) inConns() (conns []*connection) {
	for _, p := range n.inputs() {
		for _, c := range p.conns {
			conns = append(conns, c)
		}
	}
	return
}

func (n nodeBase) outConns() (conns []*connection) {
	for _, p := range n.outputs() {
		for _, c := range p.conns {
			conns = append(conns, c)
		}
	}
	return
}

func (n *nodeBase) Move(p Point) {
	n.ViewBase.Move(p)
	nodeMoved(n.self)
}

func nodeMoved(n node) {
	for _, c := range append(n.inConns(), n.outConns()...) {
		c.reform()
	}
}

func (n *nodeBase) TookKeyFocus() {
	n.focused = true
	n.text.SetTextColor(Color{.5, .5, .9, 1})
}

func (n *nodeBase) LostKeyFocus() {
	n.focused = false
	n.text.SetTextColor(Color{1, 1, 1, 1})
}

func (n *nodeBase) Paint() {
	SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[n.focused])
	for _, p := range append(ins(n), outs(n)...) {
		pt := CenterInParent(p)
		dy := n.gap
		if p.out {
			dy = -dy
		}
		y := (pt.Y-dy)/2 + dy
		DrawCubic([4]Point{Pt(0, dy), Pt(0, y), Pt(pt.X, y), pt}, int(pt.Len()/2))
	}
}

var seqType = struct{ types.Type }{}

func seqIn(n node) *port {
	for _, in := range n.inputs() {
		if in.obj.Type == seqType {
			return in
		}
	}
	return nil
}

func seqOut(n node) *port {
	for _, out := range n.outputs() {
		if out.obj.Type == seqType {
			return out
		}
	}
	return nil
}

func ins(n node) (p []*port) {
	for _, in := range n.inputs() {
		if in.obj.Type != seqType {
			p = append(p, in)
		}
	}
	return
}

func outs(n node) (p []*port) {
	for _, out := range n.outputs() {
		if out.obj.Type != seqType {
			p = append(p, out)
		}
	}
	return
}

type basicLiteralNode struct {
	*nodeBase
	kind token.Token
}

func newBasicLiteralNode(kind token.Token) *basicLiteralNode {
	n := &basicLiteralNode{kind: kind}
	n.nodeBase = newNodeBase(n)
	out := n.newOutput(nil)
	switch kind {
	case token.INT, token.FLOAT:
		if kind == token.INT {
			out.setType(types.Typ[types.UntypedInt])
		} else {
			out.setType(types.Typ[types.UntypedFloat])
		}
		n.text.SetValidator(func(s *string) bool {
			*s = strings.TrimLeft(*s, "0")
			if *s == "" || *s == "-" {
				*s = "0"
			}
			if (*s)[0] == '.' {
				*s = "0" + *s
			}
			if l := len(*s); (*s)[l-1] == '-' {
				if (*s)[0] == '-' {
					*s = (*s)[1 : l-1]
				} else {
					*s = "-" + (*s)[:l-1]
				}
			}
			if _, err := strconv.ParseInt(*s, 10, 64); err == nil {
				n.kind = token.INT
				out.setType(types.Typ[types.UntypedInt])
			} else {
				if _, err := strconv.ParseFloat(*s, 4096); err == nil {
					n.kind = token.FLOAT
					out.setType(types.Typ[types.UntypedFloat])
				} else {
					return false
				}
			}
			return true
		})
	case token.IMAG:
		// TODO
	case token.STRING:
		out.setType(types.Typ[types.UntypedString])
	case token.CHAR:
		out.setType(types.Typ[types.UntypedRune])
		n.text.SetValidator(func(s *string) bool {
			if *s == "" {
				return false
			}
			*s = (*s)[len(*s)-1:]
			return true
		})
	}
	return n
}

func (n *basicLiteralNode) KeyPress(k KeyEvent) {
	switch k.Key {
	case KeyEnter:
		s := n.text.GetText()
		t := n.outs[0].obj.Type
		n.text.Reject = func() {
			n.text.SetText(s)
			n.outs[0].setType(t)
			SetKeyFocus(n)
		}
		SetKeyFocus(n.text)
	default:
		n.nodeBase.KeyPress(k)
	}
}

type compositeLiteralNode struct {
	*nodeBase
	typ *typeView
}

func newCompositeLiteralNode() *compositeLiteralNode {
	n := &compositeLiteralNode{}
	n.nodeBase = newNodeBase(n)
	out := n.newOutput(nil)
	n.typ = newTypeView(&out.obj.Type)
	n.typ.mode = compositeOrPtrType
	n.Add(n.typ)
	return n
}
func (n *compositeLiteralNode) editType() {
	n.typ.editType(func() {
		if t := *n.typ.typ; t != nil {
			n.setType(t)
		} else {
			n.blk.removeNode(n)
			SetKeyFocus(n.blk)
		}
	})
}
func (n *compositeLiteralNode) setType(t types.Type) {
	n.typ.setType(t)
	n.outs[0].setType(t)
	if t != nil {
		n.blk.func_().addPkgRef(t)
		t, _ = indirect(t)
		local := true
		if nt, ok := t.(*types.Named); ok {
			t = nt.UnderlyingT
			local = nt.Obj.Pkg == n.blk.func_().pkg()
		}
		switch t := t.(type) {
		case *types.Struct:
			for _, f := range t.Fields {
				if local || f.IsExported() {
					n.newInput(f)
				}
			}
		case *types.Slice:
			// TODO: variable number of inputs? (same can be achieved using append.)  variable number of index/value input pairs?
		case *types.Map:
			// TODO: variable number of key/value input pairs?
		}
		MoveCenter(n.typ, ZP)
		n.gap = Height(n.typ) / 2
		n.reform()
		SetKeyFocus(n)
	}
}
