package gui

import (
	"code.google.com/p/gordon-go/ftgl"
	gl "github.com/chsc/gogl/gl21"
)

var font ftgl.Font

func initFont() {
	font = ftgl.NewTextureFont("/Library/Fonts/Times New Roman.ttf")
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
	SetValidator(func(*string) bool)
}

type TextBase struct {
	*ViewBase
	Self                Text
	text                string
	textColor           Color
	frameSize           float64
	frameColor          Color
	backgroundColor     Color
	Validate            func(*string) bool
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
	t.textColor = Color{1, 1, 1, 1}
	t.backgroundColor = Color{0, 0, 0, 1}
	t.SetText(text)
	t.Self = self
	t.ViewBase.Self = self
	return t
}

func (t TextBase) GetText() string { return t.text }
func (t *TextBase) SetText(text string) {
	t.text = text
	Resize(t, Pt(2*t.frameSize+font.Advance(t.text), 2*t.frameSize-font.Descender()+font.Ascender()))
	if t.TextChanged != nil {
		t.TextChanged(text)
	}
}

func (t TextBase) GetTextColor() Color { return t.textColor }
func (t *TextBase) SetTextColor(c Color) {
	t.textColor = c
	Repaint(t)
}

func (t *TextBase) SetBackgroundColor(c Color) {
	t.backgroundColor = c
	Repaint(t)
}

func (t *TextBase) SetFrameColor(c Color) {
	t.frameColor = c
	Repaint(t)
}

func (t *TextBase) SetFrameSize(size float64) {
	t.frameSize = size
	Resize(t, Pt(2*t.frameSize+font.Advance(t.text), 2*t.frameSize-font.Descender()+font.Ascender()))
}

func (t *TextBase) SetValidator(validate func(*string) bool) {
	t.Validate = validate
}

func (t *TextBase) KeyPress(event KeyEvent) {
	if len(event.Text) > 0 {
		text := t.text + event.Text
		if t.Validate == nil || t.Validate(&text) {
			t.Self.SetText(text)
		}
	}
	switch event.Key {
	case KeyBackspace:
		if len(t.text) > 0 {
			text := t.text[:len(t.text)-1]
			if t.Validate == nil || t.Validate(&text) {
				t.Self.SetText(text)
			}
		}
	case KeyEnter:
		if t.Accept != nil {
			t.Accept(t.text)
		}
	case KeyEscape:
		if t.Reject != nil {
			t.Reject()
		}
	}
}

func (t *TextBase) Paint() {
	SetColor(t.backgroundColor)
	FillRect(Rect(t).Inset(t.frameSize))
	if t.frameSize > 0 {
		SetColor(t.frameColor)
		SetLineWidth(t.frameSize)
		DrawRect(Rect(t))
	}

	SetColor(t.textColor)
	gl.Translated(gl.Double(t.frameSize), gl.Double(t.frameSize-font.Descender()), 0)
	font.Render(t.text)
}
