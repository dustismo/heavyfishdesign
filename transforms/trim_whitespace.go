package transforms

import "github.com/dustismo/heavyfishdesign/path"

type TrimWhitespaceTransform struct {
	SegmentOperators path.SegmentOperators
}

// PathTransform trims any whitespace by moving the path to as close to 0,0 as possible.
// Note that you should typically call simplify before triming whitespace to avoid
// things like M 0 0, M 10, 11
func (tw TrimWhitespaceTransform) PathTransform(p path.Path) (path.Path, error) {
	tl, _, err := path.BoundingBoxTrimWhitespace(p, tw.SegmentOperators)
	if err != nil {
		return p, err
	}

	return ShiftTransform{
		DeltaX:           -tl.X,
		DeltaY:           -tl.Y,
		SegmentOperators: tw.SegmentOperators,
	}.PathTransform(p)

}
