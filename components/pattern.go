package components

import (
	"fmt"
	"math"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/tdewolff/canvas"
)

// PatternComponentFactory builds Pattern Studio kinds. Closed cutouts,
// tile grids, overlapping circles, and voronoi are finished paths. Hatch,
// wave, and lattice kinds emit centerlines; the custom components stroke
// those with framed_stroke.
type PatternComponentFactory struct{}

type PatternComponent struct {
	*dom.BasicComponent
}

func (f PatternComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *dom.DocumentContext) (dom.Component, error) {
	bc := dom.AppContext().MakeBasicComponent(mp)
	return &PatternComponent{BasicComponent: bc}, nil
}

func (f PatternComponentFactory) ComponentTypes() []string {
	return []string{"pattern"}
}

func (pc *PatternComponent) Render(ctx dom.RenderContext) (path.Path, dom.RenderContext, error) {
	pc.RenderStart(ctx)
	attr := pc.Attr()
	kind := attr.MustString("kind", "")
	width := attr.MustFloat64("width", 0)
	height := attr.MustFloat64("height", 0)
	margin := attr.MustFloat64("margin", 0)
	clip := attr.MustBool("clip_to_border", false)
	if width <= 0 || height <= 0 {
		return nil, ctx, fmt.Errorf("pattern requires width and height")
	}
	if margin < 0 {
		margin = 0
	}

	var p path.Path
	switch kind {
	case "hexagon_grid":
		p = hexagonGrid(width, height, margin, attr.MustFloat64("circumradius", 0.35), attr.MustFloat64("gap", 0.2), clip)
	case "hex_circles":
		p = hexCircles(width, height, margin, attr.MustFloat64("radius", 0.35), attr.MustFloat64("gap", 0.2), clip)
	case "lens_grid":
		p = lensGrid(width, height, margin, lensParams{
			length: attr.MustFloat64("lens_length", 0.72),
			width:  attr.MustFloat64("lens_width", 0.44),
			lean:   attr.MustFloat64("lean_deg", 36),
			gapX:   attr.MustFloat64("gap_x", 0.13),
			gapY:   attr.MustFloat64("gap_y", 0.11),
		}, clip)
	case "hatch":
		p = hatchPath(width, height, margin, hatchParams{
			angleDeg:    attr.MustFloat64("angle_deg", 45),
			spacing:     attr.MustFloat64("spacing", 0.5),
			phase:       attr.MustFloat64("phase", 0),
			cross:       attr.MustBool("cross", false),
			crossOffset: attr.MustFloat64("cross_angle", 90),
		})
	case "wave_lattice":
		p = waveLattice(width, height, margin, waveParams{
			amplitude:  attr.MustFloat64("amplitude", 0.15),
			wavelength: attr.MustFloat64("wavelength", 1.2),
			thickness:  attr.MustFloat64("ribbon_thickness", 0.25),
			gap:        attr.MustFloat64("ribbon_gap", 0.3),
			phaseStep:  attr.MustFloat64("phase_step", 0.4),
			samples:    int(math.Floor(attr.MustFloat64("samples", 48))),
		}, clip)
	case "staggered_lattice_lines":
		p = staggeredLattice(width, height, margin, latticeParams{
			rectWidth:  attr.MustFloat64("rect_width", 0.5),
			rectHeight: attr.MustFloat64("rect_height", 1.25),
			vGapRatio:  attr.MustFloat64("v_gap_ratio", 0.4),
			hGapRatio:  attr.MustFloat64("h_gap_ratio", 0.95),
		}, clip)
	case "overlapping_circles":
		p = overlappingCircles(width, height, margin, overlapParams{
			radiusMin:   attr.MustFloat64("radius_min", 0.25),
			radiusMax:   attr.MustFloat64("radius_max", 0.6),
			strokeWidth: attr.MustFloat64("stroke_width", 0.08),
			count:       int(math.Floor(attr.MustFloat64("count", 30))),
			seed:        int32(attr.MustFloat64("seed", 17)),
		})
	case "voronoi":
		p = voronoiSlots(width, height, margin, voronoiParams{
			siteCount:  int(math.Floor(attr.MustFloat64("site_count", 32))),
			seed:       int32(attr.MustFloat64("seed", 7)),
			thickness:  attr.MustFloat64("thickness", 0.25),
			minSiteGap: attr.MustFloat64("min_site_gap", 0.2),
		})
	case "framed_stroke":
		var err error
		p, err = patternFramedStroke(width, height, margin, framedStrokeParams{
			motif:     attr.MustString("motif", ""),
			thickness: attr.MustFloat64("thickness", 0.1),
			fit:       attr.MustBool("fit", true),
			frame:     attr.MustBool("frame", true),
			join:      attr.MustString("join", "miter"),
		})
		if err != nil {
			return nil, ctx, err
		}
	case "tile_grid":
		var err error
		p, err = tileGrid(width, height, margin, tileParams{
			shape:    attr.MustString("shape", "quatrefoil"),
			scale:    attr.MustFloat64("scale", 0.4),
			gapX:     attr.MustFloat64("gap_x", 0.2),
			gapY:     attr.MustFloat64("gap_y", 0.2),
			stagger:  attr.MustFloat64("stagger", 0.5),
			arm:      attr.MustFloat64("arm", 0),
			corner:   attr.MustFloat64("corner_radius", 0.04),
			rotation: attr.MustFloat64("rotation_deg", 0),
			motif:    attr.MustString("motif", ""),
		}, clip)
		if err != nil {
			return nil, ctx, err
		}
	default:
		return nil, ctx, fmt.Errorf("unknown pattern kind %q", kind)
	}
	if p == nil || len(p.Segments()) == 0 {
		// PathCursor panics on an empty segment list. A lone move keeps an
		// empty grid renderable without adding a cut.
		empty := path.NewDraw()
		empty.MoveTo(path.NewPoint(0, 0))
		p = empty.Path()
	}
	return pc.HandleTransforms(pc, p, ctx)
}

