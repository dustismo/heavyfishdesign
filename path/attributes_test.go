package path

import (
	"fmt"
	"testing"
)

func TestBoxMeasurements(t *testing.T) {
	so := NewSegmentOperators()
	p, err := ParsePathFromSvg("M 2 2 l 1 1 M 4 4")
	if err != nil {
		t.Error(err)
	}

	point, err := PointPathAttribute(BottomRight, p, so)
	expected := "(X: 3.000, Y: 3.000)"
	actual := fmt.Sprintf("%s", point.StringPrecision(3))
	if expected != actual {
		t.Errorf("Expected %s, but got %s\n", expected, actual)
	}
}
