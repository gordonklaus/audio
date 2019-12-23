// +build android

#include <SLES/OpenSLES.h>
#include <SLES/OpenSLES_Android.h>
#include "_cgo_export.h"

#include <stdlib.h>

SLObjectItf engineObject;
SLEngineItf engineEngine;
SLObjectItf outputMixObject;
SLObjectItf bqPlayerObject;
SLPlayItf bqPlayerPlay;
SLAndroidSimpleBufferQueueItf bqPlayerBufferQueue;

static void bqPlayerCallback(SLAndroidSimpleBufferQueueItf bq, void *context) {
	stream_t *s = context;
	streamCallback(s);
	(*bqPlayerBufferQueue)->Enqueue(bqPlayerBufferQueue, s->outBuffer, s->outBufferByteLen);
}

void start(int channels) {
	SLresult result;
	result = slCreateEngine(&engineObject, 0, 0, 0, 0, 0);
	if (result != SL_RESULT_SUCCESS) return;
	result = (*engineObject)->Realize(engineObject, SL_BOOLEAN_FALSE);
	if (result != SL_RESULT_SUCCESS) return;
	result = (*engineObject)->GetInterface(engineObject, SL_IID_ENGINE, &engineEngine);
	if (result != SL_RESULT_SUCCESS) return;

    const SLInterfaceID ids[] = {SL_IID_VOLUME};
    const SLboolean req[] = {SL_BOOLEAN_FALSE};
    result = (*engineEngine)->CreateOutputMix(engineEngine, &(outputMixObject), 1, ids, req);
	if (result != SL_RESULT_SUCCESS) return;
    result = (*outputMixObject)->Realize(outputMixObject, SL_BOOLEAN_FALSE);
	if (result != SL_RESULT_SUCCESS) return;

	int speakers = SL_SPEAKER_FRONT_CENTER;
	if (channels > 1) 
		speakers = SL_SPEAKER_FRONT_LEFT | SL_SPEAKER_FRONT_RIGHT;
	SLDataFormat_PCM format_pcm = {SL_DATAFORMAT_PCM, (SLuint32)channels, SL_SAMPLINGRATE_48, SL_PCMSAMPLEFORMAT_FIXED_16, SL_PCMSAMPLEFORMAT_FIXED_16, speakers, SL_BYTEORDER_LITTLEENDIAN};

	SLDataLocator_AndroidSimpleBufferQueue loc_bufq = {SL_DATALOCATOR_ANDROIDSIMPLEBUFFERQUEUE, 2};
	SLDataSource audioSrc = {&loc_bufq, &format_pcm};

	// configure audio sink
	SLDataLocator_OutputMix loc_outmix = {SL_DATALOCATOR_OUTPUTMIX, outputMixObject};
	SLDataSink audioSnk = {&loc_outmix, 0};

	const SLInterfaceID ids1[] = {SL_IID_ANDROIDSIMPLEBUFFERQUEUE};
	const SLboolean req1[] = {SL_BOOLEAN_TRUE};
	result = (*engineEngine)->CreateAudioPlayer(engineEngine, &(bqPlayerObject), &audioSrc, &audioSnk, 1, ids1, req1);
	if (result != SL_RESULT_SUCCESS) return;
	result = (*bqPlayerObject)->Realize(bqPlayerObject, SL_BOOLEAN_FALSE);
	if (result != SL_RESULT_SUCCESS) return;
	result = (*bqPlayerObject)->GetInterface(bqPlayerObject, SL_IID_PLAY, &(bqPlayerPlay));
	if (result != SL_RESULT_SUCCESS) return;
	result = (*bqPlayerObject)->GetInterface(bqPlayerObject, SL_IID_ANDROIDSIMPLEBUFFERQUEUE, &bqPlayerBufferQueue);
	if (result != SL_RESULT_SUCCESS) return;

	stream_t *s = malloc(sizeof(stream_t));
	s->outBufferSampleLength = 64;
	s->outBufferByteLen = s->outBufferSampleLength * channels * sizeof(short);
	s->outBuffer = malloc(s->outBufferByteLen);

	result = (*bqPlayerBufferQueue)->RegisterCallback(bqPlayerBufferQueue, bqPlayerCallback, s);
	if (result != SL_RESULT_SUCCESS) return;
	result = (*bqPlayerPlay)->SetPlayState(bqPlayerPlay, SL_PLAYSTATE_PLAYING);
	if (result != SL_RESULT_SUCCESS) return;

	bqPlayerCallback(bqPlayerBufferQueue, s);
}

void stop() {
	(*bqPlayerObject)->Destroy(bqPlayerObject);
	(*outputMixObject)->Destroy(outputMixObject);
	(*engineObject)->Destroy(engineObject);
}
