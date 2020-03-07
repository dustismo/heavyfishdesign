package path

import (
	"fmt"
	"math"
)

// a draw is decorator of a Path which has convenient drawing operations
type Draw struct {
	path Path
}

func NewDraw() *Draw {
	d := &Draw{
		path: NewPath(),
	}

	return d
}

// Returns the underlying Path
func (d *Draw) Path() Path {
	return d.path
}

func (d *Draw) CurrentPosition() Point {
	segments := d.path.Segments()
	if len(segments) == 0 {
		return NewPoint(0, 0)
	}
	return segments[len(segments)-1].End()
}

func (d *Draw) AddSegments(segs []Segment) {
	for _, seg := range segs {
		d.AddSegment(seg)
	}
}

// tries to safely add a new segment
// if this segment doesnt start from the current position,
// then the segment start is mutated to the current position
func (d *Draw) AddSegment(seg Segment) {
	if len(d.path.Segments()) == 0 && !IsMove(seg) {
		// empty, starting with a non-move, so
		// make sure to move to the start
		// always add a move to origin to start
		d.path.AddSegments(MoveSegment{
			StartPoint: NewPoint(0, 0),
			EndPoint:   seg.Start(),
		})
	}

	cur := d.CurrentPosition()
	if !cur.Equals(seg.Start()) {
		s, err := SetSegmentStart(seg, cur)
		if err != nil {
			fmt.Printf("Error adding segment %+v -- %s", seg, err.Error())
		} else {
			seg = s
		}
	}
	d.path.AddSegments(seg)
}

func (d *Draw) MoveTo(point Point) {
	seg := MoveSegment{
		StartPoint: d.CurrentPosition(),
		EndPoint:   point,
	}
	d.path.AddSegments(seg)
}
func (d *Draw) ToAbsPosition(dxdy Point) Point {
	current := d.CurrentPosition()
	return NewPoint(
		current.X+dxdy.X,
		current.Y+dxdy.Y,
	)
}

// moves relative to current cursor position
func (d *Draw) RelMoveTo(dxdy Point) {
	// convert to absolute position.
	point := d.ToAbsPosition(dxdy)
	d.MoveTo(point)
}

// draws a smooth curve.
// Note that this is converted to a Cubic Bezier
func (d *Draw) RelSmoothCurveTo(controlPointEndDxDy, dxdy Point) {
	controlPointEnd := d.ToAbsPosition(controlPointEndDxDy)
	point := d.ToAbsPosition(dxdy)
	d.SmoothCurveTo(controlPointEnd, point)
}

// draws a smooth curve.
// Note that this is converted to a Cubic Bezier
func (d *Draw) SmoothCurveTo(controlPointEnd, point Point) {
	current := d.CurrentPosition()
	cx1 := current.X
	cy1 := current.Y
	segments := d.path.Segments()
	if len(segments) > 0 {
		seg, ok := segments[len(segments)-1].(CurveSegment)

		if ok {
			// reflect the point around the current position
			cx1 = current.X +
				(current.X - seg.ControlPointEnd.X)
			cy1 = current.Y +
				(current.Y - seg.ControlPointEnd.Y)
		}
	}
	controlPointStart := NewPoint(cx1, cy1)
	d.CurveTo(controlPointStart, controlPointEnd, point)
}

// relative Quadratic curve.
// Note - we convert to cubic curves
func (d *Draw) RelQCurveTo(controlPointEndDxDy, dxdy Point) {
	controlPointEnd := d.ToAbsPosition(controlPointEndDxDy)
	point := d.ToAbsPosition(dxdy)
	d.QCurveTo(controlPointEnd, point)
}

// Quadratic curve.
// Note - we convert to cubic curves
func (d *Draw) QCurveTo(controlPoint, point Point) {
	// Convert to Cubic Bezier
	// https://stackoverflow.com/questions/3162645/convert-a-quadratic-bezier-to-a-cubic-one
	// CP1 = QP0 + 2/3 *(QP1-QP0)
	// CP2 = QP2 + 2/3 *(QP1-QP2)

	current := d.CurrentPosition()
	// control point 1
	cpx1 := current.X + (2.0/3.0)*(controlPoint.X-current.X)
	cpy1 := current.Y + (2.0/3.0)*(controlPoint.Y-current.X)

	// control point 2
	cpx2 := point.X + (2.0/3.0)*(controlPoint.X-point.X)
	cpy2 := point.Y + (2.0/3.0)*(controlPoint.Y-point.Y)

	d.CurveTo(NewPoint(cpx1, cpy1), NewPoint(cpx2, cpy2), point)
}

// draws a Relative Curve
// this is the Cubic Bezier
func (d *Draw) RelCurveTo(controlPointStartDxDy, controlPointEndDxDy, dxdy Point) {
	d.CurveTo(
		d.ToAbsPosition(controlPointStartDxDy),
		d.ToAbsPosition(controlPointEndDxDy),
		d.ToAbsPosition(dxdy),
	)
}

