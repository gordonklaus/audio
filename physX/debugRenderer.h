#ifndef DEBUGRENDERER_H
#define DEBUGRENDERER_H


#ifdef __cplusplus
extern "C" {
#endif

typedef struct DebugRenderer {
	void* scene;
	double distance;
	double angle;
	double elevation;
	double x, y;
} DebugRenderer;

DebugRenderer* newDebugRenderer(void* scene);
void DebugRenderer_resize(DebugRenderer* r, int w, int h);
void DebugRenderer_zoom(DebugRenderer* r, double factor);
void DebugRenderer_rotate(DebugRenderer* r, double dAngle, double dElevation);
void DebugRenderer_pan(DebugRenderer* r, double right, double forward);
void DebugRenderer_render(DebugRenderer* r);

#ifdef __cplusplus
}
#endif


#endif // DEBUGRENDERER_H

