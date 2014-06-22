package ftgl

// #cgo pkg-config: ftgl
// #include "ftgl.h"
// #include <stdlib.h>
import "C"

import "unsafe"

type Font struct {
	f unsafe.Pointer
}

func NewTextureFont(filePath string) Font {
	s := C.CString(filePath)
	defer C.free(unsafe.Pointer(s))
	f := C.NewTextureFont(s)
	if f == nil {
		panic("unable to create font from file " + filePath)
	}
	return Font{f}
}

// the caller must ensure that buffer is not collected before the Font is done being used
func NewTextureFontFromBuffer(buffer []byte) Font {
	f := C.NewTextureFontFromBuffer((*C.uchar)(&buffer[0]), C.size_t(len(buffer)))
	if f == nil {
		panic("unable to create font from buffer")
	}
	return Font{f}
}

func (f Font) Nil() bool { return f.f == nil }

func (f Font) SetFaceSize(size, res uint) {
	if C.Font_SetFaceSize(f.f, C.uint(size), C.uint(res)) != 1 {
		panic("unable to set face size")
	}
}

func (f Font) Ascender() float64 {
	return float64(C.Font_Ascender(f.f))
}

func (f Font) Descender() float64 {
	return float64(C.Font_Descender(f.f))
}

func (f Font) LineHeight() float64 {
	return float64(C.Font_LineHeight(f.f))
}

func (f Font) Advance(text string) float64 {
	s := C.CString(text)
	defer C.free(unsafe.Pointer(s))
	return float64(C.Font_Advance(f.f, s))
}

func (f Font) Render(text string) {
	s := C.CString(text)
	defer C.free(unsafe.Pointer(s))
	C.Font_Render(f.f, s)
}
