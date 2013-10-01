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
	node    node
	nodes   map[node]bool
	conns   map[*connection]bool
	focused bool
	vel     map[node]Point
}

func newBlock(n node) *block {
	b := &block{}
	b.ViewBase = NewView(b)
	b.node = n
	b.nodes = map[node]bool{}
	b.conns = map[*connection]bool{}
	b.vel = map[node]Point{}
	return b
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
		b.nodes[n] = true
		n.setBlock(b)
		switch n := n.(type) {
		case *callNode:
			if _, ok := n.obj.(method); !ok && n.obj != nil {
				b.func_().addPkgRef(n.obj)
			}
		}
		if f := b.func_(); f != nil {
			f.wakeUp()
		}
	}
}

func (b *block) removeNode(n node) {
	b.Remove(n)
	delete(b.nodes, n)
	delete(b.vel, n)
	switch n := n.(type) {
	case *callNode:
		b.func_().subPkgRef(n.obj)
	}
	if f := b.func_(); f != nil {
		f.wakeUp()
	}
}

func (b *block) addConn(c *connection) {
	if c.blk != nil {
		delete(c.blk.conns, c)
	}
	c.blk = b
	b.Add(c)
	Lower(c)
	b.conns[c] = true
	if f := b.func_(); f != nil {
		f.wakeUp()
	}
}

