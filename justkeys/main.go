package main

import (
	"log"
	"math"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

var (
	print   = log.Print
	printf  = log.Printf
	println = log.Println
)

var (
	pressed   = map[event.TouchSequenceID]key{}
	scrollLoc = map[event.TouchSequenceID]geom.Pt{}
	// fingers Fingers
)

func main() {
	app.Run(app.Callbacks{
		Start: start,
		Stop:  stop,
		Touch: touch,
		Draw:  draw,
	})
}

func start() {
	initKeys()
	startAudio()
}

func stop() {
	stopAudio()

	gl.DeleteProgram(program)
	gl.DeleteBuffer(positionbuf)
	gl.DeleteBuffer(pointsizebuf)
}

func touch(t event.Touch) {
	// finger := fingers.touch(t)

	if t.Type == event.TouchStart && t.Loc.Y < 8 {
		scrollLoc[t.ID] = t.Loc.X
	}
	if _, ok := scrollLoc[t.ID]; ok {
		avg0, stddev0 := scrollStats()
		scrollLoc[t.ID] = t.Loc.X
		avg1, stddev1 := scrollStats()
		scale := 1.0
		if stddev0 > 0 && stddev1 > 0 {
			scale = float64(stddev1 / stddev0)
		}
		pitchOffset -= pitchRange * float64((avg1/geom.Pt(scale)-avg0)/geom.Width)
		pitchRange /= scale
		updateProjectionMatrix()
		if t.Type == event.TouchEnd {
			delete(scrollLoc, t.ID)
		}
		return
	}

	if t.Type == event.TouchStart {
		if k := nearestKey(t.Loc); k != nil {
			pressed[t.ID] = k
		}
	}
	if k := pressed[t.ID]; k != nil {
		switch t.Type {
		case event.TouchStart:
			k.press(t.Loc)
		case event.TouchMove:
			k.move(t.Loc)
		case event.TouchEnd:
			k.release(t.Loc)
			delete(pressed, t.ID)
		}
	}
}

func scrollStats() (avg, stddev geom.Pt) {
	n := geom.Pt(len(scrollLoc))
	for _, x := range scrollLoc {
		avg += x
	}
	avg /= n
	for _, x := range scrollLoc {
		stddev += (x - avg) * (x - avg)
	}
	return avg, geom.Pt(math.Sqrt(float64(stddev / n)))
}

func nearestKey(loc geom.Point) key {
	var key key
	dist := geom.Pt(math.MaxFloat32)
	for _, k := range keys {
		kb := k.base()
		dx := geom.Pt((kb.pitch-pitchOffset)/pitchRange)*geom.Width - loc.X
		dy := geom.Pt(1-kb.y)*geom.Height - loc.Y
		d := geom.Pt(math.Hypot(float64(dx), float64(dy)))
		if d < dist && d < geom.Pt(math.Max(8, kb.size/float64(geom.PixelsPerPt))) {
			dist = d
			key = k
		}
	}
	return key
}

func draw() {
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	drawKeys()
}

// type Fingers struct {
// 	fingers []*finger
// }
//
// func (f *Fingers) touch(t event.Touch) *finger {
// 	if t.Type == event.TouchStart {
// 		finger := &finger{t.Loc}
// 		f.fingers = append(f.fingers, finger)
// 		return finger
// 	}
// 	index := 0
// 	dist := geom.Pt(math.MaxFloat32)
// 	for i, finger := range f.fingers {
// 		dx := finger.loc.X - t.Loc.X
// 		dy := finger.loc.Y - t.Loc.Y
// 		d := dx*dx + dy*dy
// 		if d < dist {
// 			index = i
// 			dist = d
// 		}
// 	}
// 	finger := f.fingers[index]
// 	finger.loc = t.Loc
// 	if t.Type == event.TouchEnd {
// 		n := len(f.fingers)
// 		f.fingers[index] = f.fingers[n-1]
// 		f.fingers = f.fingers[:n-1]
// 	}
// 	return finger
// }
//
// type finger struct {
// 	loc geom.Point
// }
