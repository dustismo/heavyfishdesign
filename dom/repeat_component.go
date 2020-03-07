package dom

import (
	"fmt"
	"math"

	"github.com/dustismo/heavyfishdesign/transforms"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

type RepeatComponentFactory struct{}

type RepeatComponent struct {
	*BasicComponent
	component Component // the thing to be repeated
}

func (rcf RepeatComponentFactory) CreateComponent(componentType string, dm *dynmap.DynMap, dc *DocumentContext) (Component, error) {
	factory := AppContext()
	bc := factory.MakeBasicComponent(dm)

	elementDM := dm.MustDynMap("component", dynmap.New())
	element, err := factory.MakeComponent(elementDM, dc)
	if err != nil {
		return nil, err
	}

	rc := &RepeatComponent{
		BasicComponent: bc,
		component:      element,
	}
	// set the parent and children
	element.SetParent(rc)
	rc.children = []Element{element}
	return rc, nil
}

// The list of component types this Factory should be used for
func (rcf RepeatComponentFactory) ComponentTypes() []string {
	return []string{"repeat"}
}

func RepeatRender(ctx RenderContext, component Component, maxX, maxY float64) (path.Path, RenderContext, error) {
	// send in a fresh context, since we need this to render at 0,0
	context := ctx.Clone()
	context.Cursor = path.NewPoint(0, 0)
	so := AppContext().SegmentOperators()
	p, _, err := component.Render(context)
	if err != nil {
		return p, context, err
	}
	// measure what we are drawing.
	bbTL, bbBR, err := path.BoundingBoxWithWhitespace(p, so)
	if err != nil {
		return p, context, err
	}
	width := bbBR.X - bbTL.X
	height := bbBR.Y - bbTL.Y

	// find how many iterations we are going to do.
	iterationsX := 1
	if maxX > 0 {
		iterationsX = int(math.Floor(maxX / (width)))
	}

	iterationsY := 1
	if maxY > 0 {
		iterationsY = int(math.Floor(maxY / (height)))
	}

	// now render the same block N times
	paths := []path.Path{}
	for ix := 0; ix < iterationsX; ix++ {
		for iy := 0; iy < iterationsY; iy++ {
			component.Params().Put("index_x", ix)
			component.Params().Put("index_y", iy)
			pth, _, err := component.Render(context)
			if err != nil {
				return pth, context, err
			}

			// now adjust the x and y
			newOrigin := path.NewPoint(float64(ix)*width, float64(iy)*height)
			pth, err = transforms.MoveTransform{
				Point:            newOrigin,
				Handle:           path.StartPosition,
				SegmentOperators: so,
			}.PathTransform(pth)
			if err != nil {
				return pth, context, err
			}
			paths = append(paths, pth)
		}
	}
	if len(paths) == 0 {
		return p, context, fmt.Errorf("Error could not repeat, probably not enough space for 1 iteration.")
	}

	// join the paths
	pth := transforms.SimpleJoin{}.JoinPaths(paths...)
	context.Cursor = path.PathCursor(pth)

	return pth, context, nil
}
func (rc *RepeatComponent) Render(ctx RenderContext) (path.Path, RenderContext, error) {
	rc.RenderStart(ctx)
	attr := rc.Attr()
	maxX := attr.MustFloat64("max_x", 0.0)
	maxY := attr.MustFloat64("max_y", 0.0)

	p, context, err := RepeatRender(ctx, rc.component, maxX, maxY)

	if err != nil {
		return p, ctx, err
	}
	return rc.HandleTransforms(rc, p, context)
}
