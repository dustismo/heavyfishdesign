package transforms

import (
	"math"

	"github.com/dustismo/heavyfishdesign/path"
)

type RotateScaleTransform struct {
	StartPoint       path.Point
	EndPoint         path.Point
	SegmentOperators path.SegmentOperators
}

func (rt RotateScaleTransform) line(p path.Path) (path.LineSegment, error) {
	startPoint, err := path.PointPathAttribute(path.StartPoint, p, rt.SegmentOperators)
	if err != nil {
		return path.LineSegment{}, err
	}
	endPoint, err := path.PointPathAttribute(path.EndPoint, p, rt.SegmentOperators)
	if err != nil {
		return path.LineSegment{}, err
	}
	return path.LineSegment{
		StartPoint: startPoint,
		EndPoint:   endPoint,
	}, nil
}

func (rt RotateScaleTransform) PathTransform(p path.Path) (path.Path, error) {
	requestedLine := path.LineSegment{
		StartPoint: rt.StartPoint,
		EndPoint:   rt.EndPoint,
	}

	curLine, err := rt.line(p)
	if err != nil {
		return nil, err
	}
	//first rotate to the requested angle
	pth, err := RotateTransform{
		Degrees:          requestedLine.Angle() - curLine.Angle(),
		Axis:             path.TopLeft,
		SegmentOperators: rt.SegmentOperators,
	}.PathTransform(p)

	if err != nil {
		return nil, err
	}
	newX := math.Abs(requestedLine.EndPoint.X - requestedLine.StartPoint.X)
	newY := math.Abs(requestedLine.EndPoint.Y - requestedLine.StartPoint.Y)

	// now recalculate the curLine based on the rotated line
	curLine, err = rt.line(pth)
	if err != nil {
		return nil, err
	}

	oldX := math.Abs(curLine.EndPoint.X - curLine.StartPoint.X)
	oldY := math.Abs(curLine.EndPoint.Y - curLine.StartPoint.Y)

	xScale := 0.0
	if oldX != 0 {
		xScale = newX / oldX
	}
	yScale := 0.0
	if oldY != 0 {
		yScale = newY / oldY
	}
	scale := math.Max(xScale, yScale)
	// Now scale to the requested size
	pth, err = ScaleTransform{
		ScaleX:           scale,
		ScaleY:           scale,
		SegmentOperators: rt.SegmentOperators,
	}.PathTransform(pth)
	if err != nil {
		return pth, err
	}

	// move
	pth, err = MoveTransform{
		Point:            rt.StartPoint,
		Handle:           path.StartPoint,
		SegmentOperators: rt.SegmentOperators,
	}.PathTransform(pth)
	return pth, err
}
