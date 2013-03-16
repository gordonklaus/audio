package main

import (
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	"math"
	"math/rand"
)

type block struct {
	*ViewBase
	node node
	nodes map[node]bool
	conns map[*connection]bool
	focused, editing bool
	editingNode node
}

func newBlock(n node) *block {
	b := &block{}
	b.ViewBase = NewView(b)
	b.node = n
	b.nodes = map[node]bool{}
	b.conns = map[*connection]bool{}
	return b
}

func (b *block) outer() *block { return b.node.block() }
func (b *block) outermost() *block {
	if outer := b.outer(); outer != nil {
		return outer.outermost()
	}
	return b
}
func (b *block) func_() *funcNode { return b.outermost().node.(*funcNode) }

func (b *block) addNode(n node) {
	if !b.nodes[n] {
		b.AddChild(n)
		b.nodes[n] = true
		switch n := n.(type) {
		case *callNode:
			b.func_().addPkgRef(n.info)
		}
	}
}

func (b *block) removeNode(n node) {
	b.RemoveChild(n)
	delete(b.nodes, n)
	switch n := n.(type) {
	case *callNode:
		b.func_().subPkgRef(n.info)
	}
}

func (b *block) addConnection(c *connection) {
	if c.blk != nil {
		delete(c.blk.conns, c)
	}
	c.blk = b
	b.AddChild(c)
	c.Lower()
	b.conns[c] = true
}

func (b *block) removeConnection(c *connection) {
	c.disconnect()
	delete(b.conns, c)
	b.RemoveChild(c)
}

func (b block) allNodes() (nodes []node) {
	for n := range b.nodes {
		nodes = append(nodes, n)
		switch n := n.(type) {
		case *ifNode:
			nodes = append(nodes, append(n.falseblk.allNodes(), n.trueblk.allNodes()...)...)
		case *loopNode:
			nodes = append(nodes, n.loopblk.allNodes()...)
		}
	}
	return
}

func (b block) allConns() (conns []*connection) {
	for c := range b.conns {
		conns = append(conns, c)
	}
	for n := range b.nodes {
		switch n := n.(type) {
		case *ifNode:
			conns = append(conns, append(n.falseblk.allConns(), n.trueblk.allConns()...)...)
		case *loopNode:
			conns = append(conns, n.loopblk.allConns()...)
		}
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
		if visitedThisCall[n] { return false }
		visitedThisCall[n] = true
		
		if !visited[n] {
			visited[n] = true
conns:		for _, c := range n.inConns() {
				if b.conns[c] {
					src := c.src.node
					for !b.nodes[src] {
						src = src.block().node
						if src == nil { continue conns }
					}
					if !insertInOrder(src, visitedThisCall.Copy()) { return false }
				}
			}
			order = append(order, n)
		}
		return true
	}
	
	endNodes := []node{}
nx:	for n := range b.nodes {
		for _, c := range n.outConns() {
			if c.blk == b { continue nx }
		}
		endNodes = append(endNodes, n)
	}
	if len(endNodes) == 0 && len(b.nodes) > 0 { return }
	
	for _, n := range endNodes {
		if !insertInOrder(n, Set{}) { return }
	}
	ok = true
	return
}

func (b *block) startEditing() {
	b.TakeKeyboardFocus()
	b.editing = true
}

func (b *block) stopEditing() {
	b.editing = false
	b.editingNode = nil
}

