package gui

import . "github.com/chsc/gogl/gl21"

type Color struct{ R, G, B, A float64 }

func SetColor(c Color)       { Color4d(Double(c.R), Double(c.G), Double(c.B), Double(c.A)) }
func SetPointSize(x float64) { PointSize(Float(x)) }
func SetLineWidth(x float64) { LineWidth(Float(x)) }

func DrawPoint(p Point) {
	Begin(POINTS)
	defer End()
	Vertex2d(Double(p.X), Double(p.Y))
}

func DrawLine(p1, p2 Point) {
	Begin(LINES)
	defer End()
	Vertex2d(Double(p1.X), Double(p1.Y))
	Vertex2d(Double(p2.X), Double(p2.Y))
}

func DrawRect(r Rectangle) {
	p1, p2, p3, p4 := r.Min, Pt(r.Max.X, r.Min.Y), r.Max, Pt(r.Min.X, r.Max.Y)
	DrawLine(p1, p2)
	DrawLine(p2, p3)
	DrawLine(p3, p4)
	DrawLine(p4, p1)
}

func FillRect(r Rectangle) {
	Rectd(Double(r.Min.X), Double(r.Min.Y), Double(r.Max.X), Double(r.Max.Y))
}

func DrawPolygon(pts ...Point) {
	Begin(LINE_LOOP)
	defer End()
	for _, p := range pts {
		Vertex2d(Double(p.X), Double(p.Y))
	}
}

func FillPolygon(pts ...Point) {
	Begin(POLYGON)
	defer End()
	for _, p := range pts {
		Vertex2d(Double(p.X), Double(p.Y))
	}
}

func DrawBezier(ctrlPts ...Point) {
	pts := []Double{}
	steps := 0.0
	for i, p := range ctrlPts {
		pts = append(pts, Double(p.X), Double(p.Y), 0)
		if i > 0 {
			steps += ctrlPts[i].Sub(ctrlPts[i-1]).Len()
		}
	}
	Map1d(MAP1_VERTEX_3, 0, 1, 3, Int(len(ctrlPts)), &pts[0])
	Enable(MAP1_VERTEX_3)
	defer Disable(MAP1_VERTEX_3)
	MapGrid1d(Int(steps), 0, 1)
	EvalMesh1(LINE, 0, Int(steps))
}

func Rotate(rot float64) {
	Rotated(Double(rot*360), 0, 0, 1)
}
