// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/go.exp/go/types"
	. "code.google.com/p/gordon-go/gui"
	. "code.google.com/p/gordon-go/util"
	"go/token"
	"math"
	"math/rand"
	"unicode"
)

const blockRadius = 16

type block struct {
	*ViewBase
	node      node
	nodes     map[node]bool
	conns     map[*connection]bool
	localVars map[*localVar]bool
	focused   bool

	arrange, childArranged blockchan
	stop                   stopchan
}

func newBlock(n node, arranged blockchan) *block {
	b := &block{}
	b.ViewBase = NewView(b)
	b.node = n
	b.nodes = map[node]bool{}
	b.conns = map[*connection]bool{}
	b.localVars = map[*localVar]bool{}

	b.arrange = make(blockchan)
	b.childArranged = make(blockchan)
	b.stop = make(stopchan)
	go arrange(b.arrange, b.childArranged, arranged, b.stop)
	rearrange(b)

	n.Add(b)
	return b
}

func (b *block) close() {
	b.walk(func(b *block) {
		close(b.stop)
	}, nil, nil)
}

func (b *block) outer() *block { return b.node.block() }
func (b *block) outermost() *block {
	if outer := b.outer(); outer != nil {
		return outer.outermost()
	}
	return b
}
func (b *block) func_() *funcNode {
	f, _ := b.outermost().node.(*funcNode)
	return f
}

func func_(n node) *funcNode {
	if b := n.block(); b != nil {
		return b.func_()
	}
	fn, _ := n.(*funcNode)
	return fn
}

func (b *block) addNode(n node) {
	if !b.nodes[n] {
		b.Add(n)
		n.Move(Pt(rand.NormFloat64(), rand.NormFloat64()))
		b.nodes[n] = true
		n.setBlock(b)
		switch n := n.(type) {
		case *callNode:
			if _, ok := n.obj.(method); !ok && n.obj != nil {
				b.func_().addPkgRef(n.obj)
			}
		case *valueNode:
			if v, ok := n.obj.(*localVar); ok {
				v.addref(n)
			}
		}
		rearrange(b)
	}
}

func (b *block) removeNode(n node) {
	if b.nodes[n] {
		b.Remove(n)
		delete(b.nodes, n)
		switch n := n.(type) {
		case *callNode:
			if _, ok := n.obj.(method); !ok && n.obj != nil {
				b.func_().subPkgRef(n.obj)
			}
		case *valueNode:
			if v, ok := n.obj.(*localVar); ok {
				v.subref(n)
			}
		case *ifNode:
			n.falseblk.close()
			n.trueblk.close()
		case *loopNode:
			n.loopblk.close()
		case *funcNode:
			n.funcblk.close()
		}
		rearrange(b)
	}
}

func (b *block) addConn(c *connection) {
	if c.blk != nil {
		delete(c.blk.conns, c)
		c.blk.Remove(c)
		rearrange(c.blk)
	}
	c.blk = b
	b.Add(c)
	Lower(c)
	b.conns[c] = true
	rearrange(b)
}

func (b *block) removeConn(c *connection) {
	c.disconnect()
	delete(b.conns, c)
	b.Remove(c)
	rearrange(b)
}

func (b *block) walk(bf func(*block), nf func(node), cf func(*connection)) {
	if bf != nil {
		bf(b)
	}
	for n := range b.nodes {
		if nf != nil {
			nf(n)
		}
		switch n := n.(type) {
		case *ifNode:
			n.falseblk.walk(bf, nf, cf)
			n.trueblk.walk(bf, nf, cf)
		case *loopNode:
			n.loopblk.walk(bf, nf, cf)
		case *funcNode:
			n.funcblk.walk(bf, nf, cf)
		}
	}
	if cf != nil {
		for c := range b.conns {
			cf(c)
		}
	}
}

func (b *block) allNodes() (nodes []node) {
	b.walk(nil, func(n node) {
		nodes = append(nodes, n)
	}, nil)
	return
}

func (b block) inConns() (conns []*connection) {
	for n := range b.nodes {
		for _, c := range n.inConns() {
			if !b.conns[c] {
				conns = append(conns, c)
			}
		}
	}
	return
}

