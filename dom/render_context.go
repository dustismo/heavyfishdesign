package dom

import (
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/util"
)

type RenderContext struct {
	Origin path.Point
	Cursor path.Point
	Log    *util.HfdLog
}

func (c RenderContext) Clone() RenderContext {
	return RenderContext{
		Origin: c.Origin,
		Cursor: c.Cursor,
		Log:    c.Log,
	}
}
