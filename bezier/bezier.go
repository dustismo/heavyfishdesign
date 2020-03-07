package bezier

import (
	"fmt"
	"math"
	"sort"
)

// Cubic bezier operations.
// ported bezier.js (https://github.com/Pomax/bezierjs)

type CubicCurve struct {
	Start        Point
	StartControl Point
	End          Point
	EndControl   Point
}

type Point struct {
	X float64
	Y float64
}

func NewPoint(x, y float64) Point {
	return Point{x, y}
}

func (p Point) String() string {
	return fmt.Sprintf("(%.3f,%.3f)", p.X, p.Y)
}

// internal representation which also holds some extra metadata
type curveSegment struct {
	curve CubicCurve
	// the start of the curve, will be non-zero if this is a subsection of another curve
	t1 float64
	// the end of the curve, will be non-1 if this is a subsection of another curve
	t2 float64

	// lut (table of points on the curve at specific intervals)
	lut []Point
}

func normalize(curve CubicCurve) CubicCurve {
	// move control point if they are exactly on the start/end points
	// we do this because some of the algorithms below stop working in this case
	// there is probably a better way to handle this..
	if curve.Start.X == curve.StartControl.X && curve.Start.Y == curve.StartControl.Y {
		// move start control up the line
		curve.StartControl = FindPoint(curve, .1)
	}
	if curve.End.X == curve.EndControl.X && curve.End.Y == curve.EndControl.Y {
		// move end control up the line
		curve.EndControl = FindPoint(curve, 1-.01)
	}
	return curve
}

// decasteljaus algorithm to find point on a curve.  This will find the point t on the curve (where t >=0 && t <= 1)
// that is, t is the percentage along the line of the curve.
func DeCasteljau(curve CubicCurve, t float64) Point {
	// straight port of https://pomax.github.io/bezierinfo/#decasteljau
	return _deCasteljau(cubicCurveToArray(curve), t)
}

func _deCasteljau(points []Point, t float64) Point {
	if len(points) == 1 {
		return points[0]
	}
	var newPoints = points[0 : len(points)-1]
	for i := range newPoints {
		x := (1-t)*points[i].X + t*points[i+1].X
		y := (1-t)*points[i].Y + t*points[i+1].Y
		newPoints[i] = NewPoint(x, y)
	}
	return _deCasteljau(newPoints, t)
}

func arrayToCubicCurve(points []Point) CubicCurve {
	return CubicCurve{
		Start:        points[0],
		StartControl: points[1],
		EndControl:   points[2],
		End:          points[3],
	}
}

func cubicCurveToArray(curve CubicCurve) []Point {
	return []Point{
		curve.Start,
		curve.StartControl,
		curve.EndControl,
		curve.End,
	}
}

func SplitCurve(curve CubicCurve, t float64) (CubicCurve, CubicCurve) {
	left, right := _splitCurveSection(curveSegment{curve: curve}, t, -1)
	return left.curve, right.curve
}

// keeps track of the t1 and t2 internally
// set t2 < 0 to ignore it and split into two
// if both t1 and t2 are set, then only left will be returned (as a subsegment)
func _splitCurveSection(curve curveSegment, tStart float64, tEnd float64) (l, r curveSegment) {
	// shortcuts
	if tStart == 0 {
		l, _ = _splitCurveSection(curve, tEnd, -1)
		return l, r
	}
	if tEnd == 1 {
		_, r = _splitCurveSection(curve, tStart, -1)
		return r, r
	}
	q := hull(curve.curve, tStart)

	// no shortcut: use "de Casteljau" iteration.

	left := curveSegment{
		curve: arrayToCubicCurve([]Point{q[0], q[4], q[7], q[9]}),
	}
	right := curveSegment{
		curve: arrayToCubicCurve([]Point{q[9], q[8], q[6], q[3]}),
	}

	// make sure we bind _t1/_t2 information!
	left.t1 = _map(0, 0, 1, curve.t1, curve.t2)
	left.t2 = _map(tStart, 0, 1, curve.t1, curve.t2)

	right.t1 = _map(tStart, 0, 1, curve.t1, curve.t2)
	right.t2 = _map(1, 0, 1, curve.t1, curve.t2)

	// if we have no t2, we're done
	if tEnd < 0 {
		return left, right
	}

	// if we have a t2, split again:
	tEnd = _map(tEnd, tStart, 1, 0, 1)
	left, right = _splitCurveSection(right, tEnd, -1) //ignore the right side.
	return left, right
}

