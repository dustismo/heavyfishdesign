package bezier

import "math"

const MaxInt = int(^uint(0) >> 1)
const MinInt = -MaxInt - 1

// Distance finds the straightline distance between the two points
func Distance(p1 Point, p2 Point) float64 {
	x := p1.X - p2.X
	y := p1.Y - p2.Y
	return math.Sqrt((x * x) + (y * y))
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

// returns true if requested is between v1 and v2
// order of v1 and v2 does not matter
func between(requested, v1, v2 float64) bool {
	a := v1
	b := v2
	r := requested
	if Approx(r, a) || Approx(r, b) {
		return true
	}

	return (r >= a && r <= b) || (r <= a && r >= b)
}
