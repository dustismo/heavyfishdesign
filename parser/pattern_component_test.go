package parser

import (
	"math"
	"testing"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/util"
)

func renderPatternComponent(t *testing.T, componentJSON string) path.Path {
	t.Helper()
	pth, _, err := renderPatternComponentMaybeErr(t, componentJSON)
	if err != nil {
		t.Fatal(err)
	}
	return pth
}

func renderPatternComponentMaybeErr(t *testing.T, componentJSON string) (path.Path, dom.RenderContext, error) {
	t.Helper()
	InitContext()
	dm, err := dynmap.ParseJSON(`{
		"params": {},
		"parts": [{"components": [` + componentJSON + `]}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := dom.ParseDocument(dm, util.NewLog())
	if err != nil {
		t.Fatal(err)
	}
	component, ok := doc.Parts[0].Children()[0].(dom.Component)
	if !ok {
		t.Fatal("part child is not a component")
	}
	return component.Render(dom.RenderContext{})
}

func countSegments(p path.Path) (moves, lines, curves int) {
	for _, seg := range p.Segments() {
		switch seg.(type) {
		case path.MoveSegment:
			moves++
		case path.LineSegment:
			lines++
		case path.CurveSegment:
			curves++
		}
	}
	return
}

func TestLatticeOpeningsShareFrameEdge(t *testing.T) {
	pth := renderImportedComponent(t, `{
		"imports": [{"path": "../designs/component_examples/staggered_lattice.hfd"}],
		"parts": [{"components": [{
			"type": "staggered_lattice",
			"width": 4,
			"height": 3,
			"margin": 0.2,
			"thickness": 0.08,
			"rect_width": 1.15,
			"rect_height": 1.55,
			"v_gap_ratio": 0.35,
			"h_gap_ratio": 0.55
		}]}]
	}`)
	// Openings that meet the frame should share the frame's inner edge
	// (margin + half the stroke), not stop a hair inside it.
	const margin = 0.2
	const half = 0.04
	near := func(x, edge float64) bool { return math.Abs(x-edge) <= 0.005 }
	for _, seg := range pth.Segments() {
		ln, ok := seg.(path.LineSegment)
		if !ok {
			continue
		}
		dx := math.Abs(ln.EndPoint.X - ln.StartPoint.X)
		dy := math.Abs(ln.EndPoint.Y - ln.StartPoint.Y)
		if dx > 0.002 || dy < 0.05 {
			continue
		}
		x := ln.EndPoint.X
		if x > margin+half+half {
			continue
		}
		if !near(x, margin) && !near(x, margin+half) {
			t.Fatalf("left opening at x=%.3f, want the frame edge %.3f or %.3f", x, margin, margin+half)
		}
	}
}

func TestPatternHexagonGridOneCell(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "hexagon_grid",
		"width": 2,
		"height": 2,
		"margin": 0.25,
		"circumradius": 0.5,
		"gap": 0.2
	}`)
	moves, lines, curves := countSegments(pth)
	if moves != 1 || lines != 6 || curves != 0 {
		t.Fatalf("segments moves=%d lines=%d curves=%d, want 1/6/0\n%s", moves, lines, curves, path.SvgString(pth, 3))
	}
	top := pth.Segments()[0].End()
	if math.Abs(top.X-1) > 1e-6 || math.Abs(top.Y-0.5) > 1e-6 {
		t.Fatalf("top vertex = (%v, %v), want (1, 0.5)", top.X, top.Y)
	}
	closed := pth.Segments()[len(pth.Segments())-1].End()
	if math.Abs(closed.X-top.X) > 1e-9 || math.Abs(closed.Y-top.Y) > 1e-9 {
		t.Fatalf("hexagon does not return to its start")
	}
}

func TestPatternHexCirclesCount(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "hex_circles",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"radius": 0.3,
		"gap": 0.2
	}`)
	moves, _, curves := countSegments(pth)
	if moves != 6 || curves != 12 {
		t.Fatalf("segments moves=%d curves=%d, want 6 moves and 12 curves\n%s", moves, curves, path.SvgString(pth, 3))
	}
}

func TestPatternLensGridHasArcs(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "lens_grid",
		"width": 4,
		"height": 4,
		"margin": 0.4,
		"lens_length": 1,
		"lens_width": 0.4,
		"lean_deg": 0,
		"gap_x": 0.2,
		"gap_y": 0.2
	}`)
	moves, _, curves := countSegments(pth)
	if moves < 1 || curves < 2 {
		t.Fatalf("lens grid moves=%d curves=%d\n%s", moves, curves, path.SvgString(pth, 3))
	}
}