// splits the curve into a section
func SplitCurveSection(curve CubicCurve, tStart float64, tEnd float64) CubicCurve {
	segment := curveSegment{
		curve: curve,
		t1:    tStart,
		t2:    tEnd,
	}

	s, _ := _splitCurveSection(segment, tStart, tEnd)
	return s.curve
}

func _map(v, ds, de, ts, te float64) float64 {
	d1 := de - ds
	d2 := te - ts
	v2 := v - ds
	r := v2 / d1
	return ts + d2*r
}

//
func derive(curve CubicCurve) [][]Point {
	dpoints := [][]Point{}
	points := cubicCurveToArray(curve)
	p := points
	d := len(p)
	for c := d - 1; d > 1; d-- {
		list := []Point{}
		for j := 0; j < c; j++ {
			dpt := Point{
				X: float64(c) * (p[j+1].X - p[j].X),
				Y: float64(c) * (p[j+1].Y - p[j].Y),
			}
			// add to the head of the list
			list = append(list, dpt)
		}
		dpoints = append(dpoints, list)
		p = list
		c--
	}
	return dpoints
}

// Calculates the curve tangent at the specified t value. Note that this yields a not-normalized vector {x: dx, y: dy}
func Derivative(curve CubicCurve, t float64) Point {
	mt := 1 - t
	p := derive(curve)[0]

	a := mt * mt
	b := mt * t * 2
	c := t * t

	return Point{
		X: a*p[0].X + b*p[1].X + c*p[2].X,
		Y: a*p[0].Y + b*p[1].Y + c*p[2].Y,
	}
}

func droots(p []float64) []float64 {
	// quadratic roots are easy
	if len(p) == 3 {
		a := p[0]
		b := p[1]
		c := p[2]
		d := a - 2*b + c
		if d != 0 {
			m1 := -math.Sqrt(b*b - a*c)
			m2 := -a + b
			v1 := -(m1 + m2) / d
			v2 := -(-m1 + m2) / d
			return []float64{v1, v2}
		} else if b != c && d == 0 {
			return []float64{(2*b - c) / (2 * (b - c))}
		}
		return []float64{}
	}

	// linear roots are even easier
	if len(p) == 2 {
		a := p[0]
		b := p[1]
		if a != b {
			return []float64{a / (a - b)}
		}
		return []float64{}
	}
	return []float64{}
}

// finds the given t value on the curve.  Where t is between 0 and 1
func FindPoint(curve CubicCurve, t float64) Point {
	points := cubicCurveToArray(curve)
	// shortcuts
	if t == 0 {
		return curve.Start
	}

	if t == 1 {
		return curve.End
	}

	p := points
	mt := 1 - t
	mt2 := mt * mt
	t2 := t * t
	a := mt2 * mt
	b := mt2 * t * 3
	c := mt * t2 * 3
	d := t * t2
	return Point{
		X: a*p[0].X + b*p[1].X + c*p[2].X + d*p[3].X,
		Y: a*p[0].Y + b*p[1].Y + c*p[2].Y + d*p[3].Y,
	}
}

// Calculates all the extrema on a curve. Extrema are calculated for each dimension,
// rather than for the full curve, so that the result is not the number of convex/concave transitions,
// but the number of those transitions for each separate dimension.
// This function yields x: [num, num, ...], y: [...] where each dimension lists the array of t values
// at which an extremum occurs
//
// These points can be used to determine the reach of a curve.
// see https://pomax.github.io/bezierjs/#extrema
func extrema(curve CubicCurve) ([]float64, []float64) {
	dpoints := derive(curve)

	px := []float64{}
	py := []float64{}
	for _, p := range dpoints[0] {
		px = append(px, p.X)
		py = append(py, p.Y)
	}

	px1 := []float64{}
	py1 := []float64{}
	for _, p := range dpoints[1] {
		px1 = append(px1, p.X)
		py1 = append(py1, p.Y)
	}

	x := append(droots(px), droots(px1)...)
	y := append(droots(py), droots(py1)...)

	// now filter out the items outside the range
	retX := []float64{}
	retY := []float64{}
	for _, p := range x {
		if p >= 0 && p <= 1 {
			retX = append(retX, p)
		}
	}

	for _, p := range y {
		if p >= 0 && p <= 1 {
			retY = append(retY, p)
		}
	}
	return retX, retY
}

