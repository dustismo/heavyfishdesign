package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestDedup4(t *testing.T) {
	pathStr := `M 50.000 100.000 L 150.000 100.000 M 100.000 50.000 L 100.000 150.000 M 150.000 100.000 L 50.000 100.000 M 100.000 150.000 L 100.000 50.000`

	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	transform := DedupSegmentsTransform{3}
	p, _ = transform.PathTransform(p)

	expectedStr := `M 50.000 100.000 L 150.000 100.000 M 100.000 50.000 L 100.000 150.000 M 150.000 100.000 L 50.000 100.000 M 100.000 150.000 L 100.000 50.000`
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestDedup3(t *testing.T) {
	pathStr := `L 0.000 5.000 M 0.000 0.000 M 0.000 5.000 L 5.000 5.000`

	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	transform := DedupSegmentsTransform{3}
	p, _ = transform.PathTransform(p)

	expectedStr := `M 0.000 0.000 L 0.000 5.000 L 5.000 5.000`
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestDedupIdempotency(t *testing.T) {
	pathStr := `M 0.000 0.000 L 0.000 5.000 L 5.000 5.000`

	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	transform := DedupSegmentsTransform{3}
	p, _ = transform.PathTransform(p)

	expectedStr := `M 0.000 0.000 L 0.000 5.000 L 5.000 5.000`
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestDedupSquare(t *testing.T) {
	pathStr := `M 4.180 0.000  L 1.170 0.000 L 0.000 0.000 L 0.000 5.000 
				M 0.000 5.000 M 5.000 0.000 L 5.000 5.000`

	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	transform := DedupSegmentsTransform{2}
	p, _ = transform.PathTransform(p)

	expectedStr := `M 4.180 0.000 L 1.170 0.000 L 0.000 0.000 L 0.000 5.000 M 5.000 0.000 L 5.000 5.000`
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestDedupMoves(t *testing.T) {
	pathStr := "L 1 3 m 10,20 M 15 17 L 34 80"
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	transform := DedupSegmentsTransform{2}
	p, _ = transform.PathTransform(p)
	expectedStr := "M 0.000 0.000 L 1.000 3.000 M 15.000 17.000 L 34.000 80.000"
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestDedupSegmentAndReverse(t *testing.T) {

	pathStr := " M 15 17 L 34 80 L 15 17"
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	transform := DedupSegmentsTransform{2}
	p, _ = transform.PathTransform(p)
	expectedStr := "M 15 17"
	actualStr := path.SvgString(p, 0)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
