package parser

import (
	"io/ioutil"
	"math"
	"testing"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/util"
)

func TestSimpleEval(t *testing.T) {
	InitContext()

	rc := dom.RenderContext{}
	json :=
		`
	{
		"params": {
			"box_size": 10
		},
		"parts": [
			{
				"params" : {
					
				},
				"components": [
					{
						"type": "draw",
						"id" : "id_test_1",
						"commands" : [
							{
								"command" : "line",
								"to" : {"x": "box_size + 5","y":"box_size"}
							}
						]
					}
				]
			}
		]
	}
	`
	dm, err := dynmap.ParseJSON(json)
	if err != nil {
		t.Errorf("%s", err)
	}

	doc, err := dom.ParseDocument(dm, util.NewLog())
	if err != nil {
		t.Errorf("%s", err)
	}
	expected := "M 0.000 0.000 L 15.000 10.000"
	PartRenderEquals(doc.Parts[0], rc, expected, t)
}

func TestParser(t *testing.T) {
	InitContext()

	rc := dom.RenderContext{}
	json :=
		`
	{
		"params": {
			"box_size": 10
		},
		"parts": [
			{
				"params" : {
					
				},
				"components": [
					{
						"type": "draw",
						"commands" : [
							{
								"command" : "line",
								"to" : {"x": "box_size + 5 / 2","y":0}
							},
							{
								"command" : "line",
								"to" : {"x": "box_size","y":"box_size"}
							},
							{
								"command" : "line",
								"to" : {"x": 0,"y":"box_size"}
							},
							{
								"command" : "line",
								"to" : {"x": 0,"y":0}
							}
						]
					}
				]
			}
		]
	}
	`
	dm, err := dynmap.ParseJSON(json)
	if err != nil {
		t.Errorf("%s", err)
	}

	doc, err := dom.ParseDocument(dm, util.NewLog())
	if err != nil {
		t.Errorf("%s", err)
	}
	expected := "M 0.000 0.000 L 12.500 0.000 L 10.000 10.000 L 0.000 10.000 L 0.000 0.000"
	PartRenderEquals(doc.Parts[0], rc, expected, t)
}

// Regression: scale {"width":…} must not pick up the part's "height" param as the scale transform's
// target height (Attr.lookup fell through to parent params), which caused non-uniform scaling.
func TestScaleWidthOnlyDoesNotUsePartHeightParam(t *testing.T) {
	InitContext()
	rc := dom.RenderContext{}
	json := `{
		"params": {},
		"parts": [{
			"params": { "height": 100 },
			"components": [{
				"type": "draw",
				"transforms": [
					{ "type": "scale", "width": "16" }
				],
				"commands": [
					{ "command": "move", "to": "0, 0" },
					{ "command": "circle", "radius": 5 }
				]
			}]
		}]
	}`
	dm, err := dynmap.ParseJSON(json)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := dom.ParseDocument(dm, util.NewLog())
	if err != nil {
		t.Fatal(err)
	}
	pth, _, err := doc.Parts[0].Render(rc)
	if err != nil {
		t.Fatal(err)
	}
	so := path.NewSegmentOperators()
	tl, br, err := path.BoundingBoxTrimWhitespace(pth, so)
	if err != nil {
		t.Fatal(err)
	}
	w, h := br.X-tl.X, br.Y-tl.Y
	if w <= 0 || h <= 0 {
		t.Fatalf("bbox w=%v h=%v", w, h)
	}
	if math.Abs(w/h-1) > 0.05 {
		t.Fatalf("width-only scale with part param height should stay uniform; bbox w=%v h=%v (ratio %v)", w, h, w/h)
	}
}

func PartRenderEquals(p *dom.Part, rc dom.RenderContext, expected string, t *testing.T) bool {
	r, _, _ := p.Render(rc)
	actual := path.SvgString(r, 3)

	if actual != expected {
		t.Errorf("expected: %s\nactual: %s\n", expected, actual)
		return false
	}
	return true
}

func LoadPlanSet(filename string) (*dom.PlanSet, dom.RenderContext, error) {
	InitContext()
	context := dom.RenderContext{
		Origin: path.NewPoint(0, 0),
		Cursor: path.NewPoint(0, 0),
	}
	b, err := ioutil.ReadFile(filename) // just pass the file name
	if err != nil {
		return nil, context, err
	}
	json := string(b) // convert content to a 'string'
	dm, err := dynmap.ParseJSON(json)
	if err != nil {
		return nil, context, err
	}

	doc, err := dom.ParseDocument(dm, util.NewLog())
	if err != nil {
		return nil, context, err
	}

	planset := dom.NewPlanSet(doc)

	err = planset.Init(context)
	if err != nil {
		return planset, context, err
	}
	return planset, context, err
}