// Calculates the curve normal at the specified t value. Note that this yields a normalised
// vector {x: nx, y: ny}.
// the normal is simply the normalised tangent vector, rotated by a quarter turn.
// https://pomax.github.io/bezierjs/#normal
func normal(curve CubicCurve, t float64) Point {
	d := Derivative(curve, t)
	var q = math.Sqrt(d.X*d.X + d.Y*d.Y)
	return Point{X: -d.Y / q, Y: d.X / q}
}

func lerp(r float64, v1 Point, v2 Point) Point {
	return Point{
		X: v1.X + r*(v2.X-v1.X),
		Y: v1.Y + r*(v2.Y-v1.Y),
	}
}

// gets all the points along the curve at the given step
// includes the start and end point
func getLUT(curve curveSegment, steps int) []Point {
	if steps == len(curve.lut) {
		return curve.lut
	}
	curve.lut = []Point{}
	// We want a range from 0 to 1 inclusive, so
	// we decrement and then use <= rather than <:
	steps--
	for t := 0; t <= steps; t++ {
		curve.lut = append(curve.lut, FindPoint(curve.curve, (float64(t)/float64(steps))))
	}
	return curve.lut
}

// finds the closest point in the given lut to the point
func closest(lut []Point, p Point) (distance float64, point Point, lutIndex int) {
	distance = math.Pow(2, 63)

	for idx, p := range lut {
		d := Distance(point, p)
		if d < distance {
			distance = d
			lutIndex = idx
		}
	}
	return distance, lut[lutIndex], lutIndex
}

type ValuePair struct {
	Left  float64
	Right float64
}

func (c ValuePair) String() string {
	return fmt.Sprintf("%.5f/%.5f", c.Left, c.Right)
}

// dedups the ValuePair array, based on the String() method (currently truncated to 5 digits)
func ValuePairDedup(a []ValuePair) []ValuePair {
	keys := make(map[string]bool)
	list := []ValuePair{}
	for _, entry := range a {
		key := entry.String()
		if _, value := keys[key]; !value {
			keys[key] = true
			list = append(list, entry)
		}
	}
	return list
}

// returns a list of strings in the form
// numbers are truncated to 5 decimal places because thats what the javascript does.
// "0.25684/0.22961"
func pairIteration(c1 curveSegment, c2 curveSegment, curveIntersectionThreshold float64) []ValuePair {
	c1bbTopLeft, c1bbBottomRight := BoundingBox(c1.curve)
	c2bbTopLeft, c2bbBottomRight := BoundingBox(c2.curve)
	r := float64(100000)

	// check the bounding boxes
	if (c1bbBottomRight.X-c1bbTopLeft.X)+(c1bbBottomRight.Y-c1bbTopLeft.Y) < curveIntersectionThreshold &&
		(c2bbBottomRight.X-c2bbTopLeft.X)+(c2bbBottomRight.Y-c2bbTopLeft.Y) < curveIntersectionThreshold {
		return []ValuePair{
			ValuePair{
				Left:  math.Round(r*(c1.t1+c1.t2)/2) / r,
				Right: math.Round(r*(c2.t1+c2.t2)/2) / r,
			},
		}
	}

	cc1Left, cc1Right := _splitCurveSection(c1, 0.5, -1)
	cc2Left, cc2Right := _splitCurveSection(c2, 0.5, -1)

	pairs := [][]curveSegment{}
	pTemp := [][]curveSegment{
		[]curveSegment{cc1Left, cc2Left},
		[]curveSegment{cc1Left, cc2Right},
		[]curveSegment{cc1Right, cc2Right},
		[]curveSegment{cc1Right, cc2Left},
	}
	for _, p := range pTemp {
		leftTL, leftBR := BoundingBox(p[0].curve)
		rightTL, rightBR := BoundingBox(p[1].curve)
		if BoundingBoxOverlaps(leftTL, leftBR, rightTL, rightBR) {
			pairs = append(pairs, p)
		}
	}
	if len(pairs) == 0 {
		return []ValuePair{}
	}
	results := []ValuePair{}
	for _, pair := range pairs {
		results = append(results, pairIteration(pair[0], pair[1], curveIntersectionThreshold)...)
	}
	results = ValuePairDedup(results)
	return results
}

