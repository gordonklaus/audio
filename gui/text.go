package gui

import (
	."github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/ftgl"
	."code.google.com/p/gordon-go/util"
	"image"
)

type Color struct {
	Red float32
	Green float32
	Blue float32
	Alpha float32
}

var font Font = CreateTextureFont("/usr/X11R6/lib/X11/fonts/TTF/luxirr.ttf")
func init() {
	font.SetFaceSize(fontSize, 1)
}

const (
	fontSize = 18
)

type Text struct {
	ViewBase
	AggregateMouseHandler
	text string
	textColor Color
	backgroundColor Color
	editable bool
	validator func(*string)bool
	
	TextChanged *Signal
}

func NewText(text string) *Text { return NewTextBase(nil, text) }
func NewTextBase(self View, text string) *Text {
	t := &Text{}
	if self == nil { self = t }
	t.ViewBase = *NewView(self)
	t.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(t)}
	t.textColor = Color{1, 1, 1, 1}
	t.backgroundColor = Color{0, 0, 0, 1}
	t.TextChanged = NewSignal()
	
	t.SetText(text)
	return t
}

func (t Text) GetText() string { return t.text }
func (t *Text) SetText(text string) {
	t.text = text
	t.Resize(int(font.Advance(text)), int(font.LineHeight()))
	t.TextChanged.Emit(text)
}

func (t *Text) SetTextColor(color Color) {
	t.textColor = color
	t.Repaint()
}

func (t *Text) SetBackgroundColor(color Color) {
	t.backgroundColor = color
	t.Repaint()
}

func (t *Text) SetEditable(editable bool) {
	t.editable = editable
}

func (t *Text) SetValidator(validator func(*string)bool) {
	t.validator = validator
}

func (t *Text) GetMouseFocus(button int, p image.Point) MouseHandlerView {
	if t.editable { return t.ViewBase.GetMouseFocus(button, p) }
	return nil
}

func (t *Text) KeyPressed(event KeyEvent) {
	if len(event.Text) > 0 {
		text := t.text + event.Text
		if t.validator == nil || t.validator(&text) {
			t.SetText(text)
		}
	}
	switch event.Key {
	case KeyBackspace:
		if len(t.text) > 0 {
			t.SetText(t.text[:len(t.text) - 1])
		}
	}
}

func setColor(color Color) { gl.Color4d(gl.Double(color.Red), gl.Double(color.Green), gl.Double(color.Blue), gl.Double(color.Alpha)) }
func (t Text) Paint() {
	w, h := gl.Double(t.Width()), gl.Double(t.Height())
	setColor(t.backgroundColor)
	gl.Rectd(0, 0, w, h)
	
	setColor(t.textColor)
	gl.Translated(0, -gl.Double(font.Descender()), 0)
	font.Render(t.text)
}
