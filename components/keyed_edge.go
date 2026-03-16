package components

// KeyedEdgeComponent draws an edge with a key-shaped pocket at the centre of the seam.
//
// The key shape is provided as an SVG path string (key_svg).  The full SVG is split
// horizontally at its midpoint; the "plug" side uses the top half and the "socket"
// side uses the bottom half.  This means a non-symmetric key (e.g. a pawn silhouette)
// produces complementary but mirror-image pockets that the physical key fits into
// exactly.
//
// key_height is NOT a parameter — it is derived automatically from the SVG's natural
// aspect ratio, preserving the key shape exactly.
//
// Parameters:
//   from      – start point of the seam edge (provided by the splitter)
//   to        – end  point of the seam edge  (provided by the splitter)
//   key_svg   – SVG path string (or param alias) of the full key silhouette
//   key_width – width of the key / pocket along the edge
//               (default: 50% of the total edge length)
//   key_side  – "plug" (top half of key, default) or "socket" (bottom half)
//
// Usage in a splitter:
//   "plug_edge":   { "type": "keyed_edge", "key_side": "plug",   "key_svg": "...", "key_width": "spline_width" }
//   "socket_edge": { "type": "keyed_edge", "key_side": "socket", "key_svg": "...", "key_width": "spline_width" }

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

type KeyedEdgeComponentFactory struct{}

type KeyedEdgeComponent struct {
	*dom.BasicComponent
	segmentOperators path.SegmentOperators
}

func (kef KeyedEdgeComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *dom.DocumentContext) (dom.Component, error) {
	factory := dom.AppContext()
	bc := factory.MakeBasicComponent(mp)
	return &KeyedEdgeComponent{
		BasicComponent:   bc,
		segmentOperators: factory.SegmentOperators(),
	}, nil
}

func (kef KeyedEdgeComponentFactory) ComponentTypes() []string {
	return []string{"keyed_edge"}
}

