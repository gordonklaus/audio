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
	pitchRange  float32 = 4
	pitchOffset float32 = 7

	mel  = newMelody(big.NewRat(512, 1), 5)
	keys []key

	pressed = map[*event.Touch]key{}
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
	projmat.Identity()
	projmat.Translate(&projmat, -1, -1, 0)
	projmat.Scale(&projmat, 2/pitchRange, 2, 1)
	projmat.Translate(&projmat, -pitchOffset, 0, 0)

	initKeys()

	startAudio()
}

func stop() {
	stopAudio()

	gl.DeleteProgram(program)
	gl.DeleteBuffer(positionbuf)
	gl.DeleteBuffer(pointsizebuf)
}

func touch(t *event.Touch) {
	// finger := fingers.touch(t)

	// TODO: Don't manually invert or multiply projection matrix; use improved package f32.
	m := new(f32.Mat4)
	m.Identity()
	m.Translate(m, pitchOffset, 0, 0)
	m.Scale(m, pitchRange/2, .5, 1)
	m.Translate(m, 1, 1, 0)
	v := new(f32.Mat4)
	v[0][0] = float32(2*t.Loc.X/geom.Width - 1)
	v[1][0] = float32(1 - 2*t.Loc.Y/geom.Height)
	v[3][0] = 1
	v.Mul(m, v)

	pitch := float64(v[0][0])
	y := float64(v[1][0])
	var key key
	mind := math.MaxFloat64
	for _, k := range keys {
		kb := k.base()
		dx := kb.pitch - pitch
		dy := kb.y - y
		d := dx*dx + dy*dy
		if d < mind {
			mind = d
			key = k
		}
	}
	switch t.Type {
	case event.TouchStart:
		pressed[t] = key
		pressed[t].press(t.Loc)
	case event.TouchMove:
		pressed[t].move(t.Loc)
	case event.TouchEnd:
		pressed[t].release(t.Loc)
		delete(pressed, t)
	}
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
