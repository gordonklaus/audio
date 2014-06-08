// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"code.google.com/p/gordon-go/go/types"
	. "code.google.com/p/gordon-go/gui"
	. "code.google.com/p/gordon-go/util"
	"go/token"
	"math/rand"
)

const blockRadius = 8

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
	close(b.stop)
	for n := range b.nodes {
		b.removeNode(n)
	}
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
			if n.obj != nil && !isMethod(n.obj) {
				b.func_().addPkgRef(n.obj)
			}
		case *compositeLiteralNode:
			// handled in compositeLiteralNode.setType
		case *valueNode:
			switch obj := n.obj.(type) {
			case *types.Const, *types.Var:
				b.func_().addPkgRef(obj)
			case *types.Func:
				if !isMethod(obj) {
					b.func_().addPkgRef(obj)
				}
			case *localVar:
				obj.addref(n)
			}
		}
		rearrange(b)
	}
}

func (b *block) removeNode(n node) {
	if b.nodes[n] {
		for _, c := range append(n.inConns(), n.outConns()...) {
			c.blk.removeConn(c)
		}
		b.Remove(n)
		delete(b.nodes, n)
		switch n := n.(type) {
		case *callNode:
			if n.obj != nil && !isMethod(n.obj) {
				b.func_().subPkgRef(n.obj)
			}
		case *compositeLiteralNode:
			if t := *n.typ.typ; t != nil {
				b.func_().subPkgRef(t)
			}
		case *valueNode:
			switch obj := n.obj.(type) {
			case *types.Const, *types.Var:
				b.func_().subPkgRef(obj)
			case *types.Func:
				if !isMethod(obj) {
					b.func_().subPkgRef(obj)
				}
			case *localVar:
				obj.subref(n)
			}
		case *ifNode:
			for _, b := range n.blocks {
				b.close()
			}
		case *selectNode:
			for _, c := range n.cases {
				c.blk.close()
			}
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
	b = c.blk //disconnect might change c.blk
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
			for _, b := range n.blocks {
				b.walk(bf, nf, cf)
			}
		case *selectNode:
			for _, c := range n.cases {
				c.blk.walk(bf, nf, cf)
			}
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

type focuserFrom interface {
	focusFrom(v View, pass bool)
}

func (b *block) focusNearestView(viewOrPoint interface{}, dirKey int) {
	p, _ := viewOrPoint.(Point)
	if v, ok := viewOrPoint.(View); ok {
		p = MapTo(v, ZP, b)
	}

	b2 := b.outermost()
	p = MapTo(b, p, b2)
	b = b2

	views := []View{}
	for _, n := range b.allNodes() {
		views = append(views, n)
	}
	nearest := nearestView(b, views, p, dirKey)
	if nearest != nil {
		SetKeyFocus(nearest)
	}
}

func (b *block) focus() {
	if len(b.nodes) == 0 {
		SetKeyFocus(b)
	} else {
		b.focusNearestView(Center(b).Add(Pt(0, Height(b)/2)), KeyDown)
	}
}

func (b *block) TookKeyFocus() {
	for n := range b.nodes {
		SetKeyFocus(n)
		return
	}
	b.focused = true
	Repaint(b)
	panTo(b, Center(b))
}

func (b *block) LostKeyFocus() {
	b.focused = false
	Repaint(b)
}

func (b *block) KeyPress(event KeyEvent) {
	switch k := event.Key; k {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		if event.Alt && !event.Shift {
			b.focusNearestView(KeyFocus(b), k)
		} else if n, ok := KeyFocus(b).(node); ok {
			focseq := event.Alt && event.Shift
			if k == KeyUp {
				seq := seqIn(n)
				if num := len(ins(n)); seq != nil && (focseq || num == 0 && len(seq.conns) > 0) {
					seq.focusMiddle()
				} else if num > 0 {
					ins(n)[(num-1)/2].focusMiddle()
				}
			}
			if k == KeyDown {
				seq := seqOut(n)
				if num := len(outs(n)); seq != nil && (focseq || num == 0 && len(seq.conns) > 0) {
					seq.focusMiddle()
				} else if num > 0 {
					outs(n)[(num-1)/2].focusMiddle()
				}
			}
		} else if f, ok := b.node.(focuserFrom); ok {
			if event.Key == KeyUp {
				f.focusFrom(b, true)
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
			b.removeNode(v)
			SetKeyFocus(foc)
		}
	case KeyEscape:
		if n, ok := KeyFocus(b).(node); ok {
			if f, ok := n.block().node.(focuserFrom); ok {
				f.focusFrom(b, false)
			} else {
				SetKeyFocus(n.block().node)
			}
		} else {
			if f, ok := b.node.(focuserFrom); ok {
				f.focusFrom(b, false)
			} else {
				SetKeyFocus(b.node)
			}
		}
	default:
		openBrowser := func() {
			browser := newBrowser(browserOptions{enterTypes: true, canFuncAsVal: true}, b)
			b.Add(browser)
			browser.Move(Center(b))
			browser.accepted = func(obj types.Object) {
				browser.Close()
				b.newNode(obj, browser.funcAsVal, "")
			}
			oldFocus := KeyFocus(b)
			browser.canceled = func() {
				browser.Close()
				SetKeyFocus(oldFocus)
			}
			browser.KeyPress(event)
			SetKeyFocus(browser)
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
				n := newCompositeLiteralNode(b.func_().pkg())
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

func (b *block) newNode(obj types.Object, funcAsVal bool, godefer string) node {
	var n node
	currentPkg := b.func_().pkg()
	switch obj := obj.(type) {
	case special:
		switch obj.Name {
		case "break", "continue":
			n = newBranchNode(obj.Name)
		case "call":
			n = newCallNode(nil, currentPkg, godefer)
		case "convert":
			n = newConvertNode(currentPkg)
		case "func":
			n = newFuncNode(nil, b.childArranged)
		case "go", "defer":
			godefer = obj.Name + " "
			browser := newBrowser(browserOptions{objFilter: isGoDeferrable, enterTypes: true}, b)
			browser.Move(Center(b))
			browser.accepted = func(obj types.Object) {
				browser.Close()
				b.newNode(obj, false, godefer)
			}
			browser.canceled = func() {
				browser.Close()
				SetKeyFocus(b)
			}
			b.Add(browser)
			SetKeyFocus(browser)
			return nil
		case "if":
			i := newIfNode(b.childArranged)
			i.newBlock()
			n = i
		case "loop":
			n = newLoopNode(b.childArranged)
		case "select":
			n = newSelectNode(b.childArranged)
		case "typeAssert":
			n = newTypeAssertNode(currentPkg)
		}
	case *types.Func, *types.Builtin:
		if obj.GetName() == "[]" {
			n = newIndexNode(false)
		} else if obj.GetName() == "[:]" {
			n = newSliceNode()
		} else if obj.GetName() == "<-" {
			n = newChanNode(true)
		} else if isOperator(obj) {
			n = newOperatorNode(obj)
		} else if funcAsVal && obj.GetPkg() != nil { //Pkg==nil == builtin
			n = newValueNode(obj, currentPkg, false)
		} else {
			n = newCallNode(obj, currentPkg, godefer)
		}
	case *types.Var, *types.Const, field, *localVar:
		switch obj.GetName() {
		default:
			n = newValueNode(obj, currentPkg, false)
		case "=":
			n = newValueNode(nil, nil, true)
		case "*":
			n = newValueNode(nil, nil, false)
		}
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
	return n
}

func (b *block) Paint() {
	if b.focused {
		SetPointSize(2 * portSize)
		SetColor(focusColor)
		DrawPoint(Center(b))
	}
	{
		SetColor(lineColor)
		SetLineWidth(1.5)
		rect := Rect(b)
		l, r, b, t := rect.Min.X, rect.Max.X, rect.Min.Y, rect.Max.Y
		lb, bl := Pt(l, b+blockRadius), Pt(l+blockRadius, b)
		rb, br := Pt(r, b+blockRadius), Pt(r-blockRadius, b)
		rt, tr := Pt(r, t-blockRadius), Pt(r-blockRadius, t)
		lt, tl := Pt(l, t-blockRadius), Pt(l+blockRadius, t)
		DrawLine(bl, br)
		DrawBezier(br, Pt(r, b), rb)
		DrawLine(rb, rt)
		DrawBezier(rt, Pt(r, t), tr)
		DrawLine(tr, tl)
		DrawBezier(tl, Pt(l, t), lt)
		DrawLine(lt, lb)
		DrawBezier(lb, Pt(l, b), bl)
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
		sig := f.sig()

		ports := n.ins
		vars := &sig.Results
		if p.out {
			ports = n.outs
			if sig.Recv != nil { // don't remove receiver
				ports = ports[1:]
			}
			vars = &sig.Params
		}

		for i, q := range ports {
			if q == p {
				n.blk.func_().subPkgRef((*vars)[i].Type)
				*vars = append((*vars)[:i], (*vars)[i+1:]...)
				n.removePortBase(p)
				if i == len(*vars) {
					sig.IsVariadic = false
				}
				if f.obj == nil {
					f.output.setType(sig)
				}
				break
			}
		}
	}
}

func (n *portsNode) focusFrom(v View, pass bool) {
	if f, ok := n.blk.node.(focuserFrom); ok {
		f.focusFrom(n, pass)
	}
}

func (n *portsNode) KeyPress(event KeyEvent) {
	if l, ok := n.blk.node.(*loopNode); ok && event.Key == KeyUp {
		l.focusFrom(n, true)
	} else if s, ok := n.blk.node.(*selectNode); ok && event.Key == KeyUp {
		s.focusFrom(n, true)
	} else if f, ok := n.blk.node.(*funcNode); ok && f.literal && event.Key == KeyDown && n.out {
		f.focusFrom(n, true)
	} else if n.editable && event.Text == "," {
		f := n.blk.node.(*funcNode)
		sig := f.sig()

		newPort := newOutput
		ports := &n.outs
		vars := &sig.Params
		if n.out {
			newPort = newInput
			ports = &n.ins
			vars = &sig.Results
		}

		v := types.NewVar(0, n.blk.func_().pkg(), "", nil)
		p := newPort(n, v)

		i := len(*ports)
		if focus, ok := KeyFocus(n).(*port); ok {
			for j, p := range *ports {
				if p == focus {
					i = j
					break
				}
			}
			if !event.Shift || i == 0 && p.out && sig.Recv != nil {
				i++
			}
		}

		n.Add(p)
		*ports = append((*ports)[:i], append([]*port{p}, (*ports)[i:]...)...)
		n.reform()
		Show(p.valView)
		p.valView.edit(func() {
			if v.Type != nil {
				if p.out && sig.Recv != nil {
					i--
				}
				*vars = append((*vars)[:i], append([]*types.Var{v}, (*vars)[i:]...)...)
				if i == len(*vars)-1 {
					sig.IsVariadic = false
				}
				n.blk.func_().addPkgRef(v.Type)
				SetKeyFocus(p)
			} else {
				n.removePortBase(p)
			}
			if f.obj == nil {
				f.output.setType(sig)
			}
		})
	} else if n.editable && !n.out && event.Key == KeyPeriod && event.Ctrl {
		f := n.blk.node.(*funcNode)
		sig := f.sig()
		len := len(n.outs)
		if len > 0 && (sig.Recv == nil || len > 1) {
			p := n.outs[len-1]
			if KeyFocus(n) != p {
				return
			}
			sig.IsVariadic = !sig.IsVariadic
			p.valView.ellipsis = sig.IsVariadic
			if sig.IsVariadic {
				p.setType(&types.Slice{p.obj.Type})
			} else {
				p.setType(p.obj.Type.(*types.Slice).Elem)
			}
			if f.obj == nil {
				f.output.setType(sig)
			}
		}
	} else {
		n.nodeBase.KeyPress(event)
	}
}

func (n *portsNode) Paint() {
	SetColor(lineColor)
	SetLineWidth(3)
	DrawLine(Pt(-portSize/4, 0), Pt(portSize/4, 0))
	n.nodeBase.Paint()
	if n.focused {
		SetPointSize(2 * portSize)
		SetColor(focusColor)
		DrawPoint(ZP)
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
