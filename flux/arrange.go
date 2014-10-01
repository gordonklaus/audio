// Copyright 2014 Gordon Klaus. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	. "code.google.com/p/gordon-go/flux/gui"
	"math"
	"math/rand"
	"time"
)

const fps = 60.0

func animate(animate blockchan, stop stopchan) {
	b := <-animate
	n := b.block.node
	converged := make(chan bool, 1)
	for {
		next := time.After(time.Second / fps)
		select {
		case DoChan(n) <- func() {
			c := CenterInParent(n)
			converged <- b.animate()
			ResizeToFit(n, 0)
			MoveCenter(n, c)
		}:
		case <-stop:
			return
		}
		if <-converged {
			next = nil
		}

		select {
		case <-next:
		case b = <-animate:
		case <-stop:
			return
		}
	}
}

func (b *blockArrange) animate() bool {
	converged := true
	for _, n := range b.nodes {
		if !n.animate() {
			converged = false
		}
	}
	b.setRectReal()
	return converged
}

func (n *nodeArrange) animate() bool {
	converged := true
	for _, b := range n.blocks {
		if !b.animate() {
			converged = false
		}
	}
	n.arrange()
	for _, b := range n.blocks {
		b.block.Move(Pos(b))
	}
	n.node.SetRect(Rect(n))
	for _, p := range n.ports {
		p.port.Move(Pos(p))
	}

	pos := Pos(n.node)
	d := Pos(n).Sub(pos)
	if d.Len() > .1 {
		n.vel = n.vel.Add(d.Mul(60 / fps)).Mul(.8)
		n.node.Move(pos.Add(n.vel.Div(fps)))
		return false
	}
	n.node.Move(Pos(n))
	return converged
}

func (n *nodeArrange) arrange() {
	switch node := n.node.(type) {
	case *ifNode:
		x := 0.0
		for i, b := range n.blocks {
			if i > 0 {
				x += portSize + (Width(n.blocks[i-1])+Width(b))/2
			}
			cond := n.ports[i+1]
			MoveCenter(cond, Pt(x, portSize))
			b.Move(Pt(x-Width(b)/2, -Height(b)-portSize))
		}
		seqOut := n.ports[len(n.ports)-1]
		MoveCenter(seqOut, Pt(0, -Height(n.blocks[0])-portSize))
		ResizeToFit(n, 0)
	case *loopNode:
		b := n.blocks[0]
		b.Move(Pt(-Width(b)/2, -Height(b)-portSize))
		seqOut := n.ports[2]
		MoveCenter(seqOut, Pt(0, -Height(b)-portSize))
		ResizeToFit(n, 0)
	case *funcNode:
		b := n.blocks[0]
		b.Move(Pt(-Width(b)/2, portSize/2))
		ResizeToFit(n, 0)
	case *selectNode:
		seqIn := n.ports[0]
		seqOut := n.ports[len(n.ports)-1]
		if len(n.blocks) > 0 {
			x := 0.0
			i := 1
			for j, b := range n.blocks {
				if j > 0 {
					x += portSize + (Width(n.blocks[j-1])+Width(b))/2
				}
				if !n.defaultCase[b.block] {
					ch := n.ports[i]
					i++
					if n.sendCase[b.block] {
						elem := n.ports[i]
						i++
						MoveCenter(ch, Pt(x-portSize/2, portSize))
						MoveCenter(elem, Pt(x+portSize/2, portSize))
					} else {
						MoveCenter(ch, Pt(x, portSize))
					}
				}
				b.Move(Pt(x-Width(b)/2, -Height(b)-portSize))
			}
			MoveCenter(seqIn, Pt(0, -portSize))
			MoveCenter(seqOut, Pt(0, -Height(n.blocks[0])-portSize))
		} else {
			r := RectInParent(node.name)
			x := r.Center().X
			MoveCenter(seqIn, Pt(x, r.Max.Y))
			MoveCenter(seqOut, Pt(x, r.Min.Y))
		}
		ResizeToFit(n, 0)
		n.SetRect(Rect(n).Union(RectInParent(node.name)))
	}
}

// same as setRect but uses actual node positions and applies actual block rect
func (b *blockArrange) setRectReal() {
	rect := ZR
	for _, n := range b.nodes {
		r := RectInParent(n.node)
		if n, ok := n.node.(*portsNode); ok {
			// portsNodes gravitate to the boundary; thus we must adjust for the margin
			if n.out {
				r = r.Add(Pt(0, blockRadius-2))
			} else {
				r = r.Sub(Pt(0, blockRadius-2))
			}
		}
		if rect == ZR {
			rect = r
		} else {
			rect = rect.Union(r)
		}
	}
	rect = rect.Inset(-blockRadius)
	b.SetRect(rect)
	b.block.SetRect(rect)
}

