package components

import (
	"math"
	"sort"
	"strings"

	"github.com/dustismo/heavyfishdesign/path"
	"github.com/tdewolff/canvas"
)

// Hatch, wave, and lattice kinds emit centerlines. framed_stroke thickens
// them. Overlapping circles stay a panel minus the stroked rings. Voronoi
// emits the thickened cell edges itself: three edges meet at a vertex, so
// each edge is stroked and the strokes are unioned.

type hatchParams struct {
	angleDeg    float64
	spacing     float64
	phase       float64
	cross       bool
	crossOffset float64
}

type waveParams struct {
	amplitude  float64
	wavelength float64
	thickness  float64
	gap        float64
	phaseStep  float64
	samples    int
}

type latticeParams struct {
	rectWidth  float64
	rectHeight float64
	vGapRatio  float64
	hGapRatio  float64
}

type overlapParams struct {
	radiusMin   float64
	radiusMax   float64
	strokeWidth float64
	count       int
	seed        int32
}

type voronoiParams struct {
	siteCount  int
	seed       int32
	thickness  float64
	minSiteGap float64
}

func hatchPath(w, h, m float64, p hatchParams) path.Path {
	d := path.NewDraw()
	sp := math.Max(1e-6, p.spacing)
	xmin, ymin := m, m
	xmax, ymax := w-m, h-m
	if xmax <= xmin || ymax <= ymin {
		return d.Path()
	}
	cx := (xmin + xmax) / 2
	cy := (ymin + ymax) / 2
	reach := 0.5*math.Hypot(xmax-xmin, ymax-ymin) + math.Abs(p.phase)
	kMax := int(math.Ceil(reach / sp))

	addFamily := func(deg float64) {
		rad := deg * math.Pi / 180
		vx, vy := math.Cos(rad), math.Sin(rad)
		nx, ny := -math.Sin(rad), math.Cos(rad)
		for k := -kMax; k <= kMax; k++ {
			dist := float64(k)*sp + p.phase
			ax, ay, bx, by, ok := clipInfiniteLine(cx+dist*nx, cy+dist*ny, vx, vy, xmin, ymin, xmax, ymax)
			if !ok {
				continue
			}
			d.MoveTo(path.NewPoint(ax, ay))
			d.LineTo(path.NewPoint(bx, by))
		}
	}
	addFamily(p.angleDeg)
	if p.cross {
		off := p.crossOffset
		if math.IsNaN(off) || math.IsInf(off, 0) {
			off = 90
		}
		addFamily(p.angleDeg + off)
	}
	return d.Path()
}

func waveLattice(w, h, m float64, p waveParams, clip bool) path.Path {
	d := path.NewDraw()
	amp := math.Max(0, p.amplitude)
	wl := math.Max(1e-6, p.wavelength)
	thick := math.Max(0.01, p.thickness)
	gap := p.gap
	if math.IsNaN(gap) || math.IsInf(gap, 0) {
		gap = 0
	}
	pitch := math.Max(1e-6, thick+gap)
	samples := p.samples
	if samples < 16 {
		samples = 16
	}
	xmin, ymin := m, m
	xmax, ymax := w-m, h-m
	if xmax <= xmin || ymax <= ymin {
		return d.Path()
	}
	// The crest of the top ribbon and the trough of the bottom one sit half a
	// stroke outside the margin. The stroke clip then cuts them on that edge,
	// so the peaks open through the top and bottom instead of stopping short.
	past := thick / 2
	innerH := ymax - ymin
	need := innerH + 2*past - 2*amp
	intervals := 0
	pitchUsed := pitch
	if need > 1e-9 && pitch > 1e-9 {
		intervals = int(math.Round(need / pitch))
		if intervals < 1 {
			intervals = 1
		}
		pitchUsed = need / float64(intervals)
	}
	yTop := (ymin + ymax) / 2
	if intervals > 0 {
		yTop = ymin - past + amp
	}
	for i := 0; i <= intervals; i++ {
		yCL := yTop + float64(i)*pitchUsed
		appendWaveCenterline(d, xmin, xmax, yCL, amp, wl, float64(i)*p.phaseStep, samples, clip, xmin, ymin, xmax, ymax)
	}
	return d.Path()
}

func appendWaveCenterline(d *path.Draw, x0, x1, yCenter, amp, wavelength, phase float64, samples int, clip bool, rx0, ry0, rx1, ry1 float64) {
	if x1 <= x0 {
		return
	}
	n := samples
	if n < 8 {
		n = 8
	}
	span := x1 - x0
	k := 2 * math.Pi / wavelength
	yAt := func(x float64) float64 {
		return yCenter + amp*math.Sin(k*(x-x0)+phase)
	}
	slopeAt := func(x float64) float64 {
		return amp * k * math.Cos(k*(x-x0)+phase)
	}
	open := false
	var prev path.Point
	for i := 0; i < n; i++ {
		xa := x0 + span*float64(i)/float64(n)
		xb := x0 + span*float64(i+1)/float64(n)
		dx := xb - xa
		ya, yb := yAt(xa), yAt(xb)
		sa, sb := slopeAt(xa), slopeAt(xb)
		p0 := path.NewPoint(xa, ya)
		p1 := path.NewPoint(xa+dx/3, ya+sa*dx/3)
		p2 := path.NewPoint(xb-dx/3, yb-sb*dx/3)
		p3 := path.NewPoint(xb, yb)
		if !clip || cubicHullInside(p0, p1, p2, p3, rx0, ry0, rx1, ry1) {
			if !open || math.Hypot(p0.X-prev.X, p0.Y-prev.Y) > 1e-6 {
				d.MoveTo(p0)
			}
			d.CurveTo(p1, p2, p3)
			open = true
			prev = p3
			continue
		}
		if cubicHullOutside(p0, p1, p2, p3, rx0, ry0, rx1, ry1) {
			open = false
			continue
		}
		open, prev = appendClippedCubic(d, p0, p1, p2, p3, rx0, ry0, rx1, ry1, open, prev)
	}
}

func cubicHullInside(p0, p1, p2, p3 path.Point, x0, y0, x1, y1 float64) bool {
	for _, p := range []path.Point{p0, p1, p2, p3} {
		if p.X < x0 || p.X > x1 || p.Y < y0 || p.Y > y1 {
			return false
		}
	}
	return true
}

