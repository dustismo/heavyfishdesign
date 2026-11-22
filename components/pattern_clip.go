package components

import (
	"math"

	"github.com/tdewolff/canvas"
)

// clipCurvesToRect intersects p with an axis-aligned rectangle and keeps
// circular arcs as cubics. The boolean clip flattens those arcs to polygons.
func clipCurvesToRect(p *canvas.Path, x0, y0, x1, y1 float64) *canvas.Path {
	out := &canvas.Path{}
	if p == nil || p.Empty() || x1-x0 <= 1e-12 || y1-y0 <= 1e-12 {
		return out
	}
	b := p.Bounds()
	if b.X1 < x0 || b.X0 > x1 || b.Y1 < y0 || b.Y0 > y1 {
		return out
	}
	if b.X0 >= x0-1e-9 && b.Y0 >= y0-1e-9 && b.X1 <= x1+1e-9 && b.Y1 <= y1+1e-9 {
		return p.Copy()
	}

	curved := p.ReplaceArcs()
	var chains []clipChain
	s := curved.Scanner()
	for s.Scan() {
		end := s.End()
		switch s.Cmd() {
		case canvas.MoveToCmd:
			flushClipSubpath(out, p, chains, x0, y0, x1, y1)
			chains = nil
		case canvas.LineToCmd, canvas.CloseCmd:
			if a, b, ok := clipLineToRect(s.Start(), end, x0, y0, x1, y1); ok {
				chains = appendClipLine(chains, a, b)
			}
		case canvas.QuadToCmd:
			c1, c2 := quadToCubic(s.Start(), s.CP1(), end)
			chains = appendClipCubics(chains, clipCubic(cubic{s.Start(), c1, c2, end}, x0, y0, x1, y1, 0), x0, y0, x1, y1)
		case canvas.CubeToCmd:
			chains = appendClipCubics(chains, clipCubic(cubic{s.Start(), s.CP1(), s.CP2(), end}, x0, y0, x1, y1, 0), x0, y0, x1, y1)
		}
	}
	flushClipSubpath(out, p, chains, x0, y0, x1, y1)
	if out.Empty() && p.ContainsPoint((x0+x1)/2, (y0+y1)/2, canvas.NonZero) {
		out.MoveTo(x0, y0)
		out.LineTo(x1, y0)
		out.LineTo(x1, y1)
		out.LineTo(x0, y1)
		out.Close()
	}
	return out
}

type cubic [4]canvas.Point

type clipOp struct {
	cubic bool
	c1    canvas.Point
	c2    canvas.Point
	end   canvas.Point
}

type clipChain struct {
	start canvas.Point
	end   canvas.Point
	ops   []clipOp
}

func appendClipLine(chains []clipChain, a, b canvas.Point) []clipChain {
	if nearPoint(a, b) {
		return chains
	}
	if n := len(chains); n > 0 && nearPoint(chains[n-1].end, a) {
		chains[n-1].ops = append(chains[n-1].ops, clipOp{end: b})
		chains[n-1].end = b
		return chains
	}
	return append(chains, clipChain{
		start: a,
		end:   b,
		ops:   []clipOp{{end: b}},
	})
}

func appendClipCubics(chains []clipChain, parts []cubic, x0, y0, x1, y1 float64) []clipChain {
	for _, c := range parts {
		c[0] = snapRectPoint(c[0], x0, y0, x1, y1)
		c[3] = snapRectPoint(c[3], x0, y0, x1, y1)
		if nearPoint(c[0], c[3]) {
			continue
		}
		op := clipOp{cubic: true, c1: c[1], c2: c[2], end: c[3]}
		if n := len(chains); n > 0 && nearPoint(chains[n-1].end, c[0]) {
			chains[n-1].ops = append(chains[n-1].ops, op)
			chains[n-1].end = c[3]
			continue
		}
		chains = append(chains, clipChain{start: c[0], end: c[3], ops: []clipOp{op}})
	}
	return chains
}

func flushClipSubpath(out *canvas.Path, shape *canvas.Path, chains []clipChain, x0, y0, x1, y1 float64) {
	if len(chains) == 0 {
		return
	}
	// A loop that never left the rectangle comes back to its first point.
	closed := len(chains) == 1 && nearPoint(chains[0].start, chains[0].end)
	out.MoveTo(chains[0].start.X, chains[0].start.Y)
	for i := range chains {
		for _, op := range chains[i].ops {
			if op.cubic {
				out.CubeTo(op.c1.X, op.c1.Y, op.c2.X, op.c2.Y, op.end.X, op.end.Y)
			} else {
				out.LineTo(op.end.X, op.end.Y)
			}
		}
		if closed {
			break
		}
		next := chains[(i+1)%len(chains)]
		if nearPoint(chains[i].end, next.start) {
			continue
		}
		for _, pt := range boundaryLink(chains[i].end, next.start, x0, y0, x1, y1, shape) {
			out.LineTo(pt.X, pt.Y)
		}
	}
	out.Close()
}