func (b *block) removeConn(c *connection) {
	c.disconnect()
	delete(b.conns, c)
	b.Remove(c)
	if f := b.func_(); f != nil {
		f.wakeUp()
	}
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

func (b *block) allBlocks() (blocks []*block) {
	b.walk(func(b *block) {
		blocks = append(blocks, b)
	}, nil, nil)
	return
}

func (b *block) allNodes() (nodes []node) {
	b.walk(nil, func(n node) {
		nodes = append(nodes, n)
	}, nil)
	return
}

func (b block) allConns() (conns []*connection) {
	b.walk(nil, nil, func(c *connection) {
		conns = append(conns, c)
	})
	return
}

func (b block) allFocusableViews() (views []View) {
	for _, b := range b.allBlocks() {
		if len(b.nodes) == 0 {
			views = append(views, b)
		}
	}
	for _, n := range b.allNodes() {
		views = append(views, n)
		for _, p := range append(n.inputs(), n.outputs()...) {
			views = append(views, p)
		}
	}
	for _, c := range b.allConns() {
		views = append(views, c)
	}
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

func (b *block) nodeOrder() (order []node, ok bool) {
	visited := Set{}
	var insertInOrder func(n node, visitedThisCall Set) bool
	insertInOrder = func(n node, visitedThisCall Set) bool {
		if visitedThisCall[n] {
			return false
		}
		visitedThisCall[n] = true

		if !visited[n] {
			visited[n] = true
		conns:
			for _, c := range n.inConns() {
				if b.conns[c] {
					src := c.src.node
					for !b.nodes[src] {
						src = src.block().node
						if src == nil {
							continue conns
						}
					}
					if !insertInOrder(src, visitedThisCall.Copy()) {
						return false
					}
				}
			}
			order = append(order, n)
		}
		return true
	}

	endNodes := []node{}
nx:
	for n := range b.nodes {
		for _, c := range n.outConns() {
			if c.blk == b {
				continue nx
			}
		}
		endNodes = append(endNodes, n)
	}
	if len(endNodes) == 0 && len(b.nodes) > 0 {
		return
	}

	for _, n := range endNodes {
		if !insertInOrder(n, Set{}) {
			return
		}
	}
	ok = true
	return
}

// TODO: consider not repositioning portsNodes, as doing so may contribute to poor convergence and poor measure of meanSpeed
// TODO: consider using an EA to lay out nodes
func (b *block) update() (updated bool) {
	for n := range b.nodes {
		if n, ok := n.(interface {
			update() bool
		}); ok {
			updated = n.update() || updated
		}
	}

	const (
		nodeCenterCoef  = .5
		nodeSep         = 16
		nodeSepCoef     = 4
		nodeConnSep     = 8
		nodeConnSepCoef = 1
		connLen         = 32
		connLenFB       = -256
		connLenCoef     = .33
		connConnSep     = 4
		connConnSepCoef = 1
		topSpeed        = 200
		speedCompress   = 1
		dragCoef        = .85
	)

	addVel := func(n node, dv Point) {
		n.block().vel[n] = n.block().vel[n].Add(dv)
	}
	subVel := func(n node, dv Point) {
		n.block().vel[n] = n.block().vel[n].Sub(dv)
	}

	for n1 := range b.nodes {
		for n2 := range b.nodes {
			if n2 == n1 {
				continue
			}
			dir := CenterInParent(n1).Sub(CenterInParent(n2))
			if dir == ZP {
				dir = Pt(rand.NormFloat64(), rand.NormFloat64())
			}
			d := dir.Len() - Size(n1).Add(Size(n2)).Len()/2 - nodeSep
			if d >= 0 {
				continue
			}
			addVel(n1, dir.Mul(-nodeSepCoef*d/dir.Len()))
		}
		for b := b; b != nil; b = b.outer() {
		conns:
			for c := range b.conns {
				if c.src == nil || c.dst == nil {
					continue
				}
				for n := c.src.node; n.block() != nil; n = n.block().node {
					if n == n1 {
						continue conns
					}
				}
				for n := c.dst.node; n.block() != nil; n = n.block().node {
					if n == n1 {
						continue conns
					}
				}
				x := c.srcPt
				y := c.dstPt
				p := CenterInParent(n1)
				dir := p.Sub(PointToLine(p, x, y))
				if dir == ZP {
					dir = Pt(rand.NormFloat64(), rand.NormFloat64())
				}
				d := dir.Len() - Size(n1).Len()/2 - nodeConnSep
				if d > 0 {
					continue
				}
				delta := dir.Mul(-nodeConnSepCoef * d / dir.Len())
				addVel(n1, delta)
				subVel(c.src.node, delta)
				subVel(c.dst.node, delta)
			}
		}
	}
	for c := range b.conns {
		if c.src == nil || c.dst == nil {
			continue
		}
		// for b := b; b != nil; b = b.outer() {
		// 	for c2 := range b.conns {
		// 		if c == c2 ||
		// 		   c2.src == nil || c2.dst == nil ||
		// 		   c.src.node == c2.src.node || c.dst.node == c2.dst.node || c.src.node == c2.dst.node || c.dst.node == c2.src.node {
		// 			continue
		// 		}
		// 		p1, p2 := LineToLine(c.srcPt, c.dstPt, c2.srcPt, c2.dstPt)
		// 		d := p1.Sub(p2)
		// 		if l := d.Len(); l == 0 {
		// 			d = c.MapToParent(c.Center()).Sub(c2.MapToParent(c2.Center()))
		// 			if d.Len() == 0 { continue }
		// 			d = d.Mul(connConnSepCoef / d.Len())
		// 			addVel(c.src.node, d)
		// 			addVel(c.dst.node, d)
		// 			subVel(c2.src.node, d)
		// 			subVel(c2.dst.node, d)
		// 		} else if l < connConnSep {
		// 			d = d.Mul((connConnSep - l) / l)
		// 			addVel(c.src.node, d)
		// 			addVel(c.dst.node, d)
		// 			subVel(c2.src.node, d)
		// 			subVel(c2.dst.node, d)
		// 		}
		// 	}
		// }
		d := c.dstPt.Sub(c.srcPt)
		connLen := float64(connLen)
		if c.feedback {
			connLen = connLenFB
		}
		d.X -= connLen + math.Abs(d.Y)/2 // the nonlinearity abs(d.Y) can contribute to oscillations
		if d.X < 0 {
			d.X *= -d.X
		}
		d = d.Mul(connLenCoef)

		srcNode := c.src.node
		for srcNode.block() != b {
			srcNode = srcNode.block().node
		}
		dstNode := c.dst.node
		for dstNode.block() != b {
			dstNode = dstNode.block().node
		}
		addVel(srcNode, d)
		subVel(dstNode, d)
	}

	center := ZP
	for n := range b.nodes {
		center = center.Add(Pos(n))
	}
	center = center.Div(float64(len(b.nodes)))
	for n := range b.nodes {
		addVel(n, center.Sub(Pos(n).Mul(nodeCenterCoef)))
		l := b.vel[n].Len()
		if l > 0 {
			b.vel[n] = b.vel[n].Mul(math.Tanh(speedCompress*l/topSpeed) * topSpeed / l)
		}
	}

	meanVel := ZP
	nVel := 0.0
	for n := range b.nodes {
		if _, ok := n.(*portsNode); ok {
			continue
		}
		b.vel[n] = b.vel[n].Mul(dragCoef)
		meanVel = meanVel.Add(b.vel[n])
		nVel++
	}
	// if there is only one (non-port) node then it is definitely stationary
	// but we still have to check if the rect needs updating
	if nVel > 1 {
		meanVel = meanVel.Div(nVel)
		meanSpeed := 0.0
		for n := range b.nodes {
			if _, ok := n.(*portsNode); ok {
				continue
			}
			subVel(n, meanVel)
			meanSpeed += b.vel[n].Len()
		}
		meanSpeed /= nVel
		if meanSpeed < .01 {
			// NOTE:  meanSpeed is not reliable (probably due to portsNode interactions), which is why we also check rect below
			b.vel = map[node]Point{}
			return
		} else {
			updated = true
		}

		dt := math.Min(1.0/fps, 100/meanSpeed) // slow down time at high speeds to avoid oscillation
		for n := range b.nodes {
			b.vel[n] = b.vel[n].Mul(.5 + rand.Float64()) // a little noise to break up small oscillations
			n.Move(Pos(n).Add(b.vel[n].Mul(dt)).Sub(center))
		}
	}

	rect := ZR
	for n := range b.nodes {
		r := RectInParent(n)
		if n, ok := n.(*portsNode); ok {
			// portsNodes are later reposition()ed to the boundary; thus we must adjust for the margin (blockRadius)
			if n.out {
				r = r.Sub(Pt(blockRadius-2, 0))
			} else {
				r = r.Add(Pt(blockRadius-2, 0))
			}
		}
		if rect == ZR {
			rect = r
		} else {
			rect = rect.Union(r)
		}
	}
	for c := range b.conns {
		rect = rect.Union(RectInParent(c))
	}
	if rect == ZR {
		rect = Rectangle{ZP, Pt(16, 0)}
	}
	rect = rect.Inset(-blockRadius)
	if Rect(b).Size().Sub(rect.Size()).Len() < .01 {
		b.vel = map[node]Point{}
		return
	}
	b.SetRect(rect)

	return true
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
	nearest := nearestView(b, b.allFocusableViews(), MapTo(v, ZP, b), dirKey)
	if nearest != nil {
		SetKeyFocus(nearest)
	}
}

func (b *block) TookKeyFocus() { b.focused = true; Repaint(b) }
func (b *block) LostKeyFocus() { b.focused = false; Repaint(b) }

func (b *block) KeyPress(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		b.outermost().focusNearestView(KeyFocus(b), event.Key)
	case KeyBackspace, KeyDelete:
		switch v := KeyFocus(b).(type) {
		case *block:
			if v.node != nil {
				SetKeyFocus(v.node)
			}
		case *portsNode:
		case node:
			foc := View(b)
			in, out := v.inConns(), v.outConns()
			if len(in) > 0 {
				foc = in[len(in)-1].src.node
			}
			if (len(in) == 0 || event.Key == KeyDelete) && len(out) > 0 {
				foc = out[len(out)-1].dst.node
			}
			for _, c := range append(in, out...) {
				c.blk.removeConn(c)
			}
			b.removeNode(v)
			SetKeyFocus(foc)
		}
	case KeyEscape:
		if outer := b.outer(); outer != nil {
			SetKeyFocus(outer)
		} else {
			b.func_().Close()
		}
	default:
		if !(event.Ctrl || event.Alt || event.Super) {
			switch event.Text {
			default:
				f := b.func_()
				browser := newBrowser(browse, f.pkg(), f.imports())
				b.Add(browser)
				browser.Move(Center(b))
				browser.accepted = func(obj types.Object) {
					browser.Close()
					newNode(b, obj)
				}
				browser.canceled = func() {
					browser.Close()
					SetKeyFocus(b)
				}
				browser.text.KeyPress(event)
				SetKeyFocus(browser.text)
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

func newNode(b *block, obj types.Object) {
	var n node
	switch obj := obj.(type) {
	case special:
		switch obj.name {
		case "[]":
			n = newIndexNode(false)
		case "[]=":
			n = newIndexNode(true)
		case "addr":
			n = newValueNode(nil, true, false, false)
		case "call":
			n = newCallNode(nil)
		case "func":
			n = newFuncLiteralNode()
		case "if":
			n = newIfNode()
		case "indirect":
			n = newValueNode(nil, false, true, false)
		case "loop":
			n = newLoopNode()
		case "typeAssert":
			n = newTypeAssertNode()
		}
	case *types.Func, method:
		if unicode.IsLetter([]rune(obj.GetName())[0]) {
			n = newCallNode(obj)
		} else {
			n = newOperatorNode(obj)
		}
	case *types.Var, *types.Const, field:
		n = newValueNode(obj, false, false, false)
	default:
		panic("bad obj")
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

func (n *portsNode) reposition() {
	b := n.blk
	y := Center(b).Y
	if n.out {
		MoveOrigin(n, Pt(Rect(b).Max.X-2, y))
	} else {
		MoveOrigin(n, Pt(Rect(b).Min.X+2, y))
	}
}

func (n *portsNode) removePort(p *port) {
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
			f.subPkgRef((*vars)[i].Type)
			*vars = append((*vars)[:i], (*vars)[i+1:]...)
			n.removePortBase(p)
			SetKeyFocus(n)
			if f.obj == nil {
				f.output.setType(f.output.obj.Type)
			}
			n.reposition()
			break
		}
	}
}

func (n *portsNode) KeyPress(event KeyEvent) {
	if n.editable && event.Text == "," {
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
		n.reposition()
		Show(p.valView)
		p.valView.edit(func() {
			if v.Type != nil {
				*vars = append(*vars, v)
				f.addPkgRef(v.Type)
				SetKeyFocus(p)
			} else {
				n.removePortBase(p)
				n.reposition()
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
