package main

import (
	"log"
	"math"
	"math/big"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/f32"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/gl/glutil"
)

var (
	print   = log.Print
	printf  = log.Printf
	println = log.Println
)

var (
	program    gl.Program
	projection gl.Uniform

	projmat     f32.Mat4
	pitchRange  float64 = 4
	pitchOffset float64 = 7

	mel  = newMelody(big.NewRat(512, 1), 5)
	keys []key

	pressed   = map[*event.Touch]key{}
	scrollLoc = map[*event.Touch]geom.Pt{}
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
	var err error
	program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	gl.Enable(34370) // GL_PROGRAM_POINT_SIZE; apparently not necessary on Android
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_COLOR)

	projection = gl.GetUniformLocation(program, "projection")
	updateProjectionMatrix()

	initKeys()

	startAudio()
}

func updateProjectionMatrix() {
	projmat.Identity()
	projmat.Translate(&projmat, -1, -1, 0)
	projmat.Scale(&projmat, 2/float32(pitchRange), 2, 1)
	projmat.Translate(&projmat, -float32(pitchOffset), 0, 0)
}

func stop() {
	stopAudio()

	gl.DeleteProgram(program)
	gl.DeleteBuffer(positionbuf)
	gl.DeleteBuffer(pointsizebuf)
}

func touch(t *event.Touch) {
	// finger := fingers.touch(t)

	if t.Type == event.TouchStart && t.Loc.Y < 36 {
		scrollLoc[t] = t.Loc.X
	}
	if _, ok := scrollLoc[t]; ok {
		avg0, stddev0 := scrollStats()
		scrollLoc[t] = t.Loc.X
		avg1, stddev1 := scrollStats()
		scale := 1.0
		if stddev0 > 0 && stddev1 > 0 {
			scale = float64(stddev1 / stddev0)
		}
		pitchOffset -= pitchRange * float64((avg1/geom.Pt(scale)-avg0)/geom.Width)
		pitchRange /= scale
		updateProjectionMatrix()
		if t.Type == event.TouchEnd {
			delete(scrollLoc, t)
		}
		return
	}

	switch t.Type {
	case event.TouchStart:
		pressed[t] = nearestKey(t.Loc)
		pressed[t].press(t.Loc)
	case event.TouchMove:
		pressed[t].move(t.Loc)
	case event.TouchEnd:
		pressed[t].release(t.Loc)
		delete(pressed, t)
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
	p := pitchOffset + pitchRange*float64(loc.X/geom.Width)
	y := 1 - float64(loc.Y/geom.Height)
	var key key
	min := math.MaxFloat64
	for _, k := range keys {
		kb := k.base()
		dx := kb.pitch - p
		dy := kb.y - y
		d := dx*dx + dy*dy
		if d < min {
			min = d
			key = k
		}
	}
	return key
}

func draw() {
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(program)
	projection.WriteMat4(&projmat)
	drawKeys()
}

const vertexShader = `#version 100
uniform mat4 projection;
attribute vec3 position;
attribute float pointsize;
void main() {
	gl_Position = projection * vec4(position, 1);
	gl_PointSize = pointsize;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main(void)
{
    vec2 v = 2.0*gl_PointCoord.xy - vec2(1.0);
	float r2 = dot(v, v);
	float a = (1.0 / (r2 + .25) - .8) / 3.2;
    gl_FragColor = a * color;
}
`

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
