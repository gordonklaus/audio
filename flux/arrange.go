package main

import (
	. "code.google.com/p/gordon-go/gui"
	"math"
	// "math/rand"
)

type blockArrange struct {
	*ViewBase
	block *block
	node  *nodeArrange
	nodes []*nodeArrange
	conns []*connArrange

	// cross int
}

func newBlockArrange(b *block, n *nodeArrange, ports map[*port]*portArrange) *blockArrange {
	a := &blockArrange{block: b, node: n}
	a.ViewBase = NewView(a)
	a.Move(Pos(b))
	a.SetRect(Rect(b))
	for n := range b.nodes {
		n := newNodeArrange(n, a, ports)
		a.nodes = append(a.nodes, n)
		a.Add(n)
	}
	for c := range b.conns {
		c := newConnArrange(c, ports)
		a.conns = append(a.conns, c)
		a.Add(c)
	}
	// a.cross = math.MaxInt32
	return a
}

func (b *blockArrange) find(n *nodeArrange) *nodeArrange {
	for n.block != b {
		n = n.block.node
	}
	return n
}

type nodeArrange struct {
	*ViewBase
	node     node
	block    *blockArrange
	ports    []*portArrange
	blocks   []*blockArrange
	hasConns bool

	g, g_, step Point

	vel Point
}

func newNodeArrange(n node, b *blockArrange, ports map[*port]*portArrange) *nodeArrange {
	a := &nodeArrange{node: n, block: b}
	a.ViewBase = NewView(a)
	a.Move(Pos(n))
	a.SetRect(Rect(n))
	for _, port := range append(n.inputs(), n.outputs()...) {
		p := newPortArrange(port, a)
		a.ports = append(a.ports, p)
		a.Add(p)
		if len(port.conns) > 0 {
			a.hasConns = true
		}
		ports[port] = p
	}
	switch n := n.(type) {
	case *ifNode:
		f, t := newBlockArrange(n.falseblk, a, ports), newBlockArrange(n.trueblk, a, ports)
		a.blocks = []*blockArrange{f, t}
		a.Add(f)
		a.Add(t)
	case *loopNode:
		b := newBlockArrange(n.loopblk, a, ports)
		a.blocks = []*blockArrange{b}
		a.Add(b)
	case *funcNode:
		b := newBlockArrange(n.funcblk, a, ports)
		a.blocks = []*blockArrange{b}
		a.Add(b)
	}
	a.step = Pt(1, 1)
	return a
}

func (n *nodeArrange) add(d Point) { n.g = n.g.Add(d) }
func (n *nodeArrange) sub(d Point) { n.g = n.g.Sub(d) }

func (n *nodeArrange) arrange() bool {
	done := true
	for _, b := range n.blocks {
		if !b.arrange() {
			done = false
		}
	}
	if done {
		return true
	}
	switch node := n.node.(type) {
	case *ifNode:
		falseblk, trueblk := n.blocks[0], n.blocks[1]
		falseblk.Move(Pt(-blockRadius, -4-Height(falseblk)))
		trueblk.Move(Pt(-blockRadius, 4))
		ResizeToFit(n, 0)
	case *loopNode:
		b := n.blocks[0]
		b.Move(Pt(portSize/2, -Height(b)/2))
		seqOut := n.ports[2]
		MoveCenter(seqOut, Pt(portSize/2+Width(b), 0))
		ResizeToFit(n, 0)
	case *funcNode:
		b := n.blocks[0]
		b.Move(Pt(-Width(b)-portSize/2, -Height(b)/2))
		if node.lit() {
			output := n.ports[0]
			MoveCenter(output, Pt(portSize/2, 0))
		}
		ResizeToFit(n, 0)
	}
	return false
}

func (n *nodeArrange) animate() bool {
	done := true
	pos := Pos(n.node)
	d := Pos(n).Sub(pos)
	if n.block != nil && d.Len() > .1 { // check for nil block because top level funcNode position isn't stable
		n.vel = n.vel.Add(d.Mul(60 / fps)).Mul(.8)
		n.node.Move(pos.Add(n.vel.Div(fps)))
		done = false
	} else {
		n.node.Move(Pos(n))
	}
	n.node.SetRect(Rect(n))
	for _, p := range n.ports {
		p.port.Move(Pos(p))
	}
	for _, b := range n.blocks {
		if !b.animate() {
			done = false
		}
	}
	return done
}

