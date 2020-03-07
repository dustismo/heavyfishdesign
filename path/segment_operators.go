package path

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/bezier"
)

// various operations we should be able to perform on segments
type SegmentOperators interface {
	Reverse(segment Segment) (Segment, error)
	BoundingBox(segment Segment) (topLeft, bottomRight Point, err error)
	// find the intersection between the two segments
	Intersect(segment1, segment2 Segment) ([]Point, error)
	// Splits the segment at the requested point
	// if point does not exist on the segment then ??
	Split(segment Segment, point Point) ([]Segment, error)
	Offset(segment Segment, distance float64) ([]Segment, error)
	// Joins the two disjoint segments into on continuous path
	// possibly returns multiple segments
	Join(segment1, segment2 Segment) ([]Segment, error)
	// transforms all the points associated with the segment
	TransformPoints(segment Segment, pt PointTransform) (Segment, error)
}

type PointTransform func(p Point) Point

type Curve interface {
	Start() Point
	End() Point
	ControlStart() Point
	ControlEnd() Point
}

type CurveOperators interface {
	BoundingBox(curve Curve) (topLeft, bottomRight Point, err error)
	IntersectLine(curve Curve, line LineSegment) ([]Point, error)
	IntersectProjectedLine(curve Curve, line LineSegment) ([]Point, error)
	IntersectCurve(c1 Curve, c2 Curve) ([]Point, error)
	Split(curve Curve, point Point) ([]Curve, error)
	Offset(curve Curve, distance float64) ([]Curve, error)
}

// contains the segment operators for all the segments except
// some special Curve operators
type DefaultSegmentOperators struct {
	CurveOperators CurveOperators
	Precision      int
}

func NewSegmentOperators() SegmentOperators {
	return DefaultSegmentOperators{
		CurveOperators: NewBezierCurveOperators(),
		Precision:      3,
	}
}

// the default implementation of curve operators.
type BezierCurveOperators struct {
	// 0.5?
	CurveIntersectionThreshold float64
}

func NewBezierCurveOperators() BezierCurveOperators {
	return BezierCurveOperators{
		CurveIntersectionThreshold: .5,
	}
}

func (do DefaultSegmentOperators) TransformPoints(s Segment, pt PointTransform) (Segment, error) {
	switch seg := s.(type) {
	case MoveSegment:
		return MoveSegment{
			StartPoint: pt(seg.Start()),
			EndPoint:   pt(seg.End()),
		}, nil

	case LineSegment:
		return LineSegment{
			StartPoint: pt(seg.Start()),
			EndPoint:   pt(seg.End()),
		}, nil

	case CurveSegment:
		return CurveSegment{
			StartPoint:        pt(seg.Start()),
			EndPoint:          pt(seg.End()),
			ControlPointStart: pt(seg.ControlPointStart),
			ControlPointEnd:   pt(seg.ControlPointEnd),
		}, nil
	}
	return nil, fmt.Errorf("Error unable to transform segment %+v", s)
}

func (do DefaultSegmentOperators) Move(s Segment, amount Point) (Segment, error) {
	pt := func(p Point) Point {
		return NewPoint(p.X+amount.X, p.Y+amount.Y)
	}
	return do.TransformPoints(s, pt)
}

// Joins the two lines by finding the intersection
func (do DefaultSegmentOperators) JoinLines(l1, l2 LineSegment) ([]Segment, error) {
	pnt, success := LineIntersection(l1, l2, do.Precision)

	if !success {
		if l1.End().EqualsPrecision(l2.Start(), do.Precision) {
			// line slopes are equal and they are joined, so combine into a single point
			return []Segment{
				LineSegment{
					StartPoint: l1.Start(),
					EndPoint:   l2.End(),
				},
			}, nil
		}

		// lines are parallel, so cannot be joined
		return []Segment{l1, l2}, nil
	}
	return []Segment{
		LineSegment{
			StartPoint: l1.Start().Clone(),
			EndPoint:   pnt.Clone(),
		},
		LineSegment{
			StartPoint: pnt.Clone(),
			EndPoint:   l2.End(),
		},
	}, nil

}

