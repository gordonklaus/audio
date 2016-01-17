// +build js

package audio

import (
	"errors"

	"github.com/gopherjs/gopherjs/js"
)

var node *js.Object

func startPlaying(v Voice, c PlayControl) error {
	contextType := js.Global.Get("AudioContext")
	if contextType == js.Undefined {
		contextType = js.Global.Get("webkitAudioContext")
	}
	if contextType == js.Undefined {
		s := "The Web Audio API is apparently not supported in this browser."
		js.Global.Get("document").Call("write", "<p>"+s+"</p>")
		return errors.New(s)
	}
	context := contextType.New()
	Init(v, Params{SampleRate: context.Get("sampleRate").Float()})
	node = context.Call("createScriptProcessor", 1024, 0, 1)
	node.Set("onaudioprocess", func(e *js.Object) {
		out := e.Get("outputBuffer").Call("getChannelData", 0).Interface().([]float32)
		for i := range out {
			out[i] = float32(v.Sing())
		}
		if v.Done() {
			c.Stop()
		}
	})
	node.Call("connect", context.Get("destination"))
	return nil
}

func stopPlaying() error {
	node.Call("disconnect")
	return nil
}
