package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

type Axis int

const (
	Horizontal Axis = iota
	Vertical
)

// This moves the path origin to the requested point
type MirrorTransform struct {
	Axis             Axis
	Handle           path.PathAttr
	SegmentOperators path.SegmentOperators
}

func (mt MirrorTransform) PathTransform(p path.Path) (path.Path, error) {
	segments := []path.Segment{}
	if len(mt.Handle) == 0 {
		// handle should be TOP_LEFT by default..
		mt.Handle = path.TopLeft
	}
	axisPoint, err := path.PointPathAttribute(mt.Handle, p, mt.SegmentOperators)
	if err != nil {
		return p, err
	}
	// first move to 0,0 then mirror then move back.

	p, err = MoveTransform{
		Point:            path.NewPoint(0, 0),
		Handle:           mt.Handle,
		SegmentOperators: mt.SegmentOperators,
	}.PathTransform(p)

	if err != nil {
		return p, err
	}
	_, br, err := path.BoundingBoxWithWhitespace(p, mt.SegmentOperators)
	if err != nil {
		return p, err
	}

	pt := func(p path.Point) path.Point {
		newPoint := p.Clone()
		if mt.Axis == Horizontal {
			// switch Y
			newPoint.Y = br.Y - p.Y
		}
		if mt.Axis == Vertical {
			// switch X
			newPoint.X = br.X - p.X
		}
		return newPoint
	}
	for _, seg := range p.Segments() {
		s, err := mt.SegmentOperators.TransformPoints(seg, pt)
		if err != nil {
			return nil, err
		}
		segments = append(segments, s)
	}

	pth, err := path.NewPathFromSegments(segments), nil
	if err != nil {
		return pth, err
	}

	// now move back to the original origin
	pth, err = ShiftTransform{
		DeltaX:           axisPoint.X,
		DeltaY:           axisPoint.Y,
		SegmentOperators: mt.SegmentOperators,
	}.PathTransform(pth)

	return pth, err
}
