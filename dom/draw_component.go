package dom

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/transforms"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

type DrawComponentFactory struct{}

type DrawComponent struct {
	*BasicComponent
	commands []*dynmap.DynMap
}

func (dcf DrawComponentFactory) CreateComponent(componentType string, dm *dynmap.DynMap, dc *DocumentContext) (Component, error) {
	commands := dm.MustDynMapSlice("commands", []*dynmap.DynMap{})
	factory := AppContext()
	bc := factory.MakeBasicComponent(dm)

	return &DrawComponent{
		BasicComponent: bc,
		commands:       commands,
	}, nil
}

// The list of component types this Factory should be used for
func (dcf DrawComponentFactory) ComponentTypes() []string {
	return []string{"draw"}
}
func (dc *DrawComponent) svgRotateScaleTo(draw *path.Draw, ctx RenderContext, to path.Point, svg string, reverse bool, svgFrom path.Point, svgTo path.Point) error {
	_, endPoint := path.GetStartAndEnd(draw.Path().Segments())
	p, err := path.ParsePathFromSvg(svg)
	if err != nil {
		return err
	}
	if reverse {
		p, err = transforms.PathReverse{}.PathTransform(p)
		if err != nil {
			return err
		}
	}

	pth, err := transforms.RotateScaleTransform{
		StartPoint:       endPoint,
		EndPoint:         to,
		PathStartPoint:   svgFrom,
		PathEndPoint:     svgTo,
		SegmentOperators: AppContext().SegmentOperators(),
	}.PathTransform(p)
	if err != nil {
		return err
	}
	draw.AddSegments(pth.Segments())
	return nil
}

func (dc *DrawComponent) svgScaleTo(draw *path.Draw, ctx RenderContext, to path.Point, svg string, reverse bool) error {
	_, endPoint := path.GetStartAndEnd(draw.Path().Segments())
	p, err := path.ParsePathFromSvg(svg)
	if err != nil {
		return err
	}
	if reverse {
		p, err = transforms.PathReverse{}.PathTransform(p)
		if err != nil {
			return err
		}
	}

	pth, err := transforms.ScaleTransform{
		StartPoint:       endPoint,
		EndPoint:         to,
		SegmentOperators: AppContext().SegmentOperators(),
	}.PathTransform(p)
	if err != nil {
		return err
	}
	pth, err = transforms.MoveTransform{
		Point:            endPoint,
		Handle:           path.StartPoint,
		SegmentOperators: AppContext().SegmentOperators(),
	}.PathTransform(pth)
	if err != nil {
		return err
	}
	draw.AddSegments(pth.Segments())
	return nil
}