func TestPatternClipKeepsPartialHexesInsideMargin(t *testing.T) {
	unclipped := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "hexagon_grid",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"circumradius": 0.9,
		"gap": 0.05
	}`)
	_, lines, curves := countSegments(unclipped)
	if lines != 0 || curves != 0 {
		t.Fatalf("uncropped grid should place nothing, got %s", path.SvgString(unclipped, 3))
	}

	clipped := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "hexagon_grid",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"circumradius": 0.9,
		"gap": 0.05,
		"clip_to_border": true
	}`)
	if len(clipped.Segments()) == 0 {
		t.Fatal("clip_to_border should keep hexes that cross the margin")
	}
	for _, seg := range clipped.Segments() {
		pt := seg.End()
		if pt.X < 0.2-1e-4 || pt.X > 1.8+1e-4 || pt.Y < 0.2-1e-4 || pt.Y > 1.8+1e-4 {
			t.Fatalf("clipped point (%v, %v) is outside the inner margin\n%s", pt.X, pt.Y, path.SvgString(clipped, 3))
		}
	}
}

func TestPatternExpressionsAndUnknownKind(t *testing.T) {
	InitContext()
	dm, err := dynmap.ParseJSON(`{
		"params": {"panel_width": 2, "panel_height": 2},
		"parts": [{
			"components": [{
				"type": "pattern",
				"kind": "hexagon_grid",
				"width": "panel_width",
				"height": "panel_height",
				"margin": 0.25,
				"circumradius": 0.5,
				"gap": 0.2
			}]
		}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := dom.ParseDocument(dm, util.NewLog())
	if err != nil {
		t.Fatal(err)
	}
	pth, _, err := doc.Parts[0].Render(dom.RenderContext{})
	if err != nil {
		t.Fatal(err)
	}
	moves, _, _ := countSegments(pth)
	if moves != 1 {
		t.Fatalf("expression-sized hexagon moves=%d, want 1", moves)
	}

	bad, err := dynmap.ParseJSON(`{
		"params": {},
		"parts": [{"components": [{"type": "pattern", "kind": "nope", "width": 1, "height": 1}]}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	doc, err = dom.ParseDocument(bad, util.NewLog())
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = doc.Parts[0].Render(dom.RenderContext{})
	if err == nil {
		t.Fatal("unknown pattern kind should fail the render")
	}
}

func endpointsInside(t *testing.T, p path.Path, minX, minY, maxX, maxY, tol float64) {
	t.Helper()
	for _, seg := range p.Segments() {
		end := seg.End()
		if end.X < minX-tol || end.X > maxX+tol || end.Y < minY-tol || end.Y > maxY+tol {
			t.Fatalf("endpoint (%v, %v) outside [%v,%v]–[%v,%v]", end.X, end.Y, minX, minY, maxX, maxY)
		}
	}
}

func TestPatternHatchOneHorizontalSlot(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "hatch",
		"width": 2,
		"height": 2,
		"margin": 0.25,
		"angle_deg": 0,
		"spacing": 10
	}`)
	moves, lines, curves := countSegments(pth)
	if moves != 1 || lines != 1 || curves != 0 {
		t.Fatalf("hatch moves=%d lines=%d curves=%d\n%s", moves, lines, curves, path.SvgString(pth, 3))
	}
	start := pth.Segments()[0].End()
	end := pth.Segments()[1].End()
	if math.Abs(start.Y-1) > 1e-4 || math.Abs(end.Y-1) > 1e-4 {
		t.Fatalf("hatch y = %v, %v", start.Y, end.Y)
	}
	xs := []float64{start.X, end.X}
	if math.Abs(math.Min(xs[0], xs[1])-0.25) > 1e-4 || math.Abs(math.Max(xs[0], xs[1])-1.75) > 1e-4 {
		t.Fatalf("hatch x = %v, want 0.25 and 1.75", xs)
	}
}

func TestPatternHatchCrossHasMoreSlots(t *testing.T) {
	one := renderPatternComponent(t, `{
		"type": "pattern", "kind": "hatch",
		"width": 3, "height": 3, "margin": 0.2,
		"angle_deg": 0, "spacing": 0.6, "thickness": 0.08
	}`)
	both := renderPatternComponent(t, `{
		"type": "pattern", "kind": "hatch",
		"width": 3, "height": 3, "margin": 0.2,
		"angle_deg": 0, "spacing": 0.6, "thickness": 0.08,
		"cross": "true"
	}`)
	if len(both.Segments()) <= len(one.Segments()) {
		t.Fatalf("cross hatch has %d segments, single has %d", len(both.Segments()), len(one.Segments()))
	}
}

func TestPatternWaveLatticeRibbons(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "wave_lattice",
		"width": 4,
		"height": 3,
		"margin": 0.2,
		"amplitude": 0.1,
		"wavelength": 1.2,
		"ribbon_thickness": 0.2,
		"ribbon_gap": 0.4,
		"phase_step": 0.4,
		"samples": 16
	}`)
	moves, _, curves := countSegments(pth)
	if moves < 1 || curves < 16 {
		t.Fatalf("wave moves=%d curves=%d", moves, curves)
	}
}

func TestWavePeaksCrossTopAndBottom(t *testing.T) {
	center := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "wave_lattice",
		"width": 4,
		"height": 3,
		"margin": 0.2,
		"amplitude": 0.15,
		"wavelength": 1.1,
		"ribbon_thickness": 0.18,
		"ribbon_gap": 0.22,
		"phase_step": 0.5,
		"samples": 32
	}`)
	above, below := false, false
	for _, seg := range center.Segments() {
		y := seg.End().Y
		if y < 0.2-1e-3 {
			above = true
		}
		if y > 2.8+1e-3 {
			below = true
		}
	}
	if !above || !below {
		t.Fatalf("centerline peaks stay inside the margin above=%v below=%v", above, below)
	}

	stroked := renderImportedComponent(t, `{
		"imports": [{"path": "../designs/component_examples/wave_lattice.hfd"}],
		"parts": [{"components": [{
			"type": "wave_lattice",
			"width": 4,
			"height": 3,
			"margin": 0.2,
			"thickness": 0.18,
			"amplitude": 0.15,
			"wavelength": 1.1,
			"ribbon_gap": 0.22,
			"phase_step": 0.5,
			"samples": 32,
			"frame": false
		}]}]
	}`)
	top, bot := false, false
	for _, seg := range stroked.Segments() {
		y := seg.End().Y
		if math.Abs(y-0.2) < 0.03 {
			top = true
		}
		if math.Abs(y-2.8) < 0.03 {
			bot = true
		}
		if y < 0.2-0.03 || y > 2.8+0.03 {
			t.Fatalf("stroked wave left the margin at y=%v", y)
		}
	}
	if !top || !bot {
		t.Fatalf("stroked peaks do not meet the edges top=%v bottom=%v", top, bot)
	}
}

func TestPatternWaveClipStaysInsideMargin(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "wave_lattice",
		"width": 3,
		"height": 2,
		"margin": 0.3,
		"amplitude": 0.5,
		"wavelength": 1,
		"ribbon_thickness": 0.2,
		"ribbon_gap": 0.15,
		"samples": 16,
		"clip_to_border": "true"
	}`)
	moves, _, _ := countSegments(pth)
	if moves < 1 {
		t.Fatal("clipped waves produced no path")
	}
	endpointsInside(t, pth, 0.3, 0.3, 2.7, 1.7, 1e-4)
}