func rearrange(b *block) {
	if b == nil {
		return
	}
	ba := newBlockArrange(b, nil, portmap{})
	go func() {
		select {
		case b.arrange <- ba:
		case <-b.stop:
		}
	}()
}

// TODO: treat conns as curves, not lines (or make them more linear?)
func arrange(arrange, childArranged, arranged blockchan, stop stopchan) {
	b := <-arrange
	minCross := math.MaxInt32
	stagnant := 0
	for iter := 0; ; iter++ {
		const (
			portsNodeCoef   = 1024
			nodeCenterCoef  = .5
			nodeSep         = 8
			nodeSepCoef     = 1024
			nodeConnSep     = 8
			nodeConnSepCoef = 8
			connLen         = 16
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
					p.Y = Rect(b).Min.Y + 2
				} else {
					p.Y = Rect(b).Max.Y - 2
				}
				dir := p.Sub(MapToParent(ZP, n1))
				if n1.hasConns {
					dir.X = 0
				}
				n1.add(dir.Mul(portsNodeCoef * dir.Len()))
			}
			for _, n2 := range b.nodes {
				if n2 == n1 {
					continue
				}
				dir := CenterInParent(n1).Sub(CenterInParent(n2))
				r1 := RectInParent(n1).Inset(-nodeSep / 2)
				r2 := RectInParent(n2).Inset(-nodeSep / 2)
				sep := math.Min(r1.Intersect(r2).Size().XY())
				n1.add(dir.Mul(nodeSepCoef * math.Min(1, sep/nodeSep) / dir.Len()))
			}
			for _, c := range b.conns {
				if c.hidden {
					continue
				}
				srcNode := c.src.node.in(b)
				dstNode := c.dst.node.in(b)
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
		// TODO: feedback conns are in an outer block, so arranging them has no effect.  do something
		for _, c := range b.conns {
			// TODO: connections that cross block boundaries should be longer.  map src and dst points to the boundaries
			srcNode := c.src.node.in(b)
			dstNode := c.dst.node.in(b)
			d := c.dst.centerIn(b).Sub(c.src.centerIn(b))
			if c.conn.feedback {
				d.Y = -d.Y
				d.Y -= connLen + Height(srcNode) + Height(dstNode)
			} else {
				d.Y += connLen
			}
			if c.hidden {
				d.X = 0
			} else {
				d.Y += math.Abs(d.X) / 2
				d.X *= math.Abs(d.X) / 16
			}
			if d.Y > 0 {
				d.Y *= d.Y * d.Y
			}
			if c.conn.feedback {
				d.Y = -d.Y
			}
			d = d.Mul(connLenCoef)
			srcNode.add(d)
			dstNode.sub(d)
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

		converged := true
		for _, n := range b.nodes {
			dx := rprop(&n.g_.X, &n.g.X, &n.step.X)
			dy := rprop(&n.g_.Y, &n.g.Y, &n.step.Y)
			d := Pt(dx, dy)
			n.Move(Pos(n).Add(d))
			if d.Len() > .01 {
				converged = false
			}
		}
		b.setRect()

		always := make(chan bool, 1)
		if converged || iter > 5000 {
			iter = 0
			cross := 0
			b.forCross(func(src1, dst1, src2, dst2 *nodeArrange, srcP1, dstP1, srcP2, dstP2 Point) {
				cross++
			})
			if cross < minCross {
				minCross = cross
				stagnant = 0
				select {
				case arranged <- b.copy(nil, portmap{}):
				case <-stop:
					return
				}
			} else {
				stagnant++
			}
			if stagnant < 100 {
				b.uncross()
			} else {
				stagnant = 0
				always = nil
			}
		}

		select {
		case newb := <-arrange:
			b = newb.replaceAll(b)
			minCross = math.MaxInt32
		case newChild := <-childArranged:
			child := b.find(newChild.block)
			for child == nil {
				b = (<-arrange).replaceAll(b)
				child = b.find(newChild.block)
			}
			child.replace(newChild)
			minCross = math.MaxInt32
		case <-stop:
			return
		case always <- true:
		}
	}
}

func (n *nodeArrange) add(d Point) { n.g = n.g.Add(d) }
func (n *nodeArrange) sub(d Point) { n.g = n.g.Sub(d) }

func (n *nodeArrange) in(b *blockArrange) *nodeArrange {
	for n.block != b {
		n = n.block.node
	}
	return n
}

func (p *portArrange) centerIn(b *blockArrange) Point {
	return Map(Center(p), p, b)
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

// keep in sync with setRectReal
func (b *blockArrange) setRect() {
	rect := ZR
	for _, n := range b.nodes {
		r := RectInParent(n)
		if n, ok := n.node.(*portsNode); ok {
			// portsNodes gravitate to the boundary; thus we must adjust for the margin
			if n.out {
				r = r.Add(Pt(0, blockRadius-2))
			} else {
				r = r.Sub(Pt(0, blockRadius-2))
			}
		}
		if rect == ZR {
			rect = r
		} else {
			rect = rect.Union(r)
		}
	}
	rect = rect.Inset(-blockRadius)
	b.SetRect(rect)
}

func (b *blockArrange) replaceAll(b2 *blockArrange) *blockArrange {
	for _, n := range b.nodes {
		for _, b := range n.blocks {
			if b2 := b2.find(b.block); b2 != nil {
				b.replace(b2)
			}
		}
	}
	return b
}

func (b *blockArrange) find(block *block) *blockArrange {
	for _, n := range b.nodes {
		for _, b := range n.blocks {
			if b.block == block {
				return b
			}
		}
	}
	return nil
}

func (b *blockArrange) replace(b2 *blockArrange) {
	n := b.node
	ports := portmap{}
	b2 = b2.copy(n, ports) //must copy to avoid mutating b2 and interfering in replaceAll; incidentally, this is a good way to set node and get ports
	for _, c := range n.block.conns {
		if p, ok := ports[c.src.port]; ok {
			c.src = p
		}
		if p, ok := ports[c.dst.port]; ok {
			c.dst = p
		}
	}
	n.Remove(b)
	n.Add(b2)
	*b = *b2
	n.arrange()
}

func (b *blockArrange) forCross(f func(src1, dst1, src2, dst2 *nodeArrange, srcP1, dstP1, srcP2, dstP2 Point)) {
	for i, c1 := range b.conns {
		if c1.hidden {
			continue
		}
		for _, c2 := range b.conns[i+1:] {
			if c2.hidden || c1.src == c2.src || c1.dst == c2.dst {
				continue
			}
			src1 := c1.src.node.in(b)
			dst1 := c1.dst.node.in(b)
			src2 := c2.src.node.in(b)
			dst2 := c2.dst.node.in(b)
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

func (b *blockArrange) uncross() {
	b.forCross(func(src1, dst1, src2, dst2 *nodeArrange, srcP1, dstP1, srcP2, dstP2 Point) {
		d1 := dstP1.Sub(srcP1)
		d2 := dstP2.Sub(srcP2)
		d := d1.Add(d2)
		d = d.Div(d.Len())
		d1 = d.Mul(d1.Len()).Sub(d1)
		d2 = d.Mul(d2.Len()).Sub(d2)
		if dst1 == dst2 || (src1 != src2 && rand.Float64() < .5) {
			src1.Move(Pos(src1).Sub(d1))
			src2.Move(Pos(src2).Sub(d2))
		} else {
			dst1.Move(Pos(dst1).Add(d1))
			dst2.Move(Pos(dst2).Add(d2))
		}
	})
}

type blockArrange struct {
	*ViewBase
	block *block
	node  *nodeArrange
	nodes []*nodeArrange
	conns []*connArrange
}

type nodeArrange struct {
	*ViewBase
	node   node
	block  *blockArrange
	ports  []*portArrange
	blocks []*blockArrange

	hasConns              bool
	defaultCase, sendCase map[*block]bool

	g, g_, step Point

	vel Point
}

type portArrange struct {
	*ViewBase
	port *port
	node *nodeArrange
}

type connArrange struct {
	*ViewBase
	conn     *connection
	src, dst *portArrange
	hidden   bool
}

func newBlockArrange(block *block, n *nodeArrange, ports portmap) *blockArrange {
	b := &blockArrange{block: block, node: n}
	b.ViewBase = NewView(b)
	b.Move(Pos(block))
	b.SetRect(Rect(block))
	for n := range block.nodes {
		n := newNodeArrange(n, b, ports)
		b.nodes = append(b.nodes, n)
		b.Add(n)
	}
	for c := range block.conns {
		if c.src == nil || c.dst == nil {
			continue
		}
		c := newConnArrange(c, ports)
		b.conns = append(b.conns, c)
		b.Add(c)
	}
	return b
}

func newNodeArrange(node node, b *blockArrange, ports portmap) *nodeArrange {
	n := &nodeArrange{node: node, block: b}
	n.ViewBase = NewView(n)
	n.Move(Pos(node))
	n.SetRect(Rect(node))
	for _, port := range append(node.inputs(), node.outputs()...) {
		p := newPortArrange(port, n)
		n.ports = append(n.ports, p)
		n.Add(p)
		if len(port.conns) > 0 {
			n.hasConns = true
		}
		ports[port] = p
	}
	switch node := node.(type) {
	case *ifNode:
		for _, b := range node.blocks {
			b := newBlockArrange(b, n, ports)
			n.blocks = append(n.blocks, b)
			n.Add(b)
		}
	case *loopNode:
		b := newBlockArrange(node.loopblk, n, ports)
		n.blocks = []*blockArrange{b}
		n.Add(b)
	case *funcNode:
		b := newBlockArrange(node.funcblk, n, ports)
		n.blocks = []*blockArrange{b}
		n.Add(b)
	case *selectNode:
		n.defaultCase = map[*block]bool{}
		n.sendCase = map[*block]bool{}
		for _, c := range node.cases {
			b := newBlockArrange(c.blk, n, ports)
			n.blocks = append(n.blocks, b)
			n.Add(b)
			n.defaultCase[c.blk] = c.ch == nil
			n.sendCase[c.blk] = c.send
		}
	}
	n.step = Pt(1, 1)
	return n
}

func newPortArrange(port *port, n *nodeArrange) *portArrange {
	p := &portArrange{port: port, node: n}
	p.ViewBase = NewView(p)
	p.Move(Pos(port))
	p.SetRect(Rect(port))
	return p
}

func newConnArrange(conn *connection, ports portmap) *connArrange {
	c := &connArrange{conn: conn}
	c.ViewBase = NewView(c)
	c.src = ports[conn.src]
	c.dst = ports[conn.dst]
	c.hidden = conn.hidden
	return c
}

func (b *blockArrange) copy(n *nodeArrange, ports portmap) *blockArrange {
	b2 := &blockArrange{block: b.block, node: n}
	b2.ViewBase = NewView(b2)
	b2.Move(Pos(b))
	b2.SetRect(Rect(b))
	for _, n := range b.nodes {
		n := n.copy(b2, ports)
		b2.nodes = append(b2.nodes, n)
		b2.Add(n)
	}
	for _, c := range b.conns {
		c := c.copy(ports)
		if c == nil {
			continue
		}
		b2.conns = append(b2.conns, c)
		b2.Add(c)
	}
	return b2
}

func (n *nodeArrange) copy(b *blockArrange, ports portmap) *nodeArrange {
	n2 := &nodeArrange{node: n.node, block: b}
	n2.ViewBase = NewView(n2)
	n2.Move(Pos(n))
	n2.SetRect(Rect(n))
	for _, p := range n.ports {
		p := p.copy(n2)
		n2.ports = append(n2.ports, p)
		n2.Add(p)
		ports[p.port] = p
	}
	for _, b := range n.blocks {
		b := b.copy(n2, ports)
		n2.blocks = append(n2.blocks, b)
		n2.Add(b)
	}
	n2.hasConns = n.hasConns
	n2.defaultCase = n.defaultCase
	n2.sendCase = n.sendCase
	n2.step = Pt(1, 1)
	return n2
}

func (p *portArrange) copy(n *nodeArrange) *portArrange {
	p2 := &portArrange{port: p.port, node: n}
	p2.ViewBase = NewView(p2)
	p2.Move(Pos(p))
	p2.SetRect(Rect(p))
	return p2
}

func (c *connArrange) copy(ports portmap) *connArrange {
	c2 := &connArrange{conn: c.conn, hidden: c.hidden}
	c2.ViewBase = NewView(c2)
	c2.src = ports[c.src.port]
	c2.dst = ports[c.dst.port]
	if c2.src == nil || c2.dst == nil {
		return nil // an out-of-date inner block may lack the ports for a new connection; ignore
	}
	return c2
}

type portmap map[*port]*portArrange

type blockchan chan *blockArrange
type stopchan chan struct{}

func (c stopchan) stop() {
	c <- struct{}{}
}
