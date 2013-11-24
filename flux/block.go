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

	g    map[node]*Point
	step map[node]*Point
}

func newBlock(n node) *block {
	b := &block{}
	b.ViewBase = NewView(b)
	b.node = n
	b.nodes = map[node]bool{}
	b.conns = map[*connection]bool{}
	b.g = make(map[node]*Point)
	b.step = make(map[node]*Point)
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
		n.Move(Pt(100*rand.NormFloat64(), 100*rand.NormFloat64()))
		b.nodes[n] = true
		n.setBlock(b)
		switch n := n.(type) {
		case *callNode:
			if _, ok := n.obj.(method); !ok && n.obj != nil {
				b.func_().addPkgRef(n.obj)
			}
		}
		b.g[n] = &Point{}
		b.step[n] = &Point{.01, .01}
		if f := b.func_(); f != nil {
			f.wakeUp()
		}
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
		}
		delete(b.g, n)
		delete(b.step, n)
		if f := b.func_(); f != nil {
			f.wakeUp()
		}
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
		if src := parentOrSelfInBlock(c.src.node, b); src != nil {
			srcs = append(srcs, src)
		}
	}
	for _, c := range n.outConns() {
		if !c.feedback || c.dst == nil {
			continue
		}
		if dst := parentOrSelfInBlock(c.dst.node, b); dst != nil && dst != n {
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
		if dst := parentOrSelfInBlock(c.dst.node, b); dst != nil {
			dsts = append(dsts, dst)
		}
	}
	for _, c := range n.inConns() {
		if !c.feedback || c.src == nil {
			continue
		}
		if src := parentOrSelfInBlock(c.src.node, b); src != nil && src != n {
			dsts = append(dsts, src)
		}
	}
	return
}

func parentOrSelfInBlock(n node, blk *block) node {
	for b := n.block(); b != nil; n, b = b.node, b.outer() {
		if b == blk {
			return n
		}
	}
	return nil
}

