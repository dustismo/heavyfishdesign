package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestSliceTransform(t *testing.T) {
	// https://codepen.io/dustismo/pen/xxxvgWm?editors=1001
	pathStr := "M 0,0 C 46.434 102.260 139.261 169.468 163.395 119.000 C 170.800 103.516 171.739 76.955 160.000 30.000"
	originalPath, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	transform := HSliceTransform{
		Y:                70,
		SegmentOperators: path.NewSegmentOperators(),
		Precision:        3,
	}

	p, err := transform.PathTransform(originalPath)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	expectedStr := "M 0.000 0.000 C 11.837 26.067 26.688 49.857 42.647 70.000 M 167.684 70.000 C 166.344 58.647 163.856 45.424 160.000 30.000"
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