func TestPatternLatticeCenterlines(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "staggered_lattice_lines",
		"width": 3,
		"height": 3,
		"margin": 0.2,
		"rect_width": 1.2,
		"rect_height": 1.4,
		"v_gap_ratio": 0.3,
		"h_gap_ratio": 0.4,
		"line_thickness": 0
	}`)
	moves, lines, curves := countSegments(pth)
	if moves < 1 || lines < moves || curves != 0 {
		t.Fatalf("centerlines moves=%d lines=%d curves=%d", moves, lines, curves)
	}
}

func TestPatternLatticeCenterlinesInsideMargin(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "staggered_lattice_lines",
		"width": 3,
		"height": 3,
		"margin": 0.25,
		"rect_width": 1.6,
		"rect_height": 1.8,
		"v_gap_ratio": 0.2,
		"h_gap_ratio": 0.25,
		"line_thickness": 0.08
	}`)
	moves, lines, curves := countSegments(pth)
	// Cells are closed loops, so they have one extra line. They may cross the
	// margin; framed_stroke clips the stroke.
	if moves < 1 || lines <= moves || curves != 0 {
		t.Fatalf("lattice centerlines moves=%d lines=%d curves=%d", moves, lines, curves)
	}
}

func TestPatternOverlappingCirclesPanelOnly(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "overlapping_circles",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"count": 0
	}`)
	moves, lines, curves := countSegments(pth)
	if moves != 1 || lines != 4 || curves != 0 {
		t.Fatalf("empty circles moves=%d lines=%d curves=%d\n%s", moves, lines, curves, path.SvgString(pth, 3))
	}
	endpointsInside(t, pth, 0.2, 0.2, 1.8, 1.8, 1e-6)
}

func TestPatternOverlappingCirclesStable(t *testing.T) {
	js := `{
		"type": "pattern",
		"kind": "overlapping_circles",
		"width": 3,
		"height": 2.5,
		"margin": 0.15,
		"radius_min": 0.25,
		"radius_max": 0.45,
		"stroke_width": 0.06,
		"count": 6,
		"seed": 17
	}`
	a := path.SvgString(renderPatternComponent(t, js), 3)
	b := path.SvgString(renderPatternComponent(t, js), 3)
	if a != b {
		t.Fatal("overlapping circles changed between renders with the same seed")
	}
	checkRings := func(label, js string) {
		t.Helper()
		pth := renderPatternComponent(t, js)
		_, lines, curves := countSegments(pth)
		if curves < 8 {
			t.Fatalf("%s lost curves: lines=%d curves=%d", label, lines, curves)
		}
		var open int
		var worst float64
		var loops int
		var start, prev path.Point
		has := false
		flush := func() {
			if !has {
				return
			}
			loops++
			gap := math.Hypot(prev.X-start.X, prev.Y-start.Y)
			if gap > worst {
				worst = gap
			}
			if gap > 0.002 {
				open++
			}
		}
		for _, seg := range pth.Segments() {
			if _, ok := seg.(path.MoveSegment); ok {
				flush()
				start = seg.End()
				prev = start
				has = true
				continue
			}
			prev = seg.End()
		}
		flush()
		t.Logf("%s loops=%d open=%d worst=%.5f lines=%d curves=%d", label, loops, open, worst, lines, curves)
		if open > 0 {
			t.Fatalf("%s: %d ring loops do not close (worst %.4f)", label, open, worst)
		}
	}
	checkRings("stable", js)
	checkRings("design", `{
		"type": "pattern",
		"kind": "overlapping_circles",
		"width": 4,
		"height": 3,
		"margin": 0.2,
		"radius_min": 0.22,
		"radius_max": 0.48,
		"stroke_width": 0.07,
		"count": 12,
		"seed": 17
	}`)
}

func TestHatchCustomComponentStrokesCenterline(t *testing.T) {
	pth := renderImportedComponent(t, `{
		"imports": [{"path": "../designs/component_examples/hatch.hfd"}],
		"parts": [{"components": [{
			"type": "hatch",
			"width": 2,
			"height": 2,
			"margin": 0.25,
			"thickness": 0.2,
			"angle_deg": 0,
			"spacing": 10,
			"frame": false
		}]}]
	}`)
	sawLow, sawHigh := false, false
	for _, seg := range pth.Segments() {
		y := seg.End().Y
		if math.Abs(y-0.9) < 1e-3 {
			sawLow = true
		}
		if math.Abs(y-1.1) < 1e-3 {
			sawHigh = true
		}
	}
	if !sawLow || !sawHigh {
		t.Fatalf("hatch stroke missing the slot edges\n%s", path.SvgString(pth, 3))
	}
	endpointsInside(t, pth, 0.25, 0.25, 1.75, 1.75, 0.02)
}

func TestWaveAndLatticeCustomComponents(t *testing.T) {
	wave := renderImportedComponent(t, `{
		"imports": [{"path": "../designs/component_examples/wave_lattice.hfd"}],
		"parts": [{"components": [{
			"type": "wave_lattice",
			"width": 4,
			"height": 3,
			"margin": 0.2,
			"thickness": 0.18,
			"amplitude": 0.1,
			"wavelength": 1.2,
			"ribbon_gap": 0.4,
			"samples": 16,
			"frame": false
		}]}]
	}`)
	if moves, lines, _ := countSegments(wave); moves < 1 || lines < 8 {
		t.Fatalf("wave stroke moves=%d lines=%d", moves, lines)
	}
	endpointsInside(t, wave, 0.2, 0.2, 3.8, 2.8, 0.02)

	lattice := renderImportedComponent(t, `{
		"imports": [{"path": "../designs/component_examples/staggered_lattice.hfd"}],
		"parts": [{"components": [{
			"type": "staggered_lattice",
			"width": 3,
			"height": 3,
			"margin": 0.25,
			"thickness": 0.08,
			"rect_width": 1.6,
			"rect_height": 1.8,
			"v_gap_ratio": 0.2,
			"h_gap_ratio": 0.25
		}]}]
	}`)
	if moves, lines, _ := countSegments(lattice); moves < 1 || lines < 8 {
		t.Fatalf("lattice stroke moves=%d lines=%d", moves, lines)
	}
	endpointsInside(t, lattice, 0.25, 0.25, 2.75, 2.75, 0.02)
}

func TestTileGridCirclesAndStagger(t *testing.T) {
	straight := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "circle",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"scale": 0.5,
		"gap_x": 0.2,
		"gap_y": 0.2,
		"stagger": 0
	}`)
	moves, _, curves := countSegments(straight)
	if moves != 4 || curves < 8 {
		t.Fatalf("circle grid moves=%d curves=%d, want 4 tiles\n%s", moves, curves, path.SvgString(straight, 3))
	}
	if !subpathsClosed(straight, 1e-3) {
		t.Fatalf("circle tiles do not close\n%s", path.SvgString(straight, 3))
	}

	staggered := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "circle",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"scale": 0.5,
		"gap_x": 0.2,
		"gap_y": 0.2,
		"stagger": 0.5
	}`)
	moves, _, _ = countSegments(staggered)
	if moves != 3 {
		t.Fatalf("staggered circle grid moves=%d, want 3\n%s", moves, path.SvgString(staggered, 3))
	}
}

func TestTileGridQuatrefoilIsRound(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "quatrefoil",
		"width": 3,
		"height": 3,
		"margin": 0.2,
		"scale": 0.4,
		"gap_x": 0.2,
		"gap_y": 0.2,
		"stagger": 0,
		"arm": 0.2
	}`)
	moves, lines, curves := countSegments(pth)
	if moves != 1 || lines != 0 || curves < 4 {
		t.Fatalf("quatrefoil moves=%d lines=%d curves=%d\n%s", moves, lines, curves, path.SvgString(pth, 3))
	}
	if !subpathsClosed(pth, 1e-3) {
		t.Fatalf("quatrefoil does not close\n%s", path.SvgString(pth, 3))
	}
	tl, br := pathBox(t, pth)
	// Four outer semicircles of radius 0.2 span 0.8 on each axis. The cusp
	// square alone would span 0.4, so this catches an arc that turned inward.
	if math.Abs((br.X-tl.X)-0.8) > 0.02 || math.Abs((br.Y-tl.Y)-0.8) > 0.02 {
		t.Fatalf("quatrefoil bbox %v x %v, want about 0.8", br.X-tl.X, br.Y-tl.Y)
	}
}

