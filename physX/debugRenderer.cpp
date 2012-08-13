#include <PxPhysicsAPI.h>

#define _USE_MATH_DEFINES
#include <cmath>

#include <GL/gl.h>
#include <GL/glu.h>

#include "debugRenderer.h"

using namespace physx;

void renderBuffer(float* pVertList, float* pColorList, int type, int num) {
	glEnableClientState(GL_VERTEX_ARRAY);
	glVertexPointer(3,GL_FLOAT, 0, pVertList);
	glEnableClientState(GL_COLOR_ARRAY);
	glColorPointer(4, GL_FLOAT, 0, pColorList);
	glDrawArrays(type, 0, num);
	glDisableClientState(GL_COLOR_ARRAY);
    glDisableClientState(GL_VERTEX_ARRAY);
}

DebugRenderer* newDebugRenderer(void* scene) {
	PxScene* s = (PxScene*)scene;
	s->setVisualizationParameter(PxVisualizationParameter::eSCALE, 1);
	s->setVisualizationParameter(PxVisualizationParameter::eWORLD_AXES, 1);
	s->setVisualizationParameter(PxVisualizationParameter::eCOLLISION_SHAPES, 1);
	
	DebugRenderer* r = new DebugRenderer;
	r->scene = scene;
	r->distance = 5;
	r->angle = .9;
	r->elevation = .05;
	r->x = 0;
	r->y = 0;
	return r;
}

void DebugRenderer_resize(DebugRenderer* r, int w, int h) {
	glClearColor(0, 0, 0, 1);
	
	glViewport(0, 0, w, h);
	glMatrixMode(GL_PROJECTION);
	glLoadIdentity();
	gluPerspective(65, double(w) / h, 1, 10000);
}

void DebugRenderer_zoom(DebugRenderer* r, double factor) {
	r->distance *= factor;
}

void DebugRenderer_rotate(DebugRenderer* r, double dAngle, double dElevation) {
	r->angle += dAngle;
	r->elevation += dElevation;
	if(r->elevation > .249) r->elevation = .249;
	if(r->elevation < -.249) r->elevation = -.249;
}

void DebugRenderer_pan(DebugRenderer* r, double right, double forward) {
	double rightAngle = r->angle + .25;
	double forwardAngle = r->angle + .5;
	r->x += cosf(2*M_PI*rightAngle)*right + cosf(2*M_PI*forwardAngle)*forward;
	r->y += sinf(2*M_PI*rightAngle)*right + sinf(2*M_PI*forwardAngle)*forward;
}

