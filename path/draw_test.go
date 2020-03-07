package path

import "testing"

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