func TestTileGridClipKeepsPartialCircle(t *testing.T) {
	open := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "circle",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"scale": 2,
		"gap_x": 0,
		"gap_y": 0,
		"stagger": 0
	}`)
	moves, _, curves := countSegments(open)
	if moves != 1 || curves != 0 {
		t.Fatalf("uncropped circle should be dropped, moves=%d curves=%d", moves, curves)
	}
	clipped := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "circle",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"scale": 2,
		"gap_x": 0,
		"gap_y": 0,
		"stagger": 0,
		"clip_to_border": true
	}`)
	moves, _, curves = countSegments(clipped)
	if moves < 1 || curves < 2 {
		t.Fatalf("clipped circle moves=%d curves=%d, the arc was flattened\n%s", moves, curves, path.SvgString(clipped, 3))
	}
	endpointsInside(t, clipped, 0.2, 0.2, 1.8, 1.8, 0.03)
}

func TestTileGridDiamondRotationPacksMore(t *testing.T) {
	upright := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "diamond",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"scale": 0.8,
		"gap_x": 0.05,
		"gap_y": 0.05,
		"stagger": 0,
		"rotation_deg": 0
	}`)
	turned := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "diamond",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"scale": 0.8,
		"gap_x": 0.05,
		"gap_y": 0.05,
		"stagger": 0,
		"rotation_deg": 45
	}`)
	upMoves, _, _ := countSegments(upright)
	turnMoves, _, _ := countSegments(turned)
	if upMoves != 1 || turnMoves != 4 {
		t.Fatalf("diamond tiles upright=%d turned=%d, want 1 and 4", upMoves, turnMoves)
	}
}

