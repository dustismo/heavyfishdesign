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
		"spacing": 10,
		"thickness": 0.2
	}`)
	moves, lines, curves := countSegments(pth)
	if moves != 1 || lines < 4 || curves != 0 {
		t.Fatalf("hatch moves=%d lines=%d curves=%d\n%s", moves, lines, curves, path.SvgString(pth, 3))
	}
	sawLow, sawHigh := false, false
	for _, seg := range pth.Segments() {
		y := seg.End().Y
		if math.Abs(y-0.9) < 1e-6 {
			sawLow = true
		}
		if math.Abs(y-1.1) < 1e-6 {
			sawHigh = true
		}
		if seg.End().X < 0.25-1e-6 || seg.End().X > 1.75+1e-6 {
			t.Fatalf("hatch x %v outside the inner rect", seg.End().X)
		}
	}
	if !sawLow || !sawHigh {
		t.Fatalf("hatch slot edges missing\n%s", path.SvgString(pth, 3))
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
	moves, lines, _ := countSegments(pth)
	if moves < 1 || lines < 16 {
		t.Fatalf("wave moves=%d lines=%d", moves, lines)
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
	if moves < 4 || moves != lines || curves != 0 {
		t.Fatalf("centerlines moves=%d lines=%d curves=%d", moves, lines, curves)
	}
}

func TestPatternLatticeStrokeInsideMargin(t *testing.T) {
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
	moves, lines, _ := countSegments(pth)
	if moves < 1 || lines < 8 {
		t.Fatalf("lattice stroke moves=%d lines=%d\n%s", moves, lines, path.SvgString(pth, 3))
	}
	endpointsInside(t, pth, 0.25, 0.25, 2.75, 2.75, 0.02)
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
		"seed": 17,
		"segments": 12
	}`
	a := path.SvgString(renderPatternComponent(t, js), 3)
	b := path.SvgString(renderPatternComponent(t, js), 3)
	if a != b {
		t.Fatal("overlapping circles changed between renders with the same seed")
	}
	if len(a) < 40 {
		t.Fatalf("overlapping circles path is too short: %s", a)
	}
}