type point struct{ x, y float64 }

type box struct {
	minX, minY, maxX, maxY float64
	ok                     bool
}

func (b *box) add(x0, y0, x1, y1 float64) {
	if !b.ok {
		b.minX, b.minY, b.maxX, b.maxY = x0, y0, x1, y1
		b.ok = true
		return
	}
	b.minX = math.Min(b.minX, x0)
	b.minY = math.Min(b.minY, y0)
	b.maxX = math.Max(b.maxX, x1)
	b.maxY = math.Max(b.maxY, y1)
}

func centeringOffset(xmin, ymin, xmax, ymax float64, b box) (ox, oy float64) {
	if !b.ok {
		return 0, 0
	}
	iw := xmax - xmin
	ih := ymax - ymin
	w := b.maxX - b.minX
	h := b.maxY - b.minY
	return xmin + (iw-w)/2 - b.minX, ymin + (ih-h)/2 - b.minY
}

func addClosedPolygon(d *path.Draw, pts []path.Point) {
	if len(pts) < 3 {
		return
	}
	d.MoveTo(pts[0])
	for i := 1; i < len(pts); i++ {
		d.LineTo(pts[i])
	}
	d.LineTo(pts[0])
}

func hexagonPoints(cx, cy, r float64) []path.Point {
	pts := make([]path.Point, 6)
	for k := 0; k < 6; k++ {
		t := -math.Pi/2 + float64(k)*math.Pi/3
		pts[k] = path.NewPoint(cx+r*math.Cos(t), cy+r*math.Sin(t))
	}
	return pts
}

func hexagonGrid(width, height, margin, circumradius, gap float64, clip bool) path.Path {
	d := path.NewDraw()
	if circumradius <= 0 {
		return d.Path()
	}
	pitchX := math.Sqrt(3)*circumradius + gap
	pitchY := 1.5*circumradius + gap
	if pitchX <= 1e-6 || pitchY <= 1e-6 {
		return d.Path()
	}
	hx := (circumradius * math.Sqrt(3)) / 2
	xInset := margin + hx
	xmin, ymin := margin, margin
	xmax, ymax := width-margin, height-margin
	if xmax <= xmin || ymax <= ymin {
		return d.Path()
	}

	var centers []point
	if !clip {
		for row := 0; row < 10000; row++ {
			cy := margin + circumradius + float64(row)*pitchY
			if cy+circumradius > height-margin {
				break
			}
			xOff := 0.0
			if row%2 == 1 {
				xOff = pitchX / 2
			}
			for col := 0; col < 10000; col++ {
				cx := xInset + float64(col)*pitchX + xOff
				if cx+hx > width-margin {
					break
				}
				centers = append(centers, point{cx, cy})
			}
		}
	} else {
		rowMin := int(math.Floor((ymin-margin-circumradius-pitchY)/pitchY)) - 1
		rowMax := int(math.Ceil((ymax-margin-circumradius+pitchY)/pitchY)) + 1
		for row := rowMin; row <= rowMax; row++ {
			cy := margin + circumradius + float64(row)*pitchY
			xOff := float64(row%2) * (pitchX / 2)
			cx0 := xInset + xOff
			if cy-circumradius > ymax+1e-6 || cy+circumradius < ymin-1e-6 {
				continue
			}
			colMin := int(math.Floor((xmin-cx0-hx-pitchX)/pitchX)) - 1
			colMax := int(math.Ceil((xmax-cx0+hx+pitchX)/pitchX)) + 1
			for col := colMin; col <= colMax; col++ {
				cx := cx0 + float64(col)*pitchX
				if cx-hx > xmax+1e-6 || cx+hx < xmin-1e-6 {
					continue
				}
				centers = append(centers, point{cx, cy})
			}
		}
	}

	var b box
	for _, c := range centers {
		b.add(c.x-hx, c.y-circumradius, c.x+hx, c.y+circumradius)
	}
	ox, oy := centeringOffset(xmin, ymin, xmax, ymax, b)
	for _, c := range centers {
		pts := hexagonPoints(c.x+ox, c.y+oy, circumradius)
		if clip {
			appendClipped(d, polygonCanvas(pts), margin, width, height)
		} else {
			addClosedPolygon(d, pts)
		}
	}
	return d.Path()
}