func boundaryLink(from, to canvas.Point, x0, y0, x1, y1 float64, shape *canvas.Path) []canvas.Point {
	cw := walkRect(from, to, x0, y0, x1, y1, true)
	ccw := walkRect(from, to, x0, y0, x1, y1, false)
	cwIn := linkInside(from, cw, shape)
	ccwIn := linkInside(from, ccw, shape)
	if cwIn && (!ccwIn || len(cw) <= len(ccw)) {
		return cw
	}
	if ccwIn {
		return ccw
	}
	if len(cw) <= len(ccw) {
		return cw
	}
	return ccw
}

func linkInside(from canvas.Point, pts []canvas.Point, shape *canvas.Path) bool {
	cur := from
	for _, p := range pts {
		mid := canvas.Point{(cur.X + p.X) / 2, (cur.Y + p.Y) / 2}
		if shape == nil || !shape.ContainsPoint(mid.X, mid.Y, canvas.NonZero) {
			return false
		}
		cur = p
	}
	return true
}

func walkRect(from, to canvas.Point, x0, y0, x1, y1 float64, cw bool) []canvas.Point {
	ta := rectParam(from, x0, y0, x1, y1)
	tb := rectParam(to, x0, y0, x1, y1)
	type corner struct {
		t float64
		p canvas.Point
	}
	var mids []corner
	if cw {
		if tb < ta-1e-9 {
			tb += 4
		}
		for k := 1; k <= 4; k++ {
			t := float64(k)
			if t > ta+1e-8 && t < tb-1e-8 {
				mids = append(mids, corner{t, rectCorner(float64(k%4), x0, y0, x1, y1)})
			}
		}
	} else {
		if tb > ta+1e-9 {
			tb -= 4
		}
		for k := 0; k < 4; k++ {
			t := float64(k)
			if t > ta {
				t -= 4
			}
			if t > tb+1e-8 && t < ta-1e-8 {
				mids = append(mids, corner{t, rectCorner(float64(k), x0, y0, x1, y1)})
			}
		}
		for i := 1; i < len(mids); i++ {
			j := i
			for j > 0 && mids[j].t > mids[j-1].t {
				mids[j], mids[j-1] = mids[j-1], mids[j]
				j--
			}
		}
	}
	pts := make([]canvas.Point, 0, len(mids)+1)
	for _, c := range mids {
		pts = append(pts, c.p)
	}
	return append(pts, to)
}

func rectParam(p canvas.Point, x0, y0, x1, y1 float64) float64 {
	const e = 1e-6
	w := math.Max(1e-12, x1-x0)
	h := math.Max(1e-12, y1-y0)
	switch {
	case math.Abs(p.X-x0) < e && math.Abs(p.Y-y0) < e:
		return 0
	case math.Abs(p.X-x1) < e && math.Abs(p.Y-y0) < e:
		return 1
	case math.Abs(p.X-x1) < e && math.Abs(p.Y-y1) < e:
		return 2
	case math.Abs(p.X-x0) < e && math.Abs(p.Y-y1) < e:
		return 3
	case math.Abs(p.Y-y0) < e:
		return (p.X - x0) / w
	case math.Abs(p.X-x1) < e:
		return 1 + (p.Y-y0)/h
	case math.Abs(p.Y-y1) < e:
		return 2 + (x1-p.X)/w
	case math.Abs(p.X-x0) < e:
		return 3 + (y1-p.Y)/h
	default:
		return 0
	}
}

func rectCorner(t, x0, y0, x1, y1 float64) canvas.Point {
	switch int(math.Round(t)) % 4 {
	case 1:
		return canvas.Point{x1, y0}
	case 2:
		return canvas.Point{x1, y1}
	case 3:
		return canvas.Point{x0, y1}
	default:
		return canvas.Point{x0, y0}
	}
}

func snapRectPoint(p canvas.Point, x0, y0, x1, y1 float64) canvas.Point {
	const e = 1e-5
	if math.Abs(p.X-x0) <= e {
		p.X = x0
	}
	if math.Abs(p.X-x1) <= e {
		p.X = x1
	}
	if math.Abs(p.Y-y0) <= e {
		p.Y = y0
	}
	if math.Abs(p.Y-y1) <= e {
		p.Y = y1
	}
	return p
}