func Align(points []Point, lineP1 Point, lineP2 Point) []Point {
	tx := lineP1.X
	ty := lineP1.Y
	a := -math.Atan2(lineP2.Y-ty, lineP2.X-tx)
	results := []Point{}
	for _, v := range points {
		results = append(results, Point{
			X: (v.X-tx)*math.Cos(a) - (v.Y-ty)*math.Sin(a),
			Y: (v.X-tx)*math.Sin(a) + (v.Y-ty)*math.Cos(a),
		})
	}
	return results
}

func Approximately(a, b, precision float64) bool {
	return math.Abs(a-b) <= precision

}

// approximately using precision = 0.000001
func Approx(a, b float64) bool {
	return Approximately(a, b, 0.000001)
}

// cube root function yielding real roots
func Crt(v float64) float64 {
	if v < 0 {
		return -math.Pow(-v, float64(1)/float64(3))
	} else {
		return math.Pow(v, float64(1)/float64(3))
	}
}

func Roots(points []Point, lineP1 Point, lineP2 Point) []float64 {

	// line = line || { p1: { x: 0, y: 0 }, p2: { x: 1, y: 0 } };
	alignedPoints := Align(points, lineP1, lineP2)

	// see http://www.trans4mind.com/personal_development/mathematics/polynomials/cubicAlgebra.htm
	pa := alignedPoints[0].Y
	pb := alignedPoints[1].Y
	pc := alignedPoints[2].Y
	pd := alignedPoints[3].Y
	d := -pa + float64(3)*pb - float64(3)*pc + pd

	a := float64(3)*pa - float64(6)*pb + float64(3)*pc
	b := -float64(3)*pa + float64(3)*pb
	c := pa

	reduce := func(nums []float64) []float64 {
		r := []float64{}
		for _, t := range nums {
			if 0 <= t && t <= 1 {
				r = append(r, t)
			}
		}
		return r
	}
	if Approx(d, 0) {
		// this is not a cubic curve.
		if Approx(a, 0) {
			// in fact, this is not a quadratic curve either.
			if Approx(b, 0) {
				// in fact in fact, there are no solutions.
				return []float64{}
			}
			// linear solution:
			return reduce([]float64{-c / b})
		}
		// quadratic solution:
		q := math.Sqrt(b*b - float64(4)*a*c)
		a2 := float64(2) * a
		return reduce([]float64{
			(q - b) / a2,
			(-b - q) / a2,
		})
	}

	// at this point, we know we need a cubic solution:

	a /= d
	b /= d
	c /= d

	p := (float64(3)*b - a*a) / float64(3)
	p3 := p / float64(3)
	q := (float64(2)*a*a*a - float64(9)*a*b + float64(27)*c) / float64(27)
	q2 := q / float64(2)
	discriminant := q2*q2 + p3*p3*p3

	if discriminant < 0 {
		mp3 := -p / float64(3)
		mp33 := mp3 * mp3 * mp3
		r := math.Sqrt(mp33)
		t := -q / (float64(2) * r)
		cosphi := t
		if t < -1 {
			cosphi = -1
		} else if t > 1 {
			cosphi = 1
		}
		phi := math.Acos(cosphi)
		tau := float64(2) * math.Pi
		crtr := Crt(r)
		t1 := float64(2) * crtr
		x1 := t1*math.Cos(phi/float64(3)) - a/float64(3)
		x2 := t1*math.Cos((phi+tau)/float64(3)) - a/float64(3)
		x3 := t1*math.Cos((phi+float64(2)*tau)/float64(3)) - a/float64(3)
		return reduce([]float64{x1, x2, x3})
	} else if discriminant == 0 {
		u1 := -Crt(q2)
		if q2 < 0 {
			Crt(-q2)
		}
		x1 := float64(2)*u1 - a/float64(3)
		x2 := -u1 - a/float64(3)
		return reduce([]float64{x1, x2})
	} else {
		sd := math.Sqrt(discriminant)
		u1 := Crt(-q2 + sd)
		v1 := Crt(q2 + sd)
		return reduce([]float64{u1 - v1 - a/float64(3)})
	}
}

