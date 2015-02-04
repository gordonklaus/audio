package main

import (
	"encoding/binary"
	"log"
	"math"
	"sort"
	"time"

	"code.google.com/p/gordon-go/audio"

	"golang.org/x/mobile/f32"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/gl/glutil"
)

// TODO: Don't use gl_PointSize; actual size in pixels is distance-attenuated.

type key interface {
	base() *keyBase
	press(loc geom.Point)
	move(loc geom.Point)
	release(loc geom.Point)
}

type keyBase struct {
	ratio      ratio
	pitch      float64
	complexity int
	y          float64
	size       float64
	voice      ampVoice
}

func (k *keyBase) base() *keyBase { return k }

type ampVoice interface {
	audio.Voice
	amp() float64
}

type pressedKey struct {
	*keyBase
}

func (k *pressedKey) press(loc geom.Point) {
	if k.voice == nil || k.voice.Done() {
		k.voice = newPressedTone(math.Exp2(k.pitch))
		multivoice.Add(k.voice)
	}
	updateKeys(k.ratio)
}

func (k *pressedKey) move(loc geom.Point) {
	k.voice.(*pressedTone).attack(-3 * float64(loc.Y/geom.Height))
}

func (k *pressedKey) release(loc geom.Point) {
	if loc.Y < geom.Height-8 {
		k.move(loc)
	} else {
		k.voice.(*pressedTone).release()
	}
}

type pluckedKey struct {
	*keyBase
	pressLoc geom.Point
}

func (k *pluckedKey) press(loc geom.Point) {
	k.pressLoc = loc
}

func (k *pluckedKey) move(loc geom.Point) {}

func (k *pluckedKey) release(loc geom.Point) {
	amp := math.Max(-6, math.Log2(math.Tanh(dist(loc, k.pressLoc)/64)))

	updateKeys(k.ratio)
	v := newPluckedTone(amp, math.Exp2(k.pitch))
	multivoice.Add(v)
	k.keyBase.voice = v
}

type bowedKey struct {
	*keyBase
	moveLoc  geom.Point
	moveTime time.Time

	amp   float64
	voice *bowedTone
}

func (k *bowedKey) press(loc geom.Point) {
	k.moveLoc = loc
	k.moveTime = time.Now()

	if k.voice == nil || k.voice.Done() {
		k.voice = newBowedTone(math.Exp2(k.pitch))
		multivoice.Add(k.voice)
		k.keyBase.voice = k.voice
	}

	updateKeys(k.ratio)
}

func (k *bowedKey) move(loc geom.Point) {
	dx := dist(loc, k.moveLoc)
	dt := time.Now().Sub(k.moveTime).Seconds()
	amp := math.Max(-12, math.Log2(math.Tanh(dx/dt/128)))
	a := math.Pow(.999, 1/dt)
	k.amp = a*amp + (1-a)*k.amp
	k.voice.attack(k.amp)

	k.moveLoc = loc
	k.moveTime = time.Now()
}

func (k *bowedKey) release(loc geom.Point) {}

func dist(a, b geom.Point) float64 {
	return math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y)) / float64(geom.PixelsPerPt)
}

var (
	keys      []key
	lastPitch = 512.0

	program      gl.Program
	projection   gl.Uniform
	projmat      f32.Mat4
	position     gl.Attrib
	pointsize    gl.Attrib
	color        gl.Uniform
	positionbuf  gl.Buffer
	pointsizebuf gl.Buffer

	pitchRange  = 2.0
	pitchOffset = 6.0
)

func initKeys() {
	var err error
	program, err = glutil.CreateProgram(
		`#version 100
		uniform mat4 projection;
		attribute vec3 position;
		attribute float pointsize;
		void main() {
			gl_Position = projection * vec4(position, 1);
			gl_PointSize = pointsize;
		}`,
		`#version 100
		precision mediump float;
		uniform vec4 color;
		void main(void)
		{
		    vec2 v = 2.0*gl_PointCoord.xy - vec2(1.0);
			float r2 = dot(v, v);
			gl_FragColor = mix(color, vec4(0), r2);
		}`,
	)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	gl.Enable(34370) // GL_PROGRAM_POINT_SIZE; apparently not necessary on Android
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_COLOR)

	projection = gl.GetUniformLocation(program, "projection")
	updateProjectionMatrix()

	position = gl.GetAttribLocation(program, "position")
	pointsize = gl.GetAttribLocation(program, "pointsize")
	color = gl.GetUniformLocation(program, "color")
	positionbuf = gl.GenBuffer()
	pointsizebuf = gl.GenBuffer()
	updateKeys(ratio{1, 1})
}

