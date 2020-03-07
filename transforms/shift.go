package transforms

import "github.com/dustismo/heavyfishdesign/path"

// This shifts the path by the requested amount in X and/or Y
type ShiftTransform struct {
	DeltaX           float64
	DeltaY           float64
	SegmentOperators path.SegmentOperators
}

func (st ShiftTransform) PathTransform(p path.Path) (path.Path, error) {
	pt := func(p path.Point) path.Point {
		return path.NewPoint(p.X+st.DeltaX, p.Y+st.DeltaY)
	}
	newPath := []path.Segment{}
	for _, seg := range p.Segments() {
		s, err := st.SegmentOperators.TransformPoints(seg, pt)
		if err != nil {
			return nil, err
		}
		newPath = append(newPath, s)
	}
	return path.NewPathFromSegments(newPath), nil
}
