package components

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

type XInterceptComponentFactory struct{}

type XInterceptComponent struct {
	*dom.BasicComponent
	Outline    dom.Component // the outline that the repeated thing will be inside
	Repeatable dom.Component // the thing to be repeated
}

func (ccf XInterceptComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *dom.DocumentContext) (dom.Component, error) {
	factory := dom.AppContext()
	dm := mp.Clone()

	repeatDM := dm.MustDynMap("repeatable", dynmap.New())
	repeat, err := factory.MakeComponent(repeatDM, dc)
	if err != nil {
		return nil, err
	}

	outlineDM := dm.MustDynMap("outline", dynmap.New())
	outline, err := factory.MakeComponent(outlineDM, dc)
	if err != nil {
		return nil, err
	}

	// remember to add components to the map
	dm.Put("repeatable", repeat)
	dm.Put("outline", outline)
	bc := factory.MakeBasicComponent(dm)

	rc := &XInterceptComponent{
		BasicComponent: bc,
		Repeatable:     repeat,
		Outline:        outline,
	}

	repeat.SetParent(rc)
	outline.SetParent(rc)

	rc.SetChildren([]dom.Element{repeat, outline})
	return rc, nil
}

// The list of component types this Factory should be used for
func (ccf XInterceptComponentFactory) ComponentTypes() []string {
	return []string{"xintercept"}
}

func (rc *XInterceptComponent) Render(ctx dom.RenderContext) (path.Path, dom.RenderContext, error) {
	rc.RenderStart(ctx)
	attr := rc.Attr()

	startY := attr.MustFloat64("initial_spacing", 0)
	repeatY := attr.MustFloat64("repeat_spacing", 0)

	if repeatY == 0 {
		return nil, ctx, fmt.Errorf("repeat_spacing is manditory %.3f", repeatY)
	}

	outline, _, err := rc.Outline.Render(ctx)
	if err != nil {
		return nil, ctx, nil
	}
	outlineTopLeft, outlineBottomRight, err := path.BoundingBoxWithWhitespace(outline, dom.AppContext().SegmentOperators())

	if err != nil {
		return nil, ctx, nil
	}

	paths := []path.Path{}

	repeatableHeight := 0.0
	for y := outlineTopLeft.Y + startY; y <= outlineBottomRight.Y; y = y + repeatableHeight + repeatY {
		points, err := path.HorizontalIntercepts(outline, y, dom.AppContext().SegmentOperators())
		if err != nil {
			return nil, ctx, nil
		}

		// now iterate by pairs across the xintercepts.
		// this allows us to do more complicated shapes
		for xIndex := 0; xIndex+1 < len(points); xIndex = xIndex + 2 {
			// render the component
			from := points[xIndex]
			to := points[xIndex+1]
			length := to.X - from.X

			// if length is zero then don't render
			if !path.PrecisionEquals(length, 0, dom.AppContext().Precision()) {

				rc.Repeatable.Params().Put("from__x", from.X)
				rc.Repeatable.Params().Put("from__y", from.Y)

				rc.Repeatable.Params().Put("to", to)
				rc.Repeatable.Params().Put("to__x", to.X)
				rc.Repeatable.Params().Put("to__y", to.Y)
				rc.Repeatable.Params().Put("length", length)
				p, _, err := rc.Repeatable.Render(ctx)
				if err != nil {
					return nil, ctx, nil
				}
				// find the repeatable height
				tl, br, err := path.BoundingBoxTrimWhitespace(p, dom.AppContext().SegmentOperators())
				if err != nil {
					return nil, ctx, nil
				}
				repeatableHeight = br.Y - tl.Y
				paths = append(paths, p)
			}
		}
	}
	if len(paths) < 1 {
		return nil, ctx, fmt.Errorf("Unable to render component with id %s, repeatable did not fit within the outline", rc.Id())
	}
	p := transforms.SimpleJoin{}.JoinPaths(paths...)

	return rc.HandleTransforms(rc, p, ctx)
}
