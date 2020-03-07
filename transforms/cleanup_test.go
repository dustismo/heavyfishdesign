package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestCleanupBadStart(t *testing.T) {

	// test that the first move segment is from 0,0 not from
	// somewhere else
	segs := []path.Segment{
		path.MoveSegment{
			StartPoint: path.NewPoint(9, 3),
			EndPoint:   path.NewPoint(0, 0),
		},
	}

	p := path.NewPathFromSegmentsWithoutMove(segs)

	pth, err := CleanupTransform{Precision: 3}.PathTransform(p)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	if pth.Segments()[0].Start().X != 0 ||
		pth.Segments()[0].Start().Y != 0 {
		t.Errorf("Expected:0,0 start point")
	}
}
