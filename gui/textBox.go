package gui

import (
	."github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/ftgl"
	."code.google.com/p/gordon-go/util"
)

const (
	fontSize = 18
	xMargin = fontSize / 8
)

var font Font = CreateTextureFont("/usr/X11R6/lib/X11/fonts/TTF/luxirr.ttf")
func init() {
	font.SetFaceSize(fontSize, 1)
}

type TextBox struct {
	ViewBase
	AggregateMouseHandler
	text string
	
	TextChanged *Signal
}

func NewTextBox() *TextBox {
	t := &TextBox{}
	t.ViewBase = *NewView(t)
	t.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(t)}
	t.TextChanged = NewSignal()
	
	t.SetText("")
	return t
}

func (t *TextBox) SetText(text string) {
	t.text = text
	t.Resize(2*xMargin + int(font.Advance(text)), int(font.LineHeight()))
	t.TextChanged.Emit(text)
}

func (t *TextBox) KeyPressed(event KeyEvent) {
	if len(event.Text) > 0 {
		t.SetText(t.text + event.Text)
	}
	switch event.Key {
	case KeyBackspace:
		if len(t.text) > 0 {
			t.SetText(t.text[:len(t.text) - 1])
		}
	}
}

func (t TextBox) Paint() {
	w, h := gl.Double(t.Width()), gl.Double(t.Height())
	gl.Color4d(1, 1, 1, 1)
	gl.Rectd(0, 0, w, h)
	gl.Color4d(0, 0, 0, 1)
	gl.Begin(gl.LINE_LOOP)
	gl.Vertex2d(0, 0)
	gl.Vertex2d(w, 0)
	gl.Vertex2d(w, h)
	gl.Vertex2d(0, h)
	gl.End()
	
	gl.Translated(xMargin, -gl.Double(font.Descender()), 0)
	font.Render(t.text)
}
