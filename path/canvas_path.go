package path

import "github.com/tdewolff/canvas"

// ToCanvasPath converts an HFD path to a canvas path. Move, line, and cubic
// segments are preserved. Canvas may drop a zero-length line or merge collinear
// lines, so this is not a byte-identical round trip.
func ToCanvasPath(p Path) *canvas.Path {
	cp := &canvas.Path{}
	if p == nil {
		return cp
	}
	for _, seg := range p.Segments() {
		switch s := seg.(type) {
		case MoveSegment:
			cp.MoveTo(s.End().X, s.End().Y)
		case LineSegment:
			cp.LineTo(s.End().X, s.End().Y)
		case CurveSegment:
			cp.CubeTo(
				s.ControlPointStart.X, s.ControlPointStart.Y,
				s.ControlPointEnd.X, s.ControlPointEnd.Y,
				s.End().X, s.End().Y,
			)
		}
	}
	return cp
}

// PathFromCanvas converts a canvas path into HFD segments. Arcs become cubics.
// Close commands are dropped so a trailing Z does not add a segment, matching
// parsePathFromSvgLegacy.
func PathFromCanvas(cp *canvas.Path) Path {
	d := NewDraw()
	if cp == nil || cp.Empty() {
		return d.Path()
	}
	cp = cp.ReplaceArcs()
	s := cp.Scanner()
	for s.Scan() {
		end := NewPoint(s.End().X, s.End().Y)
		switch s.Cmd() {
		case canvas.MoveToCmd:
			d.MoveTo(end)
		case canvas.LineToCmd:
			d.LineTo(end)
		case canvas.QuadToCmd:
			d.QCurveTo(NewPoint(s.CP1().X, s.CP1().Y), end)
		case canvas.CubeToCmd:
			d.CurveTo(
				NewPoint(s.CP1().X, s.CP1().Y),
				NewPoint(s.CP2().X, s.CP2().Y),
				end,
			)
		case canvas.CloseCmd:
		}
	}
	return d.Path()
}

// OffsetPath returns a contour offset by distance. Positive distance offsets
// to the right of the path direction. Not used by the kerf offset transform.
func OffsetPath(p Path, distance float64) Path {
	return PathFromCanvas(ToCanvasPath(p).Offset(distance, 0.001))
}

// AndPaths is the intersection of a and b.
func AndPaths(a, b Path) Path {
	return PathFromCanvas(ToCanvasPath(a).And(ToCanvasPath(b)))
}

// OrPaths is the union of a and b.
func OrPaths(a, b Path) Path {
	return PathFromCanvas(ToCanvasPath(a).Or(ToCanvasPath(b)))
}

// NotPaths is a with b removed.
func NotPaths(a, b Path) Path {
	return PathFromCanvas(ToCanvasPath(a).Not(ToCanvasPath(b)))
}