func (dc *DrawComponent) Render(ctx RenderContext) (path.Path, RenderContext, error) {
	dc.RenderStart(ctx)
	draw := path.NewDraw()
	draw.MoveTo(ctx.Cursor)
	for _, e := range dc.commands {
		command := e.MustString("command", "unknown")
		attr := dc.DmAttr(e)

		switch command {
		case "move":
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			draw.MoveTo(to)
		case "rel_move":
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			draw.RelMoveTo(to)
		case "line_by_angle":
			length, ok := attr.Float64("length")
			if !ok {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "length")
			}
			angle, ok := attr.Float64("angle")
			if !ok {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "angle")
			}
			draw.LineByAngle(length, angle)
		case "line":
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			draw.LineTo(to)
		case "rel_line":
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			draw.RelLineTo(to)
		case "curve":
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			ctrlS, found := attr.Point("ctrl_start")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "ctrl_start")
			}
			ctrlE, found := attr.Point("ctrl_end")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "ctrl_end")
			}
			draw.CurveTo(ctrlS, ctrlE, to)
		case "rel_curve":
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			ctrlS, found := attr.Point("ctrl_start")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "ctrl_start")
			}
			ctrlE, found := attr.Point("ctrl_end")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "ctrl_end")
			}
			draw.RelCurveTo(ctrlS, ctrlE, to)
		case "smooth_curve":
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			ctrlE, found := attr.Point("ctrl_end")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "ctrl_end")
			}
			draw.SmoothCurveTo(ctrlE, to)
		case "rel_smooth_curve":
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			ctrlE, found := attr.Point("ctrl_end")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "ctrl_end")
			}
			draw.RelSmoothCurveTo(ctrlE, to)
		case "circle":
			radius, found := attr.Float64("radius")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "radius")
			}
			draw.Circle(radius)
		case "rectangle":
			w, found := attr.Float64("width")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "width")
			}
			h, found := attr.Float64("height")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "height")
			}
			draw.Rect(w, h)
		case "round_corner":
			radius, found := attr.Float64("radius")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "radius")
			}
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			corner, found := attr.Point("corner")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "corner")
			}
			draw.RoundedCornerTo(to, corner, radius)
		case "rel_round_corner":
			radius, found := attr.Float64("radius")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "radius")
			}
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			corner, found := attr.Point("corner")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "corner")
			}
			draw.RelRoundedCornerTo(to, corner, radius)

		case "svg":
			svg, ok := attr.SvgString("svg")
			if !ok {
				return nil, ctx, fmt.Errorf("Cant render svg path svg value is required")
			}
			err := draw.SvgPath(svg)
			if err != nil {
				return nil, ctx, err
			}
		case "svg_scale_to":
			// Special draw command which scales the svg x and y so that
			// the start point aligns with the previous endpoint and the
			// endpoint aligns to 'to'
			svg, ok := attr.SvgString("svg")
			if !ok {
				return nil, ctx, fmt.Errorf("Cant render svg path svg value is required")
			}
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			reverse := attr.MustBool("reverse", false)
			err := dc.svgScaleTo(draw, ctx, to, svg, reverse)
			if err != nil {
				return nil, ctx, err
			}
		case "rel_svg_scale_to":
			svg, ok := attr.SvgString("svg")
			if !ok {
				return nil, ctx, fmt.Errorf("Cant render svg path svg value is required")
			}
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			reverse := attr.MustBool("reverse", false)
			err := dc.svgScaleTo(draw, ctx, draw.ToAbsPosition(to), svg, reverse)
			if err != nil {
				return nil, ctx, err
			}
		case "svg_connect_to":
			// Special draw command which scales the svg x and y so that
			// the start point aligns with the previous endpoint and the
			// endpoint aligns to 'to'
			svg, ok := attr.SvgString("svg")
			if !ok {
				return nil, ctx, fmt.Errorf("Cant render svg path svg value is required")
			}
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			reverse := attr.MustBool("reverse", false)

			svgFrom := attr.MustPoint("svg_from", path.NewPoint(0, 0))
			svgTo := attr.MustPoint("svg_to", path.NewPoint(0, 0))
			err := dc.svgRotateScaleTo(draw, ctx, to, svg, reverse, svgFrom, svgTo)
			if err != nil {
				return nil, ctx, err
			}
		case "rel_svg_connect_to":
			svg, ok := attr.SvgString("svg")
			if !ok {
				return nil, ctx, fmt.Errorf("Cant render svg path svg value is required")
			}
			to, found := attr.Point("to")
			if !found {
				return nil, ctx, fmt.Errorf("%s requires param %s", command, "to")
			}
			reverse := attr.MustBool("reverse", false)

			// we dont need to translate these points, because they are within the coordinate
			// system of the svg, not the path
			svgFrom := attr.MustPoint("svg_from", path.NewPoint(0, 0))
			svgTo := attr.MustPoint("svg_to", path.NewPoint(0, 0))

			err := dc.svgRotateScaleTo(draw, ctx, draw.ToAbsPosition(to), svg, reverse, svgFrom, svgTo)
			if err != nil {
				return nil, ctx, err
			}
		default:
			return nil, ctx, fmt.Errorf("Cant render command %s", command)
		}
	}
	pth := draw.Path()
	context := ctx.Clone()
	context.Cursor = path.PathCursor(pth)
	return dc.HandleTransforms(dc, pth, ctx)
}