// this function checks for self-intersection.
// Intersections are yielded as an array of float/float strings,
// where the two floats are separated by the character / and both floats corresponding
// to t values on the curve at which the intersection is found.
//
// Note this will only return a single intersection.  Afaik a cubic curve could not intersect itself
// more than once, so this is likely a non-issue.  (Anyone know different? Math is hard.)
func IntersectsSelf(curve CubicCurve, curveIntersectionThreshold float64) (intersect bool, t1 float64, t2 float64) {
	reduced := reduce(curve)

	// "simple" curves cannot intersect with their direct
	// neighbour, so for each segment X we check whether
	// it intersects [0:x-2][x+2:last].

	l := len(reduced) - 2
	results := []ValuePair{}
	for i := 0; i < l; i++ {
		left := reduced[i : i+1]
		right := reduced[i+2:]
		result := _curveIntersects(left, right, curveIntersectionThreshold)
		results = append(results, result...)
	}

	intersect = len(results) > 0
	if intersect {
		t1 = results[0].Left
		t2 = results[0].Right
	}
	return intersect, t1, t2
}

// Finds the intersections between this curve an some line {p1: {x:... ,y:...}, p2: ... }.
// The intersections are an array of t values on this curve.
// Curves are first aligned (translation/rotation) such that the curve's first coordinate is (0,0), and the curve is rotated so that the intersecting
// line coincides with the x-axis. Doing so turns "intersection finding" into plain "root finding".
// As a root finding solution, the roots are computed symbolically for both quadratic and cubic curves,
// using the standard square root function which you might remember from high school, and the absolutely
// not standard Cardano's algorithm for solving the cubic root function.
func IntersectsLine(curve CubicCurve, p1 Point, p2 Point) []float64 {
	minx := math.Min(p1.X, p2.X)
	miny := math.Min(p1.Y, p2.Y)
	maxx := math.Max(p1.X, p2.X)
	maxy := math.Max(p2.Y, p2.Y)
	points := cubicCurveToArray(curve)

	roots := Roots(points, p1, p2)

	results := []float64{}
	for _, t := range roots {
		p := FindPoint(curve, t)
		if between(p.X, minx, maxx) && between(p.Y, miny, maxy) {
			results = append(results, t)
		}
	}
	return results
}

func IntersectsProjectedLine(curve CubicCurve, p1 Point, p2 Point) []float64 {
	points := cubicCurveToArray(curve)
	return Roots(points, p1, p2)
}

// Finds the intersection points between the two curves.
// The resulting pairs are the T values on left (i.e c1) and right (c2)
//
// Note curveIntersectionThreshold defaults to 0.5 in the javascript impl
func IntersectsCurve(c1, c2 CubicCurve, curveIntersectionThreshold float64) []ValuePair {
	cr1 := reduce(c1)
	cr2 := reduce(c2)
	return _curveIntersects(cr1, cr2, curveIntersectionThreshold)
}

func _curveIntersects(c1 []curveSegment, c2 []curveSegment, curveIntersectionThreshold float64) []ValuePair {
	pairs := [][]curveSegment{}
	// step 1: pair off any overlapping segments
	for _, l := range c1 {
		for _, r := range c2 {
			if CurveOverlaps(l.curve, r.curve) {
				pairs = append(pairs, []curveSegment{l, r})
			}
		}
	}
	// step 2: for each pairing, run through the convergence algorithm.
	intersections := []ValuePair{}

	for _, pair := range pairs {
		result := pairIteration(pair[0], pair[1], curveIntersectionThreshold)
		if len(result) > 0 {
			intersections = append(intersections, result...)
		}
	}
	return intersections
}

func Between(v, m1, m2 float64) bool {
	return (m1 <= v && v <= m2) || Approx(v, m1) || Approx(v, m2)
}

// checks the Bounding Boxes of the two curves and tests whether they overlap
func CurveOverlaps(c1, c2 CubicCurve) bool {
	lboxTL, lboxBR := BoundingBox(c1)
	rboxTL, rboxBR := BoundingBox(c2)
	return BoundingBoxOverlaps(lboxTL, lboxBR, rboxTL, rboxBR)
}

// Finds the on-curve point closest to the specific off-curve point,
// using a two-pass projection test based on the curve's LUT.
// A distance comparison finds the closest match, after which a fine interval around
// that match is checked to see if a better projection can be found.
func Project(curve CubicCurve, point Point) (pt Point, distance float64, t float64) {
	curveSegment := curveSegment{
		curve: curve,
	}

	// step 1: coarse check
	lut := getLUT(curveSegment, 100)
	l := len(lut) - 1
	mdist, pt, mpos := closest(lut, point)
	if mpos == 0 || mpos == l {
		t = float64(mpos) / float64(l)
		return pt, t, mdist
	}

	// step 2: fine check
	t1 := (float64(mpos) - 1.0) / float64(l)
	t2 := (float64(mpos) + 1.0) / float64(l)
	step := .1 / float64(l)

	mdist++
	t = t1
	ft := t
	for ft = t; t < t2+step; t += step {
		p := FindPoint(curve, t)
		d := Distance(point, p)
		if d < mdist {
			mdist = d
			ft = t
		}
	}
	p := FindPoint(curve, ft)
	return p, mdist, ft
}

