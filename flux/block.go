package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type Block struct {
	*ViewBase
	AggregateMouseHandler
	node Node
	nodes map[Node]bool
	connections map[*Connection]bool
	focused, editing bool
	editingNode Node
	points []Point
}

func NewBlock(node Node) *Block {
	b := &Block{}
	b.ViewBase = NewView(b)
	b.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(b), NewViewPanner(b)}
	b.node = node
	b.nodes = map[Node]bool{}
	b.connections = map[*Connection]bool{}
	b.reform()
	return b
}

func (b *Block) Outer() *Block {
	if b.node == nil { return nil }
	return b.node.Block()
}
func (b *Block) Outermost() *Block {
	if outer := b.Outer(); outer != nil { return outer.Outermost() }
	return b
}

func (b *Block) AddNode(node Node) {
	b.AddChild(node)
	b.nodes[node] = true
	b.reform()
}

func (b *Block) RemoveNode(node Node) {
	b.RemoveChild(node)
	delete(b.nodes, node)
	b.reform()
}

func (b *Block) NewConnection(pt Point) *Connection {
	conn := NewConnection(b, pt)
	b.AddChild(conn)
	conn.Lower()
	b.connections[conn] = true
	return conn
}

func (b *Block) DeleteConnection(connection *Connection) {
	connection.Disconnect()
	delete(b.connections, connection)
	b.RemoveChild(connection)
	b.reform()
}

func (b Block) AllNodes() (nodes []Node) {
	for n := range b.nodes {
		nodes = append(nodes, n)
		switch n := n.(type) {
		case *IfNode:
			nodes = append(nodes, append(n.falseBlock.AllNodes(), n.trueBlock.AllNodes()...)...)
		}
	}
	return nodes
}

func (b Block) AllConnections() (conns []*Connection) {
	for c := range b.connections {
		conns = append(conns, c)
	}
	for n := range b.nodes {
		switch n := n.(type) {
		case *IfNode:
			conns = append(conns, append(n.falseBlock.AllConnections(), n.trueBlock.AllConnections()...)...)
		}
	}
	return conns
}

func (b Block) HasOutputConnections() bool {
	for n := range b.nodes {
		for _, output := range n.Outputs() {
			for _, conn := range output.connections {
				if !b.nodes[conn.dst.node] {
					return true
				}
			}
		}
	}
	return false
}

func (b *Block) StartEditing() {
	b.TakeKeyboardFocus()
	b.editing = true
	b.reform()
}

func (b *Block) StopEditing() {
	b.editing = false
	b.editingNode = nil
	b.reform()
}

func (b *Block) reform() {
	// b.points = []Point{}
	rect := ZR
	for n := range b.nodes {
		r := n.MapRectToParent(n.Rect())
		// b.points = append(b.points, p)
		if rect == ZR {
			rect = r
		} else {
			rect = rect.Union(r)
		}
	}
	if rect == ZR { rect = Rect(0, 0, 16, 16) }
	// if b.editingNode != nil && !b.nodes[b.editingNode] {
	// 	p := b.editingNode.MapTo(b.editingNode.Center(), outer)
	// 	b.points = append(b.points, p)
	// 	rect = rect.Union(ZR.Add(p))
	// }
	// if b.editing && b.editingNode == nil && len(b.nodes) == 0 {
	// 	b.points = append(b.points, p.Add(Pt(-4, 32)), p.Add(Pt(4, 32)))
	// }
	
	if b.node == nil { b.Move(b.MapToParent(rect.Min)) }
	b.Pan(rect.Min)
	b.Resize(rect.Dx(), rect.Dy())
	if n, ok := b.node.(interface{positionBlocks()}); ok { n.positionBlocks() }
	
	if b.node == nil {
		b.Repaint()	
	} else {
		b.Outer().reform()
	}
}

func (b *Block) GetNearestView(views []View, point Point, directionKey int) (nearest View) {
	dir := map[int]Point{KeyLeft:{-1, 0}, KeyRight:{1, 0}, KeyUp:{0, 1}, KeyDown:{0, -1}}[directionKey]
	bestScore := 0.0
	for _, view := range views {
		d := view.MapTo(view.Center(), b).Sub(point)
		score := (dir.X * d.X + dir.Y * d.Y) / (d.X * d.X + d.Y * d.Y);
		if (score > bestScore) {
			bestScore = score
			nearest = view
		}
	}
	return
}

