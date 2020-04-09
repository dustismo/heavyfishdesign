package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestMatrixTransform(t *testing.T) {
	pathStr := "M 0,0 C 46.434 102.260 139.261 169.468 163.395 119.000 C 170.800 103.516 171.739 76.955 160.000 30.000"
	originalPath, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	move := MatrixTransform{
		A:                3,
		C:                -1,
		E:                30,
		B:                1,
		D:                3,
		F:                40,
		SegmentOperators: path.NewSegmentOperators(),
	}

	p, err := move.PathTransform(originalPath)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	expectedStr := "M 30.000 40.000 C 67.042 393.214 278.315 687.665 401.185 560.395 C 438.884 521.348 468.262 442.604 480.000 290.000"
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
