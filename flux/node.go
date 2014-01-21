// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	. "code.google.com/p/gordon-go/util"
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

type nodeText struct {
	*TextBase
	node *nodeBase
}

func newNodeText(node *nodeBase) *nodeText {
	t := &nodeText{node: node}
	t.TextBase = NewTextBase(t, "")
	return t
}
func (t *nodeText) KeyPress(event KeyEvent) {
	switch event.Key {
	case KeyEnter:
		SetKeyFocus(t.node.self)
	default:
		t.TextBase.KeyPress(event)
	}
	t.node.reform()
}
func (t nodeText) Paint() {
	Rotate(1.0 / 12) // this should go in ViewBase; but is good enough for now
	t.TextBase.Paint()
}

type nodeBase struct {
	*ViewBase
	self node
	AggregateMouser
	blk     *block
	text    *nodeText
	ins     []*port
	outs    []*port
	focused bool
}

const (
	nodeMargin = 3
)

func newNodeBase(self node) *nodeBase {
	n := &nodeBase{self: self}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(self), NewMover(self)}
	n.text = newNodeText(n)
	n.text.SetBackgroundColor(Color{0, 0, 0, 0})
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
	seqIn := n.newInput(newVar("seq", seqType))
	MoveCenter(seqIn, Pt(-8, 0))
	seqOut := n.newOutput(newVar("seq", seqType))
	MoveCenter(seqOut, Pt(8, 0))
}

func (n *nodeBase) removePortBase(p *port) { // intentionally named to not implement interface{removePort(*port)}
	for _, c := range p.conns {
		c.blk.removeConn(c)
	}
	if p.out {
		SliceRemove(&n.outs, p)
	} else {
		SliceRemove(&n.ins, p)
	}
	n.Remove(p)
	n.reform()
}

func (n *nodeBase) reform() {
	ins, outs := ins(n), outs(n)

	numIn := float64(len(ins))
	numOut := float64(len(outs))
	rx, ry := 2.0*portSize, (math.Max(numIn, numOut)+1)*portSize/2

	rect := ZR
	for i, p := range ins {
		y := -portSize * (float64(i) - (numIn-1)/2)
		x := -rx * math.Sqrt(ry*ry-y*y) / ry
		if len(ins) > 1 {
			x -= 8
		}
		MoveCenter(p, Pt(x, y))
		rect = rect.Union(RectInParent(p))
	}
	for i, p := range outs {
		y := -portSize * (float64(i) - (numOut-1)/2)
		x := rx * math.Sqrt(ry*ry-y*y) / ry
		if len(outs) > 1 {
			x += 8
		}
		MoveCenter(p, Pt(x, y))
		rect = rect.Union(RectInParent(p))
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

func (n *nodeBase) TookKeyFocus() { n.focused = true; Repaint(n) }
func (n *nodeBase) LostKeyFocus() { n.focused = false; Repaint(n) }

func (n nodeBase) Paint() {
	const DX = 8.0
	SetColor(map[bool]Color{false: {.5, .5, .5, 1}, true: {.3, .3, .7, 1}}[n.focused])
	for _, p := range append(n.ins, n.outs...) {
		pt := CenterInParent(p)
		dx := -DX
		if p.out {
			dx = DX
		}
		x := (pt.X-dx)/2 + dx
		DrawCubic([4]Point{Pt(dx, 0), Pt(x, 0), Pt(x, pt.Y), pt}, int(pt.Len()/2))
	}
	in, out := len(n.ins) > 0, len(n.outs) > 0
	if in {
		DrawLine(Pt(-DX, 0), ZP)
	}
	if out {
		DrawLine(ZP, Pt(DX, 0))
	}
	if !in || !out {
		DrawLine(Pt(0, -3), Pt(0, 3))
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
	out := n.newOutput(&types.Var{})
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
	v := &types.Var{}
	n.newOutput(v)
	n.typ = newTypeView(&v.Type)
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
		MoveCenter(n.typ, Pt(0, Rect(n).Max.Y+Height(n.typ)/2))
		SetKeyFocus(n)
	}
}