type portArrange struct {
	*ViewBase
	port *port
	node *nodeArrange
}

func newPortArrange(p *port, n *nodeArrange) *portArrange {
	a := &portArrange{port: p, node: n}
	a.ViewBase = NewView(a)
	a.Move(Pos(p))
	a.SetRect(Rect(p))
	return a
}

func (p *portArrange) centerIn(b *blockArrange) Point {
	return MapTo(p, Center(p), b)
}

type connArrange struct {
	*ViewBase
	conn     *connection
	src, dst *portArrange
}

func newConnArrange(c *connection, ports map[*port]*portArrange) *connArrange {
	a := &connArrange{conn: c}
	a.ViewBase = NewView(a)
	a.src = ports[c.src]
	a.dst = ports[c.dst]
	return a
}

// TODO: treat conns as curves, not lines (or make them more linear?)
func (b *blockArrange) arrange() bool {
	done := true

	for _, n := range b.nodes {
		if !n.arrange() {
			done = false
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
		gMax            = 32
	)

	for _, n := range b.nodes {
		n.g = ZP
	}

	for _, n1 := range b.nodes {
		if pn, ok := n1.node.(*portsNode); ok {
			p := Center(b)
			if pn.out {
				p.X = Rect(b).Max.X - 2
			} else {
				p.X = Rect(b).Min.X + 2
			}
			dir := p.Sub(MapToParent(n1, ZP))
			if n1.hasConns {
				dir.Y = 0
			}
			n1.add(dir.Mul(portsNodeCoef * dir.Len()))
		}
		for _, n2 := range b.nodes {
			if n2 == n1 {
				continue
			}
			dir := CenterInParent(n1).Sub(CenterInParent(n2))
			r1, r2 := RectInParent(n1).Inset(-nodeSep/2), RectInParent(n2).Inset(-nodeSep/2)
			sep := math.Min(r1.Intersect(r2).Size().XY())
			n1.add(dir.Mul(nodeSepCoef * math.Min(1, sep/nodeSep) / dir.Len()))
		}
		for _, c := range b.conns {
			if c.src == nil || c.dst == nil {
				continue
			}
			srcNode := b.find(c.src.node)
			dstNode := b.find(c.dst.node)
			if srcNode == n1 || dstNode == n1 {
				continue
			}
			p := CenterInParent(n1)
			x := c.src.centerIn(b)
			y := c.dst.centerIn(b)
			dir := p.Sub(PointToLine(p, x, y))
			maxSep := Size(n1).Len()/2 + nodeConnSep
			sep := dir.Len()
			if sep < maxSep {
				d := dir.Mul(nodeConnSepCoef * (maxSep/sep - 1))
				n1.add(d)
				srcNode.sub(d)
				dstNode.sub(d)
			}
		}
	}
	for _, c := range b.conns {
		if c.src == nil || c.dst == nil {
			continue
		}
		// TODO: connections that cross block boundaries should be longer.  map src and dst points to the boundaries
		d := c.dst.centerIn(b).Sub(c.src.centerIn(b))
		if c.conn.feedback {
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
		b.find(c.src.node).add(d)
		b.find(c.dst.node).sub(d)
	}

	avg := ZP
	for _, n := range b.nodes {
		avg = avg.Add(n.g)
	}
	avg = avg.Div(float64(len(b.nodes)))
	for _, n := range b.nodes {
		n.sub(avg)
	}

	center := ZP
	for _, n := range b.nodes {
		center = center.Add(Pos(n))
	}
	center = center.Div(float64(len(b.nodes)))
	for _, n := range b.nodes {
		n.add(center.Sub(Pos(n)).Mul(nodeCenterCoef))
		l := n.g.Len()
		if l > 0 {
			n.g = n.g.Mul(math.Tanh(l/gMax) * gMax / l)
		}
	}

	for _, n := range b.nodes {
		d := Pt(rprop(&n.g_.X, &n.g.X, &n.step.X), rprop(&n.g_.Y, &n.g.Y, &n.step.Y))
		n.Move(Pos(n).Add(d))
		if d.Len() > .01 {
			done = false
		}
	}

	rect := ZR
	for _, n := range b.nodes {
		r := RectInParent(n)
		if n, ok := n.node.(*portsNode); ok {
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
	// for _, c := range b.conns {
	// 	rect = rect.Union(RectInParent(c))
	// }
	if rect == ZR {
		rect = Rectangle{ZP, Pt(16, 0)}
	}
	rect = rect.Inset(-blockRadius)
	if Rect(b).Size().Sub(rect.Size()).Len() > .01 {
		done = false
	}
	b.SetRect(rect)

	// if done {
	// 	cross := 0
	// 	b.forConns(func(src1, dst1, src2, dst2 *nodeArrange, srcP1, dstP1, srcP2, dstP2 Point) {
	// 		cross++
	// 	})
	// 	if cross < b.cross {
	// 		b.cross = cross
	// 		return true
	// 	} else {
	// 		b.forConns(func(src1, dst1, src2, dst2 *nodeArrange, srcP1, dstP1, srcP2, dstP2 Point) {
	// 			d1 := dstP1.Sub(srcP1)
	// 			d2 := dstP2.Sub(srcP2)
	// 			if d1.X < math.Abs(d1.Y / 2) || d2.X < math.Abs(d2.Y / 2) {
	// 				// continue
	// 			}
	// 			if rand.Float64() < .5 {
	// 				Ɵ1 := d1.Angle()
	// 				Ɵ2 := d2.Angle()
	// 				Ɵ := (Ɵ1 + Ɵ2) / 2
	// 				d := Pt(math.Cos(Ɵ), math.Sin(Ɵ))
	// 				d1 = d.Mul(d1.Len()).Sub(d1)
	// 				d2 = d.Mul(d2.Len()).Sub(d2)
	// 				if dst1 == dst2 || (src1 != src2 && rand.Float64() < .5) {
	// 					src1.Move(Pos(src1).Sub(d1))
	// 					src2.Move(Pos(src2).Sub(d2))
	// 					// p1 := CenterInParent(src1)
	// 					// p2 := CenterInParent(src2)
	// 					// MoveCenter(src1, Pt(p1.X, p2.Y))
	// 					// MoveCenter(src2, Pt(p2.X, p1.Y))
	// 				} else {
	// 					dst1.Move(Pos(dst1).Add(d1))
	// 					dst2.Move(Pos(dst2).Add(d2))
	// 					// p1 := CenterInParent(dst1)
	// 					// p2 := CenterInParent(dst2)
	// 					// MoveCenter(dst1, Pt(p1.X, p2.Y))
	// 					// MoveCenter(dst2, Pt(p2.X, p1.Y))
	// 				}
	// 			}
	// 		})
	// 	}
	// }

	return done
}

func (b *blockArrange) forConns(f func(src1, dst1, src2, dst2 *nodeArrange, srcP1, dstP1, srcP2, dstP2 Point)) {
	for i, c1 := range b.conns {
		if c1.src == nil || c1.dst == nil {
			continue
		}
		for _, c2 := range b.conns[i+1:] {
			if c2.src == nil || c2.dst == nil || c1.src == c2.src || c1.dst == c2.dst {
				continue
			}
			src1 := b.find(c1.src.node)
			dst1 := b.find(c1.dst.node)
			src2 := b.find(c2.src.node)
			dst2 := b.find(c2.dst.node)
			if src1 == src2 && dst1 == dst2 {
				continue
			}
			srcP1, dstP1 := c1.src.centerIn(b), c1.dst.centerIn(b)
			srcP2, dstP2 := c2.src.centerIn(b), c2.dst.centerIn(b)
			p1, p2 := LineToLine(srcP1, dstP1, srcP2, dstP2)
			if p1.Sub(p2).Len() == 0 {
				f(src1, dst1, src2, dst2, srcP1, dstP1, srcP2, dstP2)
			}
		}
	}
}

func rprop(g_, g, step *float64) float64 {
	prod := *g * *g_
	if prod > 0 {
		*step *= 1.2
	} else if prod < 0 {
		*step *= .5
		// *g = 0 //seems unnecessary
	}
	// *step = math.Max(.001, *step)
	*step = math.Min(1, *step)
	*g_ = *g
	return *g * *step
}

func (b *blockArrange) animate() bool {
	done := true
	b.block.Move(Pos(b))
	b.block.SetRect(Rect(b))
	for _, n := range b.nodes {
		if !n.animate() {
			done = false
		}
	}
	return done
}
