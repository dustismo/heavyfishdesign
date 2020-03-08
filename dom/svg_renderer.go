package dom

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/dustismo/heavyfishdesign/binpacking"
	"github.com/dustismo/heavyfishdesign/path"
)

//  What to do when a rendered piece is too big for the
//  document?
type OversizeStrategy int32

const (
	// break the piece using the specified CutLine
	Split OversizeStrategy = 0
	// leave the document the same size but spill off the edge
	Spill OversizeStrategy = 1
)

// What is the strategy when all the pieces don't fit in the
// Document?
type RenderStrategy int32

const (
	// Render into multiple documents
	MultiDocument RenderStrategy = 0
	// Resize a single document
	// THis will scale the document in both directions
	ResizeDocument RenderStrategy = 1
)

type docRenderable struct {
	// position to render at
	position         path.Point
	renderedPart     *RenderedPart
	rotate           bool // should we rotate by 90deg (used for layout)
	segmentOperators path.SegmentOperators
}

// a standard document
type SVGDocument struct {
	Name string

	// document width
	Width float64

	// document height
	Height float64

	Units Units

	// how much padding between things in the document?
	Padding float64

	// render the size into the document?
	// most cases this should be true,
	// false is useful for resizable in a browser window
	RenderSize bool

	// how many decimal places to render
	Precision int

	// this is for laying out the page
	layoutContainer *binpacking.Container

	renderables      []*docRenderable
	SegmentOperators path.SegmentOperators

	CutStyle   string
	LabelStyle string
}

func (dr *docRenderable) GetWidth() float64 {
	return dr.renderedPart.Width
}

func (dr *docRenderable) GetHeight() float64 {
	return dr.renderedPart.Height
}

func (dr *docRenderable) render(d *SVGDocument, ctx RenderContext, writer io.Writer) error {
	transforms := []string{}

	translateX := dr.position.X
	translateY := dr.position.Y

	if dr.rotate {
		// rotate 90, then shift to the right by width total
		translateX = dr.position.X + dr.GetHeight()
		transforms = append(transforms, fmt.Sprintf("rotate(90 %.3f %.3f)", translateX, translateY))
	}
	// move to the correct location
	transforms = append(transforms, fmt.Sprintf("translate(%.3f %.3f)", translateX, translateY))

	d.writeSVG(writer, fmt.Sprintf("<g transform=\"%s\">", strings.Join(transforms, " ")))

	// svgItem is guarenteed to be available

	pth := dr.renderedPart.Path

	svg := fmt.Sprintf("<path id=\"%s\" d=\"%s\" style=\"%s\" />",
		dr.renderedPart.Part.Id(),
		path.SvgString(pth, d.Precision),
		d.CutStyle)
	d.writeSVG(writer, svg)

	// Now render the label
	if len(dr.renderedPart.Label.Text) > 0 {
		// position and render
		textPos, err := path.PointPathAttribute(
			dr.renderedPart.Label.Position,
			pth,
			dr.segmentOperators)
		if err != nil {
			return err
		}
		labelSvg := fmt.Sprintf("<text x=\"%.3f\" y=\"%.3f\" style=\"%s\">%s</text>",
			textPos.X, textPos.Y,
			d.LabelStyle,
			dr.renderedPart.Label.Text)
		d.writeSVG(writer, labelSvg)
	}

	d.writeSVG(writer, "</g>")

	return nil
}

// Creates a new document
// defaults to settings for .2" Lowes style plywood
func NewSVGDocument(w float64, h float64, unit Units) *SVGDocument {
	// we want accuracy to 3 decimals
	// for reference the glowforge is supposed to be .025mm accurate

	// styles := svg.Attributes{
	// 	CutStyle:        fmt.Sprintf("fill:none;stroke:black;stroke-width:%.3f", unit.FromMM(.3)),
	// 	EngraveStyle:    fmt.Sprintf("fill:none;stroke:blue;stroke-width:%.3f", unit.FromMM(.3)),
	// 	ScoreStyle:      fmt.Sprintf("fill:none;stroke:grey;stroke-width:%.3f", unit.FromMM(.3)),
	// 	LabelTextStyle:  fmt.Sprintf("font: italic %.3fpt serif; fill: blue", unit.FromMM(5)),
	// 	LabelShapeStyle: fmt.Sprintf("fill:none;stroke:blue;stroke-width:%.3f", unit.FromMM(.3)),
	// }

	name := "laser_design.svg"
	d := &SVGDocument{
		Name:            name,
		Width:           w,
		Height:          h,
		Units:           unit,
		Padding:         unit.FromInch(.1),
		RenderSize:      true,
		layoutContainer: binpacking.NewContainer(0, 0, w, h),
		Precision:       3,
		CutStyle:        fmt.Sprintf("fill:none;stroke:black;stroke-width:%.3f", unit.FromMM(.3)),
		LabelStyle:      fmt.Sprintf("font: %.3fpt serif; fill: blue", unit.FromMM(3)),
	}
	return d
}

