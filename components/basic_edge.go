package components

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

type BasicEdgeComponentFactory struct{}

// an edge is a way to connect two parts together.
type BasicEdgeComponent struct {
	*dom.BasicComponent
	components       []dom.Component
	segmentOperators path.SegmentOperators
}

func (becf BasicEdgeComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *dom.DocumentContext) (dom.Component, error) {
	factory := dom.AppContext()
	dm := mp.Clone()
	components, err := factory.MakeComponents(dm.MustDynMapSlice("components", []*dynmap.DynMap{}), dc)
	if err != nil {
		return nil, err
	}
	dm.Put("components", dom.ComponentsToDynMap(components))
	bc := factory.MakeBasicComponent(dm)
	bc.SetComponents(components)
	gc := &BasicEdgeComponent{
		BasicComponent:   bc,
		components:       components,
		segmentOperators: factory.SegmentOperators(),
	}
	for _, c := range components {
		c.SetParent(gc)
	}
	return gc, nil
}

// The list of component types this Factory should be used for
func (becf BasicEdgeComponentFactory) ComponentTypes() []string {
	return []string{"basic_edge"}
}

func (rc *BasicEdgeComponent) Render(ctx dom.RenderContext) (path.Path, dom.RenderContext, error) {
	rc.RenderStart(ctx)
	attr := rc.Attr()
	startPoint := attr.MustPoint("from", ctx.Cursor)
	endPoint, found := attr.Point2("to", startPoint)
	if !found {
		return nil, ctx, fmt.Errorf("Error, Edge component (%s) must have 'to' attribute", rc.Id())
	}
	handle := attr.MustHandle("handle", path.Origin)
	line := path.LineSegment{
		StartPoint: startPoint,
		EndPoint:   endPoint,
	}

	variableName, ok := attr.String("edge_variable_name")
	if ok {
		// set the various params for this edge
		rc.SetGlobalVariable(fmt.Sprintf("%s__length", variableName), line.Length())
		rc.SetGlobalVariable(fmt.Sprintf("%s__angle", variableName), line.Angle())

		// TODO: a submap would be nice, but the
		// parsing lib doesnt allow for dot syntax params
		// variables := dynmap.New()
		// variables.Put("length", line.Length())
		// variables.Put("angle", line.Angle())
		// variables.PutWithDot("from.x", startPoint.X)
		// variables.PutWithDot("from.y", startPoint.Y)
		// variables.PutWithDot("to.x", endPoint.X)
		// variables.PutWithDot("to.y", endPoint.Y)
		// docCtx.Params.Put(variableName, variables)
	}
	// render all the children then merge into one path.
	paths := []path.Path{}

	context := ctx.Clone()
	rc.SetLocalVariable("width", line.Length())
	for _, e := range rc.components {
		p, c, err := e.Render(context)
		if err != nil {
			return p, c, err
		}
		paths = append(paths, p)
		context = c
		// find the new cursor location
		context.Cursor = path.PathCursor(p)
	}

	p := transforms.SimpleJoin{}.JoinPaths(paths...)
	// now rotate and move if needed

	// track the handle point before we rotate.
	axisPoint, err := path.PointPathAttribute(handle, p, dom.AppContext().SegmentOperators())
	if err != nil {
		return p, ctx, err
	}
	if line.Angle() != 0 {
		p, err = transforms.RotateTransform{
			Degrees:          line.Angle(),
			Axis:             handle,
			SegmentOperators: dom.AppContext().SegmentOperators(),
		}.PathTransform(p)
		if err != nil {
			return p, ctx, err
		}
	}
	// now move to the correct start point
	p, err = transforms.ShiftTransform{
		DeltaX:           startPoint.X - axisPoint.X,
		DeltaY:           startPoint.Y - axisPoint.Y,
		SegmentOperators: dom.AppContext().SegmentOperators(),
	}.PathTransform(p)
	if err != nil {
		return p, ctx, err
	}
	return rc.HandleTransforms(rc, p, ctx)
}
