package transforms

import (
	"math"

	"github.com/dustismo/heavyfishdesign/path"
)

type CleanupTransform struct {
	Precision int
}

// cleans up the path.
// 1. make sure the first operation is a move
// 2. Insert a Move if the segment Start is not the same as the segment end of the
//		previous
// 3. align start and end if they are not equal, but within precision
// 4. Check for NaN
func (st CleanupTransform) PathTransform(p path.Path) (path.Path, error) {
	segments := []path.Segment{}
	var previousEnd path.Point
	for i, s := range p.Segments() {
		if i == 0 {
			if !path.IsMove(s) {
				segments = append(segments, path.MoveSegment{
					EndPoint: s.Start(),
				})
			}
		} else if !s.Start().Equals(previousEnd) {
			// add a Move segment..
			if s.Start().EqualsPrecision(previousEnd, st.Precision) {
				// equals within the set precision.  update the start point
				s, _ = path.SetSegmentStart(s, previousEnd)
			}
			segments = append(segments, path.MoveSegment{
				StartPoint: previousEnd,
				EndPoint:   s.Start(),
			})
		}
		if !math.IsNaN(s.End().X) && !math.IsNaN(s.End().Y) {
			segments = append(segments, s)
			previousEnd = s.End()
		}
	}
	// make sure the first move starts at 0,0
	segments = path.FixHeadMove(segments)

	// now merge any moves
	segmentsTmp := []path.Segment{}
	for i, s := range segments {
		if i > 0 && path.IsMove(s) {
			lastInd := len(segmentsTmp) - 1
			// if the previous segment is a move, simply
			// overwrite it.
			prev := segmentsTmp[lastInd]
			if path.IsMove(prev) {
				segmentsTmp[lastInd] = path.MoveSegment{
					StartPoint: prev.Start(),
					EndPoint:   s.End(),
				}
			} else {
				segmentsTmp = append(segmentsTmp, s)
			}
		} else {
			segmentsTmp = append(segmentsTmp, s)
		}
	}
	segments = segmentsTmp

	return path.NewPathFromSegments(segments), nil
}