func (b *block) update() (updated bool) {
	for n := range b.nodes {
		if n, ok := n.(interface { update() bool }); ok {
			updated = updated || n.update()
		}
	}
	
	v := map[node]Point{}
	
	const (
		nodeCenterCoef = .25
		nodeSep = 32
		nodeSepCoef = 2
		nodeConnSep = 32
		nodeConnSepCoef = 2
		connLen = 64
		connLenFB = 256
		connAngleCoef = 16
		connContractCoef = .25
		connExpandCoef = 2
		topSpeed = 200
		speedCompress = 5
	)
	
	for n1 := range b.nodes {
		for n2 := range b.nodes {
			if n2 == n1 { continue }
			dir := n1.MapToParent(n1.Center()).Sub(n2.MapToParent(n2.Center()))
			if dir == ZP {
				dir = Pt(rand.NormFloat64(), rand.NormFloat64())
			}
			d := dir.Len() - n1.Size().Add(n2.Size()).Len() / 2 - nodeSep
			if d > 0 { continue }
			v[n1] = v[n1].Add(dir.Mul(-nodeSepCoef * d / dir.Len()))
		}
		for c := range b.conns {
			src, dst := c.src, c.dst
			if src == nil || src.node == n1 || dst == nil || dst.node == n1 { continue }
			x := src.MapTo(src.Center(), b)
			y := dst.MapTo(dst.Center(), b)
			p := n1.MapToParent(n1.Center())
			xy := y.Sub(x)
			xp := p.Sub(x)
			proj := ZP
			switch t := xp.Dot(xy) / xy.Dot(xy); {
			case t <= 0:  proj = x
			default:      proj = x.Add(xy.Mul(t))
			case t >= 1:  proj = y
			}
			dir := p.Sub(proj)
			if dir == ZP {
				dir = Pt(rand.NormFloat64(), rand.NormFloat64())
			}
			d := dir.Len() - n1.Size().Len() / 2 - nodeConnSep
			if d > 0 { continue }
			v[n1] = v[n1].Add(dir.Mul(-nodeConnSepCoef * d / dir.Len()))
		}
	}
	for c := range b.conns {
		src, dst := c.src, c.dst
		if src == nil || dst == nil { continue }
		d := dst.MapTo(dst.Center(), b).Sub(src.MapTo(src.Center(), b))
		l := d.Len()
		var offset, rot float64 = connLen, 0
		if c.feedback {
			offset, rot = connLenFB, .5
		}
		l = (offset - l) / l
		if l < 0 {
			l *= connContractCoef
		} else {
			l *= connExpandCoef
		}
		angle := math.Mod(d.Angle() / 2 / math.Pi + rot + .5, 1) - .5
		u := Pt(d.Y, -d.X).Mul(connAngleCoef * 2 * angle).Add(d.Mul(l))
		
		srcNode := src.node; for srcNode.block() != b { srcNode = srcNode.block().node }
		dstNode := dst.node; for dstNode.block() != b { dstNode = dstNode.block().node }
		v[srcNode] = v[srcNode].Sub(u)
		v[dstNode] = v[dstNode].Add(u)
	}
	
	center := ZP
	for n := range b.nodes {
		center = center.Add(n.Position())
	}
	center = center.Div(float64(len(b.nodes)))
	for n := range b.nodes {
		v[n] = v[n].Add(center.Sub(n.Position().Mul(nodeCenterCoef)))
		l := v[n].Len()
		v[n] = v[n].Mul(math.Tanh(speedCompress * l / topSpeed) * topSpeed / l)
	}
	
	meanVel := ZP
	nVel := 0.0
	for n := range b.nodes {
		if _, ok := n.(*portsNode); ok { continue }
		meanVel = meanVel.Add(v[n])
		nVel++
	}
	// if there is only one (non-port) node then it is definitely stationary
	// but we still have to check if the rect needs updating
	if nVel > 1 {
		meanVel = meanVel.Div(nVel)
		meanSpeed := 0.0
		for n := range b.nodes {
			if _, ok := n.(*portsNode); ok { continue }
			meanSpeed += v[n].Sub(meanVel).Len()
		}
		meanSpeed /= nVel
		if meanSpeed < .01 {
			return
		}
		
		for n := range b.nodes {
			n.Move(n.Position().Add(v[n].Mul(2.0 / fps)).Sub(center))
		}
	}
	
	rect := ZR
	for n := range b.nodes {
		if _, ok := n.(*portsNode); ok { continue }
		r := n.MapRectToParent(n.Rect())
		r.Min.X -= 16
		r.Max.X += 16
		if rect == ZR {
			rect = r
		} else {
			rect = rect.Union(r)
		}
	}
	if rect == ZR {
		rect = Rect(0, 0, 48, 32)
	}
	s := rect.Size().Mul((math.Sqrt2 - 1) / 2)
	rect.Min = rect.Min.Sub(s)
	rect.Max = rect.Max.Add(s)
	if b.Rect() == rect { return }
	b.Resize(rect.Dx(), rect.Dy())
	b.Pan(rect.Min)
	
	return true
}

