package main

import (
	"fmt"
	"math"

	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/gui"
	"code.google.com/p/portaudio-go/portaudio"
)

func main() {
	w := &window{voices: map[int]*sineVoice{}}
	p := audio.Params{SampleRate: 96000, BufferSize: 64}
	audio.Init(w, p)

	portaudio.Initialize()
	defer portaudio.Terminate()
	s, err := portaudio.OpenDefaultStream(0, 1, p.SampleRate, p.BufferSize, w.processAudio)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := s.Start(); err != nil {
		fmt.Println(err)
		return
	}
	defer s.Stop()

	if err := gui.Run(func() {
		gui.NewWindow(w, "beep", func(win *gui.Window) {
			w.Window = win
			gui.SetKeyFocus(w)
		})
	}); err != nil {
		fmt.Println(err)
	}
}

type window struct{
	*gui.Window
	Multi audio.MultiVoice
	voices map[int]*sineVoice
}

func (w *window) KeyPress(k gui.KeyEvent) {
	_, playing := w.voices[k.Key]
	if p, ok := keyPitch[k.Key]; !playing && ok {
		v := &sineVoice{}
		v.Sine.SetFreq(pitchToFreq(p))
		w.voices[k.Key] = v
		v.Env.SetAttackTime(.01)
		v.Env.SetReleaseTime(2)
		w.Multi.Add(v)
	}
}

func (w *window) KeyRelease(k gui.KeyEvent) {
	if v, ok := w.voices[k.Key]; ok {
		delete(w.voices, k.Key)
		v.Env.Release()
	}
}

func (w *window) processAudio(in [][]float32, out [][]float32) {
	a, _ := w.Multi.Sing()
	for i, x := range a {
		out[0][i] = float32(x)
	}
}

type sineVoice struct{Sine audio.SineOsc; Env audio.AttackReleaseEnv}

func (v *sineVoice) Sing() (audio.Audio, bool) {
	a, done := v.Env.Sing()
	return distort(a.Mul(a, v.Sine.Sine())), done
}

func distort(a audio.Audio) audio.Audio {
	for i, x := range a {
		y := math.Abs(x) + 1
		y *= y
		a[i] = math.Copysign((y-1)/y, x)
	}
	return a
}

func pitchToFreq(pitch float64) float64 { return 512 * math.Pow(2, pitch/12) }

var keyPitch = map[int]float64{
	gui.KeyZ:            -12,
	gui.KeyS:            -11,
	gui.KeyX:            -10,
	gui.KeyD:            -9,
	gui.KeyC:            -8,
	gui.KeyV:            -7,
	gui.KeyG:            -6,
	gui.KeyB:            -5,
	gui.KeyH:            -4,
	gui.KeyN:            -3,
	gui.KeyJ:            -2,
	gui.KeyM:            -1,
	gui.KeyComma:        0,
	gui.KeyL:            1,
	gui.KeyPeriod:       2,
	gui.KeySemicolon:    3,
	gui.KeySlash:        4,
	gui.KeyQ:            0,
	gui.Key2:            1,
	gui.KeyW:            2,
	gui.Key3:            3,
	gui.KeyE:            4,
	gui.KeyR:            5,
	gui.Key5:            6,
	gui.KeyT:            7,
	gui.Key6:            8,
	gui.KeyY:            9,
	gui.Key7:            10,
	gui.KeyU:            11,
	gui.KeyI:            12,
	gui.Key9:            13,
	gui.KeyO:            14,
	gui.Key0:            15,
	gui.KeyP:            16,
	gui.KeyLeftBracket:  17,
	gui.KeyEqual:        18,
	gui.KeyRightBracket: 19,
	gui.KeyBackslash:    21,
}
