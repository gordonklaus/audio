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

func (f *Function) NewNode(node *Node) {
	node.Function = f
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

// func (f *Function) GetNearestView(views []View, point Point, directionKey int) (nearest View) {
// 	dir := map[int]Point{Key_Left:{-1, 0}, Key_Right:{1, 0}, Key_Up:{0, -1}, Key_Down:{0, 1}}[directionKey]
// 	bestScore := 0.0
// 	for _, view := range views {
// 		d := f.GetViewCenter(view).Sub(point)
// 		score := float64(dir.X * d.X + dir.Y * d.Y) / float64(d.X * d.X + d.Y * d.Y);
// 		if (score > bestScore) {
// 			bestScore = score
// 			nearest = view
// 		}
// 	}
// 	return
// }
// 
// func (f *Function) FocusNearestView(v View, directionKey int) {
// 	views := make([]View, 0)
// 	for _, node := range f.nodes {
// 		views = append(views, node)
// 		views = append(views, node.Getputs()...)
// 	}
// 	for _, connection := range f.connections {
// 		views = append(views, connection.srcHandle)
// 		views = append(views, connection.dstHandle)
// 	}
// 	nearest := f.GetNearestView(views, f.GetViewCenter(v), directionKey)
// 	if nearest != nil { nearest.TakeKeyboardFocus() }
// }
// 
// func (f *Function) GetViewCenter(v View) Point {
// 	center := v.Center()
// 	for v != f && v != nil {
// 		center = v.MapToParent(center);
// 		v = v.Parent()
// 	}
// 	return center
// }

func (f *Function) KeyPressed(event KeyEvent) {
	switch event.Key {
	// case Key_Left, Key_Right, Key_Up, Key_Down:
	// 	f.FocusNearestView(f, key)
	case glfw.KeyEnter:
		creator := NewNodeCreator(f)
		creator.Created.Connect(func(info ...interface{}) { f.NewNode(info[0].(*Node)) })
		creator.Move(f.Center())
	}
}

func (Function) Paint() {}
