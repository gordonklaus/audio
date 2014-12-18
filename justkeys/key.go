package main

import (
	"encoding/binary"
	"math"
	"math/big"
	"sort"

	"golang.org/x/mobile/f32"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl"
)

type key struct {
	ratio      *big.Rat
	pitch      float64
	complexity int
	y, yTarget float64
	sizeTarget float64
	size       float64
}

var (
	position     gl.Attrib
	pointsize    gl.Attrib
	color        gl.Uniform
	positionbuf  gl.Buffer
	pointsizebuf gl.Buffer
)

func initKeys() {
	position = gl.GetAttribLocation(program, "position")
	pointsize = gl.GetAttribLocation(program, "pointsize")
	color = gl.GetUniformLocation(program, "color")
	positionbuf = gl.GenBuffer()
	pointsizebuf = gl.GenBuffer()
	updateKeys(big.NewRat(1, 1))
}

func updateKeys(last *big.Rat) {
	oldkeys := keys
	keys = nil

	rats := rats()
	cur, _ := mel.current.Float64()
	complexities := make([]int, len(rats))
	minComplexity := math.MaxInt32
	for i, r := range rats {
		c := complexity(appendRatio(mel.history, r))
		complexities[i] = c
		if c < minComplexity {
			minComplexity = c
		}
	}
	for i, r := range rats {
		f, _ := r.Float64()
		k := &key{
			ratio:      r,
			pitch:      math.Log2(cur * f),
			yTarget:    1 - math.Exp2(-float64(complexities[i]-minComplexity)/4),
			sizeTarget: math.Exp2(-float64(complexities[i]) / 4),
		}
		k.y = k.yTarget
		k.size = k.sizeTarget
		if k2 := findKey(oldkeys, new(big.Rat).Mul(r, last)); k2 != nil {
			k.y = k2.y
			k.size = k2.size
		}
		keys = append(keys, k)
	}
	sort.Sort(byRatio(keys))
}

func drawKeys() {
	data := []float32{}
	pointsizedata := []float32{}
	for _, k := range keys {
		k.y = k.yTarget + (k.y-k.yTarget)*.95
		k.size = k.sizeTarget + (k.size-k.sizeTarget)*.95
		data = append(data, float32(k.pitch), float32(k.y))
		pointsizedata = append(pointsizedata, float32(k.size)*float32(geom.Height)*geom.PixelsPerPt/2)
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

func findKey(keys []*key, r *big.Rat) *key {
	i := sort.Search(
		len(keys),
		func(i int) bool {
			return keys[i].ratio.Cmp(r) >= 0
		},
	)
	if i < len(keys) && keys[i].ratio.Cmp(r) == 0 {
		return keys[i]
	}
	return nil
}

type byRatio []*key

func (s byRatio) Len() int           { return len(s) }
func (s byRatio) Less(i, j int) bool { return s[i].ratio.Cmp(s[j].ratio) < 0 }
func (s byRatio) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
