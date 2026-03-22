package transforms

import (
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