func hexCircles(width, height, margin, radius, gap float64, clip bool) path.Path {
	d := path.NewDraw()
	if radius <= 0 {
		return d.Path()
	}
	dx := 2*radius + gap
	dy := dx * (math.Sqrt(3) / 2)
	if dx <= 1e-6 || dy <= 1e-6 {
		return d.Path()
	}
	xmin, ymin := margin, margin
	xmax, ymax := width-margin, height-margin
	if xmax <= xmin || ymax <= ymin {
		return d.Path()
	}

	var centers []point
	if !clip {
		for row := 0; row < 10000; row++ {
			y := margin + radius + float64(row)*dy
			if y+radius > height-margin {
				break
			}
			xOff := 0.0
			if row%2 == 1 {
				xOff = dx / 2
			}
			for col := 0; col < 10000; col++ {
				cx := margin + radius + float64(col)*dx + xOff
				if cx+radius > width-margin {
					break
				}
				centers = append(centers, point{cx, y})
			}
		}
	} else {
		rowMin := int(math.Floor((ymin-margin-2*radius)/dy)) - 1
		rowMax := int(math.Ceil((ymax-margin+radius)/dy)) + 1
		for row := rowMin; row <= rowMax; row++ {
			y := margin + radius + float64(row)*dy
			xOff := float64(row%2) * (dx / 2)
			if y+radius < ymin-1e-6 || y-radius > ymax+1e-6 {
				continue
			}
			colMin := int(math.Floor((xmin-margin-radius-xOff-radius)/dx)) - 1
			colMax := int(math.Ceil((xmax-margin-radius-xOff+radius)/dx)) + 1
			for col := colMin; col <= colMax; col++ {
				cx := margin + radius + float64(col)*dx + xOff
				if cx+radius < xmin-1e-6 || cx-radius > xmax+1e-6 {
					continue
				}
				centers = append(centers, point{cx, y})
			}
		}
	}

	var b box
	for _, c := range centers {
		b.add(c.x-radius, c.y-radius, c.x+radius, c.y+radius)
	}
	ox, oy := centeringOffset(xmin, ymin, xmax, ymax, b)
	for _, c := range centers {
		cx, cy := c.x+ox, c.y+oy
		if clip {
			appendClipped(d, circleCanvas(cx, cy, radius), margin, width, height)
		} else {
			d.MoveTo(path.NewPoint(cx-radius, cy-radius))
			d.Circle(radius)
		}
	}
	return d.Path()
}

type lensParams struct {
	length, width, lean, gapX, gapY float64
}

