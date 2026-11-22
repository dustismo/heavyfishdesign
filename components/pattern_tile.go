package components

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/dustismo/heavyfishdesign/path"
	"github.com/tdewolff/canvas"
)

// tileParams is the tile_grid pattern. Pitch comes from the motif bounding
// box times scale, plus the gap. A quatrefoil is drawn smaller than that
// pitch: its arm defaults to 92% of the largest radius that still fits.
type tileParams struct {
	shape    string
	scale    float64
	gapX     float64
	gapY     float64
	stagger  float64
	arm      float64
	corner   float64
	rotation float64
	motif    string
}

var svgPathAttr = regexp.MustCompile(`(?i)\bd\s*=\s*(?:"([^"]*)"|'([^']*)')`)

func tileGrid(width, height, margin float64, p tileParams, clip bool) (path.Path, error) {
	d := path.NewDraw()
	xmin, ymin := margin, margin
	xmax, ymax := width-margin, height-margin
	if xmax <= xmin || ymax <= ymin {
		return d.Path(), nil
	}
	shape := strings.TrimSpace(p.shape)
	if shape == "" {
		shape = "quatrefoil"
	}
	bw, bh, unit, err := tileUnit(shape, p.rotation*math.Pi/180, p.motif)
	if err != nil {
		return nil, err
	}
	if bw <= 0 || bh <= 0 {
		return d.Path(), nil
	}

	scale := math.Max(1e-6, p.scale)
	brickW := math.Max(1e-9, bw*scale)
	brickH := math.Max(1e-9, bh*scale)
	cellW := brickW + p.gapX
	cellH := brickH + p.gapY
	uw, uh := xmax-xmin, ymax-ymin
	cols, rows := 1, 1
	if cellW > 1e-12 && cellH > 1e-12 {
		cols = maxInt(1, int(math.Floor((uw+p.gapX)/cellW+1e-9)))
		rows = maxInt(1, int(math.Floor((uh+p.gapY)/cellH+1e-9)))
	} else {
		cellW, cellH = 1e-9, 1e-9
	}
	// brick is the drawn slot inside the pitch. The generator above rebuilds
	// it as pitch minus gap so a negative gap overlaps the next tile.
	brickW = cellW - p.gapX
	brickH = cellH - p.gapY
	if brickW <= 0 || brickH <= 0 {
		return d.Path(), nil
	}

	sx, sy := scale, scale
	switch shape {
	case "quatrefoil":
		maxArm := math.Min(brickW, brickH)/4 - 1e-6
		armIn := maxArm * 0.92
		if p.arm > 0 {
			armIn = p.arm
		}
		arm := math.Max(1e-6, math.Min(armIn, maxArm))
		sx, sy = arm, arm
	case "rounded_rect":
		sx, sy = brickW, brickH
	}

	placements := tilePlacements(width, height, margin, p.gapX, p.gapY, cellW, cellH, brickW, brickH, p.stagger, cols, rows, clip)
	var b box
	halfW := (bw * sx) / 2
	halfH := (bh * sy) / 2
	for _, c := range placements {
		b.add(c.x-halfW, c.y-halfH, c.x+halfW, c.y+halfH)
	}
	ox, oy := centeringOffset(xmin, ymin, xmax, ymax, b)

	var cx0, cy0 float64
	if unit != nil {
		bounds := unit.Bounds()
		cx0 = (bounds.X0 + bounds.X1) / 2
		cy0 = (bounds.Y0 + bounds.Y1) / 2
	}
	rxMotif := 0.0
	if shape == "rounded_rect" {
		rxMotif = math.Min(0.5, math.Max(0, p.corner)/scale)
	}
	rot := p.rotation * math.Pi / 180
	for _, c := range placements {
		tcx, tcy := c.x+ox, c.y+oy
		var shapePath *canvas.Path
		switch shape {
		case "quatrefoil":
			shapePath = quatrefoilCanvas(tcx, tcy, sx)
		case "circle":
			shapePath = circleCanvas(tcx, tcy, 0.5*sx)
		case "rounded_rect":
			shapePath = roundedRectCanvas(tcx, tcy, sx, sy, rxMotif*sx, rxMotif*sy)
		case "diamond":
			shapePath = diamondCanvas(tcx, tcy, 0.5*sx, 0.5*sy, rot)
		default:
			shapePath = stampMotif(unit, cx0, cy0, sx, sy, tcx, tcy)
		}
		if clip {
			appendClipped(d, shapePath, margin, width, height)
		} else {
			appendCanvas(d, shapePath)
		}
	}
	return d.Path(), nil
}

func tileUnit(shape string, rot float64, motif string) (bw, bh float64, unit *canvas.Path, err error) {
	switch shape {
	case "circle", "rounded_rect":
		return 1, 1, nil, nil
	case "quatrefoil":
		return 4, 4, nil, nil
	case "diamond":
		bw, bh = rhombusBox(0.5, 0.5, rot)
		return bw, bh, nil, nil
	case "custom_svg":
		data, err := motifPathData(motif)
		if err != nil || data == "" {
			return 0, 0, nil, err
		}
		unit, err = canvas.ParseSVGPath(data)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("tile_grid motif: %w", err)
		}
		if unit.Empty() {
			return 0, 0, nil, nil
		}
		bounds := unit.Bounds()
		return math.Max(1e-9, bounds.X1-bounds.X0), math.Max(1e-9, bounds.Y1-bounds.Y0), unit, nil
	default:
		return 0, 0, nil, fmt.Errorf("unknown tile shape %q", shape)
	}
}

