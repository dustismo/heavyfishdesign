package transforms

import (
	"math"

	"github.com/dustismo/heavyfishdesign/path"
)

// This changes the size and proportions of the given path
type ScaleTransform struct {
	ScaleX float64
	ScaleY float64
	// TODO: Scaling by start and end point should
	// use the rotate_scale transform.
	StartPoint       path.Point
	EndPoint         path.Point
	Width            float64
	Height           float64
	SegmentOperators path.SegmentOperators
}

func (st ScaleTransform) PathTransform(p path.Path) (path.Path, error) {

	var xScale = st.ScaleX
	var yScale = st.ScaleY
	if st.Width > 0 || st.Height > 0 {
		// measure.
		tl, br, err := path.BoundingBoxTrimWhitespace(p, st.SegmentOperators)
		if err != nil {
			return p, err
		}
		curWidth := math.Abs(br.X - tl.X)
		curHeight := math.Abs(br.Y - tl.Y)
		if st.Width > 0 {
			xScale = st.Width / curWidth
		}
		if st.Height > 0 {
			yScale = st.Height / curHeight
		} else {
			yScale = xScale
		}
		if st.Width <= 0 {
			xScale = yScale
		}
	} else if !st.StartPoint.Equals(st.EndPoint) {
		// the start and end point are set, so we set the scale factors
		newX := math.Abs(st.EndPoint.X - st.StartPoint.X)
		newY := math.Abs(st.EndPoint.Y - st.StartPoint.Y)

		s, e := path.GetStartAndEnd(p.Segments())
		oldX := math.Abs(e.X - s.X)
		oldY := math.Abs(e.Y - s.Y)

		xScale = newX / oldX
		yScale = newY / oldY

		// if we arent moving in one direction then set
		// equal scaling factors
		if newX == 0 || oldX == 0 {
			xScale = yScale
		}
		if newY == 0 || oldY == 0 {
			yScale = xScale
		}
	}

	segs := []path.Segment{}
	// function to do the scaling
	pt := func(p path.Point) path.Point {
		x := p.X * xScale
		y := p.Y * yScale
		return path.NewPoint(x, y)
	}

	for _, s := range p.Segments() {
		seg, err := st.SegmentOperators.TransformPoints(s, pt)
		if err != nil {
			return p, err
		}
		segs = append(segs, seg)
	}
	return path.NewPathFromSegments(segs), nil
}