func BoundingBoxOverlaps(box1TopLeft, box1BottomRight, box2TopLeft, box2BottomRight Point) bool {
	// process X
	b1MidX := (box1TopLeft.X + box1BottomRight.X) / 2
	b2MidX := (box2TopLeft.X + box2BottomRight.X) / 2
	b1SizeX := box1BottomRight.X - box1TopLeft.X
	b2SizeX := box2BottomRight.X - box2TopLeft.X

	dX := (b1SizeX + b2SizeX) / 2
	if math.Abs(b1MidX-b2MidX) >= dX {
		return false
	}

	// process Y
	b1MidY := (box1TopLeft.Y + box1BottomRight.Y) / 2
	b2MidY := (box2TopLeft.Y + box2BottomRight.Y) / 2
	b1SizeY := box1BottomRight.Y - box1TopLeft.Y
	b2SizeY := box2BottomRight.Y - box2TopLeft.Y

	dY := (b1SizeY + b2SizeY) / 2
	if math.Abs(b1MidY-b2MidY) >= dY {
		return false
	}
	return true
}

// the smallest box that the curve fits into
func BoundingBox(curve CubicCurve) (topLeft Point, bottomRight Point) {
	// the result format is different than what is provided in bezier.js
	//
	// js
	// x: {min: 77.05277837041058, mid: 122.93644996347699, max: 168.8201215565434, size: 91.76734318613282}
	// y: {min: 67.63285925823757, mid: 153.81642962911877, max: 240, size: 172.36714074176243}
	// __proto__: Object
	// go
	// topLeft (77.053,67.633)
	// bottemRight (168.820,240.000)

	// from js to golang:
	// js.x.min = topLeft.X
	// js.X.mid = topLeft.X + bottomRight.X / 2
	// js.x.max = bottomRight.X
	// js.x.size = bottomRight.X - topLeft.X

	// js.y.min = topLeft.Y
	// js.y.max = bottomRight.Y
	// js.y.mid = topLeft.Y + bottomRight.Y / 2
	// js.y.size = bottomRight.Y - topLeft.Y

	xMin, xMax, yMin, yMax := getMinMax(curve)
	return NewPoint(xMin.X, yMin.Y), NewPoint(xMax.X, yMax.Y)
}

func hull(curve CubicCurve, t float64) []Point {
	p := cubicCurveToArray(curve)
	q := []Point{
		p[0],
		p[1],
		p[2],
		p[3],
	}
	// we lerp between all points at each iteration, until we have 1 point left.
	for len(p) > 1 {
		_p := []Point{}
		i := 0
		for l := len(p) - 1; i < l; i++ {
			pt := lerp(t, p[i], p[i+1])
			q = append(q, pt)
			_p = append(_p, pt)
		}
		p = _p
	}
	return q
}

func angle(o, v1, v2 Point) float64 {
	dx1 := v1.X - o.X
	dy1 := v1.Y - o.Y
	dx2 := v2.X - o.X
	dy2 := v2.Y - o.Y
	cross := (dx1 * dy2) - (dy1 * dx2)
	dot := (dx1 * dx2) + (dy1 * dy2)
	return math.Atan2(cross, dot)
}

func simple(curve CubicCurve) bool {
	points := cubicCurveToArray(curve)
	a1 := angle(points[0], points[3], points[1])
	a2 := angle(points[0], points[3], points[2])
	if (a1 > 0 && a2 < 0) || (a1 < 0 && a2 > 0) {
		return false
	}

	n1 := normal(curve, 0)
	n2 := normal(curve, 1)
	s := n1.X*n2.X + n1.Y*n2.Y

	angle := math.Abs(math.Acos(s))
	return angle < math.Pi/3
}

