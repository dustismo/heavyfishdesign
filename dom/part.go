package dom

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

// individual part (like a box face)
// Each part should have its own 0,0 base coordinate system.
type Part struct {
	*BasicComponent
	PartTransformers []PartTransformer
}

type RenderedPart struct {
	Part   *Part
	Path   path.Path
	Width  float64
	Height float64
}

type PartTransformer interface {
	TransformPart(part *RenderedPart, ctx RenderContext) ([]*RenderedPart, error)
}

// since Part is also a component we use a factory for it
type PartFactory struct{}

func (pf PartFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *DocumentContext) (Component, error) {
	// create the owning element
	dm := mp.Clone()
	// part does not require a type, but we need a type
	// since it could be added via a reference.
	dm.PutIfAbsent("type", "part")
	bc := AppContext().MakeBasicComponent(dm)
	components, err := AppContext().MakeComponents(dm.MustDynMapSlice("components", []*dynmap.DynMap{}), dc)
	if err != nil {
		return nil, err
	}
	// set the components based on the parsed component so we have proper
	// represention of imported or mutations
	dm.Put("components", ComponentsToDynMap(components))

	bc.SetComponents(components)
	p := &Part{
		BasicComponent: bc,
	}
	for _, c := range components {
		c.SetParent(p)
	}
	return p, nil
}

// The list of component types this Factory should be used for
func (pf PartFactory) ComponentTypes() []string {
	return []string{"part"}
}

// gets the owning document
func (p *Part) Document() *Document {
	d := p.Parent()
	if d == nil {
		return nil
	}
	return d.(*Document)
}

func (p *Part) RenderPart(ctx RenderContext) ([]*RenderedPart, error) {
	renderedParts := []*RenderedPart{}

	attr := p.Attr()

	// TODO: Repeat could be a part transformer
	repeat := attr.MustInt("repeat.total", 1)
	for i := 0; i < repeat; i++ {
		context := ctx.Clone()
		p.SetLocalVariable("part_index", i)
		pth, _, err := p.Render(context)
		if err != nil {
			return nil, err
		}
		// trim any whitespace
		pth, err = transforms.TrimWhitespaceTransform{
			SegmentOperators: AppContext().SegmentOperators(),
		}.PathTransform(pth)
		if err != nil {
			return nil, err
		}

		tl, br, err := path.BoundingBoxTrimWhitespace(pth, AppContext().SegmentOperators())
		if err != nil {
			return nil, err
		}
		width := br.X - tl.X
		height := br.Y - tl.Y

		renderedParts = append(renderedParts, &RenderedPart{
			Part:   p,
			Path:   pth,
			Width:  width,
			Height: height,
		})
	}

	// create the part transformers
	transforms, err := AppContext().MakePartTransformers(p.originalMap.MustDynMapSlice("part_transformers", []*dynmap.DynMap{}), p)
	if err != nil {
		return nil, err
	}
	for _, pt := range transforms {
		renderedPartsTmp := []*RenderedPart{}
		for _, rp := range renderedParts {
			rps, err := pt.TransformPart(rp, ctx)
			if err != nil {
				return nil, err
			}
			renderedPartsTmp = append(renderedPartsTmp, rps...)
		}
		renderedParts = renderedPartsTmp
	}
	return renderedParts, nil
}

// render a single part.  This satisfies the Component interface,
// typically RenderPart should be used instead as that will honor the
// repeats or splits
func (p *Part) Render(ctx RenderContext) (path.Path, RenderContext, error) {
	var pth path.Path
	context := ctx.Clone()
	paths := []path.Path{}
	for _, e := range p.Children() {
		component, ok := e.(Component)
		if !ok {
			return nil, ctx, fmt.Errorf("Error, part children must be components")
		}

		p1, cTmp, err := component.Render(context)
		if err != nil {
			return p1, context, err
		}
		context = cTmp
		// find the new cursor location
		context.Cursor = path.PathCursor(p1)
		paths = append(paths, p1)
	}
	if len(paths) == 0 {
		return nil, context, fmt.Errorf("Error no path found to render in \n%s", p.ToDynMap().ToJSON())
	}
	pth = transforms.SimpleJoin{}.JoinPaths(paths...)
	context.Cursor = path.PathCursor(pth)

	p1, c1, err := p.HandleTransforms(p, pth, context)

	if err != nil {
		return p1, c1, err
	}

	// now trim any whitespace and measure
	// calculate the width and height
	p2, err := transforms.TrimWhitespaceTransform{
		SegmentOperators: AppContext().SegmentOperators(),
	}.PathTransform(p1)
	if err != nil {
		return p2, c1, err
	}

	// collapse multiple Moves
	// TODO: this should be a transform, or something..
	sgs := []path.Segment{}
	for _, s := range p2.Segments() {
		if path.IsMove(s) {
			sgs = path.TrimTailMove(sgs)
		}
		sgs = append(sgs, s)
	}
	p2 = path.NewPathFromSegments(sgs)

	return p2, c1, err
}
