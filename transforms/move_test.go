package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestMoveTransform(t *testing.T) {
	// https://codepen.io/anon/pen/orPwgd
	pathStr := "M 30.000 30.000 l 5,5 l 10,0"
	originalPath, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	move := MoveTransform{
		Point:            path.NewPoint(10, 10),
		Handle:           "$MIDDLE_MIDDLE",
		SegmentOperators: path.NewSegmentOperators(),
	}

	p, err := move.PathTransform(originalPath)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	expectedStr := "M 2.500 7.500 L 7.500 12.500 L 17.500 12.500"
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
