package gui

import (
	"code.google.com/p/gordon-go/ftgl"
	"code.google.com/p/gordon-go/gui/rsrc"
	gl "github.com/chsc/gogl/gl21"
)

var font ftgl.Font

func initFont() {
	font = ftgl.NewTextureFontFromBuffer(rsrc.FreeSerif_otf)
	font.SetFaceSize(18, 1)
}

type Text interface {
	View
	GetText() string
	SetText(string)
	GetTextColor() Color
	SetTextColor(Color)
	SetBackgroundColor(Color)
	SetFrameColor(Color)
	SetFrameSize(float64)
	SetEditable(bool)
	SetValidator(func(*string) bool)
}

type TextBase struct {
	*ViewBase
	AggregateMouseHandler
	Self                Text
	text                string
	textColor           Color
	frameSize           float64
	frameColor          Color
	backgroundColor     Color
	editable            bool
	validator           func(*string) bool
	Accept, TextChanged func(string)
	Reject              func()
}

func NewText(text string) *TextBase { return NewTextBase(nil, text) }
func NewTextBase(self Text, text string) *TextBase {
	t := &TextBase{}
	if self == nil {
		self = t
	}
	t.ViewBase = NewView(t)
	t.AggregateMouseHandler = AggregateMouseHandler{NewClickKeyboardFocuser(t)}
	t.textColor = Color{1, 1, 1, 1}
	t.backgroundColor = Color{0, 0, 0, 1}
	t.Accept, t.Reject, t.TextChanged = func(string) {}, func() {}, func(string) {}
	t.SetText(text)
	t.Self = self
	t.ViewBase.Self = self
	return t
}

func (t TextBase) GetText() string { return t.text }
func (t *TextBase) SetText(text string) {
	t.text = text
	t.Resize(2*t.frameSize+font.Advance(t.text), 2*t.frameSize-font.Descender()+font.Ascender())
	t.TextChanged(text)
}

func (t TextBase) GetTextColor() Color { return t.textColor }
func (t *TextBase) SetTextColor(c Color) {
	t.textColor = c
	t.Repaint()
}

func (t *TextBase) SetBackgroundColor(c Color) {
	t.backgroundColor = c
	t.Repaint()
}

func (t *TextBase) SetFrameColor(c Color) {
	t.frameColor = c
	t.Repaint()
}

func (t *TextBase) SetFrameSize(size float64) {
	t.frameSize = size
	t.Resize(2*t.frameSize+font.Advance(t.text), 2*t.frameSize-font.Descender()+font.Ascender())
}

func (t *TextBase) SetEditable(editable bool) {
	t.editable = editable
}

func (t *TextBase) SetValidator(validator func(*string) bool) {
	t.validator = validator
}

func (t *TextBase) GetMouseFocus(button int, p Point) MouseHandlerView {
	if t.editable {
		return t.ViewBase.GetMouseFocus(button, p)
	}
	return nil
}

func (t *TextBase) KeyPressed(event KeyEvent) {
	if len(event.Text) > 0 {
		text := t.text + event.Text
		if t.validator == nil || t.validator(&text) {
			t.Self.SetText(text)
		}
	}
	switch event.Key {
	case KeyBackspace:
		if len(t.text) > 0 {
			text := t.text[:len(t.text)-1]
			if t.validator == nil || t.validator(&text) {
				t.Self.SetText(text)
			}
		}
	case KeyEnter:
		t.Accept(t.text)
	case KeyEscape:
		t.Reject()
	}
}

func (t TextBase) Paint() {
	SetColor(t.backgroundColor)
	FillRect(t.Rect().Inset(t.frameSize))
	if t.frameSize > 0 {
		SetColor(t.frameColor)
		SetLineWidth(t.frameSize)
		DrawRect(t.Rect())
	}

	SetColor(t.textColor)
	gl.Translated(gl.Double(t.frameSize), gl.Double(t.frameSize-font.Descender()), 0)
	font.Render(t.text)
}
