package components

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

type AroundComponentFactory struct{}

type AroundComponent struct {
	*dom.BasicComponent
	Repeatable dom.Component // the thing to be repeated
}

func (ccf AroundComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *dom.DocumentContext) (dom.Component, error) {
	factory := dom.AppContext()
	dm := mp.Clone()

	repeatDM := dm.MustDynMap("repeatable", dynmap.New())
	repeat, err := factory.MakeComponent(repeatDM, dc)
	if err != nil {
		return nil, err
	}

	// remember to add components to the map
	dm.Put("repeatable", repeat)

	bc := factory.MakeBasicComponent(dm)

	rc := &AroundComponent{
		BasicComponent: bc,
		Repeatable:     repeat,
	}

	repeat.SetParent(rc)
	rc.SetChildren([]dom.Element{repeat})
	return rc, nil
}

// The list of component types this Factory should be used for
func (ccf AroundComponentFactory) ComponentTypes() []string {
	return []string{"around"}
}

func (rc *AroundComponent) Render(ctx dom.RenderContext) (path.Path, dom.RenderContext, error) {
	rc.RenderStart(ctx)
	attr := rc.Attr()

	numEdges := attr.MustInt("num_edges", 0)
	if numEdges == 0 {
		return nil, ctx, fmt.Errorf("Error, Around component (%s) must have 'num_edges' attribute", rc.Id())
	}

	radius := attr.MustFloat64("radius", 0.0)
	if radius == 0 {
		return nil, ctx, fmt.Errorf("Error, Around component (%s) must have 'radius' attribute", rc.Id())
	}

	// figure out the angle and width of each segment
	centerPoint := attr.MustPoint("center_point", ctx.Cursor)
	degrees := 360.0 / float64(numEdges)

	startPoint := path.NewLineSegmentAngle(centerPoint, radius, -(degrees/2)-90).End()
	endPoint := path.NewLineSegmentAngle(centerPoint, radius, (degrees/2)-90).End()

	line := path.LineSegment{
		StartPoint: startPoint,
		EndPoint:   endPoint,
	}

	// fmt.Printf("Center point %s\nStart point %s\nEnd Point %s\nWidth %.3f\ndegrees %.3f\n",
	// 	centerPoint.StringPrecision(3),
	// 	startPoint.StringPrecision(3),
	// 	endPoint.StringPrecision(3),
	// 	line.Length(),
	// 	degrees,
	// )

	// trianglePth := path.NewPathFromSegments([]path.Segment{
	// 	path.LineSegment{centerPoint, startPoint},
	// 	line,
	// 	path.LineSegment{endPoint, centerPoint},
	// })
	// now render a single path (this is the horizontal top path)
	rCtx := ctx.Clone()
	rc.SetLocalVariable("around__length", line.Length())
	rCtx.Cursor = startPoint
	paths := []path.Path{}
	for i := 0; i < numEdges; i++ {
		rc.SetLocalVariable("around__index", 0)
		p, c, err := rc.Repeatable.Render(rCtx)
		if err != nil {
			return p, c, err
		}

		pth, err := transforms.RotateTransform{
			Degrees:          degrees * float64(i),
			Axis:             path.ToPathAttrFromPoint(centerPoint, dom.AppContext().Precision()),
			SegmentOperators: dom.AppContext().SegmentOperators(),
		}.PathTransform(p)
		if err != nil {
			return pth, ctx, err
		}
		paths = append(paths, pth)
	}
	p := transforms.SimpleJoin{}.JoinPaths(paths...)

	return p, rCtx, nil
}
