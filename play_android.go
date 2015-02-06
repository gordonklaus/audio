// +build android

package audio

/*
#cgo LDFLAGS: -landroid -lOpenSLES
extern void start();
extern void stop();
*/
import "C"
import "unsafe"

var (
	started  bool
	out      [64]float32
	callback func(out []float32)
)

func startPlaying(v Voice, cb func(out []float32)) error {
	Init(v, Params{SampleRate: 48000}) // corresponds with SL_SAMPLINGRATE_48 in play_android.c
	if !started {
		started = true
		callback = cb
		C.start()
	}
	return nil
}

//export streamCallback
func streamCallback(buf *int16) {
	callback(out[:])
	p := uintptr(unsafe.Pointer(buf))
	for i := range out {
		*(*int16)(unsafe.Pointer(p)) = int16(out[i] * 32767)
		p += unsafe.Sizeof(int16(0))
	}
}

func stopPlaying() error {
	if started {
		started = false
		C.stop()
	}
	return nil
}
