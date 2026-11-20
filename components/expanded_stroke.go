package components

import (
	"fmt"
	"math"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

// ExpandedStrokeComponentFactory thickens a child line drawing with a joined
// stroke and clips it to the panel margin.
type ExpandedStrokeComponentFactory struct{}

type ExpandedStrokeComponent struct {
	*dom.BasicComponent
	components []dom.Component
}

func (f ExpandedStrokeComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *dom.DocumentContext) (dom.Component, error) {
	factory := dom.AppContext()
	dm := mp.Clone()
	children, err := factory.MakeComponents(dm.MustDynMapSlice("components", []*dynmap.DynMap{}), dc)
	if err != nil {
		return nil, err
	}
	dm.Put("components", dom.ComponentsToDynMap(children))
	bc := factory.MakeBasicComponent(dm)
	bc.SetComponents(children)
	component := &ExpandedStrokeComponent{
		BasicComponent: bc,
		components:     children,
	}
	for _, child := range children {
		child.SetParent(component)
	}
	return component, nil
}

func (f ExpandedStrokeComponentFactory) ComponentTypes() []string {
	return []string{"expanded_stroke"}
}

func (c *ExpandedStrokeComponent) Children() []dom.Element {
	return dom.CtoE(c.components)
}

func (c *ExpandedStrokeComponent) Render(ctx dom.RenderContext) (path.Path, dom.RenderContext, error) {
	c.RenderStart(ctx)
	attr := c.Attr()
	width := attr.MustFloat64("width", 0)
	height := attr.MustFloat64("height", 0)
	margin := attr.MustFloat64("margin", 0)
	thickness := attr.MustFloat64("thickness", 0.1)
	if width <= 0 || height <= 0 {
		return nil, ctx, fmt.Errorf("expanded_stroke requires width and height")
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

	joined := fitDrawingToPanel(transforms.SimpleJoin{}.JoinPaths(paths...), margin, width, height)
	result := strokeCenterline(joined, thickness, margin, width, height)
	if result == nil || len(result.Segments()) == 0 {
		empty := path.NewDraw()
		empty.MoveTo(path.NewPoint(0, 0))
		result = empty.Path()
	}
	return c.HandleTransforms(c, result, context)
}

// fitDrawingToPanel scales the child drawing, keeping its proportions, so its
// bounding box sits inside the inner margin and is centered there.
func fitDrawingToPanel(drawing path.Path, margin, width, height float64) path.Path {
	if drawing == nil || len(drawing.Segments()) == 0 {
		return drawing
	}
	innerX0, innerY0 := margin, margin
	innerX1, innerY1 := width-margin, height-margin
	innerW, innerH := innerX1-innerX0, innerY1-innerY0
	if innerW <= 1e-9 || innerH <= 1e-9 {
		return drawing
	}
	tl, br, err := path.BoundingBoxTrimWhitespace(drawing, path.NewSegmentOperators())
	if err != nil {
		return drawing
	}
	bw, bh := br.X-tl.X, br.Y-tl.Y
	scale := 1.0
	switch {
	case bw > 1e-9 && bh > 1e-9:
		scale = math.Min(innerW/bw, innerH/bh)
	case bw > 1e-9:
		scale = innerW / bw
	case bh > 1e-9:
		scale = innerH / bh
	}
	scaledW, scaledH := bw*scale, bh*scale
	destX := innerX0 + (innerW-scaledW)/2
	destY := innerY0 + (innerH-scaledH)/2
	ops := path.NewSegmentOperators()
	moved := make([]path.Segment, 0, len(drawing.Segments()))
	for _, seg := range drawing.Segments() {
		next, err := ops.TransformPoints(seg, func(p path.Point) path.Point {
			return path.NewPoint(destX+(p.X-tl.X)*scale, destY+(p.Y-tl.Y)*scale)
		})
		if err != nil {
			return drawing
		}
		moved = append(moved, next)
	}
	return path.NewPathFromSegments(moved)
}
