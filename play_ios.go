// +build ios

package audio

/*
#cgo LDFLAGS: -framework AVFoundation -framework AudioToolbox -framework Foundation
extern const char* start();
extern const char* stop();
*/
import "C"
import (
	"errors"
	"unsafe"
)

var (
	playing bool
	voice   Voice
	ctrl    PlayControl
)

func startPlaying(v Voice, c PlayControl) error {
	if playing {
		return errors.New("audio.Play doesn't yet support multiple simultaneous voices on iOS.")
	}
	playing = true
	voice = v
	ctrl = c
	Init(voice, Params{SampleRate: 44100})
	if err := C.start(); err != nil {
		return errors.New(C.GoString(err))
	}
	return nil
}

//export streamCallback
func streamCallback(buf *float32, len uint32) {
	for p := uintptr(unsafe.Pointer(buf)); len > 0; len-- {
		*(*float32)(unsafe.Pointer(p)) = float32(voice.Sing())
		p += unsafe.Sizeof(float32(0))
	}
	if voice.Done() {
		ctrl.Stop()
	}
}

func stopPlaying() error {
	if playing {
		playing = false
		if err := C.stop(); err != nil {
			return errors.New(C.GoString(err))
		}
	}
	return nil
}
