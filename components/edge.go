package components

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

type EdgeComponentFactory struct{}

// an edge is a way to connect two parts together.
// it contains three components:
// left -> the left most part of the edge.
//			this should be stretchable and contains a special param
//			called left_width
// repeatable -> the middle section, which should be repeated
// right -> the right side, similar to Left
//
// Each component should be drawn horizontally with an origin of 0,0
type EdgeComponent struct {
	*dom.BasicComponent
	Repeatable dom.Component // the thing to be repeated
	Left       dom.Component
	Right      dom.Component
}

func (ccf EdgeComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *dom.DocumentContext) (dom.Component, error) {
	factory := dom.AppContext()
	dm := mp.Clone()

	repeatDM := dm.MustDynMap("repeatable", dynmap.New())
	repeat, err := factory.MakeComponent(repeatDM, dc)
	if err != nil {
		return nil, err
	}

	leftDM := dm.MustDynMap("left", dynmap.New())
	left, err := factory.MakeComponent(leftDM, dc)
	if err != nil {
		return nil, err
	}
	rightDM := dm.MustDynMap("right", dynmap.New())
	right, err := factory.MakeComponent(rightDM, dc)
	if err != nil {
		return nil, err
	}
	// remember to add components to the map
	dm.Put("repeatable", repeat)
	dm.Put("left", left)
	dm.Put("right", right)

	bc := factory.MakeBasicComponent(dm)

	rc := &EdgeComponent{
		BasicComponent: bc,
		Repeatable:     repeat,
		Left:           left,
		Right:          right,
	}

	repeat.SetParent(rc)
	left.SetParent(rc)
	right.SetParent(rc)

	rc.SetChildren([]dom.Element{repeat, left, right})
	return rc, nil
}

// The list of component types this Factory should be used for
func (ccf EdgeComponentFactory) ComponentTypes() []string {
	return []string{"edge"}
}

func (rc *EdgeComponent) Render(ctx dom.RenderContext) (path.Path, dom.RenderContext, error) {
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

	paddingLeft := attr.MustFloat64("padding_left", 0)
	paddingRight := attr.MustFloat64("padding_right", 0)
	// render the connector in a horizontal line, then rotate..
	maxX := line.Length() - (paddingLeft + paddingRight)
	maxY := 0.0

	p, ctx, err := dom.RepeatRender(ctx, rc.Repeatable, maxX, maxY)
	if err != nil {
		return p, ctx, err
	}

	// measure what we got
	tl, br, err := path.BoundingBoxWithWhitespace(p, dom.AppContext().SegmentOperators())
	if err != nil {
		return p, ctx, err
	}
	repeatWidth := br.X - tl.X

	// fmt.Printf("maxX %.3f repeatWidth %.3f\n", maxX, repeatWidth)
	overflow := maxX - repeatWidth

	// now add the left and right
	leftWidth := (overflow / 2) + paddingLeft
	rightWidth := (overflow / 2) + paddingRight

	ctxLeft := ctx.Clone()
	ctxLeft.Cursor = path.NewPoint(0, 0)
	rc.Left.Params().Put("left_width", leftWidth)
	pLeft, _, err := rc.Left.Render(ctxLeft)
	if err != nil {
		return p, ctx, err
	}
	// fmt.Printf("leftWidth %.3f RightWidth %.3f\n")
	ctxRight := ctx.Clone()
	ctxRight.Cursor = path.NewPoint(0, 0)
	rc.Right.Params().Put("right_width", rightWidth)

	pRight, _, err := rc.Right.Render(ctxRight)
	if err != nil {
		return p, ctx, err
	}

	// now merge together
	// 1. left is already at 0,0
	// 2. shift repeat to leftwidth, 0
	// 3. shift right to leftwidth + repeatWidth, 0

	p, err = transforms.MoveTransform{
		Point:            path.NewPoint(leftWidth, 0),
		Handle:           path.StartPosition,
		SegmentOperators: dom.AppContext().SegmentOperators(),
	}.PathTransform(p)

	if err != nil {
		return p, ctx, err
	}

	pRight, err = transforms.MoveTransform{
		Point:            path.NewPoint(leftWidth+repeatWidth, 0),
		Handle:           path.StartPosition,
		SegmentOperators: dom.AppContext().SegmentOperators(),
	}.PathTransform(pRight)
	if err != nil {
		return p, ctx, err
	}

	// now merge the paths
	p = transforms.SimpleJoin{}.JoinPaths(pLeft, p, pRight)
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
