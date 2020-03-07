package dom

import (
	"archive/zip"
	"fmt"
)

// a collection of documents
// as you add items this will create more documents
// potentially splitting items when they wont fit.
type PlanSet struct {
	// the underlying document.
	// this should contain all the parts are settings needed
	// to render
	doc     *Document
	svgDocs []*SVGDocument
}

func NewPlanSet(doc *Document) *PlanSet {
	return &PlanSet{
		doc: doc,
	}
}

// ONLY FOR TESTING
func (p *PlanSet) Document() *Document {
	return p.doc
}

func (p *PlanSet) createSvgDoc(ctx RenderContext) *SVGDocument {
	attr := p.doc.Attr()
	svgDoc := NewSVGDocument(
		attr.MustFloat64("material_width", 20),
		attr.MustFloat64("material_height", 12),
		MustUnits(attr.MustString("measurement_units", "in"), Inches),
	)

	svgDoc.SegmentOperators = AppContext().SegmentOperators()
	svgDoc.Padding = attr.MustFloat64(
		"doc_padding",
		.1,
	)
	return svgDoc
}

func (p *PlanSet) SVGDocuments() []*SVGDocument {
	return p.svgDocs
}

func (p *PlanSet) addPart(part *RenderedPart, ctx RenderContext) (bool, error) {
	// first try to add to all the existing open docs
	for _, svgDoc := range p.svgDocs {
		added, err := svgDoc.Add(part, ctx)
		if err != nil {
			return false, err
		}
		if added {
			return true, nil
		}
	}

	// add a new svgDoc and try to add it
	svgDoc := p.createSvgDoc(ctx)
	added, err := svgDoc.Add(part, ctx)
	if err != nil {
		return false, err
	}
	if added {
		p.svgDocs = append(p.svgDocs, svgDoc)
		return true, nil
	}
	return false, nil
}

func (p *PlanSet) Init(ctx RenderContext) error {
	// create the first svgdoc
	p.svgDocs = []*SVGDocument{
		p.createSvgDoc(ctx),
	}

	// render all the parts..
	// this is necessary in order to get the measurements
	for _, part := range p.doc.Parts {
		renderedParts, err := part.RenderPart(ctx)
		if err != nil {
			println(err.Error())
			return err
		}
		for _, renderedPart := range renderedParts {
			added, err := p.addPart(renderedPart, ctx)
			if err != nil {
				println(err.Error())
				return err
			}
			if !added {
				// too big.  try again?
				return fmt.Errorf("unable to add part, it is probably too big")
			}
		}
	}
	return nil
}

func positionForSplit(width, documentWidth, documentHeight float64) float64 {
	half := width / 2
	third := width / 3
	// either split into halves or thirds
	if half < documentWidth || half < documentHeight {
		return half
	} else if third < documentWidth || third < documentHeight {
		return third
	}
	return half
}

// renders all the documents into the passed in zip writer
func (ps *PlanSet) RenderZip(filename string, w *zip.Writer, ctx RenderContext) error {
	for i, svgDoc := range ps.svgDocs {
		f, err := w.Create(fmt.Sprintf("%s.%d.svg", filename, i))
		if err != nil {
			return err
		}
		svgDoc.WriteSVG(ctx, f)
	}
	return nil
}
