package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestSegmentReverse(t *testing.T) {
	// https://codepen.io/anon/pen/orPwgd
	pathStr := "M 100.000 200.000 C 100.000 200.000 200.000 300.000 300.000 200.000"
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	segment := p.Segments()[1]
	reverse := SegmentReverse{}

	reversedSegment := reverse.SegmentTransform(segment)

	expectedStr := "C 200.000 300.000 100.000 200.000 100.000 200.000"
	actualStr := reversedSegment.SvgString(3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestPathReverse(t *testing.T) {
	// https://codepen.io/anon/pen/orPwgd
	pathStr := "M 101.7 202 C 101 202 200 300 300 200 L 312 222 L 321.9 201.98"
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	reversedPath, err := PathReverse{}.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	reversedPath, _ = DedupSegmentsTransform{3}.PathTransform(reversedPath)
	expectedStr := "M 321.900 201.980 L 312.000 222.000 L 300.000 200.000 C 200.000 300.000 101.000 202.000 101.700 202.000 M 0.000 0.000"

	actualStr := path.SvgString(reversedPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