void DebugRenderer_render(DebugRenderer* r) {
	glClear(GL_COLOR_BUFFER_BIT | GL_DEPTH_BUFFER_BIT);
	float height = 1;
	glMatrixMode(GL_MODELVIEW);
	glLoadIdentity();
	gluLookAt(r->x + cosf(2*M_PI*r->elevation)*cosf(2*M_PI*r->angle)*r->distance, r->y + cosf(2*M_PI*r->elevation)*sinf(2*M_PI*r->angle)*r->distance, height + sinf(2*M_PI*r->elevation)*r->distance, r->x, r->y, height, 0, 0, 1);


	const PxRenderBuffer& data = ((PxScene*)r->scene)->getRenderBuffer();

	glLineWidth(1.0f);
	glColor4f(0.7f,0.7f,0.7f,1.0f);
	glDisable(GL_LIGHTING);

	unsigned int NbPoints = data.getNbPoints();
	if(NbPoints)
	{
		float* pVertList = new float[NbPoints*3];
    	float* pColorList = new float[NbPoints*4];
    	int vertIndex = 0;
    	int colorIndex = 0;
		const PxDebugPoint* Points = data.getPoints();
		while(NbPoints--)
		{
        	pVertList[vertIndex++] = Points->pos.x;
        	pVertList[vertIndex++] = Points->pos.y;
        	pVertList[vertIndex++] = Points->pos.z;
        	pColorList[colorIndex++] = (float)((Points->color>>16)&0xff)/255.0f;
        	pColorList[colorIndex++] = (float)((Points->color>>8)&0xff)/255.0f;
        	pColorList[colorIndex++] = (float)(Points->color&0xff)/255.0f;
        	pColorList[colorIndex++] = 1.0f;
	      	Points++;
		}
		
		renderBuffer(pVertList, pColorList, GL_POINTS, data.getNbPoints());
    	
    	delete[] pVertList;
    	delete[] pColorList;
	}

	unsigned int NbLines = data.getNbLines();
	if(NbLines)
	{
		float* pVertList = new float[NbLines*3*2];
    	float* pColorList = new float[NbLines*4*2];
    	int vertIndex = 0;
    	int colorIndex = 0;
		const PxDebugLine* Lines = data.getLines();
		while(NbLines--)
		{
        	pVertList[vertIndex++] = Lines->pos0.x;
        	pVertList[vertIndex++] = Lines->pos0.y;
        	pVertList[vertIndex++] = Lines->pos0.z;
        	pColorList[colorIndex++] = (float)((Lines->color0>>16)&0xff)/255.0f;
        	pColorList[colorIndex++] = (float)((Lines->color0>>8)&0xff)/255.0f;
        	pColorList[colorIndex++] = (float)(Lines->color0&0xff)/255.0f;
        	pColorList[colorIndex++] = 1.0f;

        	pVertList[vertIndex++] = Lines->pos1.x;
        	pVertList[vertIndex++] = Lines->pos1.y;
        	pVertList[vertIndex++] = Lines->pos1.z;
        	pColorList[colorIndex++] = (float)((Lines->color1>>16)&0xff)/255.0f;
        	pColorList[colorIndex++] = (float)((Lines->color1>>8)&0xff)/255.0f;
        	pColorList[colorIndex++] = (float)(Lines->color1&0xff)/255.0f;
        	pColorList[colorIndex++] = 1.0f;

	      	Lines++;
		}
		
		renderBuffer(pVertList, pColorList, GL_LINES, data.getNbLines()*2);
    	
    	delete[] pVertList;
    	delete[] pColorList;
	}

	unsigned int NbTris = data.getNbTriangles();
	if(NbTris)
	{
		float* pVertList = new float[NbTris*3*3];
    	float* pColorList = new float[NbTris*4*3];
    	int vertIndex = 0;
    	int colorIndex = 0;
		const PxDebugTriangle* Triangles = data.getTriangles();
		while(NbTris--)
		{
        	pVertList[vertIndex++] = Triangles->pos0.x;
        	pVertList[vertIndex++] = Triangles->pos0.y;
        	pVertList[vertIndex++] = Triangles->pos0.z;
    		pColorList[colorIndex++] = (float)((Triangles->color0>>16)&0xff)/255.0f;
    		pColorList[colorIndex++] = (float)((Triangles->color0>>8)&0xff)/255.0f;
    		pColorList[colorIndex++] = (float)(Triangles->color0&0xff)/255.0f;
    		pColorList[colorIndex++] = 1.0f;

        	pVertList[vertIndex++] = Triangles->pos1.x;
        	pVertList[vertIndex++] = Triangles->pos1.y;
        	pVertList[vertIndex++] = Triangles->pos1.z;
    		pColorList[colorIndex++] = (float)((Triangles->color1>>16)&0xff)/255.0f;
    		pColorList[colorIndex++] = (float)((Triangles->color1>>8)&0xff)/255.0f;
    		pColorList[colorIndex++] = (float)(Triangles->color1&0xff)/255.0f;
    		pColorList[colorIndex++] = 1.0f;

        	pVertList[vertIndex++] = Triangles->pos2.x;
        	pVertList[vertIndex++] = Triangles->pos2.y;
        	pVertList[vertIndex++] = Triangles->pos2.z;
    		pColorList[colorIndex++] = (float)((Triangles->color2>>16)&0xff)/255.0f;
    		pColorList[colorIndex++] = (float)((Triangles->color2>>8)&0xff)/255.0f;
    		pColorList[colorIndex++] = (float)(Triangles->color2&0xff)/255.0f;
    		pColorList[colorIndex++] = 1.0f;

	      	Triangles++;
		}
		
		renderBuffer(pVertList, pColorList, GL_TRIANGLES, data.getNbTriangles()*3);
  	
    	delete[] pVertList;
    	delete[] pColorList;
	}
}

