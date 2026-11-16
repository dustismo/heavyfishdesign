package path

import "testing"

func TestParseHorizontalVerticalQuadraticArc(t *testing.T) {
	h, err := ParsePathFromSvg("M 0 0 H 10")
	if err != nil {
		t.Fatal(err)
	}
	if got := SvgString(h, 3); got != "M 0.000 0.000 L 10.000 0.000" {
		t.Fatalf("H: %s", got)
	}

	v, err := ParsePathFromSvg("M 0 0 V 4")
	if err != nil {
		t.Fatal(err)
	}
	if got := SvgString(v, 3); got != "M 0.000 0.000 L 0.000 4.000" {
		t.Fatalf("V: %s", got)
	}

	q, err := ParsePathFromSvg("M 0 0 Q 0 10 10 10")
	if err != nil {
		t.Fatal(err)
	}
	d := NewDraw()
	d.MoveTo(NewPoint(0, 0))
	d.QCurveTo(NewPoint(0, 10), NewPoint(10, 10))
	if got, want := SvgString(q, 3), SvgString(d.Path(), 3); got != want {
		t.Fatalf("Q:\n got %s\nwant %s", got, want)
	}

	a, err := ParsePathFromSvg("M 0 0 A 5 5 0 0 1 10 0")
	if err != nil {
		t.Fatal(err)
	}
	if SvgString(a, 3) == "" {
		t.Fatal("arc produced an empty path")
	}
	sawCurve := false
	for _, seg := range a.Segments() {
		if _, ok := seg.(CurveSegment); ok {
			sawCurve = true
		}
	}
	if !sawCurve {
		t.Fatalf("arc was not converted to cubics: %s", SvgString(a, 3))
	}
}

func TestLegacyCommandsStayOnLegacyParser(t *testing.T) {
	svg := "M 0 0 L 5 0 L 10 0 L 10 2 Z"
	got, err := ParsePathFromSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	want, err := parsePathFromSvgLegacy(svg)
	if err != nil {
		t.Fatal(err)
	}
	if SvgString(got, 6) != SvgString(want, 6) {
		t.Fatalf("legacy path changed\n got %s\nwant %s", SvgString(got, 6), SvgString(want, 6))
	}
}

func TestOrPathsOverlappingRects(t *testing.T) {
	a, err := ParsePathFromSvg("M 0 0 L 2 0 L 2 2 L 0 2")
	if err != nil {
		t.Fatal(err)
	}
	b, err := ParsePathFromSvg("M 1 1 L 3 1 L 3 3 L 1 3")
	if err != nil {
		t.Fatal(err)
	}
	u := OrPaths(a, b)
	if len(u.Segments()) == 0 {
		t.Fatal("union was empty")
	}
}
