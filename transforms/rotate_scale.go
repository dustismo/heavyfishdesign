package transforms

import (
	"math"

	"github.com/dustismo/heavyfishdesign/path"
)

// Will rotate and scale the path so that the PathStartPoint and PathEndPoint equal StartPoint
// and EndPoint.  This is useful for using svg to connect two points
type RotateScaleTransform struct {
	StartPoint path.Point
	EndPoint   path.Point
	// What point on the path should be considered origin?
	// defaults to path start
	PathStartPoint path.Point
	// What point on the path should be considered the end?
	// defaults to path end
	PathEndPoint     path.Point
	SegmentOperators path.SegmentOperators
}

func (rt RotateScaleTransform) line(p path.Path) (path.LineSegment, error) {
	startPoint := rt.PathStartPoint
	endPoint := rt.PathEndPoint
	if path.IsPoint00(startPoint) && path.IsPoint00(endPoint) {
		sp, err := path.PointPathAttribute(path.StartPoint, p, rt.SegmentOperators)
		if err != nil {
			return path.LineSegment{}, err
		}
		ep, err := path.PointPathAttribute(path.EndPoint, p, rt.SegmentOperators)
		if err != nil {
			return path.LineSegment{}, err
		}
		startPoint, endPoint = sp, ep
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
	//first rotate path to the requested angle
	pth, err := RotateTransform{
		Degrees:          requestedLine.Angle() - curLine.Angle(),
		Axis:             path.Origin,
		SegmentOperators: rt.SegmentOperators,
	}.PathTransform(p)

	if err != nil {
		return nil, err
	}
	// rotate the curLine to match the path rotation
	curLineS, err := RotateTransform{
		Degrees:          requestedLine.Angle() - curLine.Angle(),
		Axis:             path.Origin,
		SegmentOperators: rt.SegmentOperators,
	}.PathTransform(path.NewPathFromSegmentsWithoutMove([]path.Segment{
		curLine,
	}))
	if err != nil {
		return nil, err
	}
	curLine = curLineS.Segments()[1].(path.LineSegment)

	newX := math.Abs(requestedLine.EndPoint.X - requestedLine.StartPoint.X)
	newY := math.Abs(requestedLine.EndPoint.Y - requestedLine.StartPoint.Y)

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
