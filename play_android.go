// +build android

package audio

/*
#cgo LDFLAGS: -landroid -lOpenSLES

typedef struct {
	short *outBuffer;
	int outBufferSampleLength;
	int outBufferByteLen;
} stream_t;

extern void start(int channels);
extern void stop();
*/
import "C"
import (
	"unsafe"
)

var (
	voice interface{}
	ctrl  PlayControl
)

func startPlaying(v interface{}, c PlayControl) error {
	voice = v
	ctrl = c
	Init(voice, Params{SampleRate: 48000}) // corresponds with SL_SAMPLINGRATE_48 in play_android.c

	channels := 1
	if _, ok := voice.(StereoVoice); ok {
		channels = 2
	}

	C.start(C.int(channels))
	return nil
}

//export streamCallback
func streamCallback(s *C.stream_t) {
	p := uintptr(unsafe.Pointer(s.outBuffer))
	switch voice := voice.(type) {
	case Voice:
		for i := s.outBufferSampleLength; i > 0; i-- {
			*(*int16)(unsafe.Pointer(p)) = int16(voice.Sing() * 32767)
			p += unsafe.Sizeof(int16(0))
		}
		if voice.Done() {
			ctrl.Stop()
		}
	case StereoVoice:
		for i := s.outBufferSampleLength; i > 0; i-- {
			left, right := voice.Sing()
			*(*int16)(unsafe.Pointer(p)) = int16(left * 32767)
			p += unsafe.Sizeof(int16(0))
			*(*int16)(unsafe.Pointer(p)) = int16(right * 32767)
			p += unsafe.Sizeof(int16(0))
		}
		if voice.Done() {
			ctrl.Stop()
		}
	}
}

func stopPlaying() error {
	C.stop()
	return nil
}
