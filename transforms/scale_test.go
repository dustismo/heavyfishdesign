package transforms

import (
	"math"
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestScaleTransform(t *testing.T) {
	pathStr := "M348.7,980.332L257.96,1029.19L329.328,881.936 C638.607,1066.56 477.01,980.332 477.01,980.332"
	originalPath, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	transform := ScaleTransform{
		StartPoint:       path.NewPoint(0, 0),
		EndPoint:         path.NewPoint(0.5, 0),
		SegmentOperators: path.NewSegmentOperators(),
	}

	p, err := transform.PathTransform(originalPath)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	expectedStr := "M 1.359 3.820 L 1.005 4.011 L 1.283 3.437 C 2.489 4.156 1.859 3.820 1.859 3.820"
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

// Lamp side pattern: scale by measured bbox width only must keep X/Y scale equal (uniform),
// so circular holes stay round. This guards ScaleTransform width-only + zero height behavior.
func TestScaleTransformWidthOnlyUniformKeepsCircleBboxSquare(t *testing.T) {
	so := path.NewSegmentOperators()
	svg := "M 58 50 C 58 54.418 54.418 58 50 58 C 45.582 58 42 54.418 42 50 C 42 45.582 45.582 42 50 42 C 54.418 42 58 45.582 58 50"
	p, err := path.ParsePathFromSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	st := ScaleTransform{
		Width:            16,
		Height:           0,
		SegmentOperators: so,
	}
	out, err := st.PathTransform(p)
	if err != nil {
		t.Fatal(err)
	}
	tl, br, err := path.BoundingBoxTrimWhitespace(out, so)
	if err != nil {
		t.Fatal(err)
	}
	w, h := br.X-tl.X, br.Y-tl.Y
	if w <= 0 || h <= 0 {
		t.Fatalf("bbox w=%v h=%v", w, h)
	}
	if math.Abs(w/h-1) > 0.02 {
		t.Fatalf("width-only scale should preserve aspect ratio; got w=%v h=%v", w, h)
	}
}

// If both width and height targets are set to different values, a nominally square path becomes non-square.
func TestScaleTransformWidthAndHeightNonUniformDistortsSquareBbox(t *testing.T) {
	so := path.NewSegmentOperators()
	svg := "M 58 50 C 58 54.418 54.418 58 50 58 C 45.582 58 42 54.418 42 50 C 42 45.582 45.582 42 50 42 C 54.418 42 58 45.582 58 50"
	p, err := path.ParsePathFromSvg(svg)
	if err != nil {
		t.Fatal(err)
	}
	st := ScaleTransform{
		Width:            16,
		Height:           48,
		SegmentOperators: so,
	}
	out, err := st.PathTransform(p)
	if err != nil {
		t.Fatal(err)
	}
	tl, br, err := path.BoundingBoxTrimWhitespace(out, so)
	if err != nil {
		t.Fatal(err)
	}
	w, h := br.X-tl.X, br.Y-tl.Y
	if math.Abs(w/h-1) < 0.15 {
		t.Fatalf("expected visible stretch with width!=height scale; got w=%v h=%v ratio=%v", w, h, w/h)
	}
}
