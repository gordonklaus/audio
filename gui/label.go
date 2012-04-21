package gui

import (
	gl "github.com/chsc/gogl/gl21"
	// ."code.google.com/p/gordon-go/ftgl"
)

type Label struct {
	ViewBase
	text string
}

func NewLabel(text string) *Label {
	l := &Label{}
	l.ViewBase = *NewView(l)
	l.SetText(text)
	return l
}

func (l *Label) GetText() string { return l.text }
func (l *Label) SetText(text string) {
	l.text = text
	l.Resize(int(font.Advance(text)), int(font.LineHeight()))
}

func (l Label) Paint() {
	gl.Translated(xMargin, -gl.Double(font.Descender()), 0)
	gl.Color4d(1, 1, 1, 1)
	font.Render(l.text)
}
