package components

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

// FramedStrokeComponentFactory thickens a child line drawing with a joined
// stroke and clips it to the panel margin.
type FramedStrokeComponentFactory struct{}

type FramedStrokeComponent struct {
	*dom.BasicComponent
	components []dom.Component
}

func (f FramedStrokeComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *dom.DocumentContext) (dom.Component, error) {
	factory := dom.AppContext()
	dm := mp.Clone()
	children, err := factory.MakeComponents(dm.MustDynMapSlice("components", []*dynmap.DynMap{}), dc)
	if err != nil {
		return nil, err
	}
	dm.Put("components", dom.ComponentsToDynMap(children))
	bc := factory.MakeBasicComponent(dm)
	bc.SetComponents(children)
	component := &FramedStrokeComponent{
		BasicComponent: bc,
		components:     children,
	}
	for _, child := range children {
		child.SetParent(component)
	}
	return component, nil
}

func (f FramedStrokeComponentFactory) ComponentTypes() []string {
	return []string{"framed_stroke"}
}

func (c *FramedStrokeComponent) Children() []dom.Element {
	return dom.CtoE(c.components)
}

func (c *FramedStrokeComponent) Render(ctx dom.RenderContext) (path.Path, dom.RenderContext, error) {
	c.RenderStart(ctx)
	attr := c.Attr()
	width := attr.MustFloat64("width", 0)
	height := attr.MustFloat64("height", 0)
	margin := attr.MustFloat64("margin", 0)
	thickness := attr.MustFloat64("thickness", 0.1)
	if width <= 0 || height <= 0 {
		return nil, ctx, fmt.Errorf("framed_stroke requires width and height")
	}
	if margin < 0 {
		margin = 0
	}

	context := ctx.Clone()
	paths := make([]path.Path, 0, len(c.components))
	for _, child := range c.components {
		rendered, next, err := child.Render(context)
		if err != nil {
			return rendered, next, err
		}
		paths = append(paths, rendered)
		context = next
		if rendered != nil && len(rendered.Segments()) > 0 {
			context.Cursor = path.PathCursor(rendered)
		}
	}

	joined, err := fitDrawingToPanel(transforms.SimpleJoin{}.JoinPaths(paths...), margin, width, height)
	if err != nil {
		return nil, ctx, err
	}
	result := strokeCenterline(joined, thickness, margin, width, height)
	if result == nil || len(result.Segments()) == 0 {
		empty := path.NewDraw()
		empty.MoveTo(path.NewPoint(0, 0))
		result = empty.Path()
	}
	return c.HandleTransforms(c, result, context)
}

// fitDrawingToPanel uses the shared scale and move transforms. Both width and
// height map the bounding box onto the inner margin. A flat drawing sets only
// the axis it has, which keeps that scale uniform.
func fitDrawingToPanel(drawing path.Path, margin, width, height float64) (path.Path, error) {
	if drawing == nil || len(drawing.Segments()) == 0 {
		return drawing, nil
	}
	innerW, innerH := width-2*margin, height-2*margin
	if innerW <= 1e-9 || innerH <= 1e-9 {
		return drawing, nil
	}
	so := path.NewSegmentOperators()
	tl, br, err := path.BoundingBoxTrimWhitespace(drawing, so)
	if err != nil {
		return nil, err
	}
	scale := transforms.ScaleTransform{SegmentOperators: so}
	if br.X-tl.X > 1e-9 {
		scale.Width = innerW
	}
	if br.Y-tl.Y > 1e-9 {
		scale.Height = innerH
	}
	scaled := drawing
	if scale.Width > 0 || scale.Height > 0 {
		scaled, err = scale.PathTransform(drawing)
		if err != nil {
			return nil, err
		}
	}
	return transforms.MoveTransform{
		Point:            path.NewPoint(margin, margin),
		Handle:           path.TopLeft,
		SegmentOperators: so,
	}.PathTransform(scaled)
}
