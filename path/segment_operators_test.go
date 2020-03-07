package path

import (
	"testing"
)

func TestOffsetLine(t *testing.T) {
	ops := NewSegmentOperators()
	line := LineSegment{
		StartPoint: NewPoint(10, 10),
		EndPoint:   NewPoint(15, 30),
	}

	segs, err := ops.Offset(line, 5)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	expected := "L 10.149 31.213"
	actual := segs[0].SvgString(3)

	if expected != actual {
		t.Errorf("Error: expected %s but got %s", expected, actual)
	}
}

// func TestSplit(t *testing.T) {
// 	ops := NewSegmentOperators()
// 	curve := CurveSegment{
// 		StartPoint:        NewPoint(101.7, 202),
// 		ControlPointStart: NewPoint(101, 202),
// 		ControlPointEnd:   NewPoint(200, 300),
// 		EndPoint:          NewPoint(300, 200),
// 	}
// 	// try to split on a point not on the line.
// 	segs, err := ops.Split(curve, NewPoint(0, 15))
// 	if err != nil {
// 		t.Errorf("Error %v", err)
// 	}
// 	path := NewPathFromSegments(segs)

// 	expected := ""
// 	actual := SvgString(path, 3)

// 	if expected != actual {
// 		t.Errorf("Error: expected %s\nbut got %s\n", expected, actual)
// 	}
// }

func TestOffsetCurve(t *testing.T) {
	ops := NewSegmentOperators()
	curve := CurveSegment{
		StartPoint:        NewPoint(101.7, 202),
		ControlPointStart: NewPoint(101, 202),
		ControlPointEnd:   NewPoint(200, 300),
		EndPoint:          NewPoint(300, 200),
	}
	segs, err := ops.Offset(curve, 20)
	if err != nil {
		t.Errorf("Error %v", err)
	}
	path := NewPathFromSegments(segs)

	expected := "M 101.700 182.000 C 92.423 182.000 87.620 186.571 84.530 191.739 C 82.948 194.387 81.696 197.711 81.696 202.004 C 81.696 202.337 87.844 221.539 119.977 240.669 C 140.564 252.926 169.576 264.969 202.797 264.969 C 236.732 264.969 275.443 252.841 314.142 214.142"
	actual := SvgString(path, 3)

	if expected != actual {
		t.Errorf("Error: expected %s but got %s", expected, actual)
	}
}
