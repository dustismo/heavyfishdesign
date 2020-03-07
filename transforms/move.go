package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

// This moves the path origin to the requested point
type MoveTransform struct {
	Point path.Point
	// if a handle is specified, than this move operation will move the
	// handle to the requested point. by default the handle is the topleft
	Handle           path.PathAttr
	SegmentOperators path.SegmentOperators
}

func (mt MoveTransform) PathTransform(p path.Path) (path.Path, error) {
	if len(mt.Handle) == 0 {
		// handle should be TOP_LEFT by default..
		mt.Handle = path.TopLeft
	}
	handle, err := path.PointPathAttribute(mt.Handle, p, mt.SegmentOperators)
	if err != nil {
		return p, err
	}
	mvX := mt.Point.X - handle.X
	mvY := mt.Point.Y - handle.Y

	return ShiftTransform{
		DeltaX:           mvX,
		DeltaY:           mvY,
		SegmentOperators: mt.SegmentOperators,
	}.PathTransform(p)
}