func reduce(curve CubicCurve) []curveSegment {

	extremaX, extremaY := extrema(curve)

	// sort and dedup
	extrema := append(extremaX, extremaY...)
	extrema = append(extrema, 0, 1) // needs to contain 0 and 1
	extrema = Float64ArrayDeDup(extrema)
	sort.Float64s(extrema)

	pass1 := []curveSegment{}
	// first pass: split on extrema
	i := 1
	for t1 := extrema[0]; i < len(extrema); i++ {
		t2 := extrema[i]
		segment, _ := _splitCurveSection(curveSegment{curve: curve}, t1, t2)
		segment.t1 = t1
		segment.t2 = t2
		pass1 = append(pass1, segment)
		t1 = t2
	}
	// second pass: further reduce these segments to simple segments
	step := 0.01
	pass2 := []curveSegment{}
	for _, p1 := range pass1 {
		t1 := 0.0
		t2 := 0.0
		for t2 <= 1 {
			for t2 = t1 + step; t2 <= 1+step; t2 += step {
				segment, _ := _splitCurveSection(p1, t1, t2)
				if !simple(segment.curve) {
					t2 -= step
					if math.Abs(t1-t2) < step {
						// we can never form a reduction
						return []curveSegment{}
					}

					segment, _ = _splitCurveSection(p1, t1, t2)

					segment.t1 = _map(t1, 0, 1, p1.t1, p1.t2)
					segment.t2 = _map(t2, 0, 1, p1.t1, p1.t2)
					pass2 = append(pass2, segment)
					t1 = t2
					break
				}
			}
		}

		if t1 < 1 {
			segment, _ := _splitCurveSection(p1, t1, 1)
			segment.t1 = _map(t1, 0, 1, p1.t1, p1.t2)
			segment.t2 = p1.t2
			pass2 = append(pass2, segment)
		}

	} //end pass 2
	return pass2
}

func curveToString(curve CubicCurve) string {
	ar := cubicCurveToArray(curve)
	return fmt.Sprintf("Array: [\n\t%s, \n\t%s, \n\t%s, \n\t%s]\n",
		ar[0].String(),
		ar[1].String(),
		ar[2].String(),
		ar[3].String())
}

func getMinMax(curve CubicCurve) (xMin, xMax, yMin, yMax Point) {
	listX, listY := extrema(curve)

	// each list must contain 0 and 1 t values (the start and endpoints)
	listX = Float64ArrayInsertIfAbsent(listX, 0)
	listX = Float64ArrayInsertIfAbsent(listX, 1)
	listY = Float64ArrayInsertIfAbsent(listY, 0)
	listY = Float64ArrayInsertIfAbsent(listY, 1)

	xMin.X = float64(MaxInt)
	xMax.X = float64(MinInt)
	yMin.Y = float64(MaxInt)
	yMax.Y = float64(MinInt)

	for _, t := range listX {
		p := FindPoint(curve, t)
		if p.X < xMin.X {
			xMin = p
		}
		if p.X > xMax.X {
			xMax = p
		}
	}

	for _, t := range listY {
		p := FindPoint(curve, t)
		if p.Y < yMin.Y {
			yMin = p
		}
		if p.Y > yMax.Y {
			yMax = p
		}
	}
	return xMin, xMax, yMin, yMax
}

//the point on the curve at t=..., offset along its normal by a distance d.
func OffsetPoint(curve CubicCurve, t, d float64) Point {
	curve = normalize(curve)
	c := FindPoint(curve, t)
	n := normal(curve, t)

	return Point{
		X: c.X + n.X*d,
		Y: c.Y + n.Y*d,
	}
}

// This function creates a new curve,
// offset along the curve normals, at distance d. Note that deep magic lies here
// and the offset curve of a Bezier curve cannot ever be another Bezier curve.
// As such, this function "cheats" and yields an array of curves which, taken
// together, form a single continuous curve equivalent to what a theoretical offset
// curve would be.
func Offset(curve CubicCurve, d float64) []CubicCurve {
	curve = normalize(curve)
	isLinear := false // TODO
	if isLinear {
		nv := normal(curve, 0)
		coords := []Point{}
		for _, p := range cubicCurveToArray(curve) {
			ret := Point{
				X: p.X + d*nv.X,
				Y: p.Y + d*nv.Y,
			}
			coords = append(coords, ret)
		}
		return []CubicCurve{arrayToCubicCurve(coords)}
	}
	reduced := reduce(curve)
	curves := []CubicCurve{}
	for _, s := range reduced {
		scaled, err := Scale(s.curve, d)
		if err != nil {
			fmt.Printf("ERROR! %s", err.Error())
			return curves
		}
		curves = append(curves, scaled)
	}
	return curves
}

