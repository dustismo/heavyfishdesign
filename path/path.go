package path

import (
	"fmt"
	"math"
	"strings"

	"github.com/dustismo/heavyfishdesign/dynmap"
)

const (
	DefaultPrecision = 3 // how many decimal places do we want to consider
)

type Point struct {
	X float64
	Y float64
	// t value on a curve. (This is optional,
	// and we ignore 0 values for this)
	t float64
}

type Path interface {
	Segments() []Segment
	AddSegments(seg ...Segment)
	Clone() Path
}

type Segment interface {
	// SetStart(p Point) Segment
	Start() Point
	End() Point
	SvgString(numDecimals int) string
	UniqueString(numDecimals int) string
	Clone() Segment
}

type MoveSegment struct {
	StartPoint Point
	EndPoint   Point
}

type LineSegment struct {
	StartPoint Point
	EndPoint   Point
}

type CurveSegment struct {
	StartPoint        Point
	ControlPointStart Point
	EndPoint          Point
	ControlPointEnd   Point
}

func NewPath() Path {
	p := &PathImpl{}
	return p
}

// creates a new Path from the passed in segments without adding a
// move at the beginning.
func NewPathFromSegmentsWithoutMove(segments []Segment) Path {
	segs := []Segment{}
	segs = append(segs, segments...)
	segs = FixHeadMove(segs)
	return &PathImpl{
		segments: segs,
	}
}

func NewPathFromSegments(segments []Segment) Path {
	segs := []Segment{}
	if len(segments) > 0 {
		if !IsMove(segments[0]) {
			// didn't start with a move so we need to move to the start
			segs = append(segs, MoveSegment{
				StartPoint: NewPoint(0, 0),
				EndPoint:   segments[0].Start(),
			})
		}
	}
	segs = append(segs, segments...)
	segs = FixHeadMove(segs)
	return &PathImpl{
		segments: segs,
	}
}

// creates a line segment based on start point, length and angle
func NewLineSegmentAngle(start Point, length, angle float64) LineSegment {
	y := (math.Sin(angle) * length) + start.Y
	x := (length / math.Cos(angle)) + start.X
	return LineSegment{
		StartPoint: start,
		EndPoint:   NewPoint(x, y),
	}
}

func SetSegmentStart(segment Segment, start Point) (Segment, error) {
	switch s := segment.(type) {
	case MoveSegment:
		s.StartPoint = start
		return s, nil
	case LineSegment:
		s.StartPoint = start
		return s, nil
	case CurveSegment:
		s.StartPoint = start
		return s, nil
	}
	return segment, fmt.Errorf("Unable to set segment start %+v", segment)
}

func SvgString(path Path, numDecimals int) string {
	str := []string{}
	for _, s := range path.Segments() {
		str = append(str, s.SvgString(numDecimals))
	}
	return strings.Join(str, " ")
}

type PathImpl struct {
	segments []Segment
}

func (p *PathImpl) SvgString(decimals int) string {
	return ""
}

func (p *PathImpl) Segments() []Segment {
	return p.segments
}

func (p *PathImpl) AddSegments(seg ...Segment) {
	// update the startpoint

	p.segments = append(p.segments, seg...)
}

func (p *PathImpl) Clone() Path {
	segs := []Segment{}
	for _, s := range p.segments {
		segs = append(segs, s.Clone())
	}
	return &PathImpl{
		segments: segs,
	}
}
func (m MoveSegment) Start() Point {
	return m.StartPoint
}
func (m MoveSegment) SetStart(p Point) Segment {
	m.StartPoint = p
	return m
}

func (m MoveSegment) End() Point {
	return m.EndPoint
}

func (m MoveSegment) Clone() Segment {
	return MoveSegment{
		StartPoint: m.StartPoint.Clone(),
		EndPoint:   m.EndPoint.Clone(),
	}
}

func (m MoveSegment) SvgString(numDecimals int) string {
	return fmt.Sprintf(precisionStr("M %.3f %.3f", numDecimals), m.End().X, m.End().Y)
}

func (m MoveSegment) UniqueString(numDecimals int) string {
	return fmt.Sprintf(precisionStr("MOVE (%.3f, %.3f) (%.3f, %.3f)", numDecimals),
		m.Start().X,
		m.Start().Y,
		m.End().X,
		m.End().Y,
	)
}
func (l LineSegment) Start() Point {
	return l.StartPoint
}
func (l LineSegment) SetStart(p Point) Segment {
	l.StartPoint = p
	return l
}
func (l LineSegment) End() Point {
	return l.EndPoint
}

func (l LineSegment) Slope() float64 {
	a := l.Start()
	b := l.End()
	return (b.Y - a.Y) / (b.X - a.X)
}

// returns true if this line is vertical
func (l LineSegment) IsVerticalPrecision(precision int) bool {
	a := l.Start()
	b := l.End()

	return PrecisionEquals(a.X, b.X, precision)
}

func (l LineSegment) IsHorizontalPrecision(precision int) bool {
	a := l.Start()
	b := l.End()
	return PrecisionEquals(a.Y, b.Y, precision)
}

func (l LineSegment) YIntercept() float64 {
	a := l.Start()
	return a.Y - l.Slope()*a.X
}

// gets the value of Y for the given X
func (l LineSegment) EvalX(x float64) float64 {
	return l.Slope()*x + l.YIntercept()
}

func (l LineSegment) Length() float64 {
	return Distance(l.Start(), l.End())
}

