package ftgl

// #cgo pkg-config: ftgl
// #include "FTGL/ftgl.h"
import "C"

import "unsafe"

type Font struct {
	font *C.FTGLfont
}

func CreateTextureFont(filePath string) Font {
	s := C.CString(filePath)
	defer C.free(unsafe.Pointer(s))
	font := C.ftglCreateTextureFont(s)
	if font == nil { panic("couldn't create font") }
	return Font{font}
}

func (font Font) SetFaceSize(size, res uint) bool {
	return C.ftglSetFontFaceSize(font.font, C.uint(size), C.uint(res)) == 1
}

func (font Font) Ascender() float64 {
	return float64(C.ftglGetFontAscender(font.font))
}

func (font Font) Descender() float64 {
	return float64(C.ftglGetFontDescender(font.font))
}

func (font Font) LineHeight() float64 {
	return float64(C.ftglGetFontLineHeight(font.font))
}

func (font Font) Advance(text string) float64 {
	s := C.CString(text)
	defer C.free(unsafe.Pointer(s))
	return float64(C.ftglGetFontAdvance(font.font, s))
}

func (font Font) Render(text string) {
	s := C.CString(text)
	defer C.free(unsafe.Pointer(s))
	C.ftglRenderFont(font.font, s, C.FTGL_RENDER_ALL)
}