package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

type SegmentReverse struct {
}

type PathReverse struct {
}

// Reverses a single segment.  Note that this flips the segment, but
// will need a Move to start if you expect it to render properly within a path
func (s SegmentReverse) SegmentTransform(segment path.Segment) path.Segment {
	// line to
	switch seg := segment.(type) {
	case path.MoveSegment:
		// easy, just switch the points
		return path.MoveSegment{
			StartPoint: seg.End(),
			EndPoint:   seg.Start(),
		}

	case path.LineSegment:
		// easy, just switch the points
		return path.LineSegment{
			StartPoint: seg.End(),
			EndPoint:   seg.Start(),
		}
	case path.CurveSegment:
		return path.CurveSegment{
			StartPoint:        seg.End(),
			EndPoint:          seg.Start(),
			ControlPointStart: seg.ControlPointEnd,
			ControlPointEnd:   seg.ControlPointStart,
		}
	}
	return segment
}

func (pr PathReverse) PathTransform(p path.Path) (path.Path, error) {
	originalSegments := p.Segments()
	// copy into a new array, so we don't mutate the original
	segments := make([]path.Segment, len(originalSegments))
	copy(segments, originalSegments)
	reverser := SegmentReverse{}
	// reverse the segment order
	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}

	// now convert back to a list of commands
	d := path.NewDraw()
	for _, seg := range segments {
		seg = reverser.SegmentTransform(seg)
		d.AddSegment(seg)
	}
	return d.Path(), nil
}