func cubicHullOutside(p0, p1, p2, p3 path.Point, x0, y0, x1, y1 float64) bool {
	minX, maxX := p0.X, p0.X
	minY, maxY := p0.Y, p0.Y
	for _, p := range []path.Point{p1, p2, p3} {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return maxX < x0 || minX > x1 || maxY < y0 || minY > y1
}

func appendClippedCubic(d *path.Draw, p0, p1, p2, p3 path.Point, x0, y0, x1, y1 float64, open bool, prev path.Point) (bool, path.Point) {
	const steps = 8
	last := prev
	for i := 1; i <= steps; i++ {
		a := cubicAt(p0, p1, p2, p3, float64(i-1)/steps)
		b := cubicAt(p0, p1, p2, p3, float64(i)/steps)
		ax, ay, bx, by, ok := clipSegmentToRect(a.X, a.Y, b.X, b.Y, x0, y0, x1, y1)
		if !ok {
			open = false
			continue
		}
		start := path.NewPoint(ax, ay)
		if !open || math.Hypot(start.X-last.X, start.Y-last.Y) > 1e-6 {
			d.MoveTo(start)
		}
		end := path.NewPoint(bx, by)
		d.LineTo(end)
		open = true
		last = end
	}
	return open, last
}

func cubicAt(p0, p1, p2, p3 path.Point, t float64) path.Point {
	u := 1 - t
	uu, tt := u*u, t*t
	w0 := uu * u
	w1 := 3 * uu * t
	w2 := 3 * u * tt
	w3 := tt * t
	return path.NewPoint(
		w0*p0.X+w1*p1.X+w2*p2.X+w3*p3.X,
		w0*p0.Y+w1*p1.Y+w2*p2.Y+w3*p3.Y,
	)
}

type seg struct{ ax, ay, bx, by float64 }

func staggeredLattice(w, h, m float64, p latticeParams, clip bool) path.Path {
	d := path.NewDraw()
	rw := math.Max(1e-6, p.rectWidth)
	rh := math.Max(1e-6, p.rectHeight)
	vRatio := p.vGapRatio
	if vRatio < 0 {
		vRatio = 0
	}
	if vRatio > 0.49 {
		vRatio = 0.49
	}
	hRatio := p.hGapRatio
	if hRatio < 0 {
		hRatio = 0
	}
	vGap := vRatio * rh
	hGap := hRatio * rw
	armHalf := hGap / 2
	rungD := math.Max(0, (rh-vGap)/4)
	stagger := (rh + vGap) / 2
	pitchY := rh + vGap
	pitchX := rw + hGap
	vStub := vGap / 2
	midReach := pitchX - rw/2

	innerX0, innerY0 := m, m
	innerX1, innerY1 := w-m, h-m
	if innerX1 <= innerX0 || innerY1 <= innerY0 {
		return d.Path()
	}

	pad := 1
	if clip {
		pad = 2
	}
	i0 := int(math.Floor((innerX0-m-rw-midReach)/pitchX)) - pad
	i1 := int(math.Ceil((innerX1-m+midReach)/pitchX)) + pad
	j0 := int(math.Floor((innerY0-m-rh-vStub-stagger)/pitchY)) - pad
	j1 := int(math.Ceil((innerY1-m+vStub+stagger)/pitchY)) + pad

	var bb box
	for i := i0; i < i1; i++ {
		for j := j0; j < j1; j++ {
			tx := m + float64(i)*pitchX
			// Signed remainder matches the studio generator for negative columns.
			ty := m + float64(j)*pitchY + float64(i%2)*stagger
			bb.add(tx-midReach, ty-vStub, tx+rw+midReach, ty+rh+vStub)
		}
	}
	if !bb.ok {
		return d.Path()
	}
	ox, oy := centeringOffset(innerX0, innerY0, innerX1, innerY1, bb)

	var segments []seg
	addClipped := func(ax, ay, bx, by float64) {
		cx0, cy0, cx1, cy1, ok := clipSegmentToRect(ax, ay, bx, by, innerX0, innerY0, innerX1, innerY1)
		if !ok || math.Hypot(cx1-cx0, cy1-cy0) < 1e-9 {
			return
		}
		segments = append(segments, seg{cx0, cy0, cx1, cy1})
	}
	yUpper := rungD
	yLower := rh - rungD
	for i := i0; i < i1; i++ {
		for j := j0; j < j1; j++ {
			tx := m + float64(i)*pitchX + ox
			ty := m + float64(j)*pitchY + float64(i%2)*stagger + oy
			// One closed loop per cell so the stroke joins the corners. The
			// loop may cross the margin; framed_stroke clips the stroke.
			d.MoveTo(path.NewPoint(tx, ty))
			d.LineTo(path.NewPoint(tx+rw, ty))
			d.LineTo(path.NewPoint(tx+rw, ty+rh))
			d.LineTo(path.NewPoint(tx, ty+rh))
			d.LineTo(path.NewPoint(tx, ty))
			cx := tx + rw/2
			addClipped(cx, ty, cx, ty-vStub)
			addClipped(cx, ty+rh, cx, ty+rh+vStub)
			addClipped(tx, ty+yUpper, tx-armHalf, ty+yUpper)
			addClipped(tx+rw, ty+yUpper, tx+rw+armHalf, ty+yUpper)
			addClipped(tx, ty+yLower, tx-armHalf, ty+yLower)
			addClipped(tx+rw, ty+yLower, tx+rw+armHalf, ty+yLower)
			yMid := rh / 2
			addClipped(tx, ty+yMid, tx-pitchX+rw/2, ty+yMid)
			addClipped(tx+rw, ty+yMid, tx+pitchX+rw/2, ty+yMid)
		}
	}

	for _, s := range segments {
		d.MoveTo(path.NewPoint(s.ax, s.ay))
		d.LineTo(path.NewPoint(s.bx, s.by))
	}
	return d.Path()
}

// expandStroke clips centerlines to the inner margin, expands each one by
// thickness, unions the ribbons, and opens the margin the same way the
// staggered lattice does: butt caps on the border, endcap slivers, a
// perimeter band, then the bare inner rectangle is dropped.
func expandStroke(segments []seg, thickness, margin, width, height float64) path.Path {
	return strokeOpenSegments(segments, thickness, margin, width, height, true)
}

// strokeOpenSegments thickens independent segments and unions them, so a
// vertex where three edges meet fills instead of leaving butt-cap gaps.
// Ends on the margin get a butt cap. frame adds the perimeter band.
func strokeOpenSegments(segments []seg, thickness, margin, width, height float64, frame bool) path.Path {
	d := path.NewDraw()
	innerX0, innerY0 := margin, margin
	innerX1, innerY1 := width-margin, height-margin
	if innerX1 <= innerX0 || innerY1 <= innerY0 {
		return d.Path()
	}
	clipped := make([]seg, 0, len(segments))
	for _, s := range segments {
		ax, ay, bx, by, ok := clipSegmentToRect(s.ax, s.ay, s.bx, s.by, innerX0, innerY0, innerX1, innerY1)
		if !ok || math.Hypot(bx-ax, by-ay) < 1e-9 {
			continue
		}
		clipped = append(clipped, seg{ax, ay, bx, by})
	}
	if len(clipped) == 0 {
		return d.Path()
	}
	if !math.IsNaN(thickness) && !math.IsInf(thickness, 0) && thickness <= 1e-9 {
		for _, s := range clipped {
			d.MoveTo(path.NewPoint(s.ax, s.ay))
			d.LineTo(path.NewPoint(s.bx, s.by))
		}
		return d.Path()
	}

	hw := math.Max(1e-6, thickness) / 2
	bboxTol := math.Max(1e-6, 1e-5*math.Max(1, math.Max(width, height)))
	var quads []*canvas.Path
	for _, s := range clipped {
		capA, capB := hw, hw
		if onRectBoundary(s.ax, s.ay, innerX0, innerY0, innerX1, innerY1, bboxTol) {
			capA = 0
		}
		if onRectBoundary(s.bx, s.by, innerX0, innerY0, innerX1, innerY1, bboxTol) {
			capB = 0
		}
		ring := segmentRing(s.ax, s.ay, s.bx, s.by, hw, capA, capB)
		if ring != nil && !ring.Empty() {
			quads = append(quads, ring)
		}
	}
	geom := unionCanvas(quads)
	if geom == nil || geom.Empty() {
		return d.Path()
	}
	geom = geom.And(innerRectCanvas(margin, width, height))
	if geom == nil || geom.Empty() {
		return d.Path()
	}
	// A border replaces those open ends: the band covers the margin, and a
	// sliver cut there would sever the slot from the border.
	if !frame {
		var slivers []*canvas.Path
		for _, s := range clipped {
			if sl := endcapSliver(s.ax, s.ay, innerX0, innerY0, innerX1, innerY1, hw, bboxTol); sl != nil {
				slivers = append(slivers, sl)
			}
			if sl := endcapSliver(s.bx, s.by, innerX0, innerY0, innerX1, innerY1, hw, bboxTol); sl != nil {
				slivers = append(slivers, sl)
			}
		}
		if len(slivers) > 0 {
			if cut := unionCanvas(slivers); cut != nil && !cut.Empty() {
				geom = geom.Not(cut)
			}
		}
		if geom == nil || geom.Empty() {
			return d.Path()
		}
	}
	if frame {
		if bands := perimeterBands(innerX0, innerY0, innerX1, innerY1, hw); len(bands) > 0 {
			geom = unionCanvas(append([]*canvas.Path{geom}, bands...))
		}
	}
	appendCanvasExceptInnerRect(d, geom, innerX0, innerY0, innerX1, innerY1)
	return d.Path()
}

// strokeCenterline thickens a drawing with canvas's stroke: both sides are
// offset and corners are joined. Open ends that meet the margin are extended
// through that edge first, so the stroke crosses the frame on the line's
// angle. The result is clipped to the inner margin and unioned with the frame.
func strokeCenterline(centerline path.Path, thickness, margin, width, height float64, frame bool, join string) path.Path {
	if centerline == nil || len(centerline.Segments()) == 0 {
		return path.NewPath()
	}
	if !math.IsNaN(thickness) && !math.IsInf(thickness, 0) && thickness <= 1e-9 {
		return expandStroke(segmentsFromPath(centerline), 0, margin, width, height)
	}
	d := path.NewDraw()
	innerX0, innerY0 := margin, margin
	innerX1, innerY1 := width-margin, height-margin
	if innerX1 <= innerX0+1e-9 || innerY1 <= innerY0+1e-9 {
		return d.Path()
	}
	hw := math.Max(1e-6, thickness) / 2
	extended := extendCenterlinePastMargin(centerline, innerX0, innerY0, innerX1, innerY1, hw)
	joiner := canvas.Joiner(canvas.MiterJoin)
	strokeTol := 0.001
	if strings.EqualFold(strings.TrimSpace(join), "round") {
		joiner = canvas.RoundJoin
		strokeTol = 0.0001
		prevTol := canvas.Tolerance
		canvas.Tolerance = 0.0001
		defer func() { canvas.Tolerance = prevTol }()
	}
	stroked := path.ToCanvasPathJoinLoops(extended).Stroke(math.Max(1e-6, thickness), canvas.ButtCap, joiner, strokeTol)
	if stroked == nil || stroked.Empty() {
		return d.Path()
	}
	clipped := stroked.And(innerRectCanvas(margin, width, height))
	if clipped == nil || clipped.Empty() {
		return d.Path()
	}
	if frame {
		if bands := perimeterBands(innerX0, innerY0, innerX1, innerY1, hw); len(bands) > 0 {
			clipped = unionCanvas(append([]*canvas.Path{clipped}, bands...))
		}
		clipped = flushCutsToFrame(clipped, innerX0, innerY0, innerX1, innerY1, hw)
	}
	appendCanvasExceptInnerRect(d, clipped, innerX0, innerY0, innerX1, innerY1)
	return d.Path()
}

// extendCenterlinePastMargin pushes an open end through the inner margin when
// that end already meets the frame, so the later stroke crosses the band
// instead of stopping on it.
func extendCenterlinePastMargin(centerline path.Path, x0, y0, x1, y1, hw float64) path.Path {
	segs := centerline.Segments()
	if len(segs) == 0 {
		return centerline
	}
	out := path.NewPath()
	var sub []path.Segment
	flush := func() {
		if len(sub) == 0 {
			return
		}
		start := sub[0].End()
		end := sub[len(sub)-1].End()
		closed := len(sub) > 2 && math.Hypot(end.X-start.X, end.Y-start.Y) < 1e-6
		var lead, tail *path.Point
		if !closed && len(sub) >= 2 {
			if dx, dy, ok := outwardFromStart(sub[1]); ok {
				if q, ok := extendPastRect(start, dx, dy, x0, y0, x1, y1, hw); ok {
					lead = &q
				}
			}
			if dx, dy, ok := outwardFromEnd(sub[len(sub)-1]); ok {
				if q, ok := extendPastRect(end, dx, dy, x0, y0, x1, y1, hw); ok {
					tail = &q
				}
			}
		}
		if lead != nil {
			out.AddSegments(path.MoveSegment{EndPoint: *lead})
			out.AddSegments(path.LineSegment{StartPoint: *lead, EndPoint: start})
			out.AddSegments(sub[1:]...)
		} else {
			out.AddSegments(sub...)
		}
		if tail != nil {
			out.AddSegments(path.LineSegment{StartPoint: end, EndPoint: *tail})
		}
		sub = nil
	}
	for _, s := range segs {
		if _, isMove := s.(path.MoveSegment); isMove && len(sub) > 0 {
			flush()
		}
		sub = append(sub, s)
	}
	flush()
	return out
}

func outwardFromStart(seg path.Segment) (float64, float64, bool) {
	switch s := seg.(type) {
	case path.LineSegment:
		return s.StartPoint.X - s.EndPoint.X, s.StartPoint.Y - s.EndPoint.Y, true
	case path.CurveSegment:
		return s.StartPoint.X - s.ControlPointStart.X, s.StartPoint.Y - s.ControlPointStart.Y, true
	default:
		return 0, 0, false
	}
}

func outwardFromEnd(seg path.Segment) (float64, float64, bool) {
	switch s := seg.(type) {
	case path.LineSegment:
		return s.EndPoint.X - s.StartPoint.X, s.EndPoint.Y - s.StartPoint.Y, true
	case path.CurveSegment:
		return s.EndPoint.X - s.ControlPointEnd.X, s.EndPoint.Y - s.ControlPointEnd.Y, true
	default:
		return 0, 0, false
	}
}

// extendPastRect returns a point further along (dx, dy) once that ray has
// cleared the rectangle by hw. Ends that stop well inside the panel are left
// alone.
func extendPastRect(p path.Point, dx, dy, x0, y0, x1, y1, hw float64) (path.Point, bool) {
	length := math.Hypot(dx, dy)
	if length < 1e-12 || hw < 1e-12 {
		return p, false
	}
	ux, uy := dx/length, dy/length
	if dist, outside := outsideRect(p.X, p.Y, x0, y0, x1, y1); outside {
		if dist >= hw-1e-9 {
			return p, false
		}
		t, ok := tToClearRect(p.X, p.Y, ux, uy, x0, y0, x1, y1, hw)
		if !ok {
			return p, false
		}
		return path.NewPoint(p.X+ux*(t+1e-3), p.Y+uy*(t+1e-3)), true
	}
	// A stroke of width 2*hw around an end this close to the frame overlaps
	// the border band. Push the end through every nearby edge it is heading
	// toward so the cap is outside the panel instead of pinching on the band.
	reach := 2 * hw
	tClear := 0.0
	near := false
	consider := func(gap, outward float64) {
		if gap > reach+1e-9 || outward <= 1e-3 {
			return
		}
		need := (gap + hw) / outward
		if need > tClear {
			tClear = need
		}
		near = true
	}
	consider(p.X-x0, -ux)
	consider(x1-p.X, ux)
	consider(p.Y-y0, -uy)
	consider(y1-p.Y, uy)
	if !near {
		return p, false
	}
	return path.NewPoint(p.X+ux*(tClear+1e-3), p.Y+uy*(tClear+1e-3)), true
}

func outsideRect(x, y, x0, y0, x1, y1 float64) (float64, bool) {
	ox, oy := 0.0, 0.0
	if x < x0 {
		ox = x0 - x
	} else if x > x1 {
		ox = x - x1
	}
	if y < y0 {
		oy = y0 - y
	} else if y > y1 {
		oy = y - y1
	}
	if ox == 0 && oy == 0 {
		return 0, false
	}
	if ox > 0 && oy > 0 {
		return math.Hypot(ox, oy), true
	}
	return math.Max(ox, oy), true
}

func tToClearRect(x, y, ux, uy, x0, y0, x1, y1, hw float64) (float64, bool) {
	t := 0.0
	set := false
	need := func(num, den float64) {
		if den < 1e-3 {
			return
		}
		if num/den > t {
			t = num / den
		}
		set = true
	}
	if x < x0 {
		need(x-(x0-hw), -ux)
	} else if x > x1 {
		need((x1+hw)-x, ux)
	}
	if y < y0 {
		need(y-(y0-hw), -uy)
	} else if y > y1 {
		need((y1+hw)-y, uy)
	}
	if !set || t < 1e-12 {
		return 0, false
	}
	return t, true
}

// segmentsFromPath flattens a rendered drawing into centerline segments.
// A close command that does not already return to the subpath start is kept.
func segmentsFromPath(p path.Path) []seg {
	if p == nil || len(p.Segments()) == 0 {
		return nil
	}
	flat := path.ToCanvasPath(p).Flatten(0.01)
	if flat == nil || flat.Empty() {
		return nil
	}
	var out []seg
	var curX, curY, subX, subY float64
	has := false
	s := flat.Scanner()
	for s.Scan() {
		x, y := s.End().X, s.End().Y
		switch s.Cmd() {
		case canvas.MoveToCmd:
			curX, curY = x, y
			subX, subY = x, y
			has = true
		case canvas.LineToCmd:
			if has && math.Hypot(x-curX, y-curY) >= 1e-9 {
				out = append(out, seg{curX, curY, x, y})
			}
			curX, curY = x, y
			has = true
		case canvas.CloseCmd:
			if has && math.Hypot(subX-curX, subY-curY) >= 1e-9 {
				out = append(out, seg{curX, curY, subX, subY})
			}
			has = false
		}
	}
	return out
}

type circle struct{ cx, cy, r float64 }

func overlappingCircles(w, h, m float64, p overlapParams) path.Path {
	d := path.NewDraw()
	xmin, ymin := m, m
	xmax, ymax := w-m, h-m
	if xmax <= xmin || ymax <= ymin {
		return d.Path()
	}
	panel := innerRectCanvas(m, w, h)
	rings := overlapAnnuli(w, h, m, p)
	if len(rings) == 0 {
		appendCanvas(d, panel)
		return d.Path()
	}
	// Union through the polygon boolean flattens every circle first. The
	// outline is the arcs themselves, cut where the circles and the panel meet.
	return ringBoundary(rings, xmin, ymin, xmax, ymax)
}

func overlapAnnuli(w, h, m float64, p overlapParams) []annulus {
	xmin, ymin := m, m
	xmax, ymax := w-m, h-m
	target := p.count
	if target < 0 {
		target = 0
	}
	if target == 0 {
		return nil
	}
	lo := math.Max(1e-6, p.radiusMin)
	hi := p.radiusMax
	if math.IsNaN(hi) || math.IsInf(hi, 0) {
		hi = lo
	}
	if hi < lo {
		lo, hi = hi, lo
	}
	sw := math.Max(1e-6, p.strokeWidth)
	innerW := xmax - xmin
	innerH := ymax - ymin
	rng := mulberry32(p.seed)
	pickR := func() float64 {
		if hi <= lo {
			return lo
		}
		return lo + rng()*(hi-lo)
	}
	outlineCrosses := func(a, b circle) bool {
		dx := a.cx - b.cx
		dy := a.cy - b.cy
		d2 := dx*dx + dy*dy
		sumR := a.r + b.r
		diffR := math.Abs(a.r - b.r)
		upper := (sumR - 1e-6) * (sumR - 1e-6)
		lower := (diffR + 1e-6) * (diffR + 1e-6)
		return d2 < upper && d2 > lower
	}

	aspect := 1.0
	if innerH > 0 {
		aspect = innerW / innerH
	}
	cols := int(math.Max(2, math.Ceil(math.Sqrt(math.Max(2, float64(target))*aspect))))
	rows := int(math.Max(2, math.Ceil(math.Max(2, float64(target))/float64(cols))))
	cellW := innerW / math.Max(1, float64(cols-1))
	cellH := innerH / math.Max(1, float64(rows-1))
	const jitterFrac = 0.3
	const outerJitterScale = 0.4

	grid := make([][]circle, rows)
	var placed []circle
	for row := 0; row < rows; row++ {
		grid[row] = make([]circle, cols)
		for col := 0; col < cols; col++ {
			baseX := xmin + float64(col)*cellW
			baseY := ymin + float64(row)*cellH
			scaleX, scaleY := 1.0, 1.0
			if col == 0 || col == cols-1 {
				scaleX = outerJitterScale
			}
			if row == 0 || row == rows-1 {
				scaleY = outerJitterScale
			}
			c := circle{
				cx: baseX + (rng()-0.5)*2*cellW*jitterFrac*scaleX,
				cy: baseY + (rng()-0.5)*2*cellH*jitterFrac*scaleY,
				r:  pickR(),
			}
			grid[row][col] = c
			placed = append(placed, c)
		}
	}
	addConnector := func(a, b circle) {
		if outlineCrosses(a, b) {
			return
		}
		dist := math.Hypot(a.cx-b.cx, a.cy-b.cy)
		minR := dist/2 - math.Min(a.r, b.r) + 0.05
		rLow := math.Max(lo, minR)
		rHigh := math.Max(rLow, hi)
		placed = append(placed, circle{
			cx: (a.cx + b.cx) / 2,
			cy: (a.cy + b.cy) / 2,
			r:  rLow + rng()*(rHigh-rLow),
		})
	}
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			c := grid[row][col]
			if col+1 < cols {
				addConnector(c, grid[row][col+1])
			}
			if row+1 < rows {
				addConnector(c, grid[row+1][col])
			}
		}
	}
	neighbor := make([]bool, len(placed))
	for i := 0; i < len(placed); i++ {
		for j := i + 1; j < len(placed); j++ {
			if outlineCrosses(placed[i], placed[j]) {
				neighbor[i] = true
				neighbor[j] = true
			}
		}
	}
	half := sw / 2
	var rings []annulus
	for i, c := range placed {
		if !neighbor[i] {
			continue
		}
		inner := c.r - half
		if inner < 0 {
			inner = 0
		}
		rings = append(rings, annulus{cx: c.cx, cy: c.cy, inner: inner, outer: c.r + half})
	}
	return rings
}

