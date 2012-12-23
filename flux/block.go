package main

import (
	."math"
	"time"
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
)

type block struct {
	*ViewBase
	node node
	nodes map[node]bool
	conns map[*connection]bool
	focused, editing bool
	editingNode node
	points []Point
	intermediatePoints []Point
}

func newBlock(n node) *block {
	b := &block{}
	b.ViewBase = NewView(b)
	b.node = n
	b.nodes = map[node]bool{}
	b.conns = map[*connection]bool{}
	b.points = []Point{ZP}
	b.Pan(Pt(-400, -300))
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
			b.func_().addPackageRef(n.info)
		}
	}
}

func (b *block) removeNode(n node) {
	b.RemoveChild(n)
	delete(b.nodes, n)
	switch n := n.(type) {
	case *callNode:
		b.func_().subPackageRef(n.info)
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

func (b *block) animate() {
	for {
		v := map[node]Point{}
		center := ZP
		for n := range b.nodes {
			v[n] = ZP
			center = center.Add(n.Position())
		}
		center = center.Div(float64(len(b.nodes)))
		for n1 := range b.nodes {
			if _, ok := n1.(*portsNode); ok { continue }
			for n2 := range b.nodes {
				if _, ok := n2.(*portsNode); ok { continue }
				if n2 == n1 { continue }
				dir := n1.MapToParent(n1.Center()).Sub(n2.MapToParent(n2.Center()))
				d := Sqrt(dir.X * dir.X + dir.Y * dir.Y)
				if d > 128 { continue }
				v[n1] = v[n1].Add(dir.Mul(2 * (128 - d) / (1 + d)))
			}
		}
		for c := range b.conns {
			xOffset := 64.0; if c.feedback { xOffset = -256 }
			src, dst := c.src, c.dst
			if src == nil || dst == nil { continue }
			d := dst.MapTo(dst.Position(), b).Sub(src.MapTo(src.Position(), b).Add(Pt(xOffset, 0)))
			d.X *= 2
			d.Y /= 2
			
			srcNode := src.node; for srcNode.block() != b { srcNode = srcNode.block().node }
			dstNode := dst.node; for dstNode.block() != b { dstNode = dstNode.block().node }
			if _, ok := srcNode.(*portsNode); ok { continue }
			if _, ok := dstNode.(*portsNode); ok { continue }
			v[srcNode] = v[srcNode].Add(d)
			v[dstNode] = v[dstNode].Sub(d)
		}
		for n, v := range v {
			v = v.Add(center.Sub(n.Position()).Div(4))
			n.Move(n.Position().Add(v.Mul(2 * .033)))
		}
		
		pts := []Point{}
		for n := range b.nodes {
			if _, ok := n.(*portsNode); ok { continue }
			r := n.MapRectToParent(n.Rect())
			pts = append(pts, r.Min, r.Max, Pt(r.Min.X, r.Max.Y), Pt(r.Max.X, r.Min.Y))
		}
		if len(pts) == 0 { pts = append(pts, ZP, Pt(0, 37), Pt(32, 18)) }
		iLowerLeft, lowerLeft := 0, pts[0]
		for i, p := range pts {
			if p.Y < lowerLeft.Y || p.Y == lowerLeft.Y && p.X < lowerLeft.X {
				iLowerLeft, lowerLeft = i, p
			}
		}
		pts[0], pts[iLowerLeft] = pts[iLowerLeft], pts[0]
		Sort(pts[1:], func(p1, p2 Point) bool {
			x := p1.Sub(lowerLeft).Cross(p2.Sub(lowerLeft))
			if x > 0 { return true }
			if x == 0 { return p1.X < p2.X }
			return false
		})
		N := len(pts)
		pts = append([]Point{pts[N-1]}, pts...)
		m := 1
		for i := 2; i <= N; i++ {
			for pts[m].Sub(pts[m - 1]).Cross(pts[i].Sub(pts[m - 1])) <= 0 {
				if m > 1 { m-- } else if i == N { break } else { i++ }
			}
			m++
			pts[m], pts[i] = pts[i], pts[m]
		}
		pts = pts[:m]
		center = ZP
		for _, p := range pts { center = center.Add(p) }
		center = center.Div(float64(len(pts)))
		for i, p := range pts {
			dir := p.Sub(center)
			d := dir.Len()
			if len(pts) > 3 { pts[i] = p.Add(dir.Mul(32 / d)) }
		}
		b.intermediatePoints = pts
		b.points = []Point{}
		for i := range pts {
			p1, p2 := pts[i], pts[(i + 1) % len(pts)]
			b.points = append(b.points, p1.Add(p2).Div(2))
		}
		
		rect := Rect(pts[0].X, pts[0].Y, pts[0].X, pts[0].Y)
		for _, p := range append(b.points, b.intermediatePoints...) {
			rect = rect.Union(Rect(p.X, p.Y, p.X, p.Y))
		}
		b.Resize(rect.Dx(), rect.Dy())
		b.Pan(rect.Min)
		
		if n, ok := b.node.(interface{positionblocks()}); ok { n.positionblocks() }
		
		time.Sleep(33 * time.Millisecond)
	}
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
	case KeyEsc:
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
					n := newNode(info[0].(Info), b)
					b.addNode(n)
					n.MoveCenter(b.Center())
					n.TakeKeyboardFocus()
				})
				browser.canceled.Connect(func(...interface{}) { b.TakeKeyboardFocus() })
				browser.text.KeyPressed(event)
				browser.text.TakeKeyboardFocus()
			case "\"":
				n := newStringConstantNode(b)
				n.text.SetEditable(true)
				b.addNode(n)
				n.MoveCenter(b.Center())
				n.text.TakeKeyboardFocus()
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
	n := len(b.points)
	for i := range b.points {
		p1, p2, p3 := b.points[i], b.intermediatePoints[(i + 1) % n], b.points[(i + 1) % n]
		DrawQuadratic([3]Point{p1, p2, p3}, int(p3.Sub(p2).Len() + p2.Sub(p1).Len()) / 8)
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
			p = n.newInput(&ValueInfo{}).port
		} else {
			p = n.newOutput(&ValueInfo{}).port
		}
		p.valueView.Show()
		p.valueView.edit(func() {
			if p.info.typ != nil {
				f := n.blk.func_()
				if n.out {
					f.info.typ.results = append(f.info.typ.results, p.info)
				} else {
					f.info.typ.parameters = append(f.info.typ.parameters, p.info)
				}
				f.addPackageRef(p.info.typ)
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

func (n portsNode) Paint() {
	SetColor(map[bool]Color{true:{.5, .5, 1, .5}, false:{1, 1, 1, .25}}[n.focused])
	// TODO:  draw half-circle instead
	for f := 1.0; f > .1; f /= 2 {
		SetPointSize(f * 12)
		DrawPoint(ZP)
	}
	n.nodeBase.Paint()
}
