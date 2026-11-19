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
	pth, _, err := component.Render(dom.RenderContext{})
	if err != nil {
		t.Fatal(err)
	}
	return pth
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
