package path

import (
	"testing"
)

func TestDistance(t *testing.T) {
	l := LineSegment{
		StartPoint: NewPoint(0, 0),
		EndPoint:   NewPoint(5, 5),
	}
	length := l.Length()
	expected := 7.071067811865475
	if !PrecisionEquals(length, expected, 3) {
		t.Errorf("Expected length %.3f but got %.3f\n", expected, length)
	}
}

func TestSlope(t *testing.T) {
	l := LineSegment{
		StartPoint: NewPoint(0, 0),
		EndPoint:   NewPoint(5, 5),
	}
	slope := l.Slope()
	if slope != 1 {
		t.Errorf("Expected slope %.3f but got %.3f\n", 1.0, slope)
	}
}

func TestLineAngle(t *testing.T) {
	l := LineSegment{
		StartPoint: NewPoint(0, 0),
		EndPoint:   NewPoint(3, 2),
	}
	angle := l.Angle()
	expected := 33.690
	if !PrecisionEquals(angle, expected, 3) {
		t.Errorf("Expected angle %.3f but got %.3f\n", expected, angle)
	}

	l = LineSegment{
		StartPoint: NewPoint(3, 2),
		EndPoint:   NewPoint(0, 0),
	}

	angle = l.Angle()
	expected = -146.310
	if !PrecisionEquals(angle, expected, 3) {
		t.Errorf("Expected angle %.3f but got %.3f\n", expected, angle)
	}
}

func TestLineByAngle(t *testing.T) {
	startPoint := NewPoint(0, 0)
	length := 2.0
	angle := 18.0
	l := NewLineSegmentAngle(startPoint, length, angle)
	if !PrecisionEquals(l.Angle(), angle, 3) {
		t.Errorf("Expected angle %.3f but got %.3f\n", angle, l.Angle())
	}
}
