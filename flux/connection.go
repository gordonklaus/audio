package main

import (
	."code.google.com/p/gordon-go/gui"
	."math"
)

type connection struct {
	*ViewBase
	AggregateMouseHandler
	blk *block
	src *output
	dst *input
	feedback bool
	
	srcHandle *connectionSourceHandle
	dstHandle *connectionDestinationHandle
	
	focused bool
	srcPt Point
	dstPt Point
}

const connectionThickness = 7

func newConnection() *connection {
	c := &connection{}
	c.ViewBase = NewView(c)
	c.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(c)}
	c.srcHandle = newConnectionSourceHandle(c)
	c.dstHandle = newConnectionDestinationHandle(c)
	c.AddChild(c.srcHandle)
	c.AddChild(c.dstHandle)
	return c
}

func (c connection) connected() bool { return c.src != nil && c.dst != nil }
func (c *connection) disconnect() { c.setSrc(nil); c.setDst(nil) }

func (c *connection) setSrc(src *output) {
	if c.src != nil { c.src.disconnect(c) }
	c.src = src
	if src != nil { src.connect(c) }
	if c.dst != nil { c.dst.connsChanged() }
	c.reblock()
	c.reform()
}

func (c *connection) setDst(dst *input) {
	if c.dst != nil { c.dst.disconnect(c) }
	c.dst = dst
	if dst != nil { dst.connect(c) }
	if c.src != nil { c.src.connsChanged() }
	c.reblock()
	c.reform()
}

func (c *connection) reblock() {
	var newblk *block
	if c.src == nil && c.dst == nil {
		return
	} else if c.src == nil {
		newblk = c.dst.node.block()
	} else if c.dst == nil {
		newblk = c.src.node.block()
	} else {
loop:	for srcblk := c.src.node.block(); srcblk != nil; srcblk = srcblk.outer() {
			for dstblk := c.dst.node.block(); dstblk != nil; dstblk = dstblk.outer() {
				if srcblk == dstblk {
					newblk = srcblk
					break loop
				}
			}
		}
	}
	newblk.addConnection(c)
}

func (c *connection) reform() {
	unconnectedOffset := Pt(48, 0)
	if c.feedback { unconnectedOffset.X = -208 }
	if c.src != nil {
		c.srcPt = c.src.MapTo(c.src.Center(), c.blk)
	} else {
		c.srcPt = c.dstPt.Sub(unconnectedOffset)
	}
	if c.dst != nil {
		c.dstPt = c.dst.MapTo(c.dst.Center(), c.blk)
	} else {
		c.dstPt = c.srcPt.Add(unconnectedOffset)
	}
	
	rect := Rectangle{c.srcPt, c.dstPt}.Canon().Inset(-connectionThickness / 2)
	c.Move(rect.Min)
	c.Resize(rect.Dx(), rect.Dy())
	
	handleOffset := c.dstPt.Sub(c.srcPt).Div(4)
	if c.srcHandle.editing {
		c.srcHandle.MoveCenter(c.MapFromParent(c.srcPt))
	} else {
		c.srcHandle.MoveCenter(c.MapFromParent(c.srcPt.Add(handleOffset)))
	}
	if c.dstHandle.editing {
		c.dstHandle.MoveCenter(c.MapFromParent(c.dstPt))
	} else {
		c.dstHandle.MoveCenter(c.MapFromParent(c.dstPt.Sub(handleOffset)))
	}
}

func (c *connection) startEditing() {
	if c.src == nil {
		c.srcHandle.startEditing()
	} else {
		c.dstHandle.startEditing()
	}
}

func (c *connection) TookKeyboardFocus() { c.focused = true; c.Repaint() }
func (c *connection) LostKeyboardFocus() { c.focused = false; c.Repaint() }

func (c *connection) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft:
		c.src.TakeKeyboardFocus()
	case KeyRight:
		c.dst.TakeKeyboardFocus()
	case KeyEscape:
		c.Parent().TakeKeyboardFocus()
	default:
		if event.Text == "\\" {
			c.feedback = !c.feedback
			c.reform()
		} else {
			c.ViewBase.KeyPressed(event)
		}
	}
}

func (c connection) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[c.focused])
	start, end := c.MapFromParent(c.srcPt), c.MapFromParent(c.dstPt)
	d := end.Sub(start)
	mid := start.Add(d.Div(2))
	if c.feedback { mid.Y = Max(start.Y, end.Y) + 128 }
	dx := Abs(d.X / 3)
	p1 := start.Add(Pt(dx, 0))
	p2 := mid
	p3 := end.Sub(Pt(dx, 0))
	pts := []Point{start, p1, p2, p3, end}
	len := 0.0
	for i := range pts {
		if i > 0 {
			len += pts[i].Sub(pts[i-1]).Len()
		}
	}
	DrawBezier(pts, int(len))
}