func (b block) outConns() (conns []*connection) {
	for n := range b.nodes {
		for _, c := range n.outConns() {
			if !b.conns[c] {
				conns = append(conns, c)
			}
		}
	}
	return
}

func (b *block) nodeOrder() []node {
	order := []node{}
	var inputsNode *portsNode

	visited := Set{}
	var insertInOrder func(n node, visitedThisCall Set)
	insertInOrder = func(n node, visitedThisCall Set) {
		if visitedThisCall[n] {
			panic("cyclic")
		}
		visitedThisCall[n] = true

		if !visited[n] {
			visited[n] = true
			for _, src := range srcsInBlock(n) {
				insertInOrder(src, visitedThisCall.Copy())
			}
			if pn, ok := n.(*portsNode); ok {
				if !pn.out {
					inputsNode = pn
				}
			} else {
				order = append(order, n)
			}
		}
	}

	endNodes := []node{}
	for n := range b.nodes {
		if len(dstsInBlock(n)) == 0 {
			endNodes = append(endNodes, n)
		}
	}
	if len(endNodes) == 0 && len(b.nodes) > 0 {
		panic("cyclic")
	}

	for _, n := range endNodes {
		insertInOrder(n, Set{})
	}
	if inputsNode != nil {
		order = append([]node{inputsNode}, order...)
	}
	return order
}

func srcsInBlock(n node) (srcs []node) {
	b := n.block()
	for _, c := range n.inConns() {
		if c.feedback || c.src == nil {
			continue
		}
		if src := b.find(c.src.node); src != nil {
			srcs = append(srcs, src)
		}
	}
	for _, c := range n.outConns() {
		if !c.feedback || c.dst == nil {
			continue
		}
		if dst := b.find(c.dst.node); dst != nil && dst != n {
			srcs = append(srcs, dst)
		}
	}
	return
}

func dstsInBlock(n node) (dsts []node) {
	b := n.block()
	for _, c := range n.outConns() {
		if c.feedback || c.dst == nil {
			continue
		}
		if dst := b.find(c.dst.node); dst != nil {
			dsts = append(dsts, dst)
		}
	}
	for _, c := range n.inConns() {
		if !c.feedback || c.src == nil {
			continue
		}
		if src := b.find(c.src.node); src != nil && src != n {
			dsts = append(dsts, src)
		}
	}
	return
}

func (b *block) find(n node) node {
	for b2 := n.block(); b2 != nil; n, b2 = b2.node, b2.outer() {
		if b2 == b {
			return n
		}
	}
	return nil
}

func nearestView(parent View, views []View, p Point, dirKey int) (nearest View) {
	dir := map[int]Point{KeyLeft: {-1, 0}, KeyRight: {1, 0}, KeyUp: {0, 1}, KeyDown: {0, -1}}[dirKey]
	best := 0.0
	for _, v := range views {
		d := MapTo(v, ZP, parent).Sub(p)
		score := (dir.X*d.X + dir.Y*d.Y) / (d.X*d.X + d.Y*d.Y)
		if score > best {
			best = score
			nearest = v
		}
	}
	return
}

func (b *block) focusNearestView(v View, dirKey int) {
	b = b.outermost()
	views := []View{}
	for _, n := range b.allNodes() {
		views = append(views, n)
		if n, ok := n.(*loopNode); ok {
			views = append(views, seqOut(n))
		}
	}
	nearest := nearestView(b, views, MapTo(v, ZP, b), dirKey)
	if nearest != nil {
		SetKeyFocus(nearest)
	}
}

func (b *block) TookKeyFocus() { b.focused = true; Repaint(b) }
func (b *block) LostKeyFocus() { b.focused = false; Repaint(b) }

