package main

import (
	."github.com/jteeuwen/glfw"
	."code.google.com/p/gordon-go/gui"
	."code.google.com/p/gordon-go/util"
)

type StringLiteralCreator struct {
	Text
	function *Function
	created *Signal
	canceled *Signal
}

func NewStringLiteralCreator(function *Function) *StringLiteralCreator {
	c := &StringLiteralCreator{}
	c.Text = *NewTextBase(c, "")
	c.function = function
	c.created = NewSignal()
	c.canceled = NewSignal()
	function.AddChild(c)
	c.SetFrameSize(3)
	c.SetFrameColor(Color{1, 1, 1, 1})
	c.SetBackgroundColor(Color{.5, .5, .5, .35})
	c.TakeKeyboardFocus()
	return c
}

func (c *StringLiteralCreator) Cancel() {
	c.Close()
	c.canceled.Emit()
}

func (c *StringLiteralCreator) LostKeyboardFocus() { c.Cancel() }

func (c *StringLiteralCreator) KeyPressed(event KeyEvent) {
	switch event.Key {
	case KeyEnter:
		c.Close()
		c.created.Emit(NewStringLiteralNode(c.Text.GetText()))
	case KeyEsc:
		c.Cancel()
	default:
		c.Text.KeyPressed(event)
	}
}
