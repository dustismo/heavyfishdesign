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


	varName := attr.MustString("lathe_variable_name", "")


	fmt.Printf("BOUNDING BOX: %+v  TO %+v\n", outlineTopLeft, outlineBottomRight)
	maxWidth := outlineBottomRight.X - outlineTopLeft.X

	part.Part.SetGlobalVariable(fmt.Sprintf("%s__max_width", varName), maxWidth)

	thickness := attr.MustFloat64("material_thickness", 0.0)
	if thickness <= 0 {
		return nil, fmt.Errorf("Lathe requires material_thickness be set")
	}

	repeatDM := lpt.mp.MustDynMap("repeat", dynmap.New())
	repeat, err := AppContext().MakeComponent(repeatDM, FindDocumentContext(part.Part))
	if err != nil {
		return nil, err
	}

	paddingTop := attr.MustFloat64("padding_top", 0)
	paddingBottom := attr.MustFloat64("padding_bottom", 0)

	repeatOutsideDM, shouldInvert := lpt.mp.GetDynMap("repeat_outside")
	repeatOutsidePadding := attr.MustFloat64("repeat_outside_padding", thickness)

	if shouldInvert {
		repeatOutside, err := AppContext().MakeComponent(repeatOutsideDM, FindDocumentContext(part.Part))
		if shouldInvert && err != nil {
			return nil, err
		}
		println("*************")
		println(repeatOutsideDM.ToJSON())

		println(repeatOutside)
		println("*************")
		repeatOutside.SetParent(part.Part)

		// op, _, err := repeatOutside.Render(ctx)
		// if err != nil {
		// 	return nil, err
		// }

		// get the bounding box
		// tl, br, err := path.BoundingBoxTrimWhitespace(op, so)
		// if err != nil {
		// 	return nil, err
		// }
		// now center within the bounding box...
		// xOffset = ((br.X - tl.X) - (outlineBottomRight.X - outlineTopLeft.X)) / 2
		// println("xOffset")
		// println(outlineTopLeft.X)
		// println(outlineBottomRight.X)
		// // move the outline based on xOffset
		// outline, err = path.ShiftPath(xOffset, 0, outline, so)
		// if err != nil {
		// 	return nil, err
		// }
	}

	// set the parent
	repeat.SetParent(part.Part)

	renderedParts := []*RenderedPart{}
	topLength := 0.0
	bottomLength := 0.0
	index := 0
	totalHeight := 0.0
	for y := outlineTopLeft.Y + paddingTop; y <= outlineBottomRight.Y-paddingBottom; y = y + thickness {
		points, err := path.HorizontalIntercepts(outline, y, AppContext().SegmentOperators())
		if err != nil {
			return nil, err
		}

		if len(points) < 2 {
			// we just skip missing pieces
			continue
		}

		// render the component
		from := points[0]
		to := points[1]
		length := to.X - from.X
		if path.PrecisionEquals(length, 0, AppContext().Precision()) {
			continue
		}
		fmt.Printf("from: %+v\n", from)
		fmt.Printf("to: %+v\n", to)

		repeat.SetLocalVariable("max_part_width", maxWidth)
		repeat.SetLocalVariable("from__x", from.X)
		repeat.SetLocalVariable("from__y", from.Y)

		repeat.SetLocalVariable("to", to)
		repeat.SetLocalVariable("to__x", to.X)
		repeat.SetLocalVariable("to__y", to.Y)
		repeat.SetLocalVariable("lathe__width", length)
		repeat.SetLocalVariable("lathe__index", index)
		p, _, err := repeat.Render(ctx)
		if err != nil {
			return nil, err
		}

		if shouldInvert {
			// move p to the X offset (in both X and Y)
			// TODO: we might need a different strategy for centering
			// in the Y direction
			p, err = path.ShiftPath(from.X + repeatOutsidePadding, 
				from.X + repeatOutsidePadding, p, so)
			if err != nil {
				return nil, err
			}


			repeatOutside, err := AppContext().MakeComponent(repeatOutsideDM, FindDocumentContext(part.Part))
			if shouldInvert && err != nil {
				return nil, err
			}
			repeatOutside.SetParent(part.Part)
			repeatOutside.SetLocalVariable("max_part_width", maxWidth)
			op, _, err := repeatOutside.Render(ctx)
			// now combine with p
			p.AddSegments(op.Segments()...)
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

		if index == 0 {
			topLength = length
		}
		bottomLength = length
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

	if len(varName) > 0 {
		part.Part.SetGlobalVariable(fmt.Sprintf("%s__total_height", varName), totalHeight)
		part.Part.SetGlobalVariable(fmt.Sprintf("%s__top_width", varName), topLength)
		part.Part.SetGlobalVariable(fmt.Sprintf("%s__bottom_width", varName), bottomLength)
		part.Part.SetGlobalVariable(fmt.Sprintf("%s__num_layers", varName), index)
	}

	return renderedParts, nil
}
