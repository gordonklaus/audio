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
	"time"
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
	pkg  *pkgText
	text *Text
	ins  []*port
	outs []*port

	godefer     string
	godeferText *Text
	typ         *typeView

	focused bool
	gap     float64
}

func newNodeBase(self node) *nodeBase {
	return newGoDeferNodeBase(self, "")
}

func newGoDeferNodeBase(self node, godefer string) *nodeBase {
	n := &nodeBase{self: self, godefer: godefer}
	n.ViewBase = NewView(n)
	n.AggregateMouser = AggregateMouser{NewClickFocuser(self), NewMover(self)}
	n.pkg = newPkgText()
	n.Add(n.pkg)
	n.text = NewText("")
	n.text.SetBackgroundColor(noColor)
	n.text.TextChanged = func(string) { n.reform() }
	n.Add(n.text)
	n.godeferText = NewText(godefer)
	n.godeferText.SetBackgroundColor(noColor)
	n.godeferText.SetTextColor(color(special{}, true, false))
	n.Add(n.godeferText)
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
	if n.godefer == "" {
		n.Add(p)
		n.outs = append(n.outs, p)
		n.reform()
	}
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
	if n.godefer != "" {
		n.godeferText.SetText(n.godefer)
	}
	x := -(Width(n.godeferText) + Width(n.pkg) + Width(n.text)) / 2
	if n.typ != nil {
		x -= Width(n.typ) / 2
	}
	n.godeferText.Move(Pt(x, -Height(n.godeferText)/2))
	x += Width(n.godeferText)
	n.pkg.Move(Pt(x, -Height(n.pkg)/2))
	x += Width(n.pkg)
	n.text.Move(Pt(x, -Height(n.text)/2))
	if n.typ != nil {
		x += Width(n.text)
		n.typ.Move(Pt(x, -Height(n.typ)/2))
	}
	if n.text.Text() != "" {
		n.gap = math.Max(math.Max(Height(n.godeferText), Height(n.pkg)), Height(n.text)) / 2
	}
	if n.typ != nil {
		n.gap = math.Max(n.gap, Height(n.typ)/2)
	}

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
	if _, ok := n.self.(*portsNode); ok {
		// portsNode rect must have an edge at y==0 or arrangement will diverge
		n.SetRect(rect)
	} else {
		ResizeToFit(n, 0)
	}
	rearrange(n.blk)
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
	if KeyFocus(n) == n {
		// TODO: not ZP for if and select
		panTo(n, ZP)
	}
}

func (n *nodeBase) TookKeyFocus() {
	n.focused = true
	panTo(n, ZP)
}

func (n *nodeBase) LostKeyFocus() {
	n.focused = false
}

func (n *nodeBase) Paint() {
	SetColor(lineColor)
	SetLineWidth(3)
	for _, p := range append(ins(n), outs(n)...) {
		pt := CenterInParent(p)
		dy := n.gap
		if p.out {
			dy = -dy
		}
		y := (pt.Y-dy)/2 + dy
		DrawBezier(Pt(0, dy), Pt(0, y), Pt(pt.X, y), pt)
	}
	if n.focused && (n.text.Text() != "" || n.typ != nil) {
		r := RectInParent(n.godeferText).Union(RectInParent(n.pkg)).Union(RectInParent(n.text))
		if n.typ != nil {
			r = r.Union(RectInParent(n.typ))
		}
		DrawRect(r)
	}
}

type pkgText struct {
	*ViewBase
	text  *Text
	pkg   *types.Package
	leave chan bool
}

func newPkgText() *pkgText {
	t := &pkgText{}
	t.ViewBase = NewView(t)
	t.text = NewText("")
	t.text.SetTextColor(color(&pkgObject{}, true, false))
	t.text.SetBackgroundColor(noColor)
	t.Add(t.text)
	t.leave = make(chan bool, 1)
	return t
}

func (t *pkgText) setPkg(p *types.Package) {
	t.pkg = p
	t.setText(p.Name)
}

func (t *pkgText) setText(s string) {
	t.text.SetText(s + ".")
	dx := Width(t) - Width(t.text)
	t.Move(Pos(t).Add(Pt(dx, 0)))
	Resize(t, Size(t.text))
}

func (t *pkgText) Mouse(m MouseEvent) {
	if m.Enter {
		go func() {
			select {
			case <-time.After(time.Second):
				Do(t, func() { t.setText(t.pkg.Path) })
				<-t.leave
				Do(t, func() { t.setText(t.pkg.Name) })
			case <-t.leave:
			}
		}()
	} else if m.Leave {
		t.leave <- true
	} else {
		MouseParent(t, m)
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
		n.text.Validate = func(s *string) bool {
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
		}
	case token.IMAG:
		// TODO
	case token.STRING:
		out.setType(types.Typ[types.UntypedString])
	case token.CHAR:
		out.setType(types.Typ[types.UntypedRune])
		n.text.Validate = func(s *string) bool {
			if *s == "" {
				return false
			}
			*s = (*s)[len(*s)-1:]
			return true
		}
	}
	n.text.Accept = func(string) { SetKeyFocus(n) }
	return n
}

func (n *basicLiteralNode) KeyPress(k KeyEvent) {
	switch k.Key {
	case KeyEnter:
		s := n.text.Text()
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

func newCompositeLiteralNode(currentPkg *types.Package) *compositeLiteralNode {
	n := &compositeLiteralNode{}
	n.nodeBase = newNodeBase(n)
	out := n.newOutput(nil)
	n.typ = newTypeView(&out.obj.Type, currentPkg)
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

func (n *compositeLiteralNode) removePort(p *port) {
	if p.bad {
		n.removePortBase(p)
	}
}