// joins the line and curve, by projecting the line until it intersects with the curve
func (do DefaultSegmentOperators) JoinLineAndCurve(c1 LineSegment, c2 CurveSegment) ([]Segment, error) {
	// find the intersection points
	points, err := do.CurveOperators.IntersectProjectedLine(c2, c1)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		// curve and line don't actaully meet!
		// Just add a line segment to join them
		// TODO: this should be better
		return []Segment{
			c1,
			LineSegment{
				StartPoint: c1.EndPoint,
				EndPoint:   c2.StartPoint,
			},
			c2,
		}, nil
	}

	// find the closest breakpoint
	breakPoint := points[0]
	distance := Distance(c1.EndPoint, points[0])

	for _, p := range points {
		d := Distance(c1.EndPoint, p)
		if d < distance {
			distance = d
			breakPoint = p
		}
	}

	// now break the curve
	segs, err := do.Split(c2, breakPoint)
	if err != nil {
		return nil, err
	}
	newCurve := segs[0]
	if len(segs) == 2 {
		newCurve = segs[1]
	}
	return []Segment{
		LineSegment{
			StartPoint: c1.StartPoint,
			EndPoint:   breakPoint,
		},
		newCurve,
	}, nil
}

func (do DefaultSegmentOperators) JoinCurveAndLine(c1 CurveSegment, c2 LineSegment) ([]Segment, error) {
	// find the intersection points
	points, err := do.CurveOperators.IntersectProjectedLine(c1, c2)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		// curve and line don't actaully meet!
		// Just add a line segment to join them
		// TODO: this should be better
		return []Segment{
			c1,
			LineSegment{
				StartPoint: c1.EndPoint,
				EndPoint:   c2.StartPoint,
			},
			c2,
		}, nil
	}

	// find the closest breakpoint
	breakPoint := points[0]
	distance := Distance(c1.EndPoint, points[0])

	for _, p := range points {
		d := Distance(c1.EndPoint, p)
		if d < distance {
			distance = d
			breakPoint = p
		}
	}

	// now break the curve
	segs, err := do.Split(c1, breakPoint)
	if err != nil {
		return nil, err
	}
	return []Segment{
		segs[0],
		LineSegment{
			StartPoint: breakPoint,
			EndPoint:   c2.EndPoint,
		},
	}, nil
}

func (do DefaultSegmentOperators) JoinCurves(c1, c2 CurveSegment) ([]Segment, error) {
	pnts, err := do.Intersect(c1, c2)
	if err != nil && len(pnts) > 0 {
		// the curves intersect, so just split them on that point
		c1segs, err := do.Split(c1, pnts[0])
		if err != nil {
			return nil, err
		}
		c2segs, err := do.Split(c2, pnts[0])
		if err != nil {
			return nil, err
		}
		return []Segment{
			c1segs[0],
			c2segs[1],
		}, nil
	}

	// segments don't intersect, so we need to be clever...
	// draw a line between them and split in half, then update the curves
	// to start and end at that point.  simple, and I bet this looks great, yah..
	midpoint := NewPoint(
		c1.End().X+(c2.Start().X-c1.End().X),
		c1.End().Y+(c2.Start().Y-c1.End().Y),
	)
	r1 := CurveSegment{
		StartPoint:        c1.StartPoint,
		ControlPointStart: c1.ControlPointStart,
		EndPoint:          midpoint,
		ControlPointEnd:   c1.ControlPointEnd,
	}
	r2 := CurveSegment{
		StartPoint:        midpoint,
		ControlPointStart: c2.ControlPointStart,
		EndPoint:          c2.EndPoint,
		ControlPointEnd:   c2.ControlPointEnd,
	}
	return []Segment{r1, r2}, nil
}

func (do DefaultSegmentOperators) Join(s1, s2 Segment) ([]Segment, error) {
	if s1.End().EqualsPrecision(s2.Start(), do.Precision) {
		ns2, err := SetSegmentStart(s2, s1.End())
		if err != nil {
			return []Segment{s1, s2}, err
		}
		return []Segment{s1, ns2}, nil
	}

	switch seg1 := s1.(type) {
	case LineSegment:
		switch seg2 := s2.(type) {
		case LineSegment: // yah two lines
			return do.JoinLines(seg1, seg2)
		case CurveSegment:
			return do.JoinLineAndCurve(seg1, seg2)
		}
	case CurveSegment:
		switch seg2 := s2.(type) {
		case CurveSegment: // curve and curve
			return do.JoinCurves(seg1, seg2)
		case LineSegment:
			return do.JoinCurveAndLine(seg1, seg2)
		}
	}
	// simplest thing ever, draw a connecting line.  w00t
	return []Segment{
		s1,
		LineSegment{
			StartPoint: s1.End().Clone(),
			EndPoint:   s2.End().Clone(),
		},
		s2,
	}, nil
}

