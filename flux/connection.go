package main

import (
	"github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type Connection struct {
	ViewBase
	function *Function
	src *Output
	dst *Input
	
	focused bool
	srcPt Point
	dstPt Point
}

const connectionThickness = 7

func NewConnection(function *Function, pt Point) *Connection {
	c := &Connection{}
	c.ViewBase = *NewView(c)
	c.function = function
	c.srcPt = pt
	c.dstPt = pt
	return c
}

func (c Connection) Connected() bool { return c.src != nil && c.dst != nil }
func (c *Connection) Disconnect() { c.SetSource(nil); c.SetDestination(nil) }

func (c *Connection) SetSource(src *Output) {
	if c.src != nil { c.src.DisconnectConnection(c) }
	c.src = src
	if src != nil { src.ConnectConnection(c) }
	c.reform()
}
func (c *Connection) DisconnectSource(point Point) {
	c.srcPt = point
	c.SetSource(nil)
}

func (c *Connection) SetDestination(dst *Input) {
	if c.dst != nil { c.dst.DisconnectConnection(c) }
	c.dst = dst
	if dst != nil { dst.ConnectConnection(c) }
	c.reform()
}
func (c *Connection) DisconnectDestination(point Point) {
	c.dstPt = point
	c.SetDestination(nil)
}

func (c *Connection) reform() {
	if c.src != nil { c.srcPt = c.function.GetViewCenter(c.src) }
	if c.dst != nil { c.dstPt = c.function.GetViewCenter(c.dst) }
	rect := Rect(c.srcPt.X, c.srcPt.Y, c.dstPt.X, c.dstPt.Y).Canon().Inset(-connectionThickness / 2)
	c.Move(rect.Min)
	c.Resize(rect.Dx(), rect.Dy())
	
	c.Repaint()
}

func (c *Connection) BeStraightLine() {
	if c.src != nil && c.dst == nil {
		c.dstPt = c.srcPt.Add(Pt(64, 0))
	} else if c.src == nil && c.dst != nil {
		c.srcPt = c.dstPt.Sub(Pt(64, 0))
	}
	c.reform()
}

// func (c *Connection) StartEditing() {
// 	if c.src == nil {
// 		c.srcHandle.StartEditing()
// 	} else {
// 		c.dstHandle.StartEditing()
// 	}
// }

func (c *Connection) TookKeyboardFocus() { c.focused = true; c.Repaint() }
func (c *Connection) LostKeyboardFocus() { c.focused = false; c.Repaint() }

func (c *Connection) KeyPressed(event KeyEvent) {
	switch event.Key {
	case glfw.KeyLeft, glfw.KeyRight, glfw.KeyUp, glfw.KeyDown:
		c.function.FocusNearestView(c, event.Key)
	case glfw.KeyEsc:
		c.function.TakeKeyboardFocus()
	default:
		c.ViewBase.KeyPressed(event)
	}
}

func (c Connection) Paint() {
	// edgeColor := map[bool]NRGBAColor{false:{255, 255, 255, 15}, true:{0, 0, 255, 15}}
	// painter.SetStrokeColor(edgeColor[c.focused])
	// src := c.MapPointFromParent(c.srcPt)
	// dst := c.MapPointFromParent(c.dstPt)
	// for width := float64(connectionThickness); width > 1; width /= 1.414 {
	// 	painter.SetLineWidth(width)
	// 	painter.MoveTo(float64(src.X), float64(src.Y))
	// 	painter.LineTo(float64(dst.X), float64(dst.Y))
	// 	painter.Stroke()
	// }
}