func TestTileGridRoundedRectCorner(t *testing.T) {
	sharp := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "rounded_rect",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"scale": 0.5,
		"gap_x": 1.2,
		"gap_y": 1.2,
		"stagger": 0,
		"corner_radius": 0
	}`)
	moves, lines, curves := countSegments(sharp)
	if moves != 1 || lines != 4 || curves != 0 {
		t.Fatalf("sharp rect moves=%d lines=%d curves=%d", moves, lines, curves)
	}
	round := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "rounded_rect",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"scale": 0.5,
		"gap_x": 1.2,
		"gap_y": 1.2,
		"stagger": 0,
		"corner_radius": 0.1
	}`)
	moves, _, curves = countSegments(round)
	if moves != 1 || curves < 4 {
		t.Fatalf("rounded rect moves=%d curves=%d", moves, curves)
	}
	tl, br := pathBox(t, round)
	if math.Abs((br.X-tl.X)-0.5) > 0.02 || math.Abs((br.Y-tl.Y)-0.5) > 0.02 {
		t.Fatalf("rounded rect bbox %v x %v, want 0.5", br.X-tl.X, br.Y-tl.Y)
	}
}

func TestTileGridCustomMotif(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "custom_svg",
		"width": 2,
		"height": 2,
		"margin": 0.2,
		"scale": 0.5,
		"gap_x": 0.2,
		"gap_y": 0.2,
		"stagger": 0,
		"motif": "M 0 0 L 1 0 L 1 1 L 0 1 Z"
	}`)
	moves, lines, _ := countSegments(pth)
	if moves != 4 || lines != 16 {
		t.Fatalf("custom tiles moves=%d lines=%d, want 4 closed squares", moves, lines)
	}
	if !subpathsClosed(pth, 1e-6) {
		t.Fatalf("custom tiles do not close")
	}
}

func TestTileGridUnknownShape(t *testing.T) {
	_, _, err := renderPatternComponentMaybeErr(t, `{
		"type": "pattern",
		"kind": "tile_grid",
		"shape": "nope",
		"width": 2,
		"height": 2
	}`)
	if err == nil {
		t.Fatal("expected unknown tile shape to fail")
	}
}

func subpathsClosed(p path.Path, tol float64) bool {
	var start path.Point
	var end path.Point
	open := false
	for _, seg := range p.Segments() {
		if _, ok := seg.(path.MoveSegment); ok {
			if open && (math.Abs(end.X-start.X) > tol || math.Abs(end.Y-start.Y) > tol) {
				return false
			}
			start = seg.End()
			end = start
			open = true
			continue
		}
		end = seg.End()
		open = true
	}
	if open && (math.Abs(end.X-start.X) > tol || math.Abs(end.Y-start.Y) > tol) {
		return false
	}
	return true
}

func pathBox(t *testing.T, p path.Path) (path.Point, path.Point) {
	t.Helper()
	tl, br, err := path.BoundingBoxTrimWhitespace(p, path.NewSegmentOperators())
	if err != nil {
		t.Fatal(err)
	}
	return tl, br
}

func renderImportedComponent(t *testing.T, docJSON string) path.Path {
	t.Helper()
	InitContext()
	dm, err := dynmap.ParseJSON(docJSON)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := dom.ParseDocument(dm, util.NewLog())
	if err != nil {
		t.Fatal(err)
	}
	component, ok := doc.Parts[0].Children()[0].(dom.Component)
	if !ok {
		t.Fatal("part child is not a component")
	}
	pth, _, err := component.Render(dom.RenderContext{})
	if err != nil {
		t.Fatal(err)
	}
	return pth
}