func (do DefaultSegmentOperators) Reverse(segment Segment) (Segment, error) {
	// line to
	switch seg := segment.(type) {
	case MoveSegment:
		// easy, just switch the points
		return MoveSegment{
			StartPoint: seg.End(),
			EndPoint:   seg.Start(),
		}, nil

	case LineSegment:
		// easy, just switch the points
		return LineSegment{
			StartPoint: seg.End(),
			EndPoint:   seg.Start(),
		}, nil
	case CurveSegment:
		return CurveSegment{
			StartPoint:        seg.End(),
			EndPoint:          seg.Start(),
			ControlPointStart: seg.ControlPointEnd,
			ControlPointEnd:   seg.ControlPointStart,
		}, nil
	}
	return nil, fmt.Errorf("Error, unable to reverse segment %+v", segment)
}

func (do DefaultSegmentOperators) BoundingBox(segment Segment) (topLeft, bottomRight Point, err error) {
	// line to
	switch seg := segment.(type) {
	case CurveSegment:
		return do.CurveOperators.BoundingBox(seg)

	default:
		minX := seg.Start().X
		maxX := seg.Start().X
		if seg.End().X < minX {
			minX = seg.End().X
		} else {
			maxX = seg.End().X
		}

		minY := seg.Start().Y
		maxY := seg.Start().Y
		if seg.End().Y < minY {
			minY = seg.End().Y
		} else {
			maxY = seg.End().Y
		}
		return NewPoint(minX, minY), NewPoint(maxX, maxY), nil
	}
	return topLeft, bottomRight, fmt.Errorf("Error, unable to find bounding box for segment %+v", segment)
}

func (do DefaultSegmentOperators) Intersect(s1, s2 Segment) ([]Point, error) {
	switch seg1 := s1.(type) {
	case CurveSegment:
		switch seg2 := s2.(type) {
		case LineSegment:
			return do.CurveOperators.IntersectLine(seg1, seg2)
		case CurveSegment:
			return do.CurveOperators.IntersectCurve(seg1, seg2)
		case MoveSegment:
			return []Point{}, nil
		}
	case LineSegment:
		switch seg2 := s2.(type) {
		case LineSegment:
			topL1, bottomR1, err := do.BoundingBox(seg1)
			if err != nil {
				return []Point{}, err
			}
			topL2, bottomR2, err := do.BoundingBox(seg2)
			if err != nil {
				return []Point{}, err
			}
			intersection, success := LineIntersection(seg1, seg2, do.Precision)
			if !success {
				// lines are parrallel
				return []Point{}, nil
			}
			if !PrecisionPointInBoundingBox(topL1, bottomR1, intersection, do.Precision) ||
				!PrecisionPointInBoundingBox(topL2, bottomR2, intersection, do.Precision) {
				// line intersection is outside the segments
				return []Point{}, nil
			}
			return []Point{intersection}, nil
		case CurveSegment:
			return do.CurveOperators.IntersectLine(seg2, seg1)
		case MoveSegment:
			return []Point{}, nil
		}
	case MoveSegment:
		return []Point{}, nil
	}
	return []Point{}, fmt.Errorf("Unable to intersect %+v and %+v", s1, s2)
}

func (do DefaultSegmentOperators) Split(segment Segment, point Point) (ret []Segment, err error) {
	if point.EqualsPrecision(segment.Start(), 7) ||
		point.EqualsPrecision(segment.End(), 7) {
		return []Segment{segment}, nil
	}

	switch seg := segment.(type) {
	case CurveSegment:
		curves, err := do.CurveOperators.Split(seg, point)
		if err != nil {
			return ret, err
		}
		for _, c := range curves {
			ret = append(ret,
				CurveSegment{
					StartPoint:        c.Start(),
					ControlPointStart: c.ControlStart(),
					EndPoint:          c.End(),
					ControlPointEnd:   c.ControlEnd(),
				},
			)
		}
	case MoveSegment:
		ret = append(ret,
			MoveSegment{
				StartPoint: seg.Start(),
				EndPoint:   point,
			},
			MoveSegment{
				StartPoint: point,
				EndPoint:   seg.End(),
			},
		)
	case LineSegment:
		ret = append(ret,
			LineSegment{
				StartPoint: seg.Start(),
				EndPoint:   point,
			},
			LineSegment{
				StartPoint: point,
				EndPoint:   seg.End(),
			},
		)
	}
	return ret, err
}