// clones this document meta data, minus any items in it
// TODO: do we want to clone the renderables as well?
func (d *SVGDocument) Clone() *SVGDocument {
	return &SVGDocument{
		Width:            d.Width,
		Height:           d.Height,
		Units:            d.Units,
		Padding:          d.Padding,
		RenderSize:       d.RenderSize,
		layoutContainer:  binpacking.NewContainer(0, 0, d.Width, d.Height),
		SegmentOperators: d.SegmentOperators,
		Precision:        d.Precision,
		CutStyle:         d.CutStyle,
	}
}

func (d *SVGDocument) start(writer io.Writer) {

	size := ""
	if d.RenderSize {
		size = fmt.Sprintf(`width="%.3f%s" height="%.3f%s"`, d.Width, d.Units.Abv, d.Height, d.Units.Abv)
	}

	svg := `<?xml version="1.0"?>
	<!-- Generated by github.com/dustismo/heavyfishdesign -->
	<svg %s viewBox="0.000 0.000 %.3f %.3f"
    	xmlns="http://www.w3.org/2000/svg"
		xmlns:xlink="http://www.w3.org/1999/xlink">
	`
	fmt.Fprintf(writer, svg, size, d.Width, d.Height)
}

func (d *SVGDocument) end(writer io.Writer) {
	svg := "</svg>\n"
	fmt.Fprintf(writer, svg)
}

// writes the whole svg document
func (d *SVGDocument) WriteSVG(ctx RenderContext, writer io.Writer) {
	d.start(writer)
	for _, r := range d.renderables {
		r.render(d, ctx, writer)
	}
	d.end(writer)
}

// adds a renderable creator into this document.  Returns
// true if it was able to fit, false otherwise.
func (d *SVGDocument) Add(p *RenderedPart, ctx RenderContext) (bool, error) {
	w := p.Width
	h := p.Height
	if h <= 0.0 || w <= 0.0 {
		// plenty of space for a nothing..
		return false, fmt.Errorf("Error, height or width is 0 or less")
	}

	if math.IsNaN(w) || math.IsNaN(h) {
		return false, fmt.Errorf("Error, cannot add part because width or height is NaN")
	}

	r := &docRenderable{
		renderedPart:     p,
		rotate:           false,
		segmentOperators: d.SegmentOperators,
	}
	inserted, bin := d.layoutContainer.InsertWithPadding(r, r, d.Padding)
	if !inserted && d.layoutContainer.IsEmpty() {
		// THis is kinda hacky, but here if the item doesn't fit we add
		// a new container with the oversized part in on its own.
		// eventually we should do this better, and probably flag this as being
		// oversized somehow.
		fmt.Printf("Warning: part %s did not fit, but adding to a new document anyway\n", p.Part.Id())
		d.layoutContainer = binpacking.NewSingleObjectContainer(
			r,
			d.layoutContainer.X, d.layoutContainer.Y,
			d.layoutContainer.Width, d.layoutContainer.Height)
		bin = *d.layoutContainer.Root
	} else if !inserted {
		return false, nil
	}
	r = &docRenderable{
		renderedPart:     p,
		position:         path.NewPoint(bin.X, bin.Y),
		rotate:           bin.Rotated,
		segmentOperators: d.SegmentOperators,
	}

	d.renderables = append(d.renderables, r)
	return true, nil
}

func (d *SVGDocument) writeSVG(writer io.Writer, svg string) {
	fmt.Fprintf(writer, "%s\n", svg)
}

func (d *SVGDocument) Parts() []*Part {
	p := []*Part{}
	for _, dr := range d.renderables {
		p = append(p, dr.renderedPart.Part)
	}
	return p
}
