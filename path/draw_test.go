package path

import "testing"

// Regression: Q→cubic conversion must use (controlPoint.Y - current.Y) for cpy1, not current.X.
func TestQCurveToUsesYDeltaForFirstCubicControl(t *testing.T) {
	d := NewDraw()
	d.MoveTo(NewPoint(10, 20))
	d.QCurveTo(NewPoint(50, 100), NewPoint(100, 0))
	segs := d.Path().Segments()
	if len(segs) < 2 {
		t.Fatalf("expected move + cubic, got %d segments", len(segs))
	}
	c, ok := segs[len(segs)-1].(CurveSegment)
	if !ok {
		t.Fatalf("expected CurveSegment, got %T", segs[len(segs)-1])
	}
	wantCP1Y := 20.0 + (2.0/3.0)*(100.0-20.0)
	if c.ControlPointStart.Y != wantCP1Y {
		t.Errorf("ControlPointStart.Y = %v, want %v", c.ControlPointStart.Y, wantCP1Y)
	}
}

func TestDrawCornerCurve(t *testing.T) {
	d := NewDraw()
	d.MoveTo(NewPoint(0, 0))
	d.RoundedCornerTo(NewPoint(5, 5), NewPoint(0, 5), 2)
	p := d.Path()
	actual := SvgString(p, 3)
	expected := "M 0.000 0.000 L 0.000 3.000 C 0.000 4.105 0.895 5.000 2.000 5.000 L 5.000 5.000"
	if actual != expected {
		t.Errorf("Expected: %s\nActual:%s\n", expected, actual)
	}
}
