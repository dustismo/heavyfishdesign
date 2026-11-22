package components

import (
	"math"
	"testing"

	"github.com/tdewolff/canvas"
)

func TestClipCircleKeepsTheRoundSide(t *testing.T) {
	// Radius 0.5 centered at (0.8, 1) crosses the right edge x=1 and stays
	// clear of the other three. A flattened boolean turns the left side into
	// a chord; the arc must still reach x=0.3.
	clipped := clipCurvesToRect(circleCanvas(0.8, 1, 0.5), 0, 0, 1, 2)
	b := clipped.Bounds()
	if math.Abs(b.X0-0.3) > 0.02 {
		t.Fatalf("left extent %v, want the circle at 0.3 (a chord would sit near 0.38)", b.X0)
	}
	if math.Abs(b.X1-1) > 0.02 {
		t.Fatalf("right extent %v, want the cut on x=1", b.X1)
	}
	if b.Y0 < -1e-4 || b.Y1 > 2+1e-4 || b.X0 < -1e-4 {
		t.Fatalf("clipped bounds %v extend outside the rectangle", b)
	}
	curves := 0
	s := clipped.Scanner()
	for s.Scan() {
		if s.Cmd() == canvas.CubeToCmd {
			curves++
		}
		if s.End().X > 1+1e-4 || s.End().X < -1e-4 || s.End().Y < -1e-4 || s.End().Y > 2+1e-4 {
			t.Fatalf("point %v outside the rectangle", s.End())
		}
	}
	if curves < 2 {
		t.Fatalf("clipped circle has %d cubics, the arc was flattened", curves)
	}
}

func TestClipQuatrefoilKeepsTheUncutLobe(t *testing.T) {
	// Arm 0.3 at (1, 0.5) reaches y=-0.1, so the top lobe is cut by y=0.
	// The bottom tip at y=1.1 must stay a curve, not a polygon of the lobe.
	clipped := clipCurvesToRect(quatrefoilCanvas(1, 0.5, 0.3), 0, 0, 2, 2)
	b := clipped.Bounds()
	if b.Y0 < -1e-3 {
		t.Fatalf("top %v is above the clip", b.Y0)
	}
	if math.Abs(b.Y1-1.1) > 0.03 {
		t.Fatalf("bottom tip %v, want 1.1", b.Y1)
	}
	curves := 0
	s := clipped.Scanner()
	for s.Scan() {
		if s.Cmd() == canvas.CubeToCmd {
			curves++
		}
	}
	if curves < 4 {
		t.Fatalf("clipped quatrefoil has %d cubics", curves)
	}
}
