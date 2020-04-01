package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

// connects any disjoint segments as smoothly as possible
type JoinTransform struct {
	Precision        int
	SegmentOperators path.SegmentOperators
	// should we close the path?
	ClosePath bool
}

// attempts to join two paths
// this is a recursive function that
// 1. attempts to join the end seg of pth1 to the start seg of pth2.
// 2. this will reverse either segment to join the closest edge
// 3. if success trim start of pth2
// 4. repeat until pth2 is empty
func (jt JoinTransform) join(pth1, pth2 path.Path) (path.Path, error) {
	p1 := path.NewPathFromSegmentsWithoutMove(path.TrimMove(pth1.Segments()))
	p2 := path.NewPathFromSegmentsWithoutMove(path.TrimMove(pth2.Segments()))

	if len(p2.Segments()) == 0 {
		return p1, nil
	}

	// if p1 empty or pth2 does not start with a move
	// then we can just transfer the first segment of p2 to p1
	if len(p1.Segments()) == 0 {
		// move one segment from p2 to p1 and retry
		p1New := append(p1.Segments(), p2.Segments()[0])
		return jt.join(
			path.NewPathFromSegmentsWithoutMove(p1New),
			path.NewPathFromSegmentsWithoutMove(path.TrimFirst(p2.Segments())),
		)
	}
	p1s := p1.Segments()[0].Start()
	_, p1e := path.GetStartAndEnd(p1.Segments())
	p2s := p2.Segments()[0].Start()
	_, p2e := path.GetStartAndEnd(p2.Segments())

	// measure to see if we should reverse any of the segments
	ss := path.Distance(p1s, p2s)
	minDistance := ss
	se := path.Distance(p1s, p2e)
	if se < minDistance {
		minDistance = se
	}
	es := path.Distance(p1e, p2s)
	if es < minDistance {
		minDistance = es
	}
	ee := path.Distance(p1e, p2e)
	if ee < minDistance {
		minDistance = ee
	}

	// see if we should reverse any segments..
	// p1.end -> p2.start is the situation we want
	// order of the switch statement is significant as
	// distances can be equal, and we want to chose the best-most option
	switch minDistance {
	case es: // do nothing, this is good
	case ee:
		// reverse p2
		rev2, err := PathReverse{}.PathTransform(p2)
		if err != nil {
			return nil, err
		}
		return jt.join(p1, rev2)
	case ss:
		// reverse p1
		rev, err := PathReverse{}.PathTransform(p1)
		if err != nil {
			return nil, err
		}
		return jt.join(rev, p2)
	case se:
		// reverse p1 and p2
		rev1, err := PathReverse{}.PathTransform(p1)
		if err != nil {
			return nil, err
		}
		rev2, err := PathReverse{}.PathTransform(p2)
		if err != nil {
			return nil, err
		}
		return jt.join(
			rev1,
			rev2)
	}

	// now joint the end segment of p1 with the start segment of p2.
	// then return p2 without the first segment.
	s1 := p1.Segments()[len(p1.Segments())-1] // end of p1
	s2 := p2.Segments()[0]                    // start of p2

	joined := []path.Segment{}
	if p1e.EqualsPrecision(p2s, jt.Precision) {
		// the points are equal, just set the start of s2
		s2New, err := path.SetSegmentStart(s2, p1e)
		if err != nil {
			return p1, err
		}
		joined = []path.Segment{
			s1,
			s2New,
		}
	} else {
		// join the segments in the normal way
		j, err := jt.SegmentOperators.Join(s1, s2)
		if err != nil {
			return p1, err
		}
		joined = j
	}

	newP1 := append(path.TrimLast(p1.Segments()), joined...)
	newP2 := path.TrimFirst(p2.Segments())
	return jt.join(path.NewPathFromSegmentsWithoutMove(newP1), path.NewPathFromSegmentsWithoutMove(newP2))
}

func (jt JoinTransform) PathTransform(pth path.Path) (path.Path, error) {
	p, err := DedupSegmentsTransform{Precision: jt.Precision}.PathTransform(pth)

	if err != nil {
		return pth, err
	}
	p, err = ReorderTransform{Precision: jt.Precision}.PathTransform(p)
	if err != nil {
		return pth, err
	}

	p, err = jt.join(path.NewPath(), p)
	if err != nil {
		return pth, err
	}
	if jt.ClosePath && len(p.Segments()) > 0 {
		//attempt to close the path, using the normal intersection
		joined := p.Segments()
		end := path.Tail(joined)
		start := joined[0]

		closed, err := jt.SegmentOperators.Join(end, start)
		if err != nil {
			return p, err
		}
		joined[0] = path.Tail(closed)

		// now replace the last element with the new joined list
		joined = append(path.TrimLast(joined), path.TrimLast(closed)...)
		return path.NewPathFromSegments(p.Segments()), nil
	} else {
		return path.NewPathFromSegments(p.Segments()), nil
	}
}

type SimpleJoin struct {
}

func (sj SimpleJoin) JoinPaths(paths ...path.Path) path.Path {

	// simple join.
	newPath := paths[0].Clone()
	for i := 1; i < len(paths); i++ {
		for index, s := range paths[i].Segments() {
			if index == 0 {
				prev := path.Tail(newPath.Segments())
				if prev != nil && !prev.End().Equals(s.End()) {
					// add a move segment to the start point
					newPath.AddSegments(path.MoveSegment{
						StartPoint: prev.End(),
						EndPoint:   s.End().Clone(),
					})
				}
			}
			newPath.AddSegments(s.Clone())
		}
	}
	return newPath
}