func nearestView(parent View, views []View, p Point, dirKey int) (nearest View) {
	dir := map[int]Point{KeyLeft:{-1, 0}, KeyRight:{1, 0}, KeyUp:{0, 1}, KeyDown:{0, -1}}[dirKey]
	best := 0.0
	for _, v := range views {
		d := v.MapTo(v.Center(), parent).Sub(p)
		score := (dir.X * d.X + dir.Y * d.Y) / (d.X * d.X + d.Y * d.Y)
		if (score > best) {
			best = score
			nearest = v
		}
	}
	return
}

func (b *block) focusNearestView(v View, dirKey int) {
	views := []View{}
	for _, n := range b.allNodes() {
		views = append(views, n)
		for _, p := range n.inputs() { views = append(views, p) }
		for _, p := range n.outputs() { views = append(views, p) }
	}
	for _, c := range b.allConns() {
		views = append(views, c)
	}
	nearest := nearestView(b, views, v.MapTo(v.Center(), b), dirKey)
	if nearest != nil { nearest.TakeKeyboardFocus() }
}

func (b *block) TookKeyboardFocus() { b.focused = true; b.Repaint() }
func (b *block) LostKeyboardFocus() { b.focused = false; b.stopEditing(); b.Repaint() }

func (b *block) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		outermost := b.outermost()
		if b.editing {
			var v View = b.editingNode
			if v == nil { v = b }
			views := []View{}; for _, n := range outermost.allNodes() { views = append(views, n) }
			if n := nearestView(b, views, v.MapTo(v.Center(), outermost), event.Key); n != nil { b.editingNode = n.(node) }
		} else {
			outermost.focusNearestView(b, event.Key)
		}
	case KeySpace:
		if b.editingNode != nil {
			if b.nodes[b.editingNode] {
				b.RemoveChild(b.editingNode)
				delete(b.nodes, b.editingNode)
				b.addNode(b.editingNode)
			} else {
				b.removeNode(b.editingNode)
				b.nodes[b.editingNode] = true
				b.AddChild(b.editingNode)
			}
		}
	case KeyEnter:
		if b.editing {
			if b.editingNode != nil && !b.nodes[b.editingNode] {
				b.nodes[b.editingNode] = true
			}
			b.stopEditing()
		} else {
			b.startEditing()
		}
	case KeyBackspace, KeyDelete:
		foc := View(b)
		switch v := b.GetKeyboardFocus().(type) {
		case *block:
			if v.node != nil {
				foc = v.node
			}
		case *portsNode:
			return
		case node:
			in, out := v.inConns(), v.outConns()
			switch {
			case event.Key == KeyBackspace && len(in) > 0:  foc = in[0].src.node
			case event.Key == KeyDelete       && len(out) > 0: foc = out[0].dst.node
			default:
				switch {
				case len(in) > 0:  foc = in[0].src.node
				case len(out) > 0: foc = out[0].dst.node
				}
			}
			for _, c := range append(in, out...) {
				c.blk.removeConnection(c)
			}
			b.removeNode(v)
		case *input:
			foc = v.node
			if event.Key == KeyBackspace && len(v.conns) > 0 {
				foc = v.conns[0]
			}
		case *output:
			foc = v.node
			if event.Key == KeyDelete && len(v.conns) > 0 {
				foc = v.conns[0]
			}
		case *connection:
			if event.Key == KeyBackspace {
				foc = v.src
			} else {
				foc = v.dst
			}
			v.blk.removeConnection(v)
		}
		foc.TakeKeyboardFocus()
	case KeyEscape:
		if b.editing {
			if b.editingNode != nil {
				b.editingNode = nil
			} else {
				b.stopEditing()
			}
		} else if outer := b.outer(); outer != nil {
			outer.TakeKeyboardFocus()
		}
	default:
		if !(event.Ctrl || event.Alt || event.Super) {
			switch event.Text {
			default:
				f := b.func_()
				browser := newBrowser(browse, f.pkg(), f.imports())
				b.AddChild(browser)
				browser.Move(b.Center())
				browser.accepted.Connect(func(info ...interface{}) {
					browser.Close()
					n := newNode(info[0].(Info), b)
					b.addNode(n)
					n.MoveCenter(b.Center())
					n.TakeKeyboardFocus()
				})
				browser.canceled.Connect(func(...interface{}) {
					browser.Close()
					b.TakeKeyboardFocus()
				})
				browser.text.KeyPressed(event)
				browser.text.TakeKeyboardFocus()
			case "\"":
				n := newStringConstantNode(b)
				b.addNode(n)
				n.MoveCenter(b.Center())
				n.text.TakeKeyboardFocus()
			case "{":
				n := newCompositeLiteralNode(b)
				b.addNode(n)
				n.MoveCenter(b.Center())
				n.editType()
			case "":
				b.ViewBase.KeyPressed(event)
			}
		} else {
			b.ViewBase.KeyPressed(event)
		}
	}
}

