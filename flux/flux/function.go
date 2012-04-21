package main

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	"code.google.com/p/gordon-go/flux"
)

type Function struct {
	ViewBase
	AggregateMouseHandler
	nodes []*Node
	
	function flux.Function
	// channels []*Channel
}

func NewFunction() *Function {
	f := &Function{}
	f.ViewBase = *NewView(f)
	f.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(f), NewViewPanner(f)}
	return f
}

func (f *Function) NewNode() *Node {
	node := NewNode(f)
	f.nodes = append(f.nodes, node)
	return node
}

// func (f *Function) NewChannel(pt image.Point) *Channel {
// 	ch := NewChannel(f, pt)
// 	f.AddChild(ch)
// 	ch.Lower()
// 	f.channels = append(f.channels, ch)
// 	return ch
// }
// 
// func (f *Function) DeleteChannel(channel *Channel) {
// 	for i, ch := range f.channels {
// 		if ch == channel {
// 			f.channels = append(f.channels[:i], f.channels[i+1:]...)
// 			f.RemoveChild(channel)
// 			channel.Disconnect()
// 			return
// 		}
// 	}
// }
// 
// func (f *Function) GetNearestView(views []View, point image.Point, directionKey int) (nearest View) {
// 	dir := map[int]image.Point{Key_Left:{-1, 0}, Key_Right:{1, 0}, Key_Up:{0, -1}, Key_Down:{0, 1}}[directionKey]
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
// 		views = append(views, node.GetPorts()...)
// 	}
// 	for _, channel := range f.channels {
// 		views = append(views, channel.srcHandle)
// 		views = append(views, channel.dstHandle)
// 	}
// 	nearest := f.GetNearestView(views, f.GetViewCenter(v), directionKey)
// 	if nearest != nil { nearest.TakeKeyboardFocus() }
// }
// 
// func (f *Function) GetViewCenter(v View) image.Point {
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
		creator.MoveCenter(f.Center())
	}
}

func (Function) Paint() {}
