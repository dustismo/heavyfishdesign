package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

type RotateTransform struct {
	Degrees          float64
	Axis             path.PathAttr
	SegmentOperators path.SegmentOperators
}

func (rt RotateTransform) PathTransformWithAxis(pth path.Path, axisPoint path.Point) (path.Path, error) {

	segments := []path.Segment{}
	// first move to 0,0 then rotate then move back.

	p, err := MoveTransform{
		Point:            path.NewPoint(0, 0),
		Handle:           rt.Axis,
		SegmentOperators: rt.SegmentOperators,
	}.PathTransform(pth)

	if err != nil {
		return p, err
	}
	pt := func(point path.Point) path.Point {
		return path.Rotate(rt.Degrees, point)
	}
	for _, seg := range p.Segments() {
		s, err := rt.SegmentOperators.TransformPoints(seg, pt)
		if err != nil {
			return nil, err
		}
		segments = append(segments, s)
	}

	pth, err = path.NewPathFromSegments(segments), nil
	if err != nil {
		return pth, err
	}

	// now move back to the original origin
	pth, err = ShiftTransform{
		DeltaX:           axisPoint.X,
		DeltaY:           axisPoint.Y,
		SegmentOperators: rt.SegmentOperators,
	}.PathTransform(pth)

	return pth, err
}

func (rt RotateTransform) PathTransform(p path.Path) (path.Path, error) {
	if len(rt.Axis) == 0 {
		// handle should be TOP_LEFT by default..
		rt.Axis = path.TopLeft
	}
	axisPoint, err := path.PointPathAttribute(rt.Axis, p, rt.SegmentOperators)
	if err != nil {
		return p, err
	}
	return rt.PathTransformWithAxis(p, axisPoint)
}
