package main

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type Function struct {
	ViewBase
	AggregateMouseHandler
	nodes []*Node
	
	connections []*Connection
}

func NewFunction() *Function {
	f := &Function{}
	f.ViewBase = *NewView(f)
	f.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(f), NewViewPanner(f)}
	return f
}

func (f *Function) AddNode(node *Node) {
	node.function = f
	f.AddChild(node)
	f.nodes = append(f.nodes, node)
}

func (f *Function) NewConnection(pt Point) *Connection {
	conn := NewConnection(f, pt)
	f.AddChild(conn)
	conn.Lower()
	f.connections = append(f.connections, conn)
	return conn
}

func (f *Function) DeleteConnection(connection *Connection) {
	for i, conn := range f.connections {
		if conn == connection {
			f.connections = append(f.connections[:i], f.connections[i+1:]...)
			f.RemoveChild(connection)
			connection.Disconnect()
			return
		}
	}
}

func (f *Function) GetNearestView(views []View, point Point, directionKey int) (nearest View) {
	dir := map[int]Point{glfw.KeyLeft:{-1, 0}, glfw.KeyRight:{1, 0}, glfw.KeyUp:{0, 1}, glfw.KeyDown:{0, -1}}[directionKey]
	bestScore := 0.0
	for _, view := range views {
		d := f.GetViewCenter(view).Sub(point)
		score := (dir.X * d.X + dir.Y * d.Y) / (d.X * d.X + d.Y * d.Y);
		if (score > bestScore) {
			bestScore = score
			nearest = view
		}
	}
	return
}

func (f *Function) FocusNearestView(v View, directionKey int) {
	views := []View{}
	for _, node := range f.nodes {
		views = append(views, node)
		views = append(views, node.Getputs()...)
	}
	for _, connection := range f.connections {
		views = append(views, connection)
	}
	nearest := f.GetNearestView(views, f.GetViewCenter(v), directionKey)
	if nearest != nil { nearest.TakeKeyboardFocus() }
}

func (f *Function) GetViewCenter(v View) Point {
	center := v.Center()
	for v != f && v != nil {
		center = v.MapToParent(center);
		v = v.Parent()
	}
	return center
}

func (f *Function) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		f.FocusNearestView(f, event.Key)
	}
	if len(event.Text) > 0 {
		if event.Text == "\"" {
			creator := NewStringLiteralCreator(f)
			creator.Move(f.Center())
			creator.created.Connect(func(n ...interface{}) {
				node := n[0].(*Node)
				f.AddNode(node)
				node.MoveCenter(f.Center())
				node.TakeKeyboardFocus()
			})
			creator.canceled.Connect(func(...interface{}) { f.TakeKeyboardFocus() })
		} else {
			creator := NewNodeCreator(f)
			creator.Move(f.Center())
			creator.created.Connect(func(n ...interface{}) {
				node := n[0].(*Node)
				f.AddNode(node)
				node.MoveCenter(f.Center())
				node.TakeKeyboardFocus()
			})
			creator.canceled.Connect(func(...interface{}) { f.TakeKeyboardFocus() })
			creator.text.KeyPressed(event)
		}
	}
}

func (Function) Paint() {}