func (b *Block) FocusNearestView(v View, directionKey int) {
	views := []View{}
	for _, node := range b.AllNodes() {
		views = append(views, node)
		for _, p := range node.Inputs() { views = append(views, p) }
		for _, p := range node.Outputs() { views = append(views, p) }
	}
	for _, connection := range b.AllConnections() {
		views = append(views, connection)
	}
	nearest := b.GetNearestView(views, v.MapTo(v.Center(), b), directionKey)
	if nearest != nil { nearest.TakeKeyboardFocus() }
}

func (b *Block) TookKeyboardFocus() { b.focused = true; b.Repaint() }
func (b *Block) LostKeyboardFocus() { b.focused = false; b.StopEditing(); b.Repaint() }

func (b *Block) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft, KeyRight, KeyUp, KeyDown:
		outermost := b.Outermost()
		if b.editing {
			var v View = b.editingNode
			if v == nil { v = b }
			views := []View{}; for _, n := range outermost.AllNodes() { views = append(views, n) }
			if n := outermost.GetNearestView(views, v.MapTo(v.Center(), outermost), event.Key); n != nil { b.editingNode = n.(Node) }
			b.reform()
		} else {
			outermost.FocusNearestView(b, event.Key)
		}
	case KeySpace:
		if b.editingNode != nil {
			if b.nodes[b.editingNode] {
				b.RemoveChild(b.editingNode)
				delete(b.nodes, b.editingNode)
				b.AddNode(b.editingNode)
			} else {
				b.RemoveNode(b.editingNode)
				b.nodes[b.editingNode] = true
				b.AddChild(b.editingNode)
			}
			b.reform()
		}
	case KeyEnter:
		if b.editing {
			if b.editingNode != nil && !b.nodes[b.editingNode] {
				b.nodes[b.editingNode] = true
			}
			b.StopEditing()
		} else {
			b.StartEditing()
		}
	case KeyEsc:
		if b.editing {
			if b.editingNode != nil {
				b.editingNode = nil
				b.reform()
			} else {
				b.StopEditing()
			}
		} else {
			b.node.Block().TakeKeyboardFocus()
		}
	default:
		switch event.Text {
		default:
			creator := NewNodeCreator(b)
			creator.Move(b.Center())
			creator.created.Connect(func(n ...interface{}) {
				node := n[0].(Node)
				b.AddNode(node)
				node.MoveCenter(b.Center())
				node.TakeKeyboardFocus()
			})
			creator.canceled.Connect(func(...interface{}) { b.TakeKeyboardFocus() })
			creator.text.KeyPressed(event)
		case "\"":
			node := NewStringConstantNode(b)
			b.AddNode(node)
			node.MoveCenter(b.Center())
			node.text.TakeKeyboardFocus()
		case ",":
			node := NewIfNode(b)
			b.AddNode(node)
			node.MoveCenter(b.Center())
			node.TakeKeyboardFocus()
		case "":
			b.ViewBase.KeyPressed(event)
		}
	}
}

// func (b *Block) MousePressed(button int, pt Point) {
// 	b.TakeKeyboardFocus()
// 	// conn := p.node.Block().NewConnection(p.MapTo(pt, p.node.Block()))
// 	// p.spec.ConnectTo(conn)
// 	// p.spec.PassMouseFocusToFreeConnectionHandle(conn, button)
// 	// conn.StartEditing()
// }
// func (b Block) MouseDragged(button int, pt Point) {}
// func (b Block) MouseReleased(button int, pt Point) {}
// 
func (b Block) Paint() {
	if b.editing {
		SetColor(Color{1, .5, 0, .7})
	} else if b.focused {
		SetColor(Color{.4, .4, 1, .7})
	} else {
		SetColor(Color{1, 1, 1, .5})
	}
	// DrawPolygon(b.points...)
	DrawRect(b.Rect())
}
