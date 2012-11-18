package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
)

type Connection struct {
	*ViewBase
	AggregateMouseHandler
	block *Block
	src *Output
	dst *Input
	
	srcHandle *ConnectionSourceHandle
	dstHandle *ConnectionDestinationHandle
	
	focused bool
	srcPt Point
	dstPt Point
}

const connectionThickness = 7

func NewConnection(block *Block, pt Point) *Connection {
	c := &Connection{}
	c.ViewBase = NewView(c)
	c.block = block
	c.srcHandle = NewConnectionSourceHandle(c)
	c.dstHandle = NewConnectionDestinationHandle(c)
	c.srcPt = pt
	c.dstPt = pt
	c.AddChild(c.srcHandle)
	c.AddChild(c.dstHandle)
	c.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(c)}
	return c
}

func (c Connection) Connected() bool { return c.src != nil && c.dst != nil }
func (c *Connection) Disconnect() { c.SetSource(nil); c.SetDestination(nil) }

func (c *Connection) SetSource(src *Output) {
	if c.src != nil { c.src.Disconnect(c) }
	c.src = src
	if src != nil { src.Connect(c) }
	c.reblock()
	c.reform()
}
func (c *Connection) DisconnectSource(point Point) {
	c.srcPt = point
	c.SetSource(nil)
}

func (c *Connection) SetDestination(dst *Input) {
	if c.dst != nil { c.dst.Disconnect(c) }
	c.dst = dst
	if dst != nil { dst.Connect(c) }
	c.reblock()
	c.reform()
}
func (c *Connection) DisconnectDestination(point Point) {
	c.dstPt = point
	c.SetDestination(nil)
}

func (c *Connection) reblock() {
	var newBlock *Block
	if c.src == nil && c.dst == nil {
		return
	} else if c.src == nil {
		newBlock = c.dst.node.Block()
	} else if c.dst == nil {
		newBlock = c.src.node.Block()
	} else {
loop:	for srcBlock := c.src.node.Block(); srcBlock != nil; srcBlock = srcBlock.Outer() {
			for dstBlock := c.dst.node.Block(); dstBlock != nil; dstBlock = dstBlock.Outer() {
				if srcBlock == dstBlock {
					newBlock = srcBlock
					break loop
				}
			}
		}
	}
	newBlock.AddConnection(c)
}

func (c *Connection) reform() {
	if c.src != nil { c.srcPt = c.src.MapTo(c.src.Center(), c.block) }
	if c.dst != nil { c.dstPt = c.dst.MapTo(c.dst.Center(), c.block) }
	rect := Rectangle{c.srcPt, c.dstPt}.Canon().Inset(-connectionThickness / 2)
	c.Move(rect.Min)
	c.Resize(rect.Dx(), rect.Dy())
	
	handleOffset := c.dstPt.Sub(c.srcPt).Div(4)
	if c.srcHandle.editing {
		c.srcHandle.MoveCenter(c.MapFrom(c.srcPt, c.block))
	} else {
		c.srcHandle.MoveCenter(c.MapFrom(c.srcPt.Add(handleOffset), c.block))
	}
	if c.dstHandle.editing {
		c.dstHandle.MoveCenter(c.MapFrom(c.dstPt, c.block))
	} else {
		c.dstHandle.MoveCenter(c.MapFrom(c.dstPt.Sub(handleOffset), c.block))
	}
	c.Repaint()
}

func (c *Connection) BeStraightLine() {
	if c.src != nil && c.dst == nil {
		c.dstPt = c.srcPt.Add(Pt(48, 0))
	} else if c.src == nil && c.dst != nil {
		c.srcPt = c.dstPt.Sub(Pt(48, 0))
	}
	c.reform()
}

func (c *Connection) StartEditing() {
	if c.src == nil {
		c.srcHandle.StartEditing()
	} else {
		c.dstHandle.StartEditing()
	}
}

func (c *Connection) TookKeyboardFocus() { c.focused = true; c.Repaint() }
func (c *Connection) LostKeyboardFocus() { c.focused = false; c.Repaint() }

func (c *Connection) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyLeft:
		c.src.TakeKeyboardFocus()
	case KeyRight:
		c.dst.TakeKeyboardFocus()
	case KeyUp, KeyDown:
		c.block.Outermost().FocusNearestView(c, event.Key)
	case KeyEsc:
		c.block.TakeKeyboardFocus()
	default:
		c.ViewBase.KeyPressed(event)
	}
}

func (c Connection) Paint() {
	SetColor(map[bool]Color{false:{.5, .5, .5, 1}, true:{.3, .3, .7, 1}}[c.focused])
	start, end := c.MapFrom(c.srcPt, c.block), c.MapFrom(c.dstPt, c.block)
	d := end.Sub(start)
	dx := d.X / 3
	DrawCubic([4]Point{start, start.Add(Pt(dx, 0)), end.Sub(Pt(dx, 0)), end}, int(d.Len() / 8))
}
