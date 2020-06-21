package path

import (
	"fmt"
	"math"
	"sort"
)

const MaxInt = int(^uint(0) >> 1)
const MinInt = -MaxInt - 1

// returns true if this point is at 0,0
func IsPoint00(p Point) bool {
	return p.X == 0.0 && p.Y == 0.0
}

// will split the segment wherever the knife segment intersects
func KnifeCut(seg Segment, knife Segment, so SegmentOperators) ([]Segment, error) {
	pnts, err := so.Intersect(knife, seg)
	if err != nil {
		return []Segment{seg}, err
	}
	if len(pnts) == 0 {
		return []Segment{seg}, nil
	}
	// else split on the first pnt and recure
	segs, err := so.Split(seg, pnts[0])
	if err != nil {
		return []Segment{seg}, err
	}
	if len(segs) == 1 {
		return []Segment{seg}, nil
	}
	retSegs := []Segment{}
	for _, s := range segs {
		sp, err := KnifeCut(s, knife, so)
		if err != nil {
			return []Segment{seg}, err
		}
		retSegs = append(retSegs, sp...)
	}
	return retSegs, nil
}

// The ordered points (by increasing x) of where this path intercepts the given
// y axis
func HorizontalIntercepts(p Path, y float64, so SegmentOperators) ([]Point, error) {
	topLeft, bottomRight, err := BoundingBoxWithWhitespace(p, so)
	if err != nil {
		return nil, err
	}
	segments := p.Segments()

	// horizontal line
	line := LineSegment{
		StartPoint: NewPoint(topLeft.X, y),
		EndPoint:   NewPoint(bottomRight.X, y),
	}

	points := []Point{}
	for _, seg := range segments {
		pnt, err := so.Intersect(line, seg)
		if err != nil {
			return nil, err
		}
		points = append(points, pnt...)
	}
	// sort by x
	sort.SliceStable(points, func(i, j int) bool {
		return points[i].X < points[j].X
	})

	// fmt.Printf("outling\n%s\n", SvgString(p, 3))
	// fmt.Printf("HORIZONTAL INTERCEPT %.3f --  %+v\n", y, points)
	return points, nil
}

func BoundingBoxTrimWhitespace(p Path, so SegmentOperators) (topLeft, bottomRight Point, err error) {
	segments := TrimMove(p.Segments())
	return boundingBox(NewPathFromSegmentsWithoutMove(segments), so)
}

// gets the bounding box, including any trailing or leading whitespace
func BoundingBoxWithWhitespace(p Path, so SegmentOperators) (topLeft, bottomRight Point, err error) {
	tl, br, err := boundingBox(p, so)
	if err != nil {
		return tl, br, err
	}
	tl = NewPoint(0, 0) // always use the origin
	start := p.Segments()[0]
	end := Tail(p.Segments())
	if start.Start().X < tl.X {
		tl.X = start.Start().X
	}
	if start.Start().Y < tl.Y {
		tl.Y = start.Start().Y
	}

	if end.End().X > br.X {
		br.X = end.End().X
	}
	if end.End().Y > br.Y {
		br.Y = end.End().Y
	}
	return tl, br, err
}

// finds the bounding box of the given path, ignoring whitespace
func boundingBox(p Path, so SegmentOperators) (topLeft, bottomRight Point, err error) {
	segments := p.Segments()
	if len(segments) == 0 {
		// return 0,0 (or should we return an error?)
		return topLeft, bottomRight, nil
	}
	topLeft, bottomRight, err = so.BoundingBox(segments[0])
	if len(segments) == 1 || err != nil {
		return topLeft, bottomRight, err
	}

	for i := 1; i < len(segments); i++ {
		s := segments[i]
		if IsMove(s) {
			continue
		}
		tl, br, err := so.BoundingBox(s)
		if err != nil {
			return tl, br, err
		}
		if tl.X < topLeft.X {
			topLeft.X = tl.X
		}
		if tl.Y < topLeft.Y {
			topLeft.Y = tl.Y
		}
		if br.X > bottomRight.X {
			bottomRight.X = br.X
		}
		if br.Y > bottomRight.Y {
			bottomRight.Y = br.Y
		}
	}
	return topLeft, bottomRight, nil
}

