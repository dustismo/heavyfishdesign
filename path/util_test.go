package path

import (
	"testing"
)

func TestKnifeCut2(t *testing.T) {
	so := NewSegmentOperators()

	line := LineSegment{
		StartPoint: NewPoint(6.522, 18.953),
		EndPoint:   NewPoint(6.521, 9.954),
	}

	knife := LineSegment{
		StartPoint: NewPoint(0.0, 9.577),
		EndPoint:   NewPoint(10.079, 9.577),
	}
	segments, _ := KnifeCut(line, knife, so)

	actual := SvgString(NewPathFromSegmentsWithoutMove(segments), 3)
	expected := "L 6.521 9.954"

	if actual != expected {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}

func TestKnifeCut(t *testing.T) {
	so := NewSegmentOperators()
	curve := CurveSegment{
		StartPoint:        NewPoint(100, 240),
		EndPoint:          NewPoint(160, 30),
		ControlPointStart: NewPoint(30, 60),
		ControlPointEnd:   NewPoint(210, 230),
	}

	knife := LineSegment{
		StartPoint: NewPoint(50, 119),
		EndPoint:   NewPoint(200, 119),
	}
	segments, _ := KnifeCut(curve, knife, so)
	actual := SvgString(NewPathFromSegmentsWithoutMove(segments), 3)
	expected := "C 46.434 102.260 139.261 169.468 163.395 119.000 C 170.800 103.516 171.739 76.955 160.000 30.000"

	if actual != expected {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}

func TestHorizontalIntercepts(t *testing.T) {
	so := NewSegmentOperators()
	curve := CurveSegment{
		StartPoint:        NewPoint(100, 240),
		EndPoint:          NewPoint(160, 30),
		ControlPointStart: NewPoint(30, 60),
		ControlPointEnd:   NewPoint(210, 230),
	}
	p := NewPathFromSegments([]Segment{curve})
	pnts, _ := HorizontalIntercepts(p, 119, so)
	actual := pnts[0].StringPrecision(3)
	expected := "(X: 163.395, Y: 119.000)"

	if actual != expected {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}

func TestBoundingBox(t *testing.T) {

}

func TestLineDiagonalIntersection(t *testing.T) {
	s1 := LineSegment{
		StartPoint: NewPoint(5.3825033905808803, 0.1692845249434733),
		EndPoint:   NewPoint(1.6237329074372457, 1.5373650982461482),
	}
	s2 := LineSegment{
		StartPoint: NewPoint(1.2817127641115771, 0.5976724774602398),
		EndPoint:   NewPoint(5.0404832472552110, -0.7704080958424351),
	}

	_, ok := LineIntersection(s1, s2, 3)

	if ok {
		t.Errorf("Line should not have intersected")
	}

}
func TestTrimMove(t *testing.T) {
	d := NewDraw()
	d.MoveTo(NewPoint(0, 0))
	d.LineTo(NewPoint(10, 10))
	d.LineTo(NewPoint(15, 15))
	d.MoveTo(NewPoint(0, 0))

	segs := TrimMove(d.Path().Segments())
	if len(segs) != 2 {
		t.Errorf("TrimMove Failed, expected %d but got %d", 2, len(segs))
	}

	d = NewDraw()
	d.MoveTo(NewPoint(0, 0))

	segs = TrimMove(d.Path().Segments())
	if len(segs) != 0 {
		t.Errorf("TrimMove Failed, expected %d but got %d", 0, len(segs))
	}

	d = NewDraw()
	d.MoveTo(NewPoint(0, 0))
	d.MoveTo(NewPoint(0, 10))

	segs = TrimMove(d.Path().Segments())
	if len(segs) != 0 {
		t.Errorf("TrimMove Failed, expected %d but got %d", 0, len(segs))
	}
}
func TestParallel(t *testing.T) {
	s1 := LineSegment{
		StartPoint: NewPoint(10, 10),
		EndPoint:   NewPoint(10, 100),
	}
	s2 := Parallel(s1, 10)

	if s2.End().X != 20 || s2.Start().X != 20 {
		t.Errorf("Error calculating parallel")
	}

	// Now test the line going in the opposite direction, parrallel should
	// be on the other side
	s1 = LineSegment{
		StartPoint: NewPoint(10, 100),
		EndPoint:   NewPoint(10, 10),
	}
	s2 = Parallel(s1, 10)

	if s2.End().X != 0 || s2.Start().X != 0 {
		t.Errorf("Error calculating parallel")
	}
}
func TestLineIntersection(t *testing.T) {
	s1 := LineSegment{
		StartPoint: NewPoint(10, 10),
		EndPoint:   NewPoint(150, 30),
	}
	s2 := LineSegment{
		StartPoint: NewPoint(10, 20),
		EndPoint:   NewPoint(100, 25),
	}

	p, _ := LineIntersection(s1, s2, 3)

	actual := p.StringPrecision(3)
	expected := "(X: 124.545, Y: 26.364)"

	if actual != expected {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}