// Finds the point at the specified distance from the start point in
//the direction of the end point..
func (l LineSegment) PointAtDistance(distance float64) Point {
	neg := 1.0
	if l.IsVerticalPrecision(DefaultPrecision) {
		if l.Start().Y > l.End().Y {
			neg = -1.0
		}
		return NewPoint(
			l.Start().X,
			l.Start().Y+(neg*distance),
		)
	}
	if l.IsHorizontalPrecision(DefaultPrecision) {
		if l.Start().X > l.End().X {
			neg = -1.0
		}
		return NewPoint(
			l.Start().X+(neg*distance),
			l.Start().Y,
		)
	}

	// todo: does this work when start is after end?
	m := l.Slope()
	x := distance*math.Cos(math.Atan(m)) + l.Start().X
	y := distance*math.Sin(math.Atan(m)) + l.Start().Y
	return NewPoint(x, y)
}

// the angle of the line in degrees.  where a positive horizontal line is 0
func (l LineSegment) Angle() float64 {
	xDiff := l.End().X - l.Start().X
	yDiff := l.End().Y - l.Start().Y
	return (180 / math.Pi) * math.Atan2(yDiff, xDiff)
}

func (l LineSegment) Clone() Segment {
	return LineSegment{
		StartPoint: l.StartPoint.Clone(),
		EndPoint:   l.EndPoint.Clone(),
	}
}

func (l LineSegment) SvgString(numDecimals int) string {
	return fmt.Sprintf(precisionStr("L %.3f %.3f", numDecimals), l.End().X, l.End().Y)
}

func (l LineSegment) UniqueString(numDecimals int) string {
	return fmt.Sprintf(precisionStr("LINE (%.3f, %.3f) (%.3f, %.3f)", numDecimals),
		l.Start().X,
		l.Start().Y,
		l.End().X,
		l.End().Y,
	)
}

func (c CurveSegment) Start() Point {
	return c.StartPoint
}

func (c CurveSegment) SetStart(p Point) Segment {
	c.StartPoint = p
	return c
}

func (c CurveSegment) End() Point {
	return c.EndPoint
}

func (c CurveSegment) Clone() Segment {
	return CurveSegment{
		StartPoint:        c.StartPoint.Clone(),
		ControlPointStart: c.ControlPointStart.Clone(),
		EndPoint:          c.EndPoint.Clone(),
		ControlPointEnd:   c.ControlPointEnd.Clone(),
	}
}

func (c CurveSegment) SvgString(numDecimals int) string {
	return fmt.Sprintf(precisionStr("C %.3f %.3f %.3f %.3f %.3f %.3f", numDecimals),
		c.ControlPointStart.X,
		c.ControlPointStart.Y,
		c.ControlPointEnd.X,
		c.ControlPointEnd.Y,
		c.End().X,
		c.End().Y)
}

func (c CurveSegment) UniqueString(numDecimals int) string {
	return fmt.Sprintf(precisionStr("CURVE (%.3f, %.3f) (%.3f, %.3f) (%.3f, %.3f) (%.3f, %.3f)", numDecimals),
		c.Start().X,
		c.Start().Y,
		c.ControlPointStart.X,
		c.ControlPointStart.Y,
		c.ControlPointEnd.X,
		c.ControlPointEnd.Y,
		c.End().X,
		c.End().Y,
	)
}

func (c CurveSegment) ControlStart() Point {
	return c.ControlPointStart
}

func (c CurveSegment) ControlEnd() Point {
	return c.ControlPointEnd
}

// Creates a new point, will convert -0.0 to 0.0
func NewPoint(x float64, y float64) Point {
	if x == -0.0 {
		x = 0
	}
	if y == -0.0 {
		y = 0
	}
	return Point{
		x, y, -1,
	}
}

// creates a new point, with the values rounded to 3 decimal places
func NewPointRounded(x float64, y float64) Point {
	x = math.Round(x*1000) / 1000
	y = math.Round(y*1000) / 1000
	return NewPoint(x, y)
}

func (p Point) StringRounded() string {
	x := math.Round(p.X*1000) / 1000
	y := math.Round(p.Y*1000) / 1000
	if x == -0.0 {
		x = 0
	}
	if y == -0.0 {
		y = 0
	}

	return fmt.Sprintf("(%.3f,%.3f)", x, y)
}
func (p Point) StringPrecision(numDecimals int) string {
	return fmt.Sprintf(precisionStr("(X: %.3f, Y: %.3f)", numDecimals), p.X, p.Y)
}

func (p Point) String() string {
	return fmt.Sprintf("(%.16f,%.16f)", p.X, p.Y)
}

// checks if the points are equal
// this check for exact equality (the floats must be the same)
func (p Point) Equals(other Point) bool {
	return p.X == other.X && p.Y == other.Y
}

// checks if the points are equal based to the requested number of
// decimal places
func (p Point) EqualsPrecision(other Point, numDecimals int) bool {
	if numDecimals < 0 {
		return p.Equals(other)
	}
	return p.StringPrecision(numDecimals) == other.StringPrecision(numDecimals)
}

func (p Point) Clone() Point {
	return NewPoint(p.X, p.Y)
}

func (p Point) ToDynMap() *dynmap.DynMap {
	mp := dynmap.New()
	mp.Put("x", p.X)
	mp.Put("y", p.Y)
	return mp
}

func precisionStr(str string, numDecimals int) string {
	return strings.ReplaceAll(str, "3", fmt.Sprintf("%d", numDecimals))
}
