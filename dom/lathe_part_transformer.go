package dom

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

type LathePartTransform struct {
	mp *dynmap.DynMap
}

func (lpt *LathePartTransform) TransformPart(part *RenderedPart, ctx RenderContext) ([]*RenderedPart, error) {
	so := AppContext().SegmentOperators()
	attr := part.Part.DmAttr(lpt.mp)

	outline := part.Path
	outlineTopLeft, outlineBottomRight, err := path.BoundingBoxWithWhitespace(outline, AppContext().SegmentOperators())

	if err != nil {
		return nil, err
	}

	thickness := attr.MustFloat64("material_thickness", 0.0)
	if thickness <= 0 {
		return nil, fmt.Errorf("Lathe requires material_thickness be set")
	}

	repeatDM := lpt.mp.MustDynMap("repeat", dynmap.New())
	repeat, err := AppContext().MakeComponent(repeatDM, FindDocumentContext(part.Part))
	if err != nil {
		return nil, err
	}

	// set the parent
	repeat.SetParent(part.Part)

	renderedParts := []*RenderedPart{}
	index := 0
	totalHeight := 0.0
	for y := outlineTopLeft.Y; y <= outlineBottomRight.Y; y = y + thickness {
		points, err := path.HorizontalIntercepts(outline, y, AppContext().SegmentOperators())
		if err != nil {
			return nil, err
		}

		if len(points) < 2 {
			return nil, fmt.Errorf("unable to calculate intercepts for lathe")
		}

		// render the component
		from := points[0]
		to := points[1]
		length := to.X - from.X
		if path.PrecisionEquals(length, 0, AppContext().Precision()) {
			continue
		}

		repeat.SetLocalVariable("from__x", from.X)
		repeat.SetLocalVariable("from__y", from.Y)

		repeat.SetLocalVariable("to", to)
		repeat.SetLocalVariable("to__x", to.X)
		repeat.SetLocalVariable("to__y", to.Y)
		repeat.SetLocalVariable("length", length)
		repeat.SetLocalVariable("part_index", index)
		p, _, err := repeat.Render(ctx)
		if err != nil {
			return nil, err
		}

		tl, br, err := path.BoundingBoxTrimWhitespace(p, so)
		if err != nil {
			return nil, err
		}
		width := br.X - tl.X
		height := br.Y - tl.Y

		label := part.Label
		if len(label.Text) > 0 {
			// append the index to the label
			label.Text = fmt.Sprintf("%s:%d", label.Text, index)
		}

		renderedParts = append(renderedParts, &RenderedPart{
			Part:   part.Part,
			Path:   p,
			Width:  width,
			Height: height,
			Label:  label,
		})
		index = index + 1
		totalHeight = totalHeight + thickness
	}

	// TODO: add total height to the globals

	return renderedParts, nil
}