// draws a Curve
// this is the Cubic Bezier
func (d *Draw) CurveTo(controlPointStart, controlPointEnd, point Point) {
	seg := CurveSegment{
		StartPoint:        d.CurrentPosition(),
		EndPoint:          point,
		ControlPointStart: controlPointStart,
		ControlPointEnd:   controlPointEnd,
	}
	d.path.AddSegments(seg)
}

// Relative line to
func (d *Draw) RelLineTo(dxdy Point) {
	point := d.ToAbsPosition(dxdy)
	d.LineTo(point)
}

// line to
func (d *Draw) LineTo(point Point) {
	seg := LineSegment{
		StartPoint: d.CurrentPosition(),
		EndPoint:   point,
	}
	d.path.AddSegments(seg)
}

// Draws a line from the current position based on length and angle
func (d *Draw) LineByAngle(length, angle float64) {
	d.path.AddSegments(NewLineSegmentAngle(d.CurrentPosition(), length, angle))
}

func (d *Draw) RelRoundedCornerTo(dxdy Point, dxdyCorner Point, radius float64) {
	point := d.ToAbsPosition(dxdy)
	corner := d.ToAbsPosition(dxdyCorner)
	d.RoundedCornerTo(point, corner, radius)
}

// Round corner
// draws a 90 deg corner using a curve
func (d *Draw) RoundedCornerTo(to Point, corner Point, radius float64) {

	// fractional length of the control point
	ctrl := (4 * (math.Sqrt(2) - 1) / 3)

	sPoint := d.CurrentPosition()
	ePoint := to

	sLine := LineSegment{sPoint, corner}
	// first line segment
	l1 := LineSegment{
		StartPoint: sPoint,
		EndPoint:   sLine.PointAtDistance(sLine.Length() - radius),
	}

	eLine := LineSegment{corner, ePoint}
	// end line segment
	e1 := LineSegment{
		StartPoint: eLine.PointAtDistance(radius),
		EndPoint:   ePoint,
	}

	// draw the curve from s1.End to e1.Start
	ctrlPtLineS := LineSegment{l1.End(), corner}
	ctrlPointStart := ctrlPtLineS.PointAtDistance(ctrlPtLineS.Length() * ctrl)

	ctrlPtLineE := LineSegment{e1.Start(), corner}
	ctrlPointEnd := ctrlPtLineE.PointAtDistance(ctrlPtLineE.Length() * ctrl)

	d.LineTo(l1.End())
	d.CurveTo(ctrlPointStart, ctrlPointEnd, e1.Start())
	d.LineTo(e1.End())
}

// Draws an approximation of a circle at the current location
// This uses bezier curves so a perfect circle is impossible.
// origin is top left
func (d *Draw) Circle(r float64) {
	// see: https://stackoverflow.com/questions/1734745/how-to-create-circle-with-b%C3%A9zier-curves
	ctrl := (4 * (math.Sqrt(2) - 1) / 3) * r
	d.RelMoveTo(NewPoint(r, 0))
	d.RelCurveTo(NewPoint(ctrl, 0), NewPoint(r, r-ctrl), NewPoint(r, r))      // top right
	d.RelCurveTo(NewPoint(0, ctrl), NewPoint(-r+ctrl, r), NewPoint(-r, r))    // bottom right
	d.RelCurveTo(NewPoint(-ctrl, 0), NewPoint(-r, -r+ctrl), NewPoint(-r, -r)) // bottom left
	d.RelCurveTo(NewPoint(0, -ctrl), NewPoint(r-ctrl, -r), NewPoint(r, -r))   // top left
}

// draws a rectangle from the current location, ending in the current location
func (d *Draw) Rect(w, h float64) {
	d.RelLineTo(NewPoint(w, 0))
	d.RelLineTo(NewPoint(0, h))
	d.RelLineTo(NewPoint(-w, 0))
	d.RelLineTo(NewPoint(0, -h))
}

// Draws a horizontal line relative to the current position
func (d *Draw) RelHLineTo(dx float64) {
	d.RelLineTo(NewPoint(dx, 0.0))
}

// Draws a horizontal line from the current position to the passed in position
func (d *Draw) HLineTo(dx float64) {
	d.LineTo(NewPoint(dx, d.CurrentPosition().Y))
}

// Draws a vertical line relative to the current position
func (d *Draw) RelVLineTo(dy float64) {
	d.RelLineTo(NewPoint(0.0, dy))
}

// Draws a vertical line from the current position to the passed in position
func (d *Draw) VLineTo(dy float64) {
	d.LineTo(NewPoint(d.CurrentPosition().X, dy))
}

// adds all segments from an svg path string
func (d *Draw) SvgPath(svg string) error {
	p, err := ParsePathFromSvg(svg)
	if err != nil {
		return err
	}
	for _, s := range p.Segments() {
		d.AddSegment(s)
	}
	return nil
}
