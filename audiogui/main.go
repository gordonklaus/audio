package audiogui

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/gui"

	"os"
	"path/filepath"
	"runtime"
)

func Main(score *audio.Score, band audio.Band) {
	_, path, _, _ := runtime.Caller(1)
	name := filepath.Base(path)
	name = name[:len(name)-3]
	path = filepath.Dir(path)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "edit":
			gui.Run(func() {
				gui.NewWindow(nil, name, func(w *gui.Window) {
					v := NewScoreView(score, band)
					v.path = filepath.Join(path, "score.go")
					w.SetCentralView(v)
					v.InitFocus()
				})
			})
		case "write":
			Write(audio.NewScorePlayer(score, band), filepath.Join(path, name+".wav"))
		default:
			println("unknown arg: " + os.Args[1])
		}
	} else {
		Play(audio.NewScorePlayer(score, band))
	}
}