// returns the last Segment in the list, or nil
func Tail(seg []Segment) Segment {
	if len(seg) == 0 {
		return nil
	}
	return seg[len(seg)-1]
}

// returns a segment list without the last element
// if empty, will return empty
func TrimLast(seg []Segment) []Segment {
	if len(seg) == 0 {
		return seg
	}
	return seg[:len(seg)-1]
}
func TrimFirst(seg []Segment) []Segment {
	if len(seg) < 1 {
		return seg
	}
	return seg[1:len(seg)]
}

// returns the current cursor location of the
// path.
func PathCursor(p Path) Point {
	return Tail(p.Segments()).End()
}

func HeadMove(seg []Segment) []Segment {
	// adds a move to the front of the list
	if len(seg) == 0 {
		return seg
	}
	if !IsMove(seg[0]) {
		return append([]Segment{MoveSegment{
			StartPoint: NewPoint(0, 0),
			EndPoint:   seg[0].Start(),
		}}, seg...)
	}
	return seg
}

// removes any Move segments from the tail of the list
func TrimTailMove(seg []Segment) []Segment {
	if len(seg) == 0 {
		return seg
	}
	if IsMove(seg[len(seg)-1]) {
		return TrimTailMove(seg[:len(seg)-1])
	}
	return seg
}

// removes any Moves in the front or tail of the list.
func TrimMove(seg []Segment) []Segment {
	if len(seg) == 0 {
		return seg
	}
	if IsMove(seg[0]) {
		return TrimMove(seg[1:])
	}
	return TrimTailMove(seg)
}

// if the first move does not start at 0,0 then fix it
func FixHeadMove(seg []Segment) []Segment {
	if len(seg) == 0 {
		return seg
	}
	if IsMove(seg[0]) &&
		(seg[0].Start().X != 0 || seg[0].Start().Y != 0) {
		seg[0] = MoveSegment{
			EndPoint: seg[0].End(),
		}
	}
	return seg
}

// returns true if this segment is a move
func IsMove(seg Segment) bool {
	_, isMove := seg.(MoveSegment)
	return isMove
}

// Distance finds the straightline distance between the two points
// distance will never be negative
func Distance(p1 Point, p2 Point) float64 {
	x := p1.X - p2.X
	y := p1.Y - p2.Y
	return math.Sqrt((x * x) + (y * y))
}

// Given a line segment, and distance this gives a parrallel line, at
// 90 deg and distance d
func Parallel(segment LineSegment, distance float64) LineSegment {
	// https://stackoverflow.com/questions/2825412/draw-a-parallel-line
	x1 := segment.Start().X
	y1 := segment.Start().Y
	x2 := segment.End().X
	y2 := segment.End().Y

	l := math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))

	x1p := x1 + distance*(y2-y1)/l
	x2p := x2 + distance*(y2-y1)/l
	y1p := y1 + distance*(x1-x2)/l
	y2p := y2 + distance*(x1-x2)/l
	ls := LineSegment{
		StartPoint: NewPoint(x1p, y1p),
		EndPoint:   NewPoint(x2p, y2p),
	}
	return ls
}

func PrecisionPointInBoundingBox(topL, bottomR, point Point, precision int) bool {
	return PrecisionCompare(point.X, topL.X, precision) >= 0 &&
		PrecisionCompare(point.X, bottomR.X, precision) <= 0 &&
		PrecisionCompare(point.Y, topL.Y, precision) >= 0 &&
		PrecisionCompare(point.Y, bottomR.Y, precision) <= 0
}

// returns true if the givin point is within the bounding box points
// points that lie directly on the bounding box are considered inside
func PointInBoundingBox(topL, bottomR, point Point) bool {
	return point.X >= topL.X && point.X <= bottomR.X &&
		point.Y >= topL.Y && point.Y <= bottomR.Y
}

