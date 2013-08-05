package main

import (
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
	"code.google.com/p/go.exp/go/types"
	"go/token"
	"math"
	"math/rand"
	"unicode"
)

const blockRadius = 16

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
		n.setBlock(b)
		switch n := n.(type) {
		case *callNode:
			b.func_().addPkgRef(n.obj)
		}
	}
}

func (b *block) removeNode(n node) {
	b.RemoveChild(n)
	delete(b.nodes, n)
	switch n := n.(type) {
	case *callNode:
		b.func_().subPkgRef(n.obj)
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
		speedCompress = 1
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
conns:
		for c := range b.conns {
			src, dst := c.src, c.dst
			if src == nil || dst == nil { continue }
			for n := src.node; n.block() != nil; n = n.block().node { if n == n1 { continue conns } }
			for n := dst.node; n.block() != nil; n = n.block().node { if n == n1 { continue conns } }
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
		meanVel = meanVel.Add(v[n])
		nVel++
	}
	// if there is only one (non-port) node then it is definitely stationary
	// but we still have to check if the rect needs updating
	if nVel > 1 {
		meanVel = meanVel.Div(nVel)
		meanSpeed := 0.0
		for n := range b.nodes {
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
		r := n.MapRectToParent(n.Rect())
		if n, ok := n.(*portsNode); ok {
			// portsNodes are later reposition()ed to the boundary; thus we must adjust for the margin (blockRadius)
			if n.out {
				r = r.Sub(Pt(blockRadius, 0))
			} else {
				r = r.Add(Pt(blockRadius, 0))
			}
		}
		if rect == ZR {
			rect = r
		} else {
			rect = rect.Union(r)
		}
	}
	for c := range b.conns {
		rect = rect.Union(c.MapRectToParent(c.Rect()))
	}
	if rect == ZR {
		rect = Rect(0, 0, 16, 0)
	}
	rect = rect.Inset(-blockRadius)
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
		for _, p := range append(n.inputs(), n.outputs()...) {
			views = append(views, p)
		}
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
		switch v := b.GetKeyboardFocus().(type) {
		case *block:
			if v.node != nil {
				v.node.TakeKeyboardFocus()
			}
		case *portsNode:
		case node:
			foc := View(b)
			in, out := v.inConns(), v.outConns()
			if len(in) > 0 {
				foc = in[len(in) - 1].src.node
			}
			if (len(in) == 0 || event.Key == KeyDelete) && len(out) > 0 {
				foc = out[len(out) - 1].dst.node
			}
			for _, c := range append(in, out...) {
				c.blk.removeConnection(c)
			}
			b.removeNode(v)
			foc.TakeKeyboardFocus()
		}
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
				browser.accepted = func(obj types.Object) {
					browser.Close()
					newNode(b, obj)
				}
				browser.canceled = func() {
					browser.Close()
					b.TakeKeyboardFocus()
				}
				browser.text.KeyPressed(event)
				browser.text.TakeKeyboardFocus()
			case "\"", "'", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				text := event.Text
				kind := token.INT
				switch event.Text {
				case "\"": kind, text = token.STRING, ""
				case "'":  kind = token.CHAR
				}
				n := newBasicLiteralNode(kind)
				b.addNode(n)
				n.MoveCenter(b.Center())
				n.text.SetText(text)
				n.text.Reject = func() {
					b.removeNode(n)
					b.TakeKeyboardFocus()
				}
				n.text.TakeKeyboardFocus()
			case "{":
				n := newCompositeLiteralNode()
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

func newNode(b *block, obj types.Object) {
	var n node
	switch obj := obj.(type) {
	case special:
		switch obj.name {
		case "[]":                        n = newIndexNode(false)
		case "[]=":                       n = newIndexNode(true)
		case "if":                        n = newIfNode()
		case "indirect":                  n = newValueNode(nil, true, false)
		case "loop":                      n = newLoopNode()
		}
	case *types.Func, method:
		switch unicode.IsLetter([]rune(obj.GetName())[0]) {
		case true:                        n = newCallNode(obj)
		case false:                       n = newOperatorNode(obj)
		}
	case *types.Var, *types.Const, field: n = newValueNode(obj, false, false)
	default:                              panic("bad obj")
	}
	b.addNode(n)
	n.MoveCenter(b.Center())
	if nn, ok := n.(interface{editType()}); ok {
		nn.editType()
	} else {
		n.TakeKeyboardFocus()
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
		rect := b.Rect()
		l, r, b, t := rect.Min.X, rect.Max.X, rect.Min.Y, rect.Max.Y
		lb, bl := Pt(l, b + blockRadius), Pt(l + blockRadius, b)
		rb, br := Pt(r, b + blockRadius), Pt(r - blockRadius, b)
		rt, tr := Pt(r, t - blockRadius), Pt(r - blockRadius, t)
		lt, tl := Pt(l, t - blockRadius), Pt(l + blockRadius, t)
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
	out bool
	editable bool
}

func newInputsNode() *portsNode { return newPortsNode(false) }
func newOutputsNode() *portsNode { return newPortsNode(true) }
func newPortsNode(out bool) *portsNode {
	n := &portsNode{out:out}
	n.nodeBase = newNodeBase(n)
	return n
}

func (n *portsNode) reposition() {
	b := n.blk
	y := b.Center().Y
	if n.out {
		n.MoveOrigin(Pt(b.Rect().Max.X, y))
	} else {
		n.MoveOrigin(Pt(b.Rect().Min.X, y))
	}
}

func (n *portsNode) removePort(p *port) {
	f := n.blk.func_()
	sig := f.obj.GetType().(*types.Signature)
	var ports []*port
	var vars *[]*types.Var
	if p.out {
		ports = n.outs
		if _, ok := f.obj.(method); ok { // don't remove receiver
			ports = ports[1:]
		}
		vars = &sig.Params
	} else {
		ports = n.ins
		vars = &sig.Results
	}
	for i, q := range ports {
		if q == p {
			SliceRemove(vars, (*vars)[i])
			f.subPkgRef(p.obj)
			for _, c := range p.conns {
				c.blk.removeConnection(c)
			}
			n.removePortBase(p)
			n.TakeKeyboardFocus()
			break
		}
	}
	n.reposition()
}

func (n *portsNode) KeyPressed(event KeyEvent) {
	if n.editable && event.Text == "," {
		f := n.blk.func_()
		sig := f.obj.GetType().(*types.Signature)
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
		p.valView.Show()
		p.valView.edit(func() {
			if v.Type != nil {
				*vars = append(*vars, v)
				f.addPkgRef(p.obj)
				p.TakeKeyboardFocus()
			} else {
				n.removePortBase(p)
				n.reposition()
				n.TakeKeyboardFocus()
			}
		})
	} else {
		n.nodeBase.KeyPressed(event)
	}
}
