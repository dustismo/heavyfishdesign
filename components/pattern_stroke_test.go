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

func TestVoronoiRectangleIsACross(t *testing.T) {
	edges := voronoiEdges([]point{{0, 0}, {2, 0}, {2, 2}, {0, 2}}, 0, 0, 2, 2)
	if len(edges) != 4 {
		t.Fatalf("rectangle voronoi has %d edges, want 4", len(edges))
	}
	var hits [4]bool
	for _, e := range edges {
		ax, ay, bx, by := e.ax, e.ay, e.bx, e.by
		if math.Hypot(ax-1, ay-1) > 1e-6 && math.Hypot(bx-1, by-1) > 1e-6 {
			t.Fatalf("edge %v does not meet the center", e)
		}
		end := [2]float64{ax, ay}
		if math.Hypot(ax-1, ay-1) <= 1e-6 {
			end = [2]float64{bx, by}
		}
		switch {
		case math.Abs(end[0]-1) < 1e-6 && math.Abs(end[1]-0) < 1e-6:
			hits[0] = true
		case math.Abs(end[0]-2) < 1e-6 && math.Abs(end[1]-1) < 1e-6:
			hits[1] = true
		case math.Abs(end[0]-1) < 1e-6 && math.Abs(end[1]-2) < 1e-6:
			hits[2] = true
		case math.Abs(end[0]-0) < 1e-6 && math.Abs(end[1]-1) < 1e-6:
			hits[3] = true
		default:
			t.Fatalf("edge ends at %v, want a side midpoint", end)
		}
	}
	for i, hit := range hits {
		if !hit {
			t.Fatalf("missing arm %d", i)
		}
	}
}

func TestVoronoiCrossJoinsAtCenter(t *testing.T) {
	edges := voronoiEdges([]point{{0, 0}, {2, 0}, {2, 2}, {0, 2}}, 0, 0, 2, 2)
	got := strokeOpenSegments(edges, 0.1, 0, 2, 2, false)
	moves := 0
	for _, seg := range got.Segments() {
		if _, ok := seg.(path.MoveSegment); ok {
			moves++
		}
	}
	if moves != 1 {
		t.Fatalf("cross stroke has %d loops, want 1 joined slot\n%s", moves, path.SvgString(got, 3))
	}
}

func TestVoronoiInteriorVerticesHaveDegreeThree(t *testing.T) {
	const xmin, ymin, xmax, ymax = 0.2, 0.2, 3.8, 2.8
	edges := voronoiEdges(voronoiSites(xmin, ymin, xmax, ymax, voronoiParams{
		siteCount: 28, seed: 7, minSiteGap: 0.22,
	}), xmin, ymin, xmax, ymax)
	key := func(x, y float64) [2]int {
		return [2]int{int(math.Round(x * 1e5)), int(math.Round(y * 1e5))}
	}
	deg := map[[2]int]int{}
	at := map[[2]int][2]float64{}
	for _, e := range edges {
		for _, p := range [][2]float64{{e.ax, e.ay}, {e.bx, e.by}} {
			k := key(p[0], p[1])
			deg[k]++
			at[k] = p
		}
	}
	onBorder := func(x, y float64) bool {
		const tol = 1e-4
		return math.Abs(x-xmin) <= tol || math.Abs(x-xmax) <= tol || math.Abs(y-ymin) <= tol || math.Abs(y-ymax) <= tol
	}
	lonely := 0
	for k, n := range deg {
		if onBorder(at[k][0], at[k][1]) {
			continue
		}
		if n < 3 {
			lonely++
			t.Logf("interior vertex %v degree %d", at[k], n)
		}
	}
	if lonely > 0 {
		t.Fatalf("%d interior vertices are not joined", lonely)
	}
}

func TestVoronoiBorderClosesThePanel(t *testing.T) {
	const m, w, h = 0.2, 4.0, 3.0
	got := voronoiSlots(w, h, m, voronoiParams{siteCount: 12, seed: 7, thickness: 0.1, minSiteGap: 0.3})
	// The band's outer edge is the margin rectangle, which the stroke path
	// drops. The border that remains is the band's inner edge.
	hw := 0.05
	top, bottom := m+hw, h-m-hw
	left, right := m+hw, w-m-hw
	var length [4]float64
	var prev path.Point
	has := false
	for _, seg := range got.Segments() {
		end := seg.End()
		if _, ok := seg.(path.MoveSegment); ok {
			prev = end
			has = true
			continue
		}
		if !has {
			prev = end
			continue
		}
		dx, dy := math.Abs(end.X-prev.X), math.Abs(end.Y-prev.Y)
		span := math.Hypot(end.X-prev.X, end.Y-prev.Y)
		if dy < 0.01 && dx > 0.02 {
			if math.Abs(end.Y-top) < 0.02 {
				length[0] += span
			}
			if math.Abs(end.Y-bottom) < 0.02 {
				length[1] += span
			}
		}
		if dx < 0.01 && dy > 0.02 {
			if math.Abs(end.X-left) < 0.02 {
				length[2] += span
			}
			if math.Abs(end.X-right) < 0.02 {
				length[3] += span
			}
		}
		prev = end
	}
	innerW, innerH := w-2*m-2*hw, h-2*m-2*hw
	want := [4]float64{innerW * 0.6, innerW * 0.6, innerH * 0.6, innerH * 0.6}
	names := [4]string{"top", "bottom", "left", "right"}
	for i, gotLen := range length {
		if gotLen < want[i] {
			t.Fatalf("%s border length %.3f, want at least %.3f", names[i], gotLen, want[i])
		}
	}
}

func TestVoronoiSitesKeepMinGap(t *testing.T) {
	sites := voronoiSites(0.2, 0.2, 3.8, 2.8, voronoiParams{
		siteCount: 20, seed: 7, minSiteGap: 0.45,
	})
	if len(sites) < 4 {
		t.Fatalf("placed %d sites", len(sites))
	}
	for i := 0; i < len(sites); i++ {
		for j := i + 1; j < len(sites); j++ {
			d := math.Hypot(sites[i].x-sites[j].x, sites[i].y-sites[j].y)
			if d < 0.45-1e-9 {
				t.Fatalf("sites %d and %d are %.4f apart", i, j, d)
			}
		}
	}
}

func TestVoronoiSlotsStableAndInside(t *testing.T) {
	p := voronoiParams{siteCount: 24, seed: 7, thickness: 0.08, minSiteGap: 0.2}
	a := path.SvgString(voronoiSlots(4, 3, 0.2, p), 4)
	b := path.SvgString(voronoiSlots(4, 3, 0.2, p), 4)
	if a != b {
		t.Fatal("voronoi changed between renders with the same seed")
	}
	center := voronoiSlots(4, 3, 0.2, voronoiParams{siteCount: 24, seed: 7, thickness: 0, minSiteGap: 0.2})
	thick := voronoiSlots(4, 3, 0.2, p)
	if len(thick.Segments()) <= len(center.Segments()) {
		t.Fatalf("thickened voronoi has %d segments, centerlines %d", len(thick.Segments()), len(center.Segments()))
	}
	for _, seg := range thick.Segments() {
		end := seg.End()
		if end.X < 0.2-0.02 || end.X > 3.8+0.02 || end.Y < 0.2-0.02 || end.Y > 2.8+0.02 {
			t.Fatalf("slot endpoint (%v, %v) left the margin", end.X, end.Y)
		}
	}
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