func clipLineToRect(a, b canvas.Point, x0, y0, x1, y1 float64) (canvas.Point, canvas.Point, bool) {
	dx := b.X - a.X
	dy := b.Y - a.Y
	u0, u1 := 0.0, 1.0
	clipT := func(p, q float64) bool {
		if math.Abs(p) < 1e-15 {
			return q >= -1e-12
		}
		t := q / p
		if p < 0 {
			if t > u1 {
				return false
			}
			if t > u0 {
				u0 = t
			}
		} else {
			if t < u0 {
				return false
			}
			if t < u1 {
				u1 = t
			}
		}
		return true
	}
	if !clipT(-dx, a.X-x0) || !clipT(dx, x1-a.X) || !clipT(-dy, a.Y-y0) || !clipT(dy, y1-a.Y) || u1 < u0 {
		return canvas.Point{}, canvas.Point{}, false
	}
	c := snapRectPoint(lerpPoint(a, b, u0), x0, y0, x1, y1)
	d := snapRectPoint(lerpPoint(a, b, u1), x0, y0, x1, y1)
	return c, d, true
}

func clipCubic(c cubic, x0, y0, x1, y1 float64, depth int) []cubic {
	minx, miny, maxx, maxy := cubicBox(c)
	if minx >= x0-1e-9 && miny >= y0-1e-9 && maxx <= x1+1e-9 && maxy <= y1+1e-9 {
		return []cubic{c}
	}
	if maxx < x0 || maxy < y0 || minx > x1 || miny > y1 {
		return nil
	}
	if depth >= 10 {
		return trimStraddle(c, x0, y0, x1, y1)
	}
	left, right := splitCubic(c, 0.5)
	a := clipCubic(left, x0, y0, x1, y1, depth+1)
	b := clipCubic(right, x0, y0, x1, y1, depth+1)
	if len(a) == 1 && cubicEqual(a[0], left) && len(b) == 1 && cubicEqual(b[0], right) {
		return []cubic{c}
	}
	out := make([]cubic, 0, len(a)+len(b))
	return append(append(out, a...), b...)
}

func trimStraddle(c cubic, x0, y0, x1, y1 float64) []cubic {
	in0 := inRect(c[0], x0, y0, x1, y1)
	in3 := inRect(c[3], x0, y0, x1, y1)
	if in0 && in3 {
		return []cubic{c}
	}
	if !in0 && !in3 {
		return nil
	}
	lo, hi := 0.0, 1.0
	for i := 0; i < 24; i++ {
		mid := (lo + hi) / 2
		if inRect(curvePoint(c, mid), x0, y0, x1, y1) == in0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	left, right := splitCubic(c, (lo+hi)/2)
	if in0 {
		left[3] = snapRectPoint(left[3], x0, y0, x1, y1)
		return []cubic{left}
	}
	right[0] = snapRectPoint(right[0], x0, y0, x1, y1)
	return []cubic{right}
}

func splitCubic(c cubic, t float64) (cubic, cubic) {
	p01 := lerpPoint(c[0], c[1], t)
	p12 := lerpPoint(c[1], c[2], t)
	p23 := lerpPoint(c[2], c[3], t)
	p012 := lerpPoint(p01, p12, t)
	p123 := lerpPoint(p12, p23, t)
	mid := lerpPoint(p012, p123, t)
	return cubic{c[0], p01, p012, mid}, cubic{mid, p123, p23, c[3]}
}

func curvePoint(c cubic, t float64) canvas.Point {
	u := 1 - t
	return canvas.Point{
		u*u*u*c[0].X + 3*u*u*t*c[1].X + 3*u*t*t*c[2].X + t*t*t*c[3].X,
		u*u*u*c[0].Y + 3*u*u*t*c[1].Y + 3*u*t*t*c[2].Y + t*t*t*c[3].Y,
	}
}

func cubicBox(c cubic) (minx, miny, maxx, maxy float64) {
	minx, miny = c[0].X, c[0].Y
	maxx, maxy = minx, miny
	for _, p := range c[1:] {
		minx, maxx = math.Min(minx, p.X), math.Max(maxx, p.X)
		miny, maxy = math.Min(miny, p.Y), math.Max(maxy, p.Y)
	}
	return minx, miny, maxx, maxy
}

func cubicEqual(a, b cubic) bool {
	for i := range a {
		if !nearPoint(a[i], b[i]) {
			return false
		}
	}
	return true
}

func quadToCubic(p0, p1, p2 canvas.Point) (canvas.Point, canvas.Point) {
	return lerpPoint(p0, p1, 2.0/3.0), lerpPoint(p2, p1, 2.0/3.0)
}

func lerpPoint(a, b canvas.Point, t float64) canvas.Point {
	return canvas.Point{a.X + (b.X-a.X)*t, a.Y + (b.Y-a.Y)*t}
}

func inRect(p canvas.Point, x0, y0, x1, y1 float64) bool {
	const e = 1e-8
	return p.X >= x0-e && p.X <= x1+e && p.Y >= y0-e && p.Y <= y1+e
}

func nearPoint(a, b canvas.Point) bool {
	return math.Hypot(a.X-b.X, a.Y-b.Y) <= 1e-5
}
