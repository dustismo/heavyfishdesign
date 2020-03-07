package dom

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

type PartSplitter struct {
	mp *dynmap.DynMap
}

func (ps *PartSplitter) transformPart(part *RenderedPart, y float64, ctx RenderContext) ([]*RenderedPart, error) {
	so := AppContext().SegmentOperators()
	attr := part.Part.DmAttr(ps.mp)

	//   |----|
	//   |____| (split point).  <-- plugEdge (faces down)
	//   |    | <-- socketEdge (faces up)
	//   |____|
	plugEdge, err := AppContext().MakeComponent(
		ps.mp.MustDynMap("plug_edge", dynmap.New()),
		FindDocumentContext(part.Part))
	if err != nil {
		return nil, err
	}
	socketEdge, err := AppContext().MakeComponent(
		ps.mp.MustDynMap("socket_edge", dynmap.New()),
		FindDocumentContext(part.Part))
	if err != nil {
		return nil, err
	}
	plugEdge.SetParent(part.Part)
	socketEdge.SetParent(part.Part)

	// first trim whitespace from this path.
	originalPath, err := transforms.TrimWhitespaceTransform{
		SegmentOperators: so,
	}.PathTransform(part.Path)
	if err != nil {
		return nil, err
	}

	tl, br, err := path.BoundingBoxTrimWhitespace(originalPath, so)
	if err != nil {
		return nil, err
	}

	yPos := y + attr.MustFloat64("y_offset", 0)

	// draw the plug edge
	// we swap from and to, so that the edge faces downward
	plugEdge.SetLocalVariable("from", path.NewPoint(br.X, yPos).ToDynMap())
	plugEdge.SetLocalVariable("to", path.NewPoint(tl.X, yPos).ToDynMap())

	plugEdgePath, _, err := plugEdge.Render(ctx)
	if err != nil {
		return nil, err
	}

	// Now calculate the bleed amount for the top piece
	_, brPl, err := path.BoundingBoxTrimWhitespace(plugEdgePath, so)
	if err != nil {
		return nil, err
	}
	bleedTop := attr.MustFloat64("bleed_top", brPl.Y-yPos)

	topPath, err := transforms.HSliceTransform{
		Y:                yPos + bleedTop,
		SegmentOperators: so,
		Precision:        AppContext().Precision(),
	}.PathTransform(originalPath)

	if err != nil {
		return nil, err
	}

	// combine the paths
	topPath = transforms.SimpleJoin{}.JoinPaths(topPath, plugEdgePath)

	// now do the other half...

	// flip the path and slice
	bottomY := br.Y - yPos
	flipped, err := transforms.MirrorTransform{
		Axis:             transforms.Horizontal,
		Handle:           path.TopLeft,
		SegmentOperators: so,
	}.PathTransform(originalPath)
	if err != nil {
		return nil, err
	}

	// draw the socket edge
	// we swap from and to, so that the edge faces downward
	socketEdge.SetLocalVariable("from", path.NewPoint(br.X, bottomY).ToDynMap())
	socketEdge.SetLocalVariable("to", path.NewPoint(tl.X, bottomY).ToDynMap())
	socketEdgePath, _, err := socketEdge.Render(ctx)
	if err != nil {
		return nil, err
	}

	// Now calculate the bleed amount for the bottom piece
	_, brSo, err := path.BoundingBoxTrimWhitespace(socketEdgePath, so)
	if err != nil {
		return nil, err
	}
	bleedBottom := attr.MustFloat64("bleed_bottom", brSo.Y-bottomY)
	bottomPath, err := transforms.HSliceTransform{
		Y:                bottomY + bleedBottom,
		SegmentOperators: so,
		Precision:        AppContext().Precision(),
	}.PathTransform(flipped)

	if err != nil {
		return nil, err
	}
	// combine the paths
	bottomPath = transforms.SimpleJoin{}.JoinPaths(bottomPath, socketEdgePath)
	// flip the bottom path back
	bp, err := transforms.MirrorTransform{
		Axis:             transforms.Horizontal,
		Handle:           path.TopLeft,
		SegmentOperators: so,
	}.PathTransform(bottomPath)
	if err != nil {
		return nil, err
	}
	bottomPath = bp

	ttl, tbr, err := path.BoundingBoxTrimWhitespace(topPath, so)
	if err != nil {
		return nil, err
	}
	twidth := tbr.X - ttl.X
	theight := tbr.Y - ttl.Y

	btl, bbr, err := path.BoundingBoxTrimWhitespace(bottomPath, so)
	if err != nil {
		return nil, err
	}
	bwidth := bbr.X - btl.X
	bheight := bbr.Y - btl.Y

	return []*RenderedPart{
		&RenderedPart{
			Part:   part.Part,
			Path:   topPath,
			Width:  twidth,
			Height: theight,
		},
		&RenderedPart{
			Part:   part.Part,
			Path:   bottomPath,
			Width:  bwidth,
			Height: bheight,
		},
	}, nil
}

func (ps *PartSplitter) TransformPart(part *RenderedPart, ctx RenderContext) ([]*RenderedPart, error) {
	attr := part.Part.DmAttr(ps.mp)
	so := AppContext().SegmentOperators()

	autoSplit := attr.MustBool("auto_split", false)
	yPos := 0.0
	if autoSplit {
		_, hasY := attr.Float64("y")
		if hasY {
			return nil, fmt.Errorf("param 'y' cannot be used with 'autosplit'")
		}
		// check if width of part > material height, or width
		tl, br, err := path.BoundingBoxTrimWhitespace(part.Path, so)
		if err != nil {
			return nil, err
		}
		w := br.X - tl.X
		h := br.Y - tl.Y

		mH := attr.MustFloat64("max_height", attr.MustFloat64("material_height", 0))
		mW := attr.MustFloat64("max_width", attr.MustFloat64("material_width", 0))
		if mH == 0 || mW == 0 {
			return nil, fmt.Errorf("for autosplit, material_height and material_width must be specified")
		}
		if (h >= mH && h >= mW) || (h >= mH && w >= mH) {
			yPos = h / 2
		} else if w >= mW && w >= mH {
			// rotate and split
			pth, err := transforms.RotateTransform{
				Degrees:          90,
				SegmentOperators: so,
			}.PathTransform(part.Path)
			if err != nil {
				return nil, err
			}
			newPart := &RenderedPart{
				Part:   part.Part,
				Path:   pth,
				Width:  h, // swap h and w since we rotated
				Height: w,
			}
			return ps.TransformPart(newPart, ctx)
		} else {
			// size fits, so don't do anything
			return []*RenderedPart{part}, nil
		}

	} else {
		y, ok := attr.Float64("y")
		if !ok {
			return nil, fmt.Errorf("Error in split transformer.  'y' must be specified")
		}
		yPos = y
	}

	if yPos == 0 {
		// do nothing
		return []*RenderedPart{part}, nil
	}
	parts, err := ps.transformPart(part, yPos, ctx)
	if err != nil || len(parts) <= 1 || !autoSplit {
		return parts, err
	}
	retParts := []*RenderedPart{}
	for _, p := range parts {
		pts, err := ps.TransformPart(p, ctx)
		if err != nil {
			return parts, err
		}
		retParts = append(retParts, pts...)
	}
	return retParts, nil
}
