package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

type RebuildTransform struct{}

// Simple transform to rebuild the path.
// this dumps the path to a string and reparses
// mostly this is for cleaning datastructure problems from other transforms
// (I.E internal changes mean start and end points don't align)
func (st RebuildTransform) PathTransform(p path.Path) (path.Path, error) {
	svgStr := path.SvgString(p, 7)
	return path.ParsePathFromSvg(svgStr)
}
