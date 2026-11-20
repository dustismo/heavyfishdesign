package components

import (
	"math"
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestExtendCenterlinePastMargin(t *testing.T) {
	d := path.NewDraw()
	d.MoveTo(path.NewPoint(0.25, 1))
	d.LineTo(path.NewPoint(2, 1.5))
	out := extendCenterlinePastMargin(d.Path(), 0.25, 0.25, 3.75, 1.75, 0.1)
	start := out.Segments()[0].End()
	if start.X >= 0.25-1e-4 {
		t.Fatalf("start stayed on the border at %v\n%s", start, path.SvgString(out, 3))
	}
	if start.X > 0.25-0.1+1e-3 {
		t.Fatalf("start did not clear the band, at %v", start)
	}
}

func TestMulberry32MatchesStudio(t *testing.T) {
	rng := mulberry32(17)
	want := []float64{
		0.6771502960473299,
		0.19265692122280598,
		0.5313839064911008,
		0.1654723146930337,
		0.08778517786413431,
	}
	for i, w := range want {
		got := rng()
		if math.Abs(got-w) > 1e-12 {
			t.Fatalf("sample %d = %.16f, want %.16f", i, got, w)
		}
	}
}
