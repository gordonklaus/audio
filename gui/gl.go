package gui

import ."github.com/chsc/gogl/gl21"

type Color struct {
	Red float64
	Green float64
	Blue float64
	Alpha float64
}

func SetColor(color Color) { Color4d(Double(color.Red), Double(color.Green), Double(color.Blue), Double(color.Alpha)) }
func SetPointSize(size float64) { PointSize(Float(size)) }
func SetLineWidth(size float64) { LineWidth(Float(size)) }

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

func DrawPolygon(points ...Point) {
	Begin(LINE_LOOP); defer End()
	for _, p := range points {
		Vertex2d(Double(p.X), Double(p.Y))
	}
}

func FillPolygon(points ...Point) {
	Begin(POLYGON); defer End()
	for _, p := range points {
		Vertex2d(Double(p.X), Double(p.Y))
	}
}

func DrawQuadratic(ctrlPoints [3]Point, steps int) {
	p := [9]Double{Double(ctrlPoints[0].X), Double(ctrlPoints[0].Y), 0, Double(ctrlPoints[1].X), Double(ctrlPoints[1].Y), 0, Double(ctrlPoints[2].X), Double(ctrlPoints[2].Y), 0}
	Map1d(MAP1_VERTEX_3, 0, 1, 3, 3, &p[0])
	Enable(MAP1_VERTEX_3); defer Disable(MAP1_VERTEX_3)
	MapGrid1d(Int(steps), 0, 1)
	EvalMesh1(LINE, 0, Int(steps))
}

func DrawCubic(ctrlPoints [4]Point, steps int) {
	p := [12]Double{Double(ctrlPoints[0].X), Double(ctrlPoints[0].Y), 0, Double(ctrlPoints[1].X), Double(ctrlPoints[1].Y), 0, Double(ctrlPoints[2].X), Double(ctrlPoints[2].Y), 0, Double(ctrlPoints[3].X), Double(ctrlPoints[3].Y), 0}
	Map1d(MAP1_VERTEX_3, 0, 1, 3, 4, &p[0])
	Enable(MAP1_VERTEX_3); defer Disable(MAP1_VERTEX_3)
	MapGrid1d(Int(steps), 0, 1)
	EvalMesh1(LINE, 0, Int(steps))
}

func DrawBezier(ctrlPoints []Point, steps int) {
	pts := []Double{}
	for _, p := range ctrlPoints {
		pts = append(pts, Double(p.X), Double(p.Y), 0)
	}
	Map1d(MAP1_VERTEX_3, 0, 1, 3, Int(len(ctrlPoints)), &pts[0])
	Enable(MAP1_VERTEX_3); defer Disable(MAP1_VERTEX_3)
	MapGrid1d(Int(steps), 0, 1)
	EvalMesh1(LINE, 0, Int(steps))
}