// TODO: treat conns as curves, not lines (or make them more linear?)
func (b *block) arrange() bool {
	done := true
	
	for n := range b.nodes {
		if n, ok := n.(interface {
			arrange() bool
		}); ok {
			if !n.arrange() {
				done = false
			}
		}
	}

	const (
		portsNodeCoef   = 1024
		nodeCenterCoef  = .5
		nodeSep         = 8
		nodeSepCoef     = 1024
		nodeConnSep     = 8
		nodeConnSepCoef = 8
		connLen         = 16
		connLenFB       = -256
		connLenCoef     = .1
		connConnSep     = 4
		connConnSepCoef = 1
		gMax            = 32
	)

	g := make(map[node]*Point)
	add := func(n node, dv Point) {
		*g[n] = g[n].Add(dv)
	}
	sub := func(n node, dv Point) {
		*g[n] = g[n].Sub(dv)
	}
	for n := range b.nodes {
		g[n] = &Point{}
	}

	for n1 := range b.nodes {
		if n1, ok := n1.(*portsNode); ok {
			p := Center(b)
			if n1.out {
				p.X = Rect(b).Max.X - 2
			} else {
				p.X = Rect(b).Min.X + 2
			}
			dir := p.Sub(MapToParent(n1, ZP))
			dir.X *= math.Abs(dir.X)
			if len(n1.ins)+len(n1.outs) == 0 {
				dir.Y *= math.Abs(dir.Y)
			} else {
				dir.Y = 0
			}
			add(n1, dir.Mul(portsNodeCoef))
		}
		for n2 := range b.nodes {
			if n2 == n1 {
				continue
			}
			dir := CenterInParent(n1).Sub(CenterInParent(n2))
			r1, r2 := RectInParent(n1).Inset(-nodeSep/2), RectInParent(n2).Inset(-nodeSep/2)
			sep := math.Min(r1.Intersect(r2).Size().XY())
			add(n1, dir.Mul(nodeSepCoef*math.Min(1, sep/nodeSep)/dir.Len()))
		}
		for c := range b.conns {
			if c.src == nil || c.dst == nil {
				continue
			}
			srcNode := parentOrSelfInBlock(c.src.node, b)
			dstNode := parentOrSelfInBlock(c.dst.node, b)
			if srcNode == n1 || dstNode == n1 {
				continue
			}
			p := CenterInParent(n1)
			x := MapToParent(c, c.srcPt)
			y := MapToParent(c, c.dstPt)
			dir := p.Sub(PointToLine(p, x, y))
			maxSep := Size(n1).Len()/2 + nodeConnSep
			sep := dir.Len()
			if sep < maxSep {
				d := dir.Mul(nodeConnSepCoef * (maxSep/sep - 1))
				add(n1, d)
				sub(srcNode, d)
				sub(dstNode, d)
			}
		}
	}
	for c := range b.conns {
		if c.src == nil || c.dst == nil {
			continue
		}
		// FIXME: srcPts and dstPts need to be mapped to a common parent
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
		// 			add(c.src.node, d)
		// 			add(c.dst.node, d)
		// 			sub(c2.src.node, d)
		// 			sub(c2.dst.node, d)
		// 		} else if l < connConnSep {
		// 			d = d.Mul((connConnSep - l) / l)
		// 			add(c.src.node, d)
		// 			add(c.dst.node, d)
		// 			sub(c2.src.node, d)
		// 			sub(c2.dst.node, d)
		// 		}
		// 	}
		// }

		// TODO: connections that cross block boundaries should be longer.  map src and dst points to the boundaries
		d := c.dstPt.Sub(c.srcPt)
		if c.feedback {
			d.X -= connLenFB
		} else {
			d.X -= connLen
		}
		d.X -= math.Abs(d.Y) / 2
		if d.X < 0 {
			d.X *= d.X * d.X
		}
		d.Y *= math.Abs(d.Y) / 16
		d = d.Mul(connLenCoef)
		add(parentOrSelfInBlock(c.src.node, b), d)
		sub(parentOrSelfInBlock(c.dst.node, b), d)
	}

	avg := ZP
	for n := range b.nodes {
		avg = avg.Add(*g[n])
	}
	avg = avg.Div(float64(len(b.nodes)))
	for n := range b.nodes {
		sub(n, avg)
	}

	center := ZP
	for n := range b.nodes {
		center = center.Add(Pos(n))
	}
	center = center.Div(float64(len(b.nodes)))
	for n := range b.nodes {
		add(n, center.Sub(Pos(n)).Mul(nodeCenterCoef))
		l := g[n].Len()
		if l > 0 {
			*g[n] = g[n].Mul(math.Tanh(l/gMax) * gMax / l)
		}
	}

	for n := range b.nodes {
		d := Pt(rprop(&b.g[n].X, &g[n].X, &b.step[n].X), rprop(&b.g[n].Y, &g[n].Y, &b.step[n].Y))
		n.Move(Pos(n).Add(d))
		if d.Len() > .01 {
			done = false
		}
	}

	rect := ZR
	for n := range b.nodes {
		r := RectInParent(n)
		if n, ok := n.(*portsNode); ok {
			// portsNodes gravitate to the boundary; thus we must adjust for the margin
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
	if Rect(b).Size().Sub(rect.Size()).Len() > .01 {
		done = false
	}
	b.SetRect(rect)

	return done
}

func rprop(g_, g, step *float64) float64 {
	prod := *g * *g_
	if prod > 0 {
		*step *= 1.2
	} else if prod < 0 {
		*step *= .5
		// *g = 0 //seems unnecessary
	}
	*step = math.Max(.001, *step)
	*step = math.Min(1, *step)
	*g_ = *g
	return *g * *step
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
					newNode(b, obj, browser.funcAsVal)
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
			n = newFuncLiteralNode()
		case "if":
			n = newIfNode()
		case "indirect":
			n = newValueNode(nil, false)
		case "loop":
			n = newLoopNode()
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
	case *types.Var, *types.Const, field:
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
				if f := n.blk.func_(); f != nil {
					f.wakeUp()
				}
				break
			}
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
		if f := n.blk.func_(); f != nil {
			f.wakeUp()
		}
		Show(p.valView)
		p.valView.edit(func() {
			if v.Type != nil {
				*vars = append(*vars, v)
				n.blk.func_().addPkgRef(v.Type)
				SetKeyFocus(p)
			} else {
				n.removePortBase(p)
				if f := n.blk.func_(); f != nil {
					f.wakeUp()
				}
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
