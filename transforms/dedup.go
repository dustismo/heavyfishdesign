package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

type DedupSegmentsTransform struct {
	Precision int
}

func (st DedupSegmentsTransform) doTransform(p path.Path) (path.Path, error) {
	segments := p.Segments()
	reverser := SegmentReverse{}
	if len(segments) < 2 {
		return p, nil
	}
	newSegments := make([]path.Segment, 0)
	newSegments = append(newSegments, segments[0])

	for i := 1; i < len(segments); i++ {
		cur := segments[i]
		prev := newSegments[len(newSegments)-1]
		switch cur.(type) {
		case path.MoveSegment:
			// check if the prev was a move as well.
			// if so, we can overwrite the prev.
			if path.IsMove(prev) {
				prev = path.MoveSegment{
					StartPoint: prev.Start(),
					EndPoint:   cur.End(),
				}
				newSegments[len(newSegments)-1] = prev
			} else if prev.End().StringPrecision(st.Precision) == cur.End().StringPrecision(st.Precision) {
				// this moves the point to where it already is
				// we shouldn't add the move!
			} else {
				newSegments = append(newSegments, cur)
			}
		default:
			if cur.UniqueString(st.Precision) == reverser.SegmentTransform(prev).UniqueString(st.Precision) {
				// this pair is the same back and forwards.  (like a line that writes over itself)
				if len(newSegments) > 0 {
					newSegments = newSegments[:len(newSegments)-1]
				}
			} else {
				newSegments = append(newSegments, cur)
			}
		}
	}
	return path.NewPathFromSegments(newSegments), nil
}

// this will remove any redundant operations.
// currently:
// 1. collapse multiple MOVEs in a row
// 2. Remove a MOVE to the current location
// 3. if prev segment is the same as the reverse of the current segment, remove both
//    	TODO: should this be split into a its own transform? I wonder if this is ever the intent?
func (st DedupSegmentsTransform) PathTransform(p path.Path) (path.Path, error) {
	pth, err := st.doTransform(p)
	if err != nil {
		return pth, err
	}
	// run the transform twice, since there are some
	// cases where the second pass is needed.
	// TODO: this is stupid, find a cleaner way.
	// return pth, err
	pth, err = st.doTransform(pth)
	if err != nil {
		return pth, err
	}
	return CleanupTransform{Precision: st.Precision}.PathTransform(pth)
}
