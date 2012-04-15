package flux

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type Compound struct {
	ViewBase
	AggregateMouseHandler
	components []*Component
	// channels []*Channel
}

func NewCompound() *Compound {
	c := &Compound{}
	c.ViewBase = *NewView(c)
	c.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(c), NewViewPanner(c)}
	return c
}

func (c *Compound) NewComponent() *Component {
	component := NewComponent(c)
	c.AddChild(component)
	c.components = append(c.components, component)
	return component
}

// func (c *Compound) NewChannel(pt image.Point) *Channel {
// 	ch := NewChannel(c, pt)
// 	c.AddChild(ch)
// 	ch.Lower()
// 	c.channels = append(c.channels, ch)
// 	return ch
// }
// 
// func (c *Compound) DeleteChannel(channel *Channel) {
// 	for i, ch := range c.channels {
// 		if ch == channel {
// 			c.channels = append(c.channels[:i], c.channels[i+1:]...)
// 			c.RemoveChild(channel)
// 			channel.Disconnect()
// 			return
// 		}
// 	}
// }
// 
// func (c *Compound) GetNearestView(views []View, point image.Point, directionKey int) (nearest View) {
// 	dir := map[int]image.Point{Key_Left:{-1, 0}, Key_Right:{1, 0}, Key_Up:{0, -1}, Key_Down:{0, 1}}[directionKey]
// 	bestScore := 0.0
// 	for _, view := range views {
// 		d := c.GetViewCenter(view).Sub(point)
// 		score := float64(dir.X * d.X + dir.Y * d.Y) / float64(d.X * d.X + d.Y * d.Y);
// 		if (score > bestScore) {
// 			bestScore = score
// 			nearest = view
// 		}
// 	}
// 	return
// }
// 
// func (c *Compound) FocusNearestView(v View, directionKey int) {
// 	views := make([]View, 0)
// 	for _, component := range c.components {
// 		views = append(views, component)
// 		views = append(views, component.GetPorts()...)
// 	}
// 	for _, channel := range c.channels {
// 		views = append(views, channel.srcHandle)
// 		views = append(views, channel.dstHandle)
// 	}
// 	nearest := c.GetNearestView(views, c.GetViewCenter(v), directionKey)
// 	if nearest != nil { nearest.TakeKeyboardFocus() }
// }
// 
// func (c *Compound) GetViewCenter(v View) image.Point {
// 	center := v.Center()
// 	for v != c && v != nil {
// 		center = v.MapToParent(center);
// 		v = v.Parent()
// 	}
// 	return center
// }

func (c *Compound) KeyPressed(key int) {
	switch key {
	// case Key_Left, Key_Right, Key_Up, Key_Down:
	// 	c.FocusNearestView(c, key)
	case glfw.KeyEnter:
		c.NewComponent()
		// creator := NewComponentCreator(c)
		// c.AddChild(creator)
		// creator.MoveCenter(c.Center())
		// creator.TakeKeyboardFocus()
	}
}

func (c Compound) Paint() {}