func (b block) Paint() {
	if b.editing {
		SetColor(Color{.7, .4, 0, 1})
	} else if b.focused {
		SetColor(Color{.3, .3, .7, 1})
	} else {
		SetColor(Color{.5, .5, .5, 1})
	}
	{
		x := Pt(b.Rect().Dx() / 2, 0)
		y := Pt(0, b.Rect().Dy() / 2)
		c := b.Rect().Center()
		r, t, l, b := c.Add(x), c.Add(y), c.Sub(x), c.Sub(y)
		tr, tl, bl, br := c.Add(x).Add(y), c.Sub(x).Add(y), c.Sub(x).Sub(y), c.Add(x).Sub(y)
		steps := int(x.X + y.Y)
		DrawQuadratic([3]Point{r, tr, t}, steps)
		DrawQuadratic([3]Point{t, tl, l}, steps)
		DrawQuadratic([3]Point{l, bl, b}, steps)
		DrawQuadratic([3]Point{b, br, r}, steps)
	}
}


type portsNode struct {
	*nodeBase
	out bool
	editable bool
}

func newInputsNode(b *block) *portsNode { return newPortsNode(false, b) }
func newOutputsNode(b *block) *portsNode { return newPortsNode(true, b) }
func newPortsNode(out bool, b *block) *portsNode {
	n := &portsNode{out:out}
	n.nodeBase = newNodeBase(n, b)
	return n
}

func (n *portsNode) KeyPressed(event KeyEvent) {
	if n.editable && event.Text == "," {
		var p *port
		if n.out {
			p = n.newInput(&Value{}).port
		} else {
			p = n.newOutput(&Value{}).port
		}
		p.valView.Show()
		p.valView.edit(func() {
			if p.val.typ != nil {
				f := n.blk.func_()
				if n.out {
					f.info.typ.results = append(f.info.typ.results, p.val)
				} else {
					f.info.typ.parameters = append(f.info.typ.parameters, p.val)
				}
				f.addPkgRef(p.val.typ)
				p.TakeKeyboardFocus()
			} else {
				n.RemoveChild(p.Self)
				n.TakeKeyboardFocus()
			}
		})
	} else {
		n.nodeBase.KeyPressed(event)
	}
}