func lensGrid(panelW, panelH, margin float64, p lensParams, clip bool) path.Path {
	d := path.NewDraw()
	length := math.Max(1e-6, p.length)
	lensW := math.Max(1e-6, p.width)
	lean := p.lean
	if math.IsNaN(lean) || math.IsInf(lean, 0) {
		lean = 0
	}
	hw, hh := lensHalfExtents(length, lensW, lean)
	pitchX := hw*2 + p.gapX
	pitchY := hh*2 + p.gapY
	if pitchX <= 1e-6 || pitchY <= 1e-6 {
		return d.Path()
	}
	xmin, ymin := margin, margin
	xmax, ymax := panelW-margin, panelH-margin
	innerW, innerH := xmax-xmin, ymax-ymin
	if innerW <= 1e-9 || innerH <= 1e-9 {
		return d.Path()
	}

	type place struct {
		cx, cy float64
		row    int
		col    int
	}
	var places []place
	if !clip {
		if hw*2 > innerW+1e-9 || hh*2 > innerH+1e-9 {
			return d.Path()
		}
		cols := int(math.Floor((innerW-hw*2)/pitchX+1e-9)) + 1
		rows := int(math.Floor((innerH-hh*2)/pitchY+1e-9)) + 1
		for row := 0; row < rows && row < 10000; row++ {
			cy := margin + hh + float64(row)*pitchY
			for col := 0; col < cols && col < 10000; col++ {
				places = append(places, place{margin + hw + float64(col)*pitchX, cy, row, col})
			}
		}
	} else {
		rowMin := int(math.Floor((ymin-(margin+hh)-hh)/pitchY)) - 1
		rowMax := int(math.Ceil((ymax-(margin+hh)+hh)/pitchY)) + 1
		colMin0 := int(math.Floor((xmin-(margin+hw)-hw)/pitchX)) - 1
		colMax0 := int(math.Ceil((xmax-(margin+hw)+hw)/pitchX)) + 1
		for row := rowMin; row <= rowMax; row++ {
			cy := margin + hh + float64(row)*pitchY
			if cy+hh < ymin-1e-6 || cy-hh > ymax+1e-6 {
				continue
			}
			for col := colMin0; col <= colMax0; col++ {
				cx := margin + hw + float64(col)*pitchX
				if cx+hw < xmin-1e-6 || cx-hw > xmax+1e-6 {
					continue
				}
				places = append(places, place{cx, cy, row, col})
			}
		}
	}

	var b box
	for _, pl := range places {
		b.add(pl.cx-hw, pl.cy-hh, pl.cx+hw, pl.cy+hh)
	}
	ox, oy := centeringOffset(xmin, ymin, xmax, ymax, b)
	for _, pl := range places {
		cx, cy := pl.cx+ox, pl.cy+oy
		cell := lean
		if (pl.row+pl.col)%2 != 0 {
			cell = -lean
		}
		if clip {
			ring := lensRing(cx, cy, length, lensW, cell, 24)
			if len(ring) < 4 {
				continue
			}
			appendClipped(d, polygonCanvas(ring), margin, panelW, panelH)
		} else if length <= lensW+1e-6 {
			r := lensW / 2
			d.MoveTo(path.NewPoint(cx-r, cy-r))
			d.Circle(r)
		} else {
			appendCanvas(d, lensArcPath(cx, cy, length, lensW, cell))
		}
	}
	return d.Path()
}

func lensMetrics(length, width float64) (r, halfChord, centerDist float64, ok bool) {
	w := math.Max(1e-6, width)
	l := math.Max(w+1e-6, length)
	r = (l*l + w*w) / (4 * w)
	centerDist = (l*l - w*w) / (2 * w)
	if r <= 0 || centerDist <= 0 {
		return 0, 0, 0, false
	}
	return r, l / 2, centerDist, true
}

func rotateClockwiseYDown(x, y, deg float64) (float64, float64) {
	t := deg * math.Pi / 180
	c, s := math.Cos(t), math.Sin(t)
	return x*c - y*s, x*s + y*c
}

func lensHalfExtents(length, width, leanDeg float64) (hw, hh float64) {
	ring := lensRing(0, 0, length, width, leanDeg, 32)
	maxX, maxY := 0.0, 0.0
	for _, p := range ring {
		maxX = math.Max(maxX, math.Abs(p.X))
		maxY = math.Max(maxY, math.Abs(p.Y))
	}
	return math.Max(1e-6, maxX), math.Max(1e-6, maxY)
}

func lensRing(cx, cy, length, width, leanDeg float64, seg int) []path.Point {
	if length <= width+1e-6 {
		return circleRing(cx, cy, math.Max(width, length)/2, maxInt(16, seg*2))
	}
	r, halfChord, centerDist, ok := lensMetrics(length, width)
	if !ok {
		return nil
	}
	a := math.Atan2(halfChord, centerDist/2)
	n := maxInt(8, seg)
	var ring []path.Point
	push := func(lx, ly float64) {
		rx, ry := rotateClockwiseYDown(lx, ly, leanDeg)
		pt := path.NewPoint(cx+rx, cy+ry)
		if len(ring) > 0 {
			prev := ring[len(ring)-1]
			if math.Hypot(prev.X-pt.X, prev.Y-pt.Y) < 1e-9 {
				return
			}
		}
		ring = append(ring, pt)
	}
	for i := 0; i <= n; i++ {
		t := -a + (2*a*float64(i))/float64(n)
		push(-centerDist/2+r*math.Cos(t), r*math.Sin(t))
	}
	for i := 1; i <= n; i++ {
		t := math.Pi - a + (2*a*float64(i))/float64(n)
		push(centerDist/2+r*math.Cos(t), r*math.Sin(t))
	}
	return ring
}