func updateProjectionMatrix() {
	projmat.Identity()
	projmat.Translate(&projmat, -1, -1, 0)
	projmat.Scale(&projmat, 2/float32(pitchRange), 2, 1)
	projmat.Translate(&projmat, -float32(pitchOffset), 0, 0)
}

func updateKeys(last ratio) {
	lastPitch *= last.float()
	playing := []ratio{{1, 1}}
	for _, k := range keys {
		k := k.base()
		if k.voice != nil && !k.voice.Done() {
			playing = append(playing, k.ratio.div(last))
		}
	}

	oldkeys := keys
	keys = nil
	added := map[ratio]bool{}
	for _, r := range rats {
		for _, playing := range playing {
			r := r.mul(playing)
			if added[r] {
				continue
			}
			added[r] = true
			var k key
			if oldkeys, k = findAndRemoveKey(oldkeys, last.mul(r)); k != nil {
				k.base().ratio = r
				keys = append(keys, k)
				continue
			}
			kb := &keyBase{
				ratio: r,
				pitch: math.Log2(lastPitch * r.float()),
			}
			// k = &bowedKey{keyBase: kb, amp: -12}
			// k = &pluckedKey{keyBase: kb}
			k = &pressedKey{keyBase: kb}
			keys = append(keys, k)
		}
	}
	sort.Sort(byPitch(keys))
}

func drawKeys() {
	gl.UseProgram(program)
	projection.WriteMat4(&projmat)

	iPlaying := []int{}
	playing := []ratio{}
	amps := []float64{}
	for i, k := range keys {
		k := k.base()
		if k.voice != nil && !k.voice.Done() {
			iPlaying = append(iPlaying, i)
			playing = append(playing, k.ratio)
			amps = append(amps, k.voice.amp())
		}
	}
	complexities := make([]float64, len(keys))
	minComplexity := math.MaxFloat64
	for i, k := range keys {
		k := k.base()
		c := -1.0
		for j, iPlaying := range iPlaying {
			if i == iPlaying {
				a := amps[j]
				amps[j] = 1
				c = complexity(playing, amps)
				amps[j] = a
				break
			}
		}
		if c == -1 {
			c = complexity(append(playing, k.ratio), append(amps, 1))
		}
		complexities[i] = c
		if c < minComplexity {
			minComplexity = c
		}
	}
	data := []float32{}
	pointsizedata := []float32{}
	for i, k := range keys {
		k := k.base()
		k.y = 1 - math.Exp2(-float64(complexities[i]-minComplexity)/4)
		k.size = math.Exp2(-float64(complexities[i])/4) * float64(geom.Width) * float64(geom.PixelsPerPt) / 4
		data = append(data, float32(k.pitch), float32(k.y))
		pointsizedata = append(pointsizedata, float32(k.size))
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, positionbuf)
	gl.BufferData(gl.ARRAY_BUFFER, gl.DYNAMIC_DRAW, f32.Bytes(binary.LittleEndian, data...))
	gl.BindBuffer(gl.ARRAY_BUFFER, pointsizebuf)
	gl.BufferData(gl.ARRAY_BUFFER, gl.DYNAMIC_DRAW, f32.Bytes(binary.LittleEndian, pointsizedata...))

	gl.EnableVertexAttribArray(position)
	gl.EnableVertexAttribArray(pointsize)
	gl.Uniform4f(color, 1, 1, 1, 1)
	gl.BindBuffer(gl.ARRAY_BUFFER, positionbuf)
	gl.VertexAttribPointer(position, 2, gl.FLOAT, false, 0, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, pointsizebuf)
	gl.VertexAttribPointer(pointsize, 1, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.POINTS, 0, len(keys))
	gl.DisableVertexAttribArray(position)
	gl.DisableVertexAttribArray(pointsize)
}

func findAndRemoveKey(keys []key, r ratio) ([]key, key) {
	i := sort.Search(
		len(keys),
		func(i int) bool {
			return !keys[i].base().ratio.less(r)
		},
	)
	if i < len(keys) && keys[i].base().ratio == r {
		k := keys[i]
		return append(keys[:i], keys[i+1:]...), k
	}
	return keys, nil
}

type byPitch []key

func (s byPitch) Len() int           { return len(s) }
func (s byPitch) Less(i, j int) bool { return s[i].base().ratio.less(s[j].base().ratio) }
func (s byPitch) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
