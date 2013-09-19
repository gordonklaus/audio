#include "ftgl.h"
#include <FTGL/ftgl.h>

#define F (*(FTFont*)f)

void *NewTextureFont(char const *filepath) {
	return new FTTextureFont(filepath);
}

void *NewTextureFontFromBuffer(const unsigned char *pBufferBytes, size_t bufferSizeInBytes) {
	return new FTTextureFont(pBufferBytes, bufferSizeInBytes);
}

int Font_SetFaceSize(void *f, unsigned int size, unsigned int res) {
	return F.FaceSize(size, res);
}

float Font_Ascender(void *f) {
	return F.Ascender();
}

float Font_Descender(void *f) {
	return F.Descender();
}

float Font_LineHeight(void *f) {
	return F.LineHeight();
}

float Font_Advance(void *f, const char *text) {
	return F.Advance(text);
}

void Font_Render(void *f, const char *text) {
	F.Render(text);
}
