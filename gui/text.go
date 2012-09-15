package gui

import (
	."github.com/jteeuwen/glfw"
	gl "github.com/chsc/gogl/gl21"
	."code.google.com/p/gordon-go/ftgl"
	."code.google.com/p/gordon-go/util"
)

var font Font = CreateTextureFont(fontpath)
func init() {
	font.SetFaceSize(fontSize, 1)
}

const (
	fontSize = 18
)

type Text struct {
	*ViewBase
	AggregateMouseHandler
	text string
	textColor Color
	frameSize float64
	frameColor Color
	backgroundColor Color
	editable bool
	validator func(*string)bool
	
	TextChanged *Signal
}

func NewText(text string) *Text { return NewTextBase(nil, text) }
func NewTextBase(self View, text string) *Text {
	t := &Text{}
	if self == nil { self = t }
	t.ViewBase = NewView(t)
	t.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(t)}
	t.textColor = Color{1, 1, 1, 1}
	t.backgroundColor = Color{0, 0, 0, 1}
	t.TextChanged = NewSignal()
	t.SetText(text)
	t.Self = self
	return t
}

func (t Text) GetText() string { return t.text }
func (t *Text) SetText(text string) {
	t.text = text
	t.Resize(2*t.frameSize + font.Advance(text), 2*t.frameSize + font.LineHeight())
	t.TextChanged.Emit(text)
}

func (t Text) GetTextColor() Color { return t.textColor }
func (t *Text) SetTextColor(color Color) {
	t.textColor = color
	t.Repaint()
}

func (t *Text) SetBackgroundColor(color Color) {
	t.backgroundColor = color
	t.Repaint()
}

func (t *Text) SetFrameColor(color Color) {
	t.frameColor = color
	t.Repaint()
}

func (t *Text) SetFrameSize(size float64) {
	t.frameSize = size
	t.Resize(2*t.frameSize + font.Advance(t.text), 2*t.frameSize + font.LineHeight())
}

func (t *Text) SetEditable(editable bool) {
	t.editable = editable
}

func (t *Text) SetValidator(validator func(*string)bool) {
	t.validator = validator
}

func (t *Text) GetMouseFocus(button int, p Point) MouseHandlerView {
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

func (t Text) Paint() {
	SetColor(t.backgroundColor)
	FillRect(t.Rect().Inset(t.frameSize))
	if t.frameSize > 0 {
		SetColor(t.frameColor)
		gl.LineWidth(gl.Float(t.frameSize))
		DrawRect(t.Rect())
	}
	
	SetColor(t.textColor)
	gl.Translated(gl.Double(t.frameSize), gl.Double(t.frameSize - font.Descender()), 0)
	font.Render(t.text)
}