func (do DefaultSegmentOperators) Offset(segment Segment, distance float64) (ret []Segment, err error) {
	// if start and end are the same we just remove it
	if segment.Start().EqualsPrecision(segment.End(), do.Precision) {
		return []Segment{}, nil
	}
	switch seg := segment.(type) {
	case CurveSegment:
		curves, err := do.CurveOperators.Offset(seg, distance)
		if err != nil {
			return ret, err
		}
		for _, c := range curves {
			ret = append(ret,
				CurveSegment{
					StartPoint:        c.Start(),
					ControlPointStart: c.ControlStart(),
					EndPoint:          c.End(),
					ControlPointEnd:   c.ControlEnd(),
				},
			)
		}
	/*
	 * NOTE: We reverse the distance for move and
	 * line since the curve offset is maybe backwards?
	 */
	case MoveSegment:
		result := Parallel(LineSegment{
			StartPoint: seg.Start(),
			EndPoint:   seg.End(),
		}, -distance)

		ret = append(ret,
			MoveSegment{
				result.Start(),
				result.End(),
			},
		)
	case LineSegment:
		ret = append(ret,
			Parallel(seg, -distance),
		)
	}

	// potentially need to adjust the first Move statement
	if len(ret) > 1 {
		_, ok := ret[0].(MoveSegment)
		if ok {
			ret[0] = MoveSegment{
				StartPoint: NewPoint(0, 0),
				EndPoint:   ret[1].End().Clone(),
			}
		}
	}

	return ret, err
}

func tP(curve bezier.CubicCurve, t float64) Point {
	p := bpP(bezier.FindPoint(curve, t))
	p.t = t
	return p
}
func bpP(p bezier.Point) Point {
	return NewPoint(
		p.X, p.Y,
	)
}

func bPP(p Point) bezier.Point {
	return bezier.NewPoint(
		p.X, p.Y,
	)
}
func bcC(curve bezier.CubicCurve) Curve {
	return CurveSegment{
		StartPoint:        bpP(curve.Start),
		ControlPointStart: bpP(curve.StartControl),
		EndPoint:          bpP(curve.End),
		ControlPointEnd:   bpP(curve.EndControl),
	}
}

func bCC(curve Curve) bezier.CubicCurve {
	return bezier.CubicCurve{
		Start:        bPP(curve.Start()),
		StartControl: bPP(curve.ControlStart()),
		End:          bPP(curve.End()),
		EndControl:   bPP(curve.ControlEnd()),
	}
}

func (b BezierCurveOperators) getTfromPoint(c bezier.CubicCurve, p Point) float64 {
	if p.t > 0 {
		return p.t
	}
	//estimate
	_, _, t := bezier.Project(c, bPP(p))
	return t
}

func (b BezierCurveOperators) BoundingBox(curve Curve) (topLeft, bottomRight Point, err error) {
	c := bCC(curve)
	tl, br := bezier.BoundingBox(c)
	return NewPoint(tl.X, tl.Y), NewPoint(br.X, br.Y), err
}

func (b BezierCurveOperators) IntersectLine(curve Curve, line LineSegment) ([]Point, error) {
	c := bCC(curve)
	tVals := bezier.IntersectsLine(c, bPP(line.Start()), bPP(line.End()))
	points := []Point{}
	for _, t := range tVals {
		points = append(points, tP(c, t))
	}
	return points, nil
}

// find the points where the given line *would* intersect if it were projected in
// either direction.
func (b BezierCurveOperators) IntersectProjectedLine(curve Curve, line LineSegment) ([]Point, error) {
	c := bCC(curve)
	tVals := bezier.IntersectsProjectedLine(c, bPP(line.Start()), bPP(line.End()))
	points := []Point{}
	for _, t := range tVals {
		points = append(points, tP(c, t))
	}
	return points, nil
}

func (b BezierCurveOperators) IntersectCurve(curve1 Curve, curve2 Curve) ([]Point, error) {
	c1 := bCC(curve1)
	c2 := bCC(curve2)
	pairs := bezier.IntersectsCurve(c1, c2, b.CurveIntersectionThreshold)
	points := []Point{}
	for _, p := range pairs {
		t := p.Left // the points from l and r are the same so we just use left
		points = append(points, tP(c1, t))
	}
	return points, nil
}

func (b BezierCurveOperators) Split(curve Curve, point Point) ([]Curve, error) {
	c := bCC(curve)
	t := b.getTfromPoint(c, point)
	if PrecisionEquals(t, 0, 7) || PrecisionEquals(t, 1, 7) {
		// the split point is either on the start or end of the curve so do nothing
		return []Curve{curve}, nil
	}

	c1, c2 := bezier.SplitCurve(c, t)
	return []Curve{
		bcC(c1), bcC(c2),
	}, nil
}

func (b BezierCurveOperators) Offset(curve Curve, distance float64) ([]Curve, error) {
	c := bCC(curve)
	curves := bezier.Offset(c, distance)
	ret := []Curve{}
	for _, cv := range curves {
		ret = append(ret, bcC(cv))
	}
	return ret, nil
}
