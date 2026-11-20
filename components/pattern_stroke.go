package components

import (
	"math"
	"sort"

	"github.com/dustismo/heavyfishdesign/path"
	"github.com/tdewolff/canvas"
)

// Stroke-union kinds follow the Pattern Studio generators: centerlines (or
// closed ribbons / annuli), expand, boolean-union with canvas, then the same
// output formulation (merged slots, clipped waves, panel minus bands).

type hatchParams struct {
	angleDeg    float64
	spacing     float64
	thickness   float64
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
	thickness  float64
}

type overlapParams struct {
	radiusMin   float64
	radiusMax   float64
	strokeWidth float64
	count       int
	seed        int32
	segments    int
}

func hatchPath(w, h, m float64, p hatchParams) path.Path {
	d := path.NewDraw()
	sp := math.Max(1e-6, p.spacing)
	sw := math.Max(1e-6, p.thickness)
	xmin, ymin := m, m
	xmax, ymax := w-m, h-m
	if xmax <= xmin || ymax <= ymin {
		return d.Path()
	}
	cx := (xmin + xmax) / 2
	cy := (ymin + ymax) / 2
	reach := 0.5*math.Hypot(xmax-xmin, ymax-ymin) + sw
	kMax := int(math.Ceil((reach + math.Abs(p.phase)) / sp))
	hw := sw / 2

	var polys []*canvas.Path
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
			ring := segmentRing(ax, ay, bx, by, hw, 0, 0)
			if ring != nil && !ring.Empty() {
				polys = append(polys, ring)
			}
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
	appendCanvas(d, unionCanvas(polys))
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
	innerH := math.Max(0, h-2*m)
	envelope := thick + 2*amp
	slack := innerH - envelope
	count := 1
	if slack > 0 && pitch > 1e-9 {
		count = 1 + int(math.Floor(slack/pitch+1e-9))
		for count > 1 && float64(count-1)*pitch+envelope > innerH+1e-6 {
			count--
		}
	}
	xmin, ymin := m, m
	xmax, ymax := w-m, h-m
	type ribbon struct {
		yCenter float64
		phase   float64
	}
	var ribbons []ribbon
	var bb box
	for i := 0; i < count; i++ {
		yCenter := m + (float64(i)+0.5)*pitch - thick/2
		if !clip && (yCenter+thick+amp > h-m || yCenter-amp < m) {
			continue
		}
		ribbons = append(ribbons, ribbon{yCenter: yCenter, phase: float64(i) * p.phaseStep})
		bb.add(xmin, yCenter-amp, xmax, yCenter+thick+amp)
	}
	ox, oy := centeringOffset(xmin, ymin, xmax, ymax, bb)
	x0, x1 := xmin+ox, xmax+ox
	for _, rib := range ribbons {
		ring := waveRibbon(x0, x1, rib.yCenter+oy, amp, wl, thick, rib.phase, samples)
		if clip {
			appendClipped(d, ring, m, w, h)
		} else {
			appendCanvas(d, ring)
		}
	}
	return d.Path()
}

func waveRibbon(x0, x1, yCenter, amp, wavelength, thick, phase float64, samples int) *canvas.Path {
	if x1 <= x0 || thick <= 0 {
		return &canvas.Path{}
	}
	n := samples
	if n < 8 {
		n = 8
	}
	span := x1 - x0
	pts := make([]path.Point, 0, 2*n+2)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		x := x0 + t*span
		y := yCenter + amp*math.Sin(2*math.Pi*(x-x0)/wavelength+phase)
		pts = append(pts, path.NewPoint(x, y))
	}
	for i := n; i >= 0; i-- {
		t := float64(i) / float64(n)
		x := x0 + t*span
		y := yCenter + thick + amp*math.Sin(2*math.Pi*(x-x0)/wavelength+phase)
		pts = append(pts, path.NewPoint(x, y))
	}
	return polygonCanvas(pts)
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
			addClipped(tx, ty, tx+rw, ty)
			addClipped(tx+rw, ty, tx+rw, ty+rh)
			addClipped(tx+rw, ty+rh, tx, ty+rh)
			addClipped(tx, ty+rh, tx, ty)
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

	return expandStroke(segments, p.thickness, m, w, h)
}

// expandStroke clips centerlines to the inner margin, expands each one by
// thickness, unions the ribbons, and opens the margin the same way the
// staggered lattice does: butt caps on the border, endcap slivers, a
// perimeter band, then the bare inner rectangle is dropped.
func expandStroke(segments []seg, thickness, margin, width, height float64) path.Path {
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
	if bands := perimeterBands(innerX0, innerY0, innerX1, innerY1, hw); len(bands) > 0 {
		geom = unionCanvas(append([]*canvas.Path{geom}, bands...))
	}
	appendCanvasExceptInnerRect(d, geom, innerX0, innerY0, innerX1, innerY1)
	return d.Path()
}

// strokeCenterline thickens a drawing with canvas's stroke: both sides are
// offset and corners are joined. Open ends that meet the margin are extended
// through that edge first, so the stroke crosses the frame on the line's
// angle. The result is clipped to the inner margin and unioned with the frame.
func strokeCenterline(centerline path.Path, thickness, margin, width, height float64) path.Path {
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
	stroked := path.ToCanvasPath(extended).Stroke(math.Max(1e-6, thickness), canvas.ButtCap, canvas.MiterJoin, 0.001)
	if stroked == nil || stroked.Empty() {
		return d.Path()
	}
	clipped := stroked.And(innerRectCanvas(margin, width, height))
	if clipped == nil || clipped.Empty() {
		return d.Path()
	}
	if bands := perimeterBands(innerX0, innerY0, innerX1, innerY1, hw); len(bands) > 0 {
		clipped = unionCanvas(append([]*canvas.Path{clipped}, bands...))
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
	target := p.count
	if target < 0 {
		target = 0
	}
	if target == 0 {
		appendCanvas(d, panel)
		return d.Path()
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
	half := sw / 2
	segN := p.segments
	if segN < 8 {
		segN = 8
	}
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
	var bands []*canvas.Path
	for i, c := range placed {
		if !neighbor[i] {
			continue
		}
		outerR := c.r + half
		innerR := c.r - half
		outer := polygonCanvas(circleRing(c.cx, c.cy, outerR, segN))
		if innerR <= 1e-6 {
			bands = append(bands, outer)
			continue
		}
		bands = append(bands, outer.Not(polygonCanvas(circleRing(c.cx, c.cy, innerR, segN))))
	}
	if len(bands) == 0 {
		appendCanvas(d, panel)
		return d.Path()
	}
	merged := unionCanvas(bands)
	if merged == nil || merged.Empty() {
		appendCanvas(d, panel)
		return d.Path()
	}
	wood := panel.Not(merged)
	if wood == nil || wood.Empty() {
		appendCanvas(d, panel)
		return d.Path()
	}
	appendCanvas(d, wood)
	return d.Path()
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
