package parser

import (
	"io/ioutil"
	"testing"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/util"
)

func TestEdge1(t *testing.T) {
	planset, context, err := LoadPlanSet("testdata/edge_test_1.hfd")
	if err != nil {
		t.Errorf("%s", err)
	}
	svgDocs := planset.SVGDocuments()
	part := svgDocs[0].Parts()[0]

	expected := "M 0.000 0.000 L 0.150 0.000 L 0.350 0.250 L 0.550 0.000 L 0.750 0.250 L 0.950 0.000 L 1.150 0.250 L 1.350 0.000 L 1.500 0.000"
	PartRenderEquals(part, context, expected, t)
}

func TestPartRepeat(t *testing.T) {
	planset, context, err := LoadPlanSet("testdata/part_repeat_test.hfd")
	if err != nil {
		t.Errorf("%s", err)
	}
	svgDocs := planset.SVGDocuments()
	parts := svgDocs[0].Parts()
	if len(parts) != 2 {
		t.Errorf("Expected 2 parts")
	}

	expected := "M 0.000 0.000 L 1.000 0.000 M 1.000 0.000 L 1.000 1.000 M 1.000 1.000 L 0.000 1.000 M 0.000 1.000 L 0.000 0.000"
	PartRenderEquals(parts[0], context, expected, t)

	expected = "M 0.000 0.000 L 2.000 0.000 M 2.000 0.000 L 2.000 2.000 M 2.000 2.000 L 0.000 2.000 M 0.000 2.000 L 0.000 0.000"
	PartRenderEquals(parts[1], context, expected, t)
}

//this tests importing a Part as a Component
func TestPartReference(t *testing.T) {
	planset, context, err := LoadPlanSet("testdata/part_reference_test.hfd")
	if err != nil {
		t.Errorf("%s", err)
	}
	svgDocs := planset.SVGDocuments()
	parts := svgDocs[0].Parts()
	expected := "M 0.000 0.000 L 5.000 0.000 M 5.000 0.000 L 5.000 5.000 M 5.000 5.000 L 0.000 5.000 M 0.000 5.000 L 0.000 0.000"
	PartRenderEquals(parts[0], context, expected, t)
}

func TestFingerJoint(t *testing.T) {
	planset, context, err := LoadPlanSet("testdata/finger_joint_import_test.hfd")
	if err != nil {
		t.Errorf("%s", err)
	}
	svgDocs := planset.SVGDocuments()
	parts := svgDocs[0].Parts()
	expected := "M 0.500 0.000 L 0.500 0.200 L 0.000 0.200 L 0.000 0.450 L 0.500 0.450 L 0.500 0.650 M 0.500 0.650 L 0.500 0.850 L 0.000 0.850 L 0.000 1.100 L 0.500 1.100 L 0.500 1.300 M 0.500 1.300 L 0.500 1.500 L 0.000 1.500 L 0.000 1.750 L 0.500 1.750 L 0.500 1.950 M 0.500 1.950 L 0.500 2.150 L 0.000 2.150 L 0.000 2.400 L 0.500 2.400 L 0.500 2.600 M 0.500 2.600 L 0.500 2.800 L 0.000 2.800 L 0.000 3.050 L 0.500 3.050 L 0.500 3.250 M 0.500 3.250 L 0.500 3.450 L 0.000 3.450 L 0.000 3.700 L 0.500 3.700 L 0.500 3.900 M 0.500 3.900 L 0.500 4.100 L 0.000 4.100 L 0.000 4.350 L 0.500 4.350 L 0.500 4.550 M 0.500 4.550 L 0.500 4.750 L 0.000 4.750 L 0.000 5.000 L 0.500 5.000 L 0.500 5.200 M 0.500 5.200 L 0.500 5.400 L 0.000 5.400 L 0.000 5.650 L 0.500 5.650 L 0.500 5.850 M 0.000 5.850"
	PartRenderEquals(parts[0], context, expected, t)
}

func TestBox(t *testing.T) {
	planset, context, err := LoadPlanSet("testdata/simple_square.hfd")
	if err != nil {
		t.Errorf("%s", err)
	}
	svgDocs := planset.SVGDocuments()
	parts := svgDocs[0].Parts()
	expected := "M 0.000 0.000 L 5.000 0.000 M 5.000 0.000 L 5.000 5.000 M 5.000 5.000 L 0.000 5.000 M 0.000 5.000 L 0.000 0.000"
	PartRenderEquals(parts[0], context, expected, t)
}

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

func TestReferenceComponent(t *testing.T) {
	InitContext()

	b, err := dom.AppContext().FileLoader().LoadBytes("testdata/finger_joint.hfd")
	if err != nil {
		t.Errorf("Error %s", err)
	}
	doc, err := dom.AppContext().DocumentParser().Parse(b, util.NewLog())
	if err != nil {
		t.Errorf("Error %s", err)
	}
	println(doc)
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