// returns true if this curve is clockwise
func IsClockwise(curve CubicCurve) bool {
	points := cubicCurveToArray(curve)
	angle := angle(points[0], points[3], points[1])
	return angle > 0
}

type distanceFunc func(float64) float64

var nullDistanceFunc distanceFunc = func(t float64) float64 { return -1 }

func Scale(curve CubicCurve, d float64) (CubicCurve, error) {
	return _scale(curve, nullDistanceFunc, d, false)
}

// scales by the passed in distance function
func ScaleByFunc(curve CubicCurve, distanceFn distanceFunc) (CubicCurve, error) {
	return _scale(curve, distanceFn, -1, true)
}

func _scale(curve CubicCurve, distanceFn distanceFunc, d float64, isDistanceFn bool) (CubicCurve, error) {
	clockwise := IsClockwise(curve)

	//   // TODO: add special handling for degenerate (=linear) curves.
	//   var clockwise = this.clockwise;
	r1 := d
	r2 := d
	if isDistanceFn {
		r1 = distanceFn(0)
		r2 = distanceFn(1)
	}
	v := []Point{OffsetPoint(curve, 0, 10), OffsetPoint(curve, 1, 10)}
	curveStart := FindPoint(curve, 0) // in js this is v[0].c
	curveEnd := FindPoint(curve, 1)   // in js this is v[1].c
	normalStart := normal(curve, 0)   // in js this is v[0].n
	normalEnd := normal(curve, 1)     // in js this is v[1].n

	success, o := lli4(v[0], curveStart, v[1], curveEnd)
	if !success {
		return CubicCurve{}, fmt.Errorf("cannot scale this curve. Try reducing it first")
	}
	//   // move all points by distance 'd' wrt the origin 'o'
	points := cubicCurveToArray(curve)

	np := make([]Point, 4)

	for _, t := range []float64{0, 1} {

		index := int(t * 3) // order = 3
		cp := points[index]
		p := NewPoint(cp.X, cp.Y)

		rTmp := r1
		if t == 1 {
			rTmp = r2
		}
		norm := normalStart
		if t == 1 {
			norm = normalEnd
		}
		p.X += rTmp * norm.X
		p.Y += rTmp * norm.Y
		np[index] = p
	}

	if !isDistanceFn {
		// move control points to lie on the intersection of the offset
		// derivative vector, and the origin-through-control vector
		for _, t := range []float64{0, 1} {
			index := int(t * 3) // order = 3

			p := np[index]
			derivative := Derivative(curve, t)
			p2 := Point{
				X: p.X + derivative.X,
				Y: p.Y + derivative.Y,
			}
			_, np[int(t+1)] = lli4(p, p2, o, points[int(t+1)])
		}
		return arrayToCubicCurve(np), nil
	}

	// move control points by "however much necessary to
	// ensure the correct tangent to endpoint".
	for _, t := range []float64{0, 1} {
		p := points[int(t+1)]
		ov := Point{
			X: p.X - o.X,
			Y: p.Y - o.Y,
		}
		rc := distanceFn((t + 1) / float64(3))
		if !clockwise {
			rc = -rc
		}
		m := math.Sqrt(ov.X*ov.X + ov.Y*ov.Y)
		ov.X /= m
		ov.Y /= m
		np[int(t+1)] = Point{
			X: p.X + rc*ov.X,
			Y: p.Y + rc*ov.Y,
		}
	}
	return arrayToCubicCurve(np), nil
}

func lli8(x1, y1, x2, y2, x3, y3, x4, y4 float64) (success bool, point Point) {
	nx :=
		(x1*y2-y1*x2)*(x3-x4) - (x1-x2)*(x3*y4-y3*x4)
	ny := (x1*y2-y1*x2)*(y3-y4) - (y1-y2)*(x3*y4-y3*x4)
	d := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
	if d == 0 {
		return false, point
	}
	return true, Point{X: nx / d, Y: ny / d}
}

func lli4(p1, p2, p3, p4 Point) (success bool, point Point) {
	x1 := p1.X
	y1 := p1.Y
	x2 := p2.X
	y2 := p2.Y
	x3 := p3.X
	y3 := p3.Y
	x4 := p4.X
	y4 := p4.Y
	return lli8(x1, y1, x2, y2, x3, y3, x4, y4)
}
