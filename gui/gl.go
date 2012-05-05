package gui

import ."github.com/chsc/gogl/gl21"

type Color struct {
	Red float32
	Green float32
	Blue float32
	Alpha float32
}

func SetColor(color Color) { Color4d(Double(color.Red), Double(color.Green), Double(color.Blue), Double(color.Alpha)) }

func DrawPoint(p Point) {
	Begin(POINTS); defer End()
	Vertex2d(Double(p.X), Double(p.Y))
}

func DrawLine(p1, p2 Point) {
	Begin(LINES); defer End()
	Vertex2d(Double(p1.X), Double(p1.Y))
	Vertex2d(Double(p2.X), Double(p2.Y))
}

func DrawRect(r Rectangle) {
	Begin(LINE_LOOP); defer End()
	Vertex2d(Double(r.Min.X), Double(r.Min.Y))
	Vertex2d(Double(r.Min.X), Double(r.Max.Y))
	Vertex2d(Double(r.Max.X), Double(r.Max.Y))
	Vertex2d(Double(r.Max.X), Double(r.Min.Y))
}

func FillRect(r Rectangle) {
	Rectd(Double(r.Min.X), Double(r.Min.Y), Double(r.Max.X), Double(r.Max.Y))
}