func (b *block) KeyPress(event KeyEvent) {
	switch k := event.Key; k {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		if event.Alt {
			b.focusNearestView(KeyFocus(b), k)
		} else if n, ok := KeyFocus(b).(node); ok {
			if k == KeyLeft {
				if seqIn(n) != nil {
					SetKeyFocus(seqIn(n))
				} else if len := len(ins(n)); len > 0 {
					ins(n)[len/2].focusMiddle()
				}
			}
			if k == KeyRight {
				if seqOut(n) != nil {
					SetKeyFocus(seqOut(n))
				} else if len := len(outs(n)); len > 0 {
					outs(n)[len/2].focusMiddle()
				}
			}
		} else {
			b.ViewBase.KeyPress(event)
		}
	case KeyBackspace, KeyDelete:
		switch v := KeyFocus(b).(type) {
		case *block:
			SetKeyFocus(v.node)
		case *portsNode:
		case node:
			foc := View(b)
			in, out := v.inConns(), v.outConns()
			if len(in) > 0 {
				foc = in[len(in)-1].src.node
			}
			if (len(in) == 0 || k == KeyDelete) && len(out) > 0 {
				foc = out[len(out)-1].dst.node
			}
			for _, c := range append(in, out...) {
				c.blk.removeConn(c)
			}
			b.removeNode(v)
			SetKeyFocus(foc)
		}
	case KeyEscape:
		if n, ok := KeyFocus(b).(node); ok {
			if f, ok := n.block().node.(*funcNode); ok && !f.lit() {
				f.Close()
			} else {
				SetKeyFocus(n.block().node)
			}
		} else {
			SetKeyFocus(b.node)
		}
	default:
		openBrowser := func() {
			browser := newBrowser(browse, b.func_())
			b.Add(browser)
			browser.Move(Center(b))
			browser.accepted = func(obj types.Object) {
				browser.Close()
				newNode(b, obj, browser.funcAsVal)
			}
			oldFocus := KeyFocus(b)
			browser.canceled = func() {
				browser.Close()
				SetKeyFocus(oldFocus)
			}
			browser.text.KeyPress(event)
			SetKeyFocus(browser.text)
		}
		if event.Command && event.Text == "0" {
			openBrowser()
			return
		}
		if !(event.Ctrl || event.Alt || event.Super) {
			switch event.Text {
			default:
				openBrowser()
			case "\"", "'", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				text := event.Text
				kind := token.INT
				switch event.Text {
				case "\"":
					kind, text = token.STRING, ""
				case "'":
					kind = token.CHAR
				}
				n := newBasicLiteralNode(kind)
				b.addNode(n)
				MoveCenter(n, Center(b))
				n.text.SetText(text)
				n.text.Reject = func() {
					b.removeNode(n)
					SetKeyFocus(b)
				}
				SetKeyFocus(n.text)
			case "{":
				n := newCompositeLiteralNode()
				b.addNode(n)
				MoveCenter(n, Center(b))
				n.editType()
			case "":
				b.ViewBase.KeyPress(event)
			}
		} else {
			b.ViewBase.KeyPress(event)
		}
	}
}

func newNode(b *block, obj types.Object, funcAsVal bool) {
	var n node
	switch obj := obj.(type) {
	case special:
		switch obj.name {
		case "=":
			n = newValueNode(nil, true)
		case "[]":
			n = newIndexNode(false)
		case "[]=":
			n = newIndexNode(true)
		case "break", "continue":
			n = newBranchNode(obj.name)
		case "call":
			n = newCallNode(nil)
		case "convert":
			n = newConvertNode()
		case "func":
			n = newFuncNode(nil, b.childArranged)
		case "if":
			n = newIfNode(b.childArranged)
		case "indirect":
			n = newValueNode(nil, false)
		case "loop":
			n = newLoopNode(b.childArranged)
		case "typeAssert":
			n = newTypeAssertNode()
		}
	case *types.Func, method:
		if !unicode.IsLetter([]rune(obj.GetName())[0]) {
			n = newOperatorNode(obj)
		} else if funcAsVal && obj.GetPkg() != nil { //Pkg==nil == builtin
			n = newValueNode(obj, false)
		} else {
			n = newCallNode(obj)
		}
	case *types.Var, *types.Const, field, *localVar:
		n = newValueNode(obj, false)
	}
	b.addNode(n)
	MoveCenter(n, Center(b))
	if nn, ok := n.(interface {
		editType()
	}); ok {
		nn.editType()
	} else {
		SetKeyFocus(n)
	}
}

