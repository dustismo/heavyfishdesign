package transforms

import "github.com/dustismo/heavyfishdesign/path"

// A basic affine (matrix) transform
// see: https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/transform#General_Transformation
type MatrixTransform struct {
	A, B, C, D, E, F float64
	SegmentOperators path.SegmentOperators
}

func (mt MatrixTransform) TransformPoint(p path.Point) path.Point {
	xTransformed := p.X*mt.A + p.Y*mt.C + mt.E
	yTransformed := p.X*mt.B + p.Y*mt.D + mt.F
	return path.NewPoint(xTransformed, yTransformed)
}

func (mt MatrixTransform) PathTransform(p path.Path) (path.Path, error) {
	segments := []path.Segment{}

	for _, seg := range p.Segments() {
		s, err := mt.SegmentOperators.TransformPoints(seg, mt.TransformPoint)
		if err != nil {
			return nil, err
		}
		segments = append(segments, s)
	}

	return path.NewPathFromSegments(segments), nil
}
