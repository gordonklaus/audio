package gui

import (
	"code.google.com/p/gordon-go/ftgl"
	"code.google.com/p/gordon-go/glfw"
	gl "github.com/chsc/gogl/gl21"
	"go/build"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Text struct {
	*ViewBase
	text                string
	font                ftgl.Font
	textColor           Color
	frameSize           float64
	frameColor          Color
	backgroundColor     Color
	Validate            func(*string) bool
	Accept, TextChanged func(string)
	Reject              func()

	cursor     bool
	stopCursor chan chan bool
}

func NewText(text string) *Text {
	t := &Text{}
	t.ViewBase = NewView(t)
	t.font = getFont()
	t.textColor = Color{1, 1, 1, 1}
	t.backgroundColor = Color{0, 0, 0, 1}
	t.stopCursor = make(chan chan bool)
	t.SetText(text)
	return t
}

var fontCache = struct {
	sync.Mutex
	m map[*glfw.Window]ftgl.Font
}{m: map[*glfw.Window]ftgl.Font{}}

// Must be called from a thread holding an OpenGL context, i.e., a window callback thread.
func getFont() ftgl.Font {
	w := glfw.GetCurrentContext()
	if w == nil {
		panic("no current context")
	}
	fontCache.Lock()
	defer fontCache.Unlock()
	font := fontCache.m[w]
	if font.Nil() {
		font = ftgl.NewTextureFont(filepath.Join(pkgDir(), "Times New Roman.ttf"))
		font.SetFaceSize(18, 1)
		fontCache.m[w] = font
	}
	return font
}

func pkgDir() string {
	for _, dir := range build.Default.SrcDirs() {
		dir := filepath.Join(dir, "code.google.com/p/gordon-go/gui")
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	panic("unreachable")
}

func (t Text) Text() string { return t.text }
func (t *Text) SetText(text string) {
	t.text = text
	t.Resize(2*t.frameSize+t.font.Advance(t.text), 2*t.frameSize-t.font.Descender()+t.font.Ascender())
	if t.TextChanged != nil {
		t.TextChanged(text)
	}
}

func (t *Text) SetTextColor(c Color) {
	t.textColor = c
	Repaint(t)
}

func (t *Text) SetBackgroundColor(c Color) {
	t.backgroundColor = c
	Repaint(t)
}

func (t *Text) SetFrameColor(c Color) {
	t.frameColor = c
	Repaint(t)
}

func (t *Text) SetFrameSize(size float64) {
	t.frameSize = size
	t.Resize(2*t.frameSize+t.font.Advance(t.text), 2*t.frameSize-t.font.Descender()+t.font.Ascender())
}

func (t *Text) TookKeyFocus() {
	t.cursor = true
	Repaint(t)
	go func() {
		tick := time.NewTicker(time.Second / 2)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				select {
				case DoChan(t) <- func() {
					t.cursor = !t.cursor
					Repaint(t)
				}:
				case ch := <-t.stopCursor:
					ch <- true
					return
				}
			case ch := <-t.stopCursor:
				ch <- true
				return
			}
		}
	}()
}

func (t *Text) LostKeyFocus() {
	ch := make(chan bool)
	t.stopCursor <- ch
	<-ch
	t.cursor = false
	Repaint(t)
}

func (t *Text) KeyPress(event KeyEvent) {
	if len(event.Text) > 0 {
		text := t.text + event.Text
		if t.Validate == nil || t.Validate(&text) {
			t.SetText(text)
		}
	}
	switch event.Key {
	case KeyBackspace:
		if len(t.text) > 0 {
			text := t.text[:len(t.text)-1]
			if t.Validate == nil || t.Validate(&text) {
				t.SetText(text)
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

func (t *Text) Paint() {
	SetColor(t.backgroundColor)
	FillRect(InnerRect(t).Inset(t.frameSize))
	if t.frameSize > 0 {
		SetColor(t.frameColor)
		SetLineWidth(t.frameSize)
		DrawRect(InnerRect(t))
	}

	if t.cursor {
		SetColor(t.textColor)
		SetLineWidth(2)
		x := t.frameSize + t.font.Advance(t.text)
		DrawLine(Pt(x, t.frameSize), Pt(x, Height(t)-2*t.frameSize))
	}

	SetColor(t.textColor)
	gl.Translated(gl.Double(t.frameSize), gl.Double(t.frameSize-t.font.Descender()), 0)
	t.font.Render(t.text)
}