func (b *block) Paint() {
	if b.focused {
		SetColor(Color{.3, .3, .7, 1})
	} else {
		SetColor(Color{.5, .5, .5, 1})
	}
	{
		rect := Rect(b)
		l, r, b, t := rect.Min.X, rect.Max.X, rect.Min.Y, rect.Max.Y
		lb, bl := Pt(l, b+blockRadius), Pt(l+blockRadius, b)
		rb, br := Pt(r, b+blockRadius), Pt(r-blockRadius, b)
		rt, tr := Pt(r, t-blockRadius), Pt(r-blockRadius, t)
		lt, tl := Pt(l, t-blockRadius), Pt(l+blockRadius, t)
		steps := int(math.Trunc(2 * math.Pi * blockRadius))
		DrawLine(bl, br)
		DrawQuadratic([3]Point{br, Pt(r, b), rb}, steps)
		DrawLine(rb, rt)
		DrawQuadratic([3]Point{rt, Pt(r, t), tr}, steps)
		DrawLine(tr, tl)
		DrawQuadratic([3]Point{tl, Pt(l, t), lt}, steps)
		DrawLine(lt, lb)
		DrawQuadratic([3]Point{lb, Pt(l, b), bl}, steps)
	}
}

type portsNode struct {
	*nodeBase
	out      bool
	editable bool
}

func newInputsNode() *portsNode  { return newPortsNode(false) }
func newOutputsNode() *portsNode { return newPortsNode(true) }
func newPortsNode(out bool) *portsNode {
	n := &portsNode{out: out}
	n.nodeBase = newNodeBase(n)
	return n
}

func (n *portsNode) removePort(p *port) {
	if n.editable {
		f := n.blk.node.(*funcNode)
		obj := f.obj
		if obj == nil {
			obj = f.output.obj
		}
		sig := obj.GetType().(*types.Signature)
		var ports []*port
		var vars *[]*types.Var
		if p.out {
			ports = n.outs
			if _, ok := obj.(method); ok { // don't remove receiver
				ports = ports[1:]
			}
			vars = &sig.Params
		} else {
			ports = n.ins
			vars = &sig.Results
		}
		for i, q := range ports {
			if q == p {
				n.blk.func_().subPkgRef((*vars)[i].Type)
				*vars = append((*vars)[:i], (*vars)[i+1:]...)
				n.removePortBase(p)
				SetKeyFocus(n)
				if f.obj == nil {
					f.output.setType(f.output.obj.Type)
				}
				rearrange(n.blk)
				break
			}
		}
	}
}

func (n *portsNode) KeyPress(event KeyEvent) {
	if l, ok := n.blk.node.(*loopNode); ok && event.Key == KeyLeft {
		SetKeyFocus(l)
	} else if f, ok := n.blk.node.(*funcNode); ok && f.lit() && event.Key == KeyRight && n.out {
		SetKeyFocus(f)
	} else if n.editable && event.Text == "," {
		f := n.blk.node.(*funcNode)
		obj := f.obj
		if obj == nil {
			obj = f.output.obj
		}
		sig := obj.GetType().(*types.Signature)
		var p *port
		var vars *[]*types.Var
		v := &types.Var{}
		if n.out {
			p = n.newInput(v)
			vars = &sig.Results
		} else {
			p = n.newOutput(v)
			vars = &sig.Params
		}
		rearrange(n.blk)
		Show(p.valView)
		p.valView.edit(func() {
			if v.Type != nil {
				*vars = append(*vars, v)
				n.blk.func_().addPkgRef(v.Type)
				SetKeyFocus(p)
			} else {
				n.removePortBase(p)
				rearrange(n.blk)
				SetKeyFocus(n)
			}
			if f.obj == nil {
				f.output.setType(f.output.obj.Type)
			}
		})
	} else {
		n.nodeBase.KeyPress(event)
	}
}

type localVar struct {
	types.Var
	refs map[*valueNode]bool
	blk  *block
}

func (v *localVar) addref(n *valueNode) {
	v.refs[n] = true
	v.reblock()
}

func (v *localVar) subref(n *valueNode) {
	delete(v.refs, n)
	v.reblock()
}

func (v *localVar) reblock() {
	if v.blk != nil {
		delete(v.blk.localVars, v)
	}
	v.blk = nil
	for n := range v.refs {
		if v.blk == nil {
			v.blk = n.blk
			continue
		}
		for b := v.blk; ; b = b.outer() {
			if b.find(n) != nil {
				v.blk = b
				break
			}
		}
	}
	if v.blk != nil {
		v.blk.localVars[v] = true
	}
}
