// +build android

package audio

/*
#cgo LDFLAGS: -landroid -lOpenSLES
extern void start();
extern void stop();
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
		return errors.New("audio.Play doesn't yet support multiple simultaneous voices on Android.")
	}
	playing = true
	voice = v
	ctrl = c
	Init(voice, Params{SampleRate: 48000}) // corresponds with SL_SAMPLINGRATE_48 in play_android.c
	C.start()
	return nil
}

//export streamCallback
func streamCallback(buf *int16) {
	p := uintptr(unsafe.Pointer(buf))
	for i := 64; i > 0; i-- {
		*(*int16)(unsafe.Pointer(p)) = int16(voice.Sing() * 32767)
		p += unsafe.Sizeof(int16(0))
	}
	if voice.Done() {
		ctrl.Stop()
	}
}

func stopPlaying() error {
	if playing {
		playing = false
		C.stop()
	}
	return nil
}
