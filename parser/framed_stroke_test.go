package parser

import (
	"math"
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestFramedStrokeClipsCenterline(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "framed_stroke",
		"width": 4,
		"height": 2,
		"margin": 0.25,
		"thickness": 0,
		"components": [{
			"type": "draw",
			"commands": [
				{"command": "move", "to": "0, 1"},
				{"command": "line", "to": "4, 1"}
			]
		}]
	}`)
	moves, lines, curves := countSegments(pth)
	if moves != 1 || lines != 1 || curves != 0 {
		t.Fatalf("centerline moves=%d lines=%d curves=%d\n%s", moves, lines, curves, path.SvgString(pth, 3))
	}
	start := pth.Segments()[0].End()
	end := pth.Segments()[1].End()
	if math.Abs(start.Y-0.25) > 1e-4 || math.Abs(end.Y-0.25) > 1e-4 {
		t.Fatalf("centerline y = %v, %v", start.Y, end.Y)
	}
	xs := []float64{start.X, end.X}
	if math.Abs(math.Min(xs[0], xs[1])-0.25) > 1e-4 || math.Abs(math.Max(xs[0], xs[1])-3.75) > 1e-4 {
		t.Fatalf("clipped x = %v, want 0.25 and 3.75", xs)
	}
}

func TestFramedStrokeUnionsCrossingLines(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "framed_stroke",
		"width": 4,
		"height": 2,
		"margin": 0.25,
		"thickness": 0.2,
		"components": [{
			"type": "draw",
			"commands": [
				{"command": "move", "to": "1, 1"},
				{"command": "line", "to": "3, 1"},
				{"command": "move", "to": "2, 0.5"},
				{"command": "line", "to": "2, 1.5"}
			]
		}]
	}`)
	moves, lines, _ := countSegments(pth)
	if moves < 1 || lines < 8 {
		t.Fatalf("expanded cross moves=%d lines=%d\n%s", moves, lines, path.SvgString(pth, 3))
	}
	endpointsInside(t, pth, 0.25, 0.25, 3.75, 1.75, 0.02)
	sawCross := false
	for _, seg := range pth.Segments() {
		end := seg.End()
		if end.X > 1.2 && end.X < 2.8 && end.Y > 0.7 && end.Y < 1.3 {
			sawCross = true
			break
		}
	}
	if !sawCross {
		t.Fatalf("union missed the crossing\n%s", path.SvgString(pth, 3))
	}
}

func TestFramedStrokeFlattensSvgCurve(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "framed_stroke",
		"width": 4,
		"height": 3,
		"margin": 0.2,
		"thickness": 0.12,
		"components": [{
			"type": "draw",
			"commands": [{
				"command": "svg",
				"svg": "M 0.4 1.5 C 1.2 0.4 2.4 2.6 3.6 1.5"
			}]
		}]
	}`)
	moves, lines, _ := countSegments(pth)
	if moves < 1 || lines < 8 {
		t.Fatalf("curved stroke moves=%d lines=%d\n%s", moves, lines, path.SvgString(pth, 3))
	}
	endpointsInside(t, pth, 0.2, 0.2, 3.8, 2.8, 0.02)
}

func TestFramedStrokeMeetsBorderOnAngle(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "framed_stroke",
		"width": 4,
		"height": 2,
		"margin": 0.25,
		"thickness": 0.2,
		"components": [{
			"type": "draw",
			"commands": [
				{"command": "move", "to": "0.25, 1"},
				{"command": "line", "to": "2, 1.5"}
			]
		}]
	}`)
	segs := pth.Segments()
	for i := 1; i < len(segs); i++ {
		if _, isMove := segs[i-1].(path.MoveSegment); isMove {
			continue
		}
		if _, isMove := segs[i].(path.MoveSegment); isMove {
			continue
		}
		a := segs[i-1].End()
		b := segs[i].End()
		if math.Abs(a.Y-b.Y) > 1e-3 || a.X < 0.24 || a.X > 0.37 || b.X < 0.24 || b.X > 0.37 {
			continue
		}
		if math.Abs(a.Y-0.9) < 0.03 || math.Abs(a.Y-1.1) < 0.03 {
			t.Fatalf("square notch where the diagonal meets the frame\n%s", path.SvgString(pth, 3))
		}
	}
	endpointsInside(t, pth, 0.25, 0.25, 3.75, 1.75, 0.02)
}

func TestFramedStrokePushesNearEndThroughBorder(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "framed_stroke",
		"width": 4,
		"height": 3,
		"margin": 0.2,
		"thickness": 0.12,
		"components": [{
			"type": "draw",
			"commands": [
				{"command": "move", "to": "1, 1.5"},
				{"command": "line", "to": "3.7, 1.5"}
			]
		}]
	}`)
	var ys []float64
	for _, seg := range pth.Segments() {
		end := seg.End()
		if end.X > 3.7 {
			ys = append(ys, end.Y)
		}
	}
	if len(ys) < 2 {
		t.Fatalf("end did not open through the right edge\n%s", path.SvgString(pth, 3))
	}
	minY, maxY := ys[0], ys[0]
	for _, y := range ys[1:] {
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	if maxY-minY < 0.1 {
		t.Fatalf("opening pinched to %.3f\n%s", maxY-minY, path.SvgString(pth, 3))
	}
}

func TestFramedStrokeFitsChildToPanel(t *testing.T) {
	pth := renderPatternComponent(t, `{
		"type": "framed_stroke",
		"width": 4,
		"height": 2,
		"margin": 0.25,
		"thickness": 0,
		"components": [{
			"type": "draw",
			"commands": [
				{"command": "move", "to": "10, 20"},
				{"command": "line", "to": "20, 20"},
				{"command": "move", "to": "10, 20"},
				{"command": "line", "to": "10, 40"}
			]
		}]
	}`)
	sawRight, sawBottom := false, false
	for _, seg := range pth.Segments() {
		end := seg.End()
		if math.Abs(end.X-3.75) < 1e-3 && math.Abs(end.Y-0.25) < 1e-3 {
			sawRight = true
		}
		if math.Abs(end.X-0.25) < 1e-3 && math.Abs(end.Y-1.75) < 1e-3 {
			sawBottom = true
		}
	}
	if !sawRight || !sawBottom {
		t.Fatalf("drawing was not fitted to the inner panel\n%s", path.SvgString(pth, 3))
	}
}

func TestFramedStrokeRequiresPanelSize(t *testing.T) {
	InitContext()
	_, _, err := renderPatternComponentMaybeErr(t, `{
		"type": "framed_stroke",
		"components": [{
			"type": "draw",
			"commands": [
				{"command": "move", "to": "0, 0"},
				{"command": "line", "to": "1, 0"}
			]
		}]
	}`)
	if err == nil {
		t.Fatal("framed_stroke without width and height should fail")
	}
}