// returns the point where the two lines would intersect,
// success is false if the lines are parrallel
func LineIntersection(l1, l2 LineSegment, precision int) (p Point, success bool) {
	l1S := l1.Slope()
	l2S := l2.Slope()
	if PrecisionEquals(l1S, l2S, precision) {
		return p, false
	}

	// need to handle vertical lines specially because slope is undefined
	if l1.IsVerticalPrecision(precision) {
		if l2.IsVerticalPrecision(precision) {
			return p, false
		}
		x := l1.Start().X
		y := l2.EvalX(x)
		return NewPoint(x, y), true
	}

	if l2.IsVerticalPrecision(precision) {
		x := l2.Start().X
		y := l1.EvalX(x)
		return NewPoint(x, y), true
	}

	x := (l2.YIntercept() - l1.YIntercept()) / (l1S - l2S)
	y := l1.EvalX(x)

	if math.IsNaN(x) || math.IsNaN(y) {
		return p, false
	}
	return NewPoint(x, y), true
}

// removes any duplicates from the array
// order may or may not be maintained
func Float64ArrayDeDup(a []float64) []float64 {
	keys := make(map[float64]bool)
	list := []float64{}
	for _, entry := range a {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// removes any duplicates maintaining ordering
func StringArrayDeDup(a []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range a {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// inserts the requested item if it does not already exist
func Float64ArrayInsertIfAbsent(a []float64, x float64) []float64 {
	if !Float64ArrayContains(a, x) {
		return append(a, x)
	}
	return a
}

func Float64ArrayContains(a []float64, x float64) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// returns the start point and the end point of the segment array
// note that the start point is the target point of the first segment
// I.e. if the segments start with a move then the start point is where
// it moves TO
func GetStartAndEnd(segments []Segment) (start, end Point) {
	if len(segments) == 0 {
		return start, end
	}
	return segments[0].End(), segments[len(segments)-1].End()
}

// Rotate the points by the given degrees.  (clockwise)
func Rotate(degree float64, point Point) Point {
	rad := -(degree * (math.Pi / 180.0))
	newX := (point.X * math.Cos(rad)) + (point.Y * math.Sin(rad))
	newY := (-point.X * math.Sin(rad)) + (point.Y * math.Cos(rad))
	return NewPoint(newX, newY)
}

// compares two floats based on the passed in precision
// f1 == f2 => 0
// f1 < f2 => -1
// f1 > f2 =>  1
func PrecisionCompare(f1, f2 float64, precision int) int {
	if PrecisionEquals(f1, f2, precision) {
		return 0
	} else if f1 < f2 {
		return -1
	}
	return 1
}

// determines equality based on the number of digits of precision
func PrecisionEquals(f1, f2 float64, precision int) bool {
	if precision < 0 {
		return f1 == f2
	}

	s1 := fmt.Sprintf(precisionStr("%.3f", precision), f1)
	s2 := fmt.Sprintf(precisionStr("%.3f", precision), f2)
	return s1 == s2
}

// splits the path into an array of paths, where each path
// begins with a move operation
func SplitPathOnMove(pth Path) []Path {
	p := NewDraw()
	paths := []Path{}
	for i, c := range pth.Segments() {
		if i > 0 && IsMove(c) {
			if !IsEmptyPath(p.Path()) {
				paths = append(paths, p.Path())
			}
			p = NewDraw()
		}
		p.AddSegment(c)
	}
	if !IsEmptyPath(p.Path()) {
		paths = append(paths, p.Path())
	}
	return paths
}

// returns true if the path either has no segments, or contains only Move segments
func IsEmptyPath(pth Path) bool {
	for _, c := range pth.Segments() {
		if !IsMove(c) {
			return false
		}
	}
	return true
}

// will trim a string to the requested length and add elipses
func StringElipses(str string, maxChars int) string {
	inStr := str
	if len(str) > maxChars {
		inStr = fmt.Sprintf("%s...", inStr[:(maxChars-3)])
	}
	return inStr
}

func DegreesToRadians(degrees float64) float64 {
	return (math.Pi / 180) * degrees
}

func PolarToCartesian(r, theta float64) Point {
	return NewPoint(r*math.Sin(theta), r*math.Cos(theta))
}
