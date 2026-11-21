package components

import (
	"math"
	"testing"

	"github.com/tdewolff/canvas"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestExtendCenterlinePastMargin(t *testing.T) {
	d := path.NewDraw()
	d.MoveTo(path.NewPoint(0.25, 1))
	d.LineTo(path.NewPoint(2, 1.5))
	out := extendCenterlinePastMargin(d.Path(), 0.25, 0.25, 3.75, 1.75, 0.1)
	start := out.Segments()[0].End()
	if start.X >= 0.25-1e-4 {
		t.Fatalf("start stayed on the border at %v\n%s", start, path.SvgString(out, 3))
	}
	if start.X > 0.25-0.1+1e-3 {
		t.Fatalf("start did not clear the band, at %v", start)
	}
}

func TestRingBoundaryMatchesBoolean(t *testing.T) {
	cases := []struct {
		name  string
		rings []annulus
	}{
		{"one", []annulus{{cx: 2, cy: 1.5, inner: 0.3, outer: 0.45}}},
		{"overlap", []annulus{
			{cx: 1.6, cy: 1.4, inner: 0.35, outer: 0.5},
			{cx: 2.2, cy: 1.6, inner: 0.3, outer: 0.48},
		}},
		{"crosses edge", []annulus{{cx: 0.3, cy: 1.2, inner: 0.2, outer: 0.4}}},
		{"solid", []annulus{{cx: 2, cy: 1.2, inner: 0, outer: 0.4}}},
		{"three", []annulus{
			{cx: 1.5, cy: 1.2, inner: 0.25, outer: 0.4},
			{cx: 2.0, cy: 1.5, inner: 0.2, outer: 0.38},
			{cx: 2.4, cy: 1.1, inner: 0.22, outer: 0.42},
		}},
		{"design", overlapAnnuli(4, 3, 0.2, overlapParams{
			radiusMin: 0.22, radiusMax: 0.48, strokeWidth: 0.07, count: 12, seed: 17,
		})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := path.ToCanvasPath(ringBoundary(tc.rings, 0.2, 0.2, 3.8, 2.8))
			want := booleanAnnuli(tc.rings, 0.2, 0.2, 3.8, 2.8)
			if d := maxBoundaryDistance(got, want); d > 0.008 {
				t.Fatalf("outline differs from the union by %.4f in", d)
			}
		})
	}
}

func booleanAnnuli(rings []annulus, xmin, ymin, xmax, ymax float64) *canvas.Path {
	old := canvas.Tolerance
	canvas.Tolerance = 0.0002
	defer func() { canvas.Tolerance = old }()
	panel := &canvas.Path{}
	panel.MoveTo(xmin, ymin)
	panel.LineTo(xmax, ymin)
	panel.LineTo(xmax, ymax)
	panel.LineTo(xmin, ymax)
	panel.Close()
	var acc *canvas.Path
	for _, g := range rings {
		ring := circleCanvas(g.cx, g.cy, g.outer)
		if g.inner > 1e-8 {
			ring = ring.Not(circleCanvas(g.cx, g.cy, g.inner))
		}
		if acc == nil {
			acc = ring
			continue
		}
		acc = acc.Or(ring)
	}
	if acc == nil {
		return panel
	}
	return panel.Not(acc)
}

func maxBoundaryDistance(a, b *canvas.Path) float64 {
	fa := a.Flatten(0.0005)
	fb := b.Flatten(0.0005)
	return math.Max(pathToPathDistance(fa, fb), pathToPathDistance(fb, fa))
}

func pathToPathDistance(from, to *canvas.Path) float64 {
	segs := flatSegs(to)
	worst := 0.0
	s := from.Scanner()
	var start canvas.Point
	var cur canvas.Point
	has := false
	for s.Scan() {
		end := s.End()
		switch s.Cmd() {
		case canvas.MoveToCmd:
			start = end
			cur = end
			has = true
		case canvas.LineToCmd:
			worst = math.Max(worst, distToSegments((cur.X+end.X)/2, (cur.Y+end.Y)/2, segs))
			cur = end
		case canvas.CloseCmd:
			worst = math.Max(worst, distToSegments((cur.X+start.X)/2, (cur.Y+start.Y)/2, segs))
			cur = start
		default:
			if !has {
				continue
			}
			worst = math.Max(worst, distToSegments(end.X, end.Y, segs))
			cur = end
		}
	}
	return worst
}

type flatSeg struct{ x0, y0, x1, y1 float64 }

func flatSegs(p *canvas.Path) []flatSeg {
	var out []flatSeg
	s := p.Scanner()
	var start, cur canvas.Point
	for s.Scan() {
		end := s.End()
		switch s.Cmd() {
		case canvas.MoveToCmd:
			start = end
			cur = end
		case canvas.LineToCmd:
			out = append(out, flatSeg{cur.X, cur.Y, end.X, end.Y})
			cur = end
		case canvas.CloseCmd:
			out = append(out, flatSeg{cur.X, cur.Y, start.X, start.Y})
			cur = start
		default:
			out = append(out, flatSeg{cur.X, cur.Y, end.X, end.Y})
			cur = end
		}
	}
	return out
}

func distToSegments(x, y float64, segs []flatSeg) float64 {
	best := math.MaxFloat64
	for _, s := range segs {
		dx, dy := s.x1-s.x0, s.y1-s.y0
		l2 := dx*dx + dy*dy
		t := 0.0
		if l2 > 1e-18 {
			t = ((x-s.x0)*dx + (y-s.y0)*dy) / l2
			if t < 0 {
				t = 0
			} else if t > 1 {
				t = 1
			}
		}
		px, py := s.x0+t*dx, s.y0+t*dy
		if d := math.Hypot(x-px, y-py); d < best {
			best = d
		}
	}
	return best
}

func TestMulberry32MatchesStudio(t *testing.T) {
	rng := mulberry32(17)
	want := []float64{
		0.6771502960473299,
		0.19265692122280598,
		0.5313839064911008,
		0.1654723146930337,
		0.08778517786413431,
	}
	for i, w := range want {
		got := rng()
		if math.Abs(got-w) > 1e-12 {
			t.Fatalf("sample %d = %.16f, want %.16f", i, got, w)
		}
	}
}