func voronoiSlots(w, h, m float64, p voronoiParams) path.Path {
	xmin, ymin := m, m
	xmax, ymax := w-m, h-m
	if xmax <= xmin || ymax <= ymin {
		return path.NewPath()
	}
	sites := voronoiSites(xmin, ymin, xmax, ymax, p)
	return strokeOpenSegments(voronoiEdges(sites, xmin, ymin, xmax, ymax), p.thickness, m, w, h, true)
}

// voronoiSites matches the studio placer, including the rejection loop and the
// padding that tops the list up to three points.
func voronoiSites(xmin, ymin, xmax, ymax float64, p voronoiParams) []point {
	iw, ih := xmax-xmin, ymax-ymin
	target := p.siteCount
	if target < 4 {
		target = 4
	}
	rng := mulberry32(p.seed)
	minGap := p.minSiteGap
	if minGap < 0 || math.IsNaN(minGap) || math.IsInf(minGap, 0) {
		minGap = 0
	}
	minD2 := minGap * minGap
	coords := make([]float64, 0, target*2)
	maxAttempts := target * 120
	if maxAttempts < 3000 {
		maxAttempts = 3000
	}
	attempts := 0
	for len(coords) < target*2 && attempts < maxAttempts {
		attempts++
		x := xmin + rng()*iw
		y := ymin + rng()*ih
		if minGap > 1e-12 {
			ok := true
			for i := 0; i < len(coords); i += 2 {
				dx := x - coords[i]
				dy := y - coords[i+1]
				if dx*dx+dy*dy < minD2 {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
		}
		coords = append(coords, x, y)
	}
	for len(coords) < 6 {
		coords = append(coords, xmin+(0.1+rng()*0.8)*iw, ymin+(0.1+rng()*0.8)*ih)
	}
	sites := make([]point, 0, len(coords)/2)
	for i := 0; i+1 < len(coords); i += 2 {
		sites = append(sites, point{coords[i], coords[i+1]})
	}
	return sites
}

// voronoiEdges is the Voronoi diagram of sites, clipped to the rectangle.
// Each triangle's circumcenter is a vertex. An interior edge joins the two
// neighboring circumcenters. A hull edge is the ray from its circumcenter
// along the perpendicular bisector, away from the triangle.
func voronoiEdges(sites []point, xmin, ymin, xmax, ymax float64) []seg {
	pts := make([]point, 0, len(sites))
	for _, p := range sites {
		dup := false
		for _, q := range pts {
			if math.Hypot(p.x-q.x, p.y-q.y) < 1e-9 {
				dup = true
				break
			}
		}
		if !dup {
			pts = append(pts, p)
		}
	}
	tris := delaunay(pts)
	type info struct {
		centers []point
		third   point
		a, b    point
	}
	edges := map[[2]int]*info{}
	for _, t := range tris {
		cc, ok := circumcenter(pts[t[0]], pts[t[1]], pts[t[2]])
		if !ok {
			continue
		}
		idx := [3]int{t[0], t[1], t[2]}
		for k := 0; k < 3; k++ {
			i, j := idx[k], idx[(k+1)%3]
			other := idx[(k+2)%3]
			if i > j {
				i, j = j, i
			}
			key := [2]int{i, j}
			inf := edges[key]
			if inf == nil {
				inf = &info{a: pts[i], b: pts[j], third: pts[other]}
				edges[key] = inf
			}
			inf.centers = append(inf.centers, cc)
		}
	}
	var out []seg
	seen := map[[4]int]bool{}
	add := func(x0, y0, x1, y1 float64) {
		if math.Hypot(x1-x0, y1-y0) < 1e-8 {
			return
		}
		r := func(v float64) int { return int(math.Round(v * 1e6)) }
		k := [4]int{r(x0), r(y0), r(x1), r(y1)}
		if k[0] > k[2] || (k[0] == k[2] && k[1] > k[3]) {
			k = [4]int{k[2], k[3], k[0], k[1]}
		}
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, seg{x0, y0, x1, y1})
	}
	for _, inf := range edges {
		if len(inf.centers) >= 2 {
			ax, ay, bx, by, ok := clipSegmentToRect(inf.centers[0].x, inf.centers[0].y, inf.centers[1].x, inf.centers[1].y, xmin, ymin, xmax, ymax)
			if ok {
				add(ax, ay, bx, by)
			}
			continue
		}
		if len(inf.centers) != 1 {
			continue
		}
		dx, dy := outwardBisector(inf.a, inf.b, inf.third)
		ax, ay, bx, by, ok := clipRayToRect(inf.centers[0].x, inf.centers[0].y, dx, dy, xmin, ymin, xmax, ymax)
		if ok {
			add(ax, ay, bx, by)
		}
	}
	return out
}

func outwardBisector(a, b, third point) (float64, float64) {
	px, py := -(b.y - a.y), b.x-a.x
	midX, midY := (a.x+b.x)/2, (a.y+b.y)/2
	if px*(third.x-midX)+py*(third.y-midY) > 0 {
		px, py = -px, -py
	}
	return px, py
}

func clipRayToRect(ox, oy, dx, dy, xmin, ymin, xmax, ymax float64) (float64, float64, float64, float64, bool) {
	length := math.Hypot(dx, dy)
	if length < 1e-15 {
		return 0, 0, 0, 0, false
	}
	dx, dy = dx/length, dy/length
	u0, u1 := 0.0, math.Inf(1)
	clip := func(p, q float64) bool {
		if math.Abs(p) < 1e-15 {
			return q >= -1e-9
		}
		r := q / p
		if p < 0 {
			if r > u1 {
				return false
			}
			if r > u0 {
				u0 = r
			}
		} else {
			if r < u0 {
				return false
			}
			if r < u1 {
				u1 = r
			}
		}
		return true
	}
	if !clip(-dx, ox-xmin) || !clip(dx, xmax-ox) || !clip(-dy, oy-ymin) || !clip(dy, ymax-oy) {
		return 0, 0, 0, 0, false
	}
	if u0 < 0 {
		u0 = 0
	}
	if u1 < u0 || u1-u0 < 1e-8 {
		return 0, 0, 0, 0, false
	}
	return ox + u0*dx, oy + u0*dy, ox + u1*dx, oy + u1*dy, true
}

func delaunay(pts []point) [][3]int {
	n := len(pts)
	if n < 3 {
		return nil
	}
	minX, minY := pts[0].x, pts[0].y
	maxX, maxY := minX, minY
	for _, p := range pts[1:] {
		if p.x < minX {
			minX = p.x
		}
		if p.y < minY {
			minY = p.y
		}
		if p.x > maxX {
			maxX = p.x
		}
		if p.y > maxY {
			maxY = p.y
		}
	}
	dx, dy := maxX-minX, maxY-minY
	if dx < 1 {
		dx = 1
	}
	if dy < 1 {
		dy = 1
	}
	span := math.Max(dx, dy) * 20
	midX, midY := (minX+maxX)/2, (minY+maxY)/2
	all := append(append([]point{}, pts...), point{midX - 2*span, midY - span}, point{midX + 2*span, midY - span}, point{midX, midY + 2*span})
	tris := [][3]int{{n, n + 1, n + 2}}
	for i := 0; i < n; i++ {
		bad := make([]bool, len(tris))
		for ti, t := range tris {
			if inCircumcircle(all[i], all[t[0]], all[t[1]], all[t[2]]) {
				bad[ti] = true
			}
		}
		type de struct{ a, b int }
		count := map[[2]int]int{}
		var directed []de
		for ti, t := range tris {
			if !bad[ti] {
				continue
			}
			for _, e := range []de{{t[0], t[1]}, {t[1], t[2]}, {t[2], t[0]}} {
				k := [2]int{e.a, e.b}
				if k[0] > k[1] {
					k[0], k[1] = k[1], k[0]
				}
				count[k]++
				directed = append(directed, e)
			}
		}
		next := make([][3]int, 0, len(tris)+8)
		for ti, t := range tris {
			if !bad[ti] {
				next = append(next, t)
			}
		}
		for _, e := range directed {
			k := [2]int{e.a, e.b}
			if k[0] > k[1] {
				k[0], k[1] = k[1], k[0]
			}
			if count[k] != 1 {
				continue
			}
			if orient(all[e.a], all[e.b], all[i]) < 0 {
				next = append(next, [3]int{e.b, e.a, i})
			} else {
				next = append(next, [3]int{e.a, e.b, i})
			}
		}
		tris = next
	}
	out := make([][3]int, 0, len(tris))
	for _, t := range tris {
		if t[0] >= n || t[1] >= n || t[2] >= n {
			continue
		}
		if math.Abs(orient(all[t[0]], all[t[1]], all[t[2]])) < 1e-14 {
			continue
		}
		out = append(out, t)
	}
	return out
}

func orient(a, b, c point) float64 {
	return (b.x-a.x)*(c.y-a.y) - (b.y-a.y)*(c.x-a.x)
}

func inCircumcircle(p, a, b, c point) bool {
	if orient(a, b, c) < 0 {
		b, c = c, b
	}
	a1, a2 := a.x-p.x, a.y-p.y
	b1, b2 := b.x-p.x, b.y-p.y
	c1, c2 := c.x-p.x, c.y-p.y
	det := (a1*a1+a2*a2)*(b1*c2-c1*b2) -
		(b1*b1+b2*b2)*(a1*c2-c1*a2) +
		(c1*c1+c2*c2)*(a1*b2-b1*a2)
	return det >= -1e-10
}

func circumcenter(a, b, c point) (point, bool) {
	d := 2 * (a.x*(b.y-c.y) + b.x*(c.y-a.y) + c.x*(a.y-b.y))
	if math.Abs(d) < 1e-14 {
		return point{}, false
	}
	a2 := a.x*a.x + a.y*a.y
	b2 := b.x*b.x + b.y*b.y
	c2 := c.x*c.x + c.y*c.y
	return point{
		(a2*(b.y-c.y) + b2*(c.y-a.y) + c2*(a.y-b.y)) / d,
		(a2*(c.x-b.x) + b2*(a.x-c.x) + c2*(b.x-a.x)) / d,
	}, true
}

type annulus struct {
	cx, cy, inner, outer float64
}

type ringPiece struct {
	x0, y0, x1, y1 float64
	cx, cy, r      float64
	delta          float64
}

// ringBoundary is the outline of the panel with the rings removed. Circle
// pieces stay arcs; only the panel edges are straight.
func ringBoundary(rings []annulus, xmin, ymin, xmax, ymax float64) path.Path {
	const angEps = 1e-6
	halfMin := math.MaxFloat64
	for _, g := range rings {
		h := (g.outer - g.inner) / 2
		if g.inner <= 1e-8 {
			h = g.outer
		}
		if h > 1e-8 && h < halfMin {
			halfMin = h
		}
	}
	if math.IsInf(halfMin, 0) || halfMin == math.MaxFloat64 {
		halfMin = 1e-4
	}
	eps := math.Min(1e-4, halfMin*0.25)

	inAnn := func(x, y float64) bool {
		for _, g := range rings {
			d := math.Hypot(x-g.cx, y-g.cy)
			if g.inner <= 1e-8 {
				if d <= g.outer {
					return true
				}
				continue
			}
			if d <= g.outer && d >= g.inner {
				return true
			}
		}
		return false
	}
	inPanel := func(x, y float64) bool {
		return x >= xmin-1e-7 && x <= xmax+1e-7 && y >= ymin-1e-7 && y <= ymax+1e-7
	}
	wood := func(x, y float64) bool {
		return inPanel(x, y) && !inAnn(x, y)
	}

	type circ struct {
		cx, cy, r float64
		outer     bool
	}
	var circles []circ
	for _, g := range rings {
		if g.outer > 1e-6 {
			circles = append(circles, circ{g.cx, g.cy, g.outer, true})
		}
		if g.inner > 1e-6 {
			circles = append(circles, circ{g.cx, g.cy, g.inner, false})
		}
	}

	type edge struct{ x0, y0, x1, y1, nx, ny float64 }
	edges := []edge{
		{xmin, ymin, xmax, ymin, 0, 1},
		{xmax, ymin, xmax, ymax, -1, 0},
		{xmax, ymax, xmin, ymax, 0, -1},
		{xmin, ymax, xmin, ymin, 1, 0},
	}

	var pieces []ringPiece
	for _, e := range edges {
		ts := []float64{0, 1}
		dx, dy := e.x1-e.x0, e.y1-e.y0
		for _, c := range circles {
			for _, p := range segmentCircleHits(e.x0, e.y0, e.x1, e.y1, c.cx, c.cy, c.r) {
				den := dx*dx + dy*dy
				if den < 1e-18 {
					continue
				}
				t := ((p[0]-e.x0)*dx + (p[1]-e.y0)*dy) / den
				if t > -1e-8 && t < 1+1e-8 {
					ts = append(ts, math.Min(1, math.Max(0, t)))
				}
			}
		}
		sort.Float64s(ts)
		uniq := ts[:1]
		for _, t := range ts[1:] {
			if t-uniq[len(uniq)-1] > 1e-7 {
				uniq = append(uniq, t)
			}
		}
		length := math.Hypot(dx, dy)
		for i := 0; i < len(uniq)-1; i++ {
			if (uniq[i+1]-uniq[i])*length < 1e-6 {
				continue
			}
			tm := (uniq[i] + uniq[i+1]) / 2
			mx, my := e.x0+tm*dx, e.y0+tm*dy
			if !wood(mx+eps*e.nx, my+eps*e.ny) {
				continue
			}
			pieces = append(pieces, ringPiece{
				x0: e.x0 + uniq[i]*dx, y0: e.y0 + uniq[i]*dy,
				x1: e.x0 + uniq[i+1]*dx, y1: e.y0 + uniq[i+1]*dy,
			})
		}
	}

	keep := func(c circ, ang float64) bool {
		co, si := math.Cos(ang), math.Sin(ang)
		mx, my := c.cx+c.r*co, c.cy+c.r*si
		if !inPanel(mx, my) {
			return false
		}
		ox, oy := mx+eps*co, my+eps*si
		ix, iy := mx-eps*co, my-eps*si
		if c.outer {
			return wood(ox, oy) && inAnn(ix, iy)
		}
		return wood(ix, iy) && inAnn(ox, oy)
	}
	for _, c := range circles {
		var angs []float64
		for _, o := range circles {
			if o.cx == c.cx && o.cy == c.cy && o.r == c.r {
				continue
			}
			for _, p := range circleHits(c.cx, c.cy, c.r, o.cx, o.cy, o.r) {
				angs = append(angs, math.Atan2(p[1]-c.cy, p[0]-c.cx))
			}
		}
		for _, e := range edges {
			for _, p := range segmentCircleHits(e.x0, e.y0, e.x1, e.y1, c.cx, c.cy, c.r) {
				angs = append(angs, math.Atan2(p[1]-c.cy, p[0]-c.cx))
			}
		}
		angs = uniqueAngles(angs, angEps)
		if len(angs) == 0 {
			if keep(c, 0) {
				pieces = append(pieces, fullRing(c.cx, c.cy, c.r))
			}
			continue
		}
		wrapped := append(append([]float64{}, angs...), angs[0]+2*math.Pi)
		for i := 0; i < len(wrapped)-1; i++ {
			delta := wrapped[i+1] - wrapped[i]
			if delta < angEps {
				continue
			}
			if !keep(c, wrapped[i]+delta/2) {
				continue
			}
			a0, a1 := wrapped[i], wrapped[i+1]
			pieces = append(pieces, ringPiece{
				x0: c.cx + c.r*math.Cos(a0), y0: c.cy + c.r*math.Sin(a0),
				x1: c.cx + c.r*math.Cos(a1), y1: c.cy + c.r*math.Sin(a1),
				cx: c.cx, cy: c.cy, r: c.r, delta: delta,
			})
		}
	}
	if len(pieces) == 0 {
		cp := &canvas.Path{}
		cp.MoveTo(xmin, ymin)
		cp.LineTo(xmax, ymin)
		cp.LineTo(xmax, ymax)
		cp.LineTo(xmin, ymax)
		cp.Close()
		return path.PathFromCanvas(cp)
	}
	return path.PathFromCanvas(chainRingPieces(pieces))
}

func fullRing(cx, cy, r float64) ringPiece {
	return ringPiece{x0: cx + r, y0: cy, x1: cx + r, y1: cy, cx: cx, cy: cy, r: r, delta: 2 * math.Pi}
}

func uniqueAngles(angs []float64, eps float64) []float64 {
	if len(angs) == 0 {
		return nil
	}
	norm := make([]float64, len(angs))
	for i, a := range angs {
		a = math.Mod(a, 2*math.Pi)
		if a < 0 {
			a += 2 * math.Pi
		}
		norm[i] = a
	}
	sort.Float64s(norm)
	out := norm[:1:1]
	for _, a := range norm[1:] {
		if a-out[len(out)-1] > eps {
			out = append(out, a)
		}
	}
	if len(out) > 1 && out[0]+2*math.Pi-out[len(out)-1] < eps {
		out = out[:len(out)-1]
	}
	return out
}

func circleHits(ax, ay, ar, bx, by, br float64) [][2]float64 {
	dx, dy := bx-ax, by-ay
	d := math.Hypot(dx, dy)
	if d < 1e-12 || d > ar+br+1e-8 || d < math.Abs(ar-br)-1e-8 {
		return nil
	}
	a := (ar*ar - br*br + d*d) / (2 * d)
	h2 := ar*ar - a*a
	if h2 < 0 {
		h2 = 0
	}
	h := math.Sqrt(h2)
	mx, my := ax+a*dx/d, ay+a*dy/d
	px, py := -dy/d*h, dx/d*h
	hits := [][2]float64{{mx + px, my + py}}
	if h > 1e-8 {
		hits = append(hits, [2]float64{mx - px, my - py})
	}
	return hits
}

func segmentCircleHits(x0, y0, x1, y1, cx, cy, r float64) [][2]float64 {
	dx, dy := x1-x0, y1-y0
	fx, fy := x0-cx, y0-cy
	a := dx*dx + dy*dy
	if a < 1e-18 {
		return nil
	}
	b := 2 * (fx*dx + fy*dy)
	c := fx*fx + fy*fy - r*r
	disc := b*b - 4*a*c
	if disc < 0 {
		return nil
	}
	sd := math.Sqrt(math.Max(0, disc))
	var hits [][2]float64
	for _, t := range []float64{(-b - sd) / (2 * a), (-b + sd) / (2 * a)} {
		if t < -1e-8 || t > 1+1e-8 {
			continue
		}
		hits = append(hits, [2]float64{x0 + t*dx, y0 + t*dy})
	}
	return hits
}

func chainRingPieces(pieces []ringPiece) *canvas.Path {
	key := func(x, y float64) [2]int {
		return [2]int{int(math.Round(x * 1e5)), int(math.Round(y * 1e5))}
	}
	starts := map[[2]int][]int{}
	ends := map[[2]int][]int{}
	for i, p := range pieces {
		k0, k1 := key(p.x0, p.y0), key(p.x1, p.y1)
		starts[k0] = append(starts[k0], i)
		ends[k1] = append(ends[k1], i)
	}
	used := make([]bool, len(pieces))
	take := func(m map[[2]int][]int, k [2]int) (int, bool) {
		for _, i := range m[k] {
			if !used[i] {
				return i, true
			}
		}
		return 0, false
	}
	// One coordinate per snapped key, so a panel hit and the circle point
	// reconstructed from its angle are the same vertex.
	canon := map[[2]int][2]float64{}
	remember := func(k [2]int, x, y float64) {
		if _, ok := canon[k]; !ok {
			canon[k] = [2]float64{x, y}
		}
	}
	for i, p := range pieces {
		k0, k1 := key(p.x0, p.y0), key(p.x1, p.y1)
		remember(k0, p.x0, p.y0)
		remember(k1, p.x1, p.y1)
		p.x0, p.y0 = canon[k0][0], canon[k0][1]
		p.x1, p.y1 = canon[k1][0], canon[k1][1]
		pieces[i] = p
	}
	cp := &canvas.Path{}
	for i := range pieces {
		if used[i] {
			continue
		}
		used[i] = true
		cp.MoveTo(pieces[i].x0, pieces[i].y0)
		emitRingPiece(cp, pieces[i], false)
		cur := key(pieces[i].x1, pieces[i].y1)
		start := key(pieces[i].x0, pieces[i].y0)
		for n := 0; n < len(pieces) && cur != start; n++ {
			if j, ok := take(starts, cur); ok {
				used[j] = true
				emitRingPiece(cp, pieces[j], false)
				cur = key(pieces[j].x1, pieces[j].y1)
				continue
			}
			if j, ok := take(ends, cur); ok {
				used[j] = true
				emitRingPiece(cp, pieces[j], true)
				cur = key(pieces[j].x0, pieces[j].y0)
				continue
			}
			break
		}
		// Close() would turn a closing LineTo into a Close command, and the
		// HFD conversion drops Close, opening the loop by that last edge.
		pos := cp.Pos()
		if cur == start && (pos.X != pieces[i].x0 || pos.Y != pieces[i].y0) {
			cp.LineTo(pieces[i].x0, pieces[i].y0)
		}
	}
	return cp
}

func emitRingPiece(cp *canvas.Path, p ringPiece, reverse bool) {
	if p.r <= 0 {
		if reverse {
			cp.LineTo(p.x0, p.y0)
			return
		}
		cp.LineTo(p.x1, p.y1)
		return
	}
	if p.delta >= 2*math.Pi-1e-4 {
		mx := p.cx - (p.x0 - p.cx)
		my := p.cy - (p.y0 - p.cy)
		cp.ArcTo(p.r, p.r, 0, true, true, mx, my)
		cp.ArcTo(p.r, p.r, 0, true, true, p.x0, p.y0)
		return
	}
	x, y := p.x1, p.y1
	sweep := true
	if reverse {
		x, y = p.x0, p.y0
		sweep = false
	}
	cp.ArcTo(p.r, p.r, 0, p.delta > math.Pi, sweep, x, y)
}

// mulberry32 matches the studio generator, including Math.imul's 32-bit wrap.
func mulberry32(seed int32) func() float64 {
	t := uint32(seed)
	return func() float64 {
		t += 0x6D2B79F5
		x := imul(t^(t>>15), t|1)
		x ^= x + imul(x^(x>>7), x|61)
		return float64(x^(x>>14)) / 4294967296
	}
}

func imul(a, b uint32) uint32 {
	return uint32(int32(a) * int32(b))
}

func clipInfiniteLine(ox, oy, vx, vy, xmin, ymin, xmax, ymax float64) (ax, ay, bx, by float64, ok bool) {
	const eps = 1e-9
	ts := make([]float64, 0, 4)
	addT := func(t float64) {
		if math.IsNaN(t) || math.IsInf(t, 0) {
			return
		}
		x := ox + t*vx
		y := oy + t*vy
		if x >= xmin-1e-6 && x <= xmax+1e-6 && y >= ymin-1e-6 && y <= ymax+1e-6 {
			ts = append(ts, t)
		}
	}
	if math.Abs(vx) > eps {
		addT((xmin - ox) / vx)
		addT((xmax - ox) / vx)
	}
	if math.Abs(vy) > eps {
		addT((ymin - oy) / vy)
		addT((ymax - oy) / vy)
	}
	if len(ts) < 2 {
		return
	}
	sort.Float64s(ts)
	t0, t1 := ts[0], ts[len(ts)-1]
	if t1-t0 < eps {
		return
	}
	return ox + t0*vx, oy + t0*vy, ox + t1*vx, oy + t1*vy, true
}

func clipSegmentToRect(ax, ay, bx, by, rx0, ry0, rx1, ry1 float64) (float64, float64, float64, float64, bool) {
	dx := bx - ax
	dy := by - ay
	u0, u1 := 0.0, 1.0
	clipP := func(p, q float64) bool {
		if math.Abs(p) < 1e-15 {
			return q >= -1e-9
		}
		r := q / p
		if p < 0 {
			if r > u1 {
				return false
			}
			if r > u0 {
				u0 = r
			}
		} else {
			if r < u0 {
				return false
			}
			if r < u1 {
				u1 = r
			}
		}
		return true
	}
	if !clipP(-dx, ax-rx0) || !clipP(dx, rx1-ax) || !clipP(-dy, ay-ry0) || !clipP(dy, ry1-ay) {
		return 0, 0, 0, 0, false
	}
	if u1 < u0-1e-12 {
		return 0, 0, 0, 0, false
	}
	return ax + u0*dx, ay + u0*dy, ax + u1*dx, ay + u1*dy, true
}

// segmentRing is a butt or square-capped stroke rectangle. capA/capB extend
// the centerline along its tangent (0 is a butt cap, half the stroke width is
// a square cap).
func segmentRing(ax, ay, bx, by, hw, capA, capB float64) *canvas.Path {
	dx := bx - ax
	dy := by - ay
	length := math.Hypot(dx, dy)
	if length < 1e-9 || hw <= 0 {
		return &canvas.Path{}
	}
	if capA > 0 || capB > 0 {
		ux, uy := dx/length, dy/length
		ax -= ux * capA
		ay -= uy * capA
		bx += ux * capB
		by += uy * capB
		dx = bx - ax
		dy = by - ay
		length = math.Hypot(dx, dy)
		if length < 1e-9 {
			return &canvas.Path{}
		}
	}
	nx := (-dy / length) * hw
	ny := (dx / length) * hw
	return polygonCanvas([]path.Point{
		path.NewPoint(ax+nx, ay+ny),
		path.NewPoint(bx+nx, by+ny),
		path.NewPoint(bx-nx, by-ny),
		path.NewPoint(ax-nx, ay-ny),
	})
}

func rectPath(x0, y0, x1, y1 float64) *canvas.Path {
	return polygonCanvas([]path.Point{
		path.NewPoint(x0, y0),
		path.NewPoint(x1, y0),
		path.NewPoint(x1, y1),
		path.NewPoint(x0, y1),
	})
}

func onRectBoundary(px, py, x0, y0, x1, y1, tol float64) bool {
	return math.Abs(px-x0) <= tol || math.Abs(px-x1) <= tol || math.Abs(py-y0) <= tol || math.Abs(py-y1) <= tol
}

func endcapSliver(px, py, x0, y0, x1, y1, hw, tol float64) *canvas.Path {
	if hw < 1e-12 {
		return nil
	}
	onL := math.Abs(px-x0) <= tol
	onR := math.Abs(px-x1) <= tol
	onT := math.Abs(py-y0) <= tol
	onB := math.Abs(py-y1) <= tol
	switch {
	case onL && onT:
		return rectPath(x0, y0, x0+hw, y0+hw)
	case onR && onT:
		return rectPath(x1-hw, y0, x1, y0+hw)
	case onL && onB:
		return rectPath(x0, y1-hw, x0+hw, y1)
	case onR && onB:
		return rectPath(x1-hw, y1-hw, x1, y1)
	case onT:
		return rectPath(px-hw, y0, px+hw, y0+hw)
	case onB:
		return rectPath(px-hw, y1-hw, px+hw, y1)
	case onL:
		return rectPath(x0, py-hw, x0+hw, py+hw)
	case onR:
		return rectPath(x1-hw, py-hw, x1, py+hw)
	default:
		return nil
	}
}

// flushCutsToFrame opens a cut that stops just inside the frame out to the
// frame's inner edge. A connector whose centerline lands in the frame band
// otherwise leaves the neighboring openings short of that edge, so the left
// and right ends no longer share one line.
func flushCutsToFrame(cut *canvas.Path, x0, y0, x1, y1, hw float64) *canvas.Path {
	if cut == nil || cut.Empty() || hw < 1e-6 {
		return cut
	}
	snap := hw
	// Strips sit just inside the frame. A gap that starts inside a strip is an
	// opening stopped short of the frame; the stroke between it and the frame
	// is the lip to remove.
	strips := [][4]float64{
		{x0 + hw, y0 + hw, x0 + hw + snap, y1 - hw},
		{x1 - hw - snap, y0 + hw, x1 - hw, y1 - hw},
		{x0 + hw + snap, y0 + hw, x1 - hw - snap, y0 + hw + snap},
		{x0 + hw + snap, y1 - hw - snap, x1 - hw - snap, y1 - hw},
	}
	var trims []*canvas.Path
	for i, s := range strips {
		if s[2]-s[0] < 1e-6 || s[3]-s[1] < 1e-6 {
			continue
		}
		gaps := rectPath(s[0], s[1], s[2], s[3]).Not(cut)
		if gaps == nil || gaps.Empty() {
			continue
		}
		for _, sub := range gaps.Split() {
			if sub == nil || sub.Empty() {
				continue
			}
			b := sub.Bounds()
			if b.X1-b.X0 < 1e-4 || b.Y1-b.Y0 < 1e-4 {
				continue
			}
			switch i {
			case 0:
				if b.X0 <= s[0]+1e-4 {
					continue
				}
				trims = append(trims, rectPath(s[0], b.Y0, b.X0, b.Y1))
			case 1:
				if b.X1 >= s[2]-1e-4 {
					continue
				}
				trims = append(trims, rectPath(b.X1, b.Y0, s[2], b.Y1))
			case 2:
				if b.Y0 <= s[1]+1e-4 {
					continue
				}
				trims = append(trims, rectPath(b.X0, s[1], b.X1, b.Y0))
			default:
				if b.Y1 >= s[3]-1e-4 {
					continue
				}
				trims = append(trims, rectPath(b.X0, b.Y1, b.X1, s[3]))
			}
		}
	}
	if len(trims) == 0 {
		return cut
	}
	opened := cut.Not(unionCanvas(trims))
	if opened == nil || opened.Empty() {
		return cut
	}
	return opened
}

func perimeterBands(x0, y0, x1, y1, hw float64) []*canvas.Path {
	if hw < 1e-12 || x1-x0 <= 2*hw+1e-12 || y1-y0 <= 2*hw+1e-12 {
		return nil
	}
	return []*canvas.Path{
		rectPath(x0, y0, x1, y0+hw),
		rectPath(x0, y1-hw, x1, y1),
		rectPath(x0, y0+hw, x0+hw, y1-hw),
		rectPath(x1-hw, y0+hw, x1, y1-hw),
	}
}

func unionCanvas(parts []*canvas.Path) *canvas.Path {
	live := make([]*canvas.Path, 0, len(parts))
	for _, p := range parts {
		if p != nil && !p.Empty() {
			live = append(live, p)
		}
	}
	if len(live) == 0 {
		return &canvas.Path{}
	}
	for len(live) > 1 {
		next := make([]*canvas.Path, 0, (len(live)+1)/2)
		for i := 0; i < len(live); i += 2 {
			if i+1 >= len(live) {
				next = append(next, live[i])
				continue
			}
			next = append(next, live[i].Or(live[i+1]))
		}
		live = next
	}
	if live[0] == nil {
		return &canvas.Path{}
	}
	return live[0]
}

// appendCanvasExceptInnerRect drops a subpath that is only the inner margin
// rectangle. The lattice perimeter band emits that outline; the fretwork is
// the remaining boundary.
func appendCanvasExceptInnerRect(d *path.Draw, cp *canvas.Path, x0, y0, x1, y1 float64) {
	if cp == nil || cp.Empty() {
		return
	}
	tol := math.Max(0.001, math.Max(1e-6, 1e-5*math.Max(1, math.Max(math.Abs(x1), math.Abs(y1)))))
	for _, sub := range cp.Split() {
		if subpathIsAxisRect(sub, x0, y0, x1, y1, tol) {
			continue
		}
		appendCanvas(d, sub)
	}
}

func subpathIsAxisRect(p *canvas.Path, x0, y0, x1, y1, tol float64) bool {
	if p == nil || p.Empty() {
		return false
	}
	s := p.Scanner()
	n := 0
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for s.Scan() {
		cmd := s.Cmd()
		if cmd == canvas.CloseCmd {
			continue
		}
		if cmd != canvas.MoveToCmd && cmd != canvas.LineToCmd {
			return false
		}
		x, y := s.End().X, s.End().Y
		n++
		minX = math.Min(minX, x)
		minY = math.Min(minY, y)
		maxX = math.Max(maxX, x)
		maxY = math.Max(maxY, y)
		onEdge := math.Abs(x-x0) <= tol || math.Abs(x-x1) <= tol || math.Abs(y-y0) <= tol || math.Abs(y-y1) <= tol
		if !onEdge {
			return false
		}
	}
	if n < 4 || n > 8 {
		return false
	}
	return math.Abs(minX-x0) <= tol && math.Abs(minY-y0) <= tol && math.Abs(maxX-x1) <= tol && math.Abs(maxY-y1) <= tol
}
