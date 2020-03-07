package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestRotateTransform(t *testing.T) {
	// https://codepen.io/anon/pen/orPwgd
	pathStr := "M 50.000 50.000 l 20,0"
	originalPath, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	transform := RotateTransform{
		Degrees:          130,
		SegmentOperators: path.NewSegmentOperators(),
	}

	p, err := transform.PathTransform(originalPath)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	expectedStr := "M 50.000 50.000 L 37.144 65.321"
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