func motifPathData(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if !strings.Contains(raw, "<") {
		return raw, nil
	}
	matches := svgPathAttr.FindAllStringSubmatch(raw, -1)
	var b strings.Builder
	for _, m := range matches {
		d := m[1]
		if d == "" {
			d = m[2]
		}
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(d)
	}
	if b.Len() == 0 {
		return "", fmt.Errorf("tile_grid motif has no path data")
	}
	return b.String(), nil
}

func tilePlacements(width, height, margin, gapX, gapY, cellW, cellH, brickW, brickH, stagger float64, cols, rows int, clip bool) []point {
	sf := stagger
	if sf < 0 {
		sf = 0
	}
	if sf > 1 {
		sf = 1
	}
	rowShift := sf * cellW
	xmin, ymin := margin, margin
	xmax, ymax := width-margin, height-margin
	j0, j1 := 0, rows
	i0, i1 := 0, cols
	if clip {
		j0, j1 = -1, rows+1
		i0, i1 = -2, cols+2
	}
	var out []point
	for j := j0; j < j1; j++ {
		// Negative rows use the same remainder as the studio (sign follows j).
		rowOff := float64(j%2) * rowShift
		for i := i0; i < i1; i++ {
			xLeft := margin + float64(i)*cellW + rowOff + gapX/2
			yTop := margin + float64(j)*cellH + gapY/2
			if !clip && (xLeft+brickW > width-margin+1e-9 || yTop+brickH > height-margin+1e-9) {
				continue
			}
			if clip && (xLeft+brickW < xmin-1e-6 || xLeft > xmax+1e-6 || yTop+brickH < ymin-1e-6 || yTop > ymax+1e-6) {
				continue
			}
			out = append(out, point{xLeft + brickW/2, yTop + brickH/2})
		}
	}
	return out
}

func rhombusBox(ah, av, rot float64) (bw, bh float64) {
	c, s := math.Cos(rot), math.Sin(rot)
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, p := range []point{{0, -av}, {ah, 0}, {0, av}, {-ah, 0}} {
		x := p.x*c - p.y*s
		y := p.x*s + p.y*c
		minX, maxX = math.Min(minX, x), math.Max(maxX, x)
		minY, maxY = math.Min(minY, y), math.Max(maxY, y)
	}
	return maxX - minX, maxY - minY
}

// quatrefoilCanvas is the outer boundary of four circles of radius r whose
// centers sit r away from (cx, cy). Adjacent circles meet at the cusps, and
// each outer edge is a semicircle, so the cut stays circular.
func quatrefoilCanvas(cx, cy, r float64) *canvas.Path {
	cp := &canvas.Path{}
	if r <= 0 {
		return cp
	}
	cp.MoveTo(cx-r, cy-r)
	cp.ArcTo(r, r, 0, false, true, cx+r, cy-r)
	cp.ArcTo(r, r, 0, false, true, cx+r, cy+r)
	cp.ArcTo(r, r, 0, false, true, cx-r, cy+r)
	cp.ArcTo(r, r, 0, false, true, cx-r, cy-r)
	return cp
}

func roundedRectCanvas(cx, cy, rw, rh, rx, ry float64) *canvas.Path {
	cp := &canvas.Path{}
	if rw <= 0 || rh <= 0 {
		return cp
	}
	rx = math.Max(0, math.Min(rx, rw/2))
	ry = math.Max(0, math.Min(ry, rh/2))
	x := cx - rw/2
	y := cy - rh/2
	if rx <= 1e-9 || ry <= 1e-9 {
		cp.MoveTo(x, y)
		cp.LineTo(x+rw, y)
		cp.LineTo(x+rw, y+rh)
		cp.LineTo(x, y+rh)
		cp.Close()
		return cp
	}
	cp.MoveTo(x+rx, y)
	cp.LineTo(x+rw-rx, y)
	cp.ArcTo(rx, ry, 0, false, true, x+rw, y+ry)
	cp.LineTo(x+rw, y+rh-ry)
	cp.ArcTo(rx, ry, 0, false, true, x+rw-rx, y+rh)
	cp.LineTo(x+rx, y+rh)
	cp.ArcTo(rx, ry, 0, false, true, x, y+rh-ry)
	cp.LineTo(x, y+ry)
	cp.ArcTo(rx, ry, 0, false, true, x+rx, y)
	return cp
}

func diamondCanvas(cx, cy, ah, av, rot float64) *canvas.Path {
	c, s := math.Cos(rot), math.Sin(rot)
	pts := make([]path.Point, 0, 4)
	for _, p := range []point{{0, -av}, {ah, 0}, {0, av}, {-ah, 0}} {
		pts = append(pts, path.NewPoint(cx+p.x*c-p.y*s, cy+p.x*s+p.y*c))
	}
	return polygonCanvas(pts)
}

func stampMotif(unit *canvas.Path, cx0, cy0, sx, sy, tcx, tcy float64) *canvas.Path {
	if unit == nil {
		return &canvas.Path{}
	}
	cp := unit.Copy()
	cp.Transform(canvas.Matrix{
		{sx, 0, tcx - sx*cx0},
		{0, sy, tcy - sy*cy0},
	})
	return cp
}