func (ke *KeyedEdgeComponent) Render(ctx dom.RenderContext) (path.Path, dom.RenderContext, error) {
	ke.RenderStart(ctx)

	so := ke.segmentOperators
	precision := dom.AppContext().Precision()
	attr := ke.Attr()

	startPoint := attr.MustPoint("from", ctx.Cursor)
	endPoint, found := attr.Point2("to", startPoint)
	if !found {
		return nil, ctx, fmt.Errorf("keyed_edge: 'to' attribute is required")
	}

	keySVGStr, ok := attr.SvgString("key_svg")
	if !ok || keySVGStr == "" {
		return nil, ctx, fmt.Errorf("keyed_edge: 'key_svg' attribute is required")
	}

	keySide := attr.MustString("key_side", "plug")

	edgeLine := path.LineSegment{StartPoint: startPoint, EndPoint: endPoint}
	edgeLen := edgeLine.Length()

	// ── 1. Parse the key SVG ──────────────────────────────────────────────────
	keyPath, err := path.ParsePathFromSvg(keySVGStr)
	if err != nil {
		return nil, ctx, fmt.Errorf("keyed_edge: invalid key_svg: %s", err.Error())
	}

	// ── 2. Normalise: move the bounding box to (0,0) ──────────────────────────
	keyPath, err = transforms.TrimWhitespaceTransform{SegmentOperators: so}.PathTransform(keyPath)
	if err != nil {
		return nil, ctx, err
	}

	// ── 3. Derive key_height from the SVG's natural aspect ratio ──────────────
	//      key_width defaults to 50% of the total edge length.
	//      key_height is always computed proportionally so the key shape is
	//      preserved exactly — it is not an exposed parameter.
	tl, br, err := path.BoundingBoxTrimWhitespace(keyPath, so)
	if err != nil {
		return nil, ctx, fmt.Errorf("keyed_edge: could not compute key_svg bounding box: %s", err.Error())
	}
	naturalWidth := br.X - tl.X
	naturalHeight := br.Y - tl.Y
	if naturalWidth <= 0 || naturalHeight <= 0 {
		return nil, ctx, fmt.Errorf("keyed_edge: key_svg has zero or negative bounding box (w=%.4f h=%.4f)", naturalWidth, naturalHeight)
	}

	keyWidth := attr.MustFloat64("key_width", edgeLen*0.5)
	keyHeight := keyWidth * (naturalHeight / naturalWidth)

	// ── 4. Scale to keyWidth × keyHeight ─────────────────────────────────────
	keyPath, err = transforms.ScaleTransform{
		Width:            keyWidth,
		Height:           keyHeight,
		SegmentOperators: so,
	}.PathTransform(keyPath)
	if err != nil {
		return nil, ctx, err
	}

	// ── 5. For socket: flip the key so the bottom half becomes the top half ───
	if keySide == "socket" {
		keyPath, err = transforms.MirrorTransform{
			Axis:             transforms.Horizontal,
			Handle:           path.TopLeft,
			SegmentOperators: so,
		}.PathTransform(keyPath)
		if err != nil {
			return nil, ctx, err
		}
		keyPath, err = transforms.TrimWhitespaceTransform{SegmentOperators: so}.PathTransform(keyPath)
		if err != nil {
			return nil, ctx, err
		}
	}

	// ── 6. Slice at key_height/2: keep the top half (y ≤ key_height/2) ───────
	//      After step 5 the "correct" half is always the top half regardless of
	//      whether we are building a plug or socket pocket.
	keyPath, err = transforms.HSliceTransform{
		Y:                keyHeight / 2,
		SegmentOperators: so,
		Precision:        precision,
	}.PathTransform(keyPath)
	if err != nil {
		return nil, ctx, err
	}

	// ── 7. Flip so the seam is at y=0 and the deepest cut is at y=key_height/2 ─
	//      Before this step y=0 is the deepest point and y=key_height/2 is the
	//      seam.  Mirroring around y=0 (TopLeft) swaps them.
	keyPath, err = transforms.MirrorTransform{
		Axis:             transforms.Horizontal,
		Handle:           path.TopLeft,
		SegmentOperators: so,
	}.PathTransform(keyPath)
	if err != nil {
		return nil, ctx, err
	}
	keyPath, err = transforms.TrimWhitespaceTransform{SegmentOperators: so}.PathTransform(keyPath)
	if err != nil {
		return nil, ctx, err
	}

	// ── 8. Connect any disjoint segments within the key profile ───────────────
	//      HSlice may split the key into disconnected pieces.  JoinTransform
	//      reconnects them by extending segment lines until they meet.
	keyPath, err = transforms.JoinTransform{
		Precision:        precision,
		SegmentOperators: so,
	}.PathTransform(keyPath)
	if err != nil {
		return nil, ctx, err
	}

	// ── 9. Centre the key profile along the edge ──────────────────────────────
	offset := (edgeLen - keyWidth) / 2
	if offset < 0 {
		offset = 0
	}
	keyPath, err = transforms.ShiftTransform{
		DeltaX:           offset,
		SegmentOperators: so,
	}.PathTransform(keyPath)
	if err != nil {
		return nil, ctx, err
	}

	// ── 10. Build straight sections on either side of the pocket ──────────────
	leftDraw := path.NewDraw()
	leftDraw.MoveTo(path.NewPoint(0, 0))
	leftDraw.LineTo(path.NewPoint(offset, 0))

	rightDraw := path.NewDraw()
	rightDraw.MoveTo(path.NewPoint(offset+keyWidth, 0))
	rightDraw.LineTo(path.NewPoint(edgeLen, 0))

	// ── 11. Assemble and close all gaps in one JoinTransform pass ─────────────
	//       Concatenate all three sub-paths' segments into one path, then run
	//       JoinTransform.  JoinLines extends the straight sections to meet the
	//       key profile's actual seam-level entry/exit points, filling any gap
	//       that arises when the key shape does not span the full key_width at y=0.
	var allSegs []path.Segment
	allSegs = append(allSegs, leftDraw.Path().Segments()...)
	allSegs = append(allSegs, keyPath.Segments()...)
	allSegs = append(allSegs, rightDraw.Path().Segments()...)

	fullPath, err := transforms.JoinTransform{
		Precision:        precision,
		SegmentOperators: so,
	}.PathTransform(path.NewPathFromSegments(allSegs))
	if err != nil {
		return nil, ctx, err
	}

	// ── 12. Rotate to match the from→to direction ─────────────────────────────
	handle := attr.MustHandle("handle", path.Origin)
	axisPoint, err := path.PointPathAttribute(handle, fullPath, so)
	if err != nil {
		return fullPath, ctx, err
	}
	if edgeLine.Angle() != 0 {
		fullPath, err = transforms.RotateTransform{
			Degrees:          edgeLine.Angle(),
			Axis:             handle,
			SegmentOperators: so,
		}.PathTransform(fullPath)
		if err != nil {
			return fullPath, ctx, err
		}
	}

	// ── 13. Translate to startPoint ───────────────────────────────────────────
	fullPath, err = transforms.ShiftTransform{
		DeltaX:           startPoint.X - axisPoint.X,
		DeltaY:           startPoint.Y - axisPoint.Y,
		SegmentOperators: so,
	}.PathTransform(fullPath)
	if err != nil {
		return fullPath, ctx, err
	}

	return ke.HandleTransforms(ke, fullPath, ctx)
}
