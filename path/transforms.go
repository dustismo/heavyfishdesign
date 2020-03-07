package path

// transform a Path into a new Path
type PathTransform interface {
	PathTransform(path Path) (Path, error)
}

// transforma a segment
type SegmentTransform interface {
	SegmentTransform(segment Segment) Segment
}

type Params interface {
	GetString(key string) string
}

// executes multiple transforms
func MultiTransform(p Path, transforms ...PathTransform) (Path, error) {
	pth := p
	var err error
	for _, t := range transforms {
		pth, err = t.PathTransform(pth)
		if err != nil {
			return pth, err
		}
	}
	return pth, err
}
