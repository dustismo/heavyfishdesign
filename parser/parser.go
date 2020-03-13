package parser

import (
	"io/ioutil"

	"github.com/dustismo/heavyfishdesign/components"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/util"

	"github.com/dustismo/heavyfishdesign/dom"
)

type DefaultDocumentParser struct {
}

func InitContext() {
	cf := []dom.ComponentFactory{
		dom.PartFactory{},
		dom.DrawComponentFactory{},
		dom.RepeatComponentFactory{},
		dom.GroupComponentFactory{},
		components.EdgeComponentFactory{},
		components.BasicEdgeComponentFactory{},
		components.XInterceptComponentFactory{},
	}
	tf := []dom.TransformFactory{
		dom.CleanupTransformFactory{},
		dom.RotateTransformFactory{},
		dom.ReverseTransformFactory{},
		dom.JoinTransformFactory{},
		dom.OffsetTransformFactory{},
		dom.MoveTransformFactory{},
		dom.TrimTransformFactory{},
		dom.MirrorTransformFactory{},
		dom.ScaleTransformFactory{},
		dom.SliceTransformFactory{},
		dom.RotateScaleTransformFactory{},
	}

	pf := []dom.PartTransformerFactory{
		dom.PartSplitterTransformerFactory{},
		dom.PartLatheTransformerFactory{},
	}
	docParser := NewDocumentParser()
	dom.AppContext().Init(
		cf, tf, pf,
		path.NewSegmentOperators(),
		docParser,
		docParser,
		SVGParser{},
	)
}

func NewDocumentParser() DefaultDocumentParser {
	return DefaultDocumentParser{}
}

func (p DefaultDocumentParser) Parse(bytes []byte, logger *util.HfdLog) (*dom.Document, error) {
	json := string(bytes) // convert content to a 'string'
	return dom.ParseDocumentFromJson(json, logger)
}

func (p DefaultDocumentParser) LoadBytes(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}