func circleRing(cx, cy, r float64, segments int) []path.Point {
	n := maxInt(8, segments)
	pts := make([]path.Point, 0, n)
	for i := 0; i < n; i++ {
		t := (float64(i) / float64(n)) * 2 * math.Pi
		pts = append(pts, path.NewPoint(cx+r*math.Cos(t), cy+r*math.Sin(t)))
	}
	return pts
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func lensArcPath(cx, cy, length, width, leanDeg float64) *canvas.Path {
	r, halfChord, _, ok := lensMetrics(length, width)
	if !ok {
		return &canvas.Path{}
	}
	tx, ty := rotateClockwiseYDown(0, -halfChord, leanDeg)
	bx, by := rotateClockwiseYDown(0, halfChord, leanDeg)
	x0, y0 := cx+tx, cy+ty
	x1, y1 := cx+bx, cy+by
	cp := &canvas.Path{}
	cp.MoveTo(x0, y0)
	cp.ArcTo(r, r, 0, false, true, x1, y1)
	cp.ArcTo(r, r, 0, false, true, x0, y0)
	return cp
}

func circleCanvas(cx, cy, r float64) *canvas.Path {
	cp := &canvas.Path{}
	cp.MoveTo(cx+r, cy)
	cp.ArcTo(r, r, 0, true, true, cx-r, cy)
	cp.ArcTo(r, r, 0, true, true, cx+r, cy)
	return cp
}

func polygonCanvas(pts []path.Point) *canvas.Path {
	cp := &canvas.Path{}
	if len(pts) == 0 {
		return cp
	}
	cp.MoveTo(pts[0].X, pts[0].Y)
	for i := 1; i < len(pts); i++ {
		cp.LineTo(pts[i].X, pts[i].Y)
	}
	cp.Close()
	return cp
}

func innerRectCanvas(margin, width, height float64) *canvas.Path {
	cp := &canvas.Path{}
	x1, y1 := width-margin, height-margin
	cp.MoveTo(margin, margin)
	cp.LineTo(x1, margin)
	cp.LineTo(x1, y1)
	cp.LineTo(margin, y1)
	cp.Close()
	return cp
}

func appendClipped(d *path.Draw, shape *canvas.Path, margin, width, height float64) {
	if shape == nil || shape.Empty() {
		return
	}
	if width-margin <= margin+1e-9 || height-margin <= margin+1e-9 {
		return
	}
	// And() flattens every curve to the canvas tolerance, so a circle becomes
	// a polygon. Clip the cubics instead and keep the arc.
	clipped := clipCurvesToRect(shape, margin, margin, width-margin, height-margin)
	appendCanvas(d, clipped)
}

// appendCanvas copies a canvas path onto d. Close becomes a line back to the
// subpath start so a clipped hole stays a closed cut. PathFromCanvas drops
// Close to match the legacy SVG parser, which is the wrong behavior here.
func appendCanvas(d *path.Draw, cp *canvas.Path) {
	if cp == nil || cp.Empty() {
		return
	}
	cp = cp.ReplaceArcs()
	s := cp.Scanner()
	var subStart path.Point
	hasSub := false
	for s.Scan() {
		end := path.NewPoint(s.End().X, s.End().Y)
		switch s.Cmd() {
		case canvas.MoveToCmd:
			d.MoveTo(end)
			subStart = end
			hasSub = true
		case canvas.LineToCmd:
			d.LineTo(end)
		case canvas.QuadToCmd:
			d.QCurveTo(path.NewPoint(s.CP1().X, s.CP1().Y), end)
		case canvas.CubeToCmd:
			d.CurveTo(
				path.NewPoint(s.CP1().X, s.CP1().Y),
				path.NewPoint(s.CP2().X, s.CP2().Y),
				end,
			)
		case canvas.CloseCmd:
			if hasSub && math.Hypot(subStart.X-end.X, subStart.Y-end.Y) > 1e-9 {
				d.LineTo(end)
			} else if hasSub && math.Hypot(d.CurrentPosition().X-subStart.X, d.CurrentPosition().Y-subStart.Y) > 1e-9 {
				d.LineTo(subStart)
			}
		}
	}
}
