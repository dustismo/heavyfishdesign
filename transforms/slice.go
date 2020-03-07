package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

// Cut the path horizontally at the given Y coordinate, removes everything above Y
//
type HSliceTransform struct {
	Y                float64
	SegmentOperators path.SegmentOperators
	Precision        int
}

func (hs HSliceTransform) PathTransform(p path.Path) (path.Path, error) {
	so := hs.SegmentOperators
	topLeft, bottomRight, err := path.BoundingBoxWithWhitespace(p, so)
	if err != nil {
		return nil, err
	}
	segments := p.Segments()

	// horizontal line
	line := path.LineSegment{
		StartPoint: path.NewPoint(topLeft.X, hs.Y),
		EndPoint:   path.NewPoint(bottomRight.X, hs.Y),
	}

	newSegments := []path.Segment{}
	// first cut through the horizontal line
	for _, seg := range segments {
		segs, err := path.KnifeCut(seg, line, so)
		if err != nil {
			return nil, err
		}
		newSegments = append(newSegments, segs...)
	}

	// Now remove anything above the line
	newSegs2 := []path.Segment{}
	for _, seg := range newSegments {
		// simple int
		// -1 = unknown
		// 0 = false
		// 1 = true
		isAbove := -1
		// first do the simple fast check, if all points and control points are above or
		// below the line then we know the answer
		switch s := seg.(type) {
		case path.CurveSegment:
			if path.PrecisionCompare(s.StartPoint.Y, hs.Y, hs.Precision) >= 0 &&
				path.PrecisionCompare(s.EndPoint.Y, hs.Y, hs.Precision) >= 0 &&
				path.PrecisionCompare(s.ControlPointStart.Y, hs.Y, hs.Precision) >= 0 &&
				path.PrecisionCompare(s.ControlPointEnd.Y, hs.Y, hs.Precision) >= 0 {
				isAbove = 1
			} else if path.PrecisionCompare(s.StartPoint.Y, hs.Y, hs.Precision) <= 0 &&
				path.PrecisionCompare(s.EndPoint.Y, hs.Y, hs.Precision) <= 0 &&
				path.PrecisionCompare(s.ControlPointStart.Y, hs.Y, hs.Precision) <= 0 &&
				path.PrecisionCompare(s.ControlPointEnd.Y, hs.Y, hs.Precision) <= 0 {
				isAbove = 0
			} else {
				isAbove = -1
			}
		default:
			if path.PrecisionCompare(s.Start().Y, hs.Y, hs.Precision) >= 0 &&
				path.PrecisionCompare(s.End().Y, hs.Y, hs.Precision) >= 0 {
				isAbove = 1
			} else if path.PrecisionCompare(s.Start().Y, hs.Y, hs.Precision) <= 0 &&
				path.PrecisionCompare(s.End().Y, hs.Y, hs.Precision) <= 0 {
				isAbove = 0
			} else {
				isAbove = -1
			}
		}

		if isAbove == -1 {
			// still unknown, so calculate the bounding box.
			tl, _, err := so.BoundingBox(seg)
			if err != nil {
				return nil, err
			}
			if path.PrecisionCompare(tl.Y, hs.Y, hs.Precision) >= 0 {
				isAbove = 1
			} else {
				isAbove = 0
			}
		}

		if isAbove == 0 {
			newSegs2 = append(newSegs2, seg)
		}
	}

	// now clean up the path and return
	return CleanupTransform{Precision: hs.Precision}.PathTransform(path.NewPathFromSegmentsWithoutMove(newSegs2))
}
