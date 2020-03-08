package parser

import (
	"io/ioutil"
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
