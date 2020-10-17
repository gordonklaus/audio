// +build js

package audio

import (
	"errors"

	"syscall/js"
)

var node js.Value
var callback js.Func

func startPlaying(v interface{}, c PlayControl) error {
	contextType := js.Global().Get("AudioContext")
	if contextType.IsUndefined() {
		contextType = js.Global().Get("webkitAudioContext")
	}
	if contextType.IsUndefined() {
		s := "The Web Audio API is apparently not supported in this browser."
		js.Global().Get("document").Call("write", "<p>"+s+"</p>")
		return errors.New(s)
	}
	context := contextType.New()
	Init(v, Params{SampleRate: context.Get("sampleRate").Float()})
	switch v := v.(type) {
	case Voice:
		node = context.Call("createScriptProcessor", 16384, 0, 1)
		callback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			out := args[0].Get("outputBuffer").Call("getChannelData", 0)
			for i := 0; i < out.Length(); i++ {
				out.SetIndex(i, v.Sing())
			}
			if v.Done() {
				c.Stop()
			}
			return nil
		})
	case StereoVoice:
		node = context.Call("createScriptProcessor", 16384, 0, 2)
		callback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			out := args[0].Get("outputBuffer")
			left := out.Call("getChannelData", 0)
			right := out.Call("getChannelData", 1)
			for i := 0; i < left.Length(); i++ {
				l, r := v.Sing()
				left.SetIndex(i, l)
				right.SetIndex(i, r)
			}
			if v.Done() {
				c.Stop()
			}
			return nil
		})
	}
	node.Set("onaudioprocess", callback)
	node.Call("connect", context.Get("destination"))
	return nil
}

func stopPlaying() error {
	node.Call("disconnect")
	callback.Release()
	return nil
}
