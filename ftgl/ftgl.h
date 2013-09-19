#ifndef FTGLGO_H
#define FTGLGO_H

#include <stddef.h>


#ifdef __cplusplus
extern "C" {
#endif

void *NewTextureFont(char const *filepath);
void *NewTextureFontFromBuffer(const unsigned char *pBufferBytes, size_t bufferSizeInBytes);
int Font_SetFaceSize(void *f, unsigned int size, unsigned int res);
float Font_Ascender(void *f);
float Font_Descender(void *f);
float Font_LineHeight(void *f);
float Font_Advance(void *f, const char *text);
void Font_Render(void *f, const char *text);

#ifdef __cplusplus
}
#endif


#endif // FTGLGO_H
