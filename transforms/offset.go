package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

type SizeShouldBe int

const (
	Unknown = iota
	Smaller
	Larger
)

type OffsetTransform struct {
	Precision        int
	Distance         float64
	SegmentOperators path.SegmentOperators
	SizeShouldBe     SizeShouldBe
}

// Transforms the path into an offset path at distance
// If distance is 0 than this transform does nothing.
func (ofs OffsetTransform) PathTransform(p path.Path) (path.Path, error) {
	if ofs.Distance == 0 {
		return p, nil
	}
	pths := path.SplitPathOnMove(p)

	if len(pths) > 1 {
		d := path.NewDraw()
		for _, pth := range pths {
			p1, err := ofs.PathTransform(pth)
			if err != nil {
				return p1, err
			}
			d.AddSegments(p1.Segments())
		}
		return d.Path(), nil
	}

	segments := []path.Segment{}
	segs := p.Segments()

	trimmed := path.TrimMove(segs)
	if len(trimmed) == 0 {
		return p, nil
	}

	for _, s := range segs {
		if !path.IsMove(s) {
			// we only offset non-move segments
			sgs, err := ofs.SegmentOperators.Offset(s, ofs.Distance)
			if err != nil {
				return p, err
			}
			if len(sgs) > 0 {
				segments = append(segments, path.MoveSegment{
					EndPoint: sgs[0].Start(),
				})
				segments = append(segments, sgs...)
			}
		}
	}

	// the list of segments here is disjoint, where the segment.Start does not
	// necessarily correspond to the previous End.  we need to join them or
	// else inject move segments..

	// need to close the offset path, if the original path is closed.

	shouldClose := trimmed[0].Start().Equals(trimmed[len(trimmed)-1].End())
	cleanedUpPath, err := RebuildTransform{}.PathTransform(
		path.NewPathFromSegments(segments))
	if err != nil {
		return cleanedUpPath, err
	}
	newPath, err := JoinTransform{
		Precision:        ofs.Precision,
		SegmentOperators: ofs.SegmentOperators,
		ClosePath:        shouldClose,
	}.PathTransform(cleanedUpPath)

	if ofs.SizeShouldBe != Unknown {
		//original valus
		otl, obr, err := path.BoundingBoxTrimWhitespace(p, ofs.SegmentOperators)
		if err != nil {
			return newPath, err
		}
		// original size
		oX := obr.X - otl.X
		oY := obr.Y - otl.Y
		// new values
		ntl, nbr, err := path.BoundingBoxTrimWhitespace(newPath, ofs.SegmentOperators)
		if err != nil {
			return newPath, err
		}
		nX := nbr.X - ntl.X
		nY := nbr.Y - ntl.Y

		shouldReverse := false
		switch ofs.SizeShouldBe {
		case Smaller:
			if oX <= nX || oY <= nY {
				shouldReverse = true
				// not smaller, so reverse and try again
			}
		case Larger:
			if oX >= nX || oY >= nY {
				// not larger, so reverse and try again
				shouldReverse = true
			}
		}
		if shouldReverse {
			// reverse and reexecute the offset, making sure
			// that we skip the SizeShouldBe part
			np, err := PathReverse{}.PathTransform(p)
			if err != nil {
				return np, err
			}
			return OffsetTransform{
				Precision:        ofs.Precision,
				Distance:         ofs.Distance,
				SegmentOperators: ofs.SegmentOperators,
				SizeShouldBe:     Unknown,
			}.PathTransform(np)
		}
		return newPath, nil
	}

	return newPath, err
}
