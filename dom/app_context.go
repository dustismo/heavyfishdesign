package dom

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"

	"github.com/dustismo/heavyfishdesign/util"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

type DocumentParser interface {
	Parse(bytes []byte, logger *util.HfdLog) (*Document, error)
}

type FileLoader interface {
	LoadBytes(filename string) ([]byte, error)
}

type SVGParser interface {
	ParseSVG(svg string, logger *util.HfdLog) (path.Path, error)
}

type ComponentFactory interface {
	CreateComponent(componentType string, dm *dynmap.DynMap, dc *DocumentContext) (Component, error)
	// The list of component types this Factory should be used for
	ComponentTypes() []string
}

type TransformFactory interface {
	CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error)
	// The list of component types this Factory should be used for
	TransformTypes() []string
}

type PartTransformerFactory interface {
	CreateTransformer(transformType string, dm *dynmap.DynMap, part *Part) (PartTransformer, error)
	// The list of component types this Factory should be used for
	TransformerTypes() []string
}

type Factories struct {
	componentFactories       map[string]ComponentFactory
	transformFactories       map[string]TransformFactory
	partTransformerFactories map[string]PartTransformerFactory
	segmentOperators         path.SegmentOperators
	documentParser           DocumentParser
	fileLoader               FileLoader
	precision                int
	svgParser                SVGParser
}

var appContext *Factories

// Constructs a Factories instance with the components and transforms we
// know about..
func AppContext() *Factories {
	if appContext == nil {
		segmentOperators := path.NewSegmentOperators()

		appContext = &Factories{
			segmentOperators: segmentOperators,
			precision:        path.DefaultPrecision,
		}
	}

	return appContext
}

func (c *Factories) SetFileLoader(fl FileLoader) {
	c.fileLoader = fl
}

func (c *Factories) Init(componentFactories []ComponentFactory,
	transformFactories []TransformFactory,
	partTransformerFactories []PartTransformerFactory,
	segOps path.SegmentOperators,
	documentParser DocumentParser,
	fileLoader FileLoader,
	svgParser SVGParser,
) {
	c.componentFactories = nil
	for _, cf := range componentFactories {
		c.AddComponentFactory(cf)
	}
	c.transformFactories = nil
	for _, t := range transformFactories {
		c.AddTransformFactory(t)
	}
	c.partTransformerFactories = nil
	for _, pt := range partTransformerFactories {
		c.AddPartTransformerFactory(pt)
	}
	c.segmentOperators = segOps
	c.documentParser = documentParser
	c.fileLoader = fileLoader
	c.svgParser = svgParser
}

func (c *Factories) AddTransformFactory(tf TransformFactory) {
	if c.transformFactories == nil {
		c.transformFactories = make(map[string]TransformFactory)
	}
	for _, k := range tf.TransformTypes() {
		c.transformFactories[k] = tf
	}
}

func (c *Factories) AddPartTransformerFactory(tf PartTransformerFactory) {
	if c.partTransformerFactories == nil {
		c.partTransformerFactories = make(map[string]PartTransformerFactory)
	}
	for _, k := range tf.TransformerTypes() {
		c.partTransformerFactories[k] = tf
	}
}

func (c *Factories) AddComponentFactory(cf ComponentFactory) {
	if c.componentFactories == nil {
		c.componentFactories = make(map[string]ComponentFactory)
	}
	for _, k := range cf.ComponentTypes() {
		c.componentFactories[k] = cf
	}
}
func (c *Factories) Precision() int {
	return c.precision
}
func (c *Factories) SegmentOperators() path.SegmentOperators {
	return c.segmentOperators
}

func (c *Factories) FileLoader() FileLoader {
	return c.fileLoader
}

func (c *Factories) DocumentParser() DocumentParser {
	return c.documentParser
}

func (c *Factories) ParseSVG(svg string, logger *util.HfdLog) (path.Path, error) {
	return c.svgParser.ParseSVG(svg, logger)
}

// Makes a component from the DynMap,
// there must be a field called "type"
func (c *Factories) MakeComponent(dm *dynmap.DynMap, dc *DocumentContext) (Component, error) {
	componentType, ok := dm.GetString("type")
	if !ok {
		return nil, fmt.Errorf("No component type in: %s", dm.ToJSON())
	}

	factory, ok := c.componentFactories[componentType]
	if !ok {
		// look for a custom component
		c, ok, err := dc.CreateCustomComponent(componentType, dm)
		if !ok {
			return nil, fmt.Errorf("Unable to find component of type %s", componentType)
		}
		return c, err
	}
	component, err := factory.CreateComponent(componentType, dm, dc)
	return component, err
}

func (c *Factories) MakeBasicElement(dm *dynmap.DynMap) *BasicElement {
	return &BasicElement{
		id:          dm.MustString("id", nextId(dm)),
		elementType: dm.MustString("type", ""),
		originalMap: dm,
		params:      dm.MustDynMap("params", dynmap.New()),
		defaults:    dm.MustDynMap("defaults", dynmap.New()),
	}
}

func (c *Factories) MakeBasicComponent(dm *dynmap.DynMap) *BasicComponent {
	be := c.MakeBasicElement(dm)
	return &BasicComponent{
		id:          be.id,
		elementType: be.elementType,
		originalMap: be.originalMap,
		params:      be.params,
		defaults:    be.defaults,
	}
}
func (c *Factories) MakePartTransformer(transformType string, dm *dynmap.DynMap, part *Part) (PartTransformer, error) {
	factory, ok := c.partTransformerFactories[transformType]
	if !ok {
		return nil, fmt.Errorf("Unable to parse part transformer of type %s", transformType)
	}
	transform, err := factory.CreateTransformer(transformType, dm, part)
	return transform, err
}

func (c *Factories) MakeTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	factory, ok := c.transformFactories[transformType]
	if !ok {
		return nil, fmt.Errorf("Unable to parse transform of type %s", transformType)
	}
	transform, err := factory.CreateTransform(transformType, dm, element)
	return transform, err
}

func (c *Factories) MakeComponents(dm []*dynmap.DynMap, dc *DocumentContext) ([]Component, error) {
	components := []Component{}
	for _, cdm := range dm {
		c, err := c.MakeComponent(cdm, dc)
		if err != nil {
			return nil, err
		}
		components = append(components, c)
	}
	return components, nil
}

// finds the transforms under the given key
func (c *Factories) MakeTransforms(dm []*dynmap.DynMap, element Element) ([]path.PathTransform, error) {
	transforms := []path.PathTransform{}
	for _, t := range dm {
		p, err := c.MakeTransform(t.MustString("type", "unknown"), t, element)
		if err != nil {
			return transforms, err
		}
		transforms = append(transforms, p)
	}
	return transforms, nil
}

func (c *Factories) MakePartTransformers(dm []*dynmap.DynMap, part *Part) ([]PartTransformer, error) {
	transforms := []PartTransformer{}
	for _, t := range dm {
		p, err := c.MakePartTransformer(t.MustString("type", "unknown"), t, part)
		if err != nil {
			return transforms, err
		}
		transforms = append(transforms, p)
	}
	return transforms, nil
}

// create an id based on the dynmap json rendering.
// this must be idempotent in order
// to keep rendering consistent
func nextId(dm *dynmap.DynMap) string {
	h := md5.New()
	str := strings.NewReplacer("\n", "", " ", "").Replace(dm.ToJSON())
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

func ParseDocumentFromPath(path string, logger *util.HfdLog) (*Document, error) {
	b, err := AppContext().FileLoader().LoadBytes(path)
	if err != nil {
		return nil, err
	}
	json := string(b) // convert content to a 'string'
	return ParseDocumentFromJson(json, logger)
}

// parses to an HFDMap.  This should be used instead of parse json directly as
// we may want to change the encoding in the future.
func ParseToHFDMap(raw string, logger *util.HfdLog) (*dynmap.DynMap, error) {
	dm, err := dynmap.ParseJSON(raw)
	return dm, err
}

func ParseDocumentFromJson(json string, logger *util.HfdLog) (*Document, error) {
	dm, err := ParseToHFDMap(json, logger)
	if err != nil {
		return nil, err
	}
	return ParseDocument(dm, logger)
}

func ParseDocument(dm *dynmap.DynMap, logger *util.HfdLog) (*Document, error) {
	// look for a filename.  if one exists, load it, then merge self into it
	refFilename := dm.MustString("filename", "")
	if len(refFilename) > 0 {
		b, err := AppContext().FileLoader().LoadBytes(refFilename)
		if err != nil {
			return nil, err
		}
		d1, err := dynmap.ParseJSON(string(b))
		if err != nil {
			return nil, err
		}
		dm.Remove("filename")
		// merge imports special like.
		for _, imp := range dm.MustDynMapSlice("imports", []*dynmap.DynMap{}) {
			d1.AddToSlice("imports", imp)
		}
		dm.Remove("imports")

		return ParseDocument(d1.Merge(dm), logger)
	}

	be := AppContext().MakeBasicElement(dm)

	// parse the imports
	dc := &DocumentContext{
		customComponents: map[string]Component{},
		Params:           dynmap.New(),
	}
	for _, importDm := range dm.MustDynMapSlice("imports", []*dynmap.DynMap{}) {
		pth := importDm.MustString("path", "")
		importType := importDm.MustString("type", "component")
		if importType == "component" {
			newDoc, err := ParseDocumentFromPath(pth, logger)
			if err != nil {
				return nil, fmt.Errorf("Error trying to import %s.  Error: %s", pth, err.Error())
			}
			for _, e := range newDoc.AllElements() {
				customType := e.ToDynMap().MustString("custom_component.type", "")
				if len(customType) > 0 {
					varName := importDm.MustString(fmt.Sprintf("alias.%s", customType), customType)
					dc.customComponents[varName] = e.(Component)
					// merge all the document params as the defaults of the
					// for the current element
					e.Defaults().Merge(newDoc.Params())
					// also merge in the params from the document context
					e.Defaults().Merge(newDoc.Context.Params)
				}
			}
		} else if importType == "svg" {
			svgBytes, err := AppContext().FileLoader().LoadBytes(pth)
			if err != nil {
				return nil, err
			}
			varName, ok := importDm.GetString("alias")
			if !ok {
				return nil, fmt.Errorf("Error, alias must be provided for svg import: %s", pth)
			}
			p, err := AppContext().ParseSVG(string(svgBytes), logger)
			if err != nil {
				return nil, err
			}
			s := path.SvgString(p, AppContext().Precision())
			dc.Params.Put(varName, s)
		}
	}

	parts := []*Part{}
	elementsByID := make(map[string]Element)
	for _, psDm := range dm.MustDynMapSlice("parts", []*dynmap.DynMap{}) {
		p, err := PartFactory{}.CreateComponent("part", psDm, dc)
		if err != nil {
			return nil, err
		}
		parts = append(parts, p.(*Part))
		// populate the id map
		populateElementsByID(p, elementsByID)
	}

	// reset the parts in the dynmap, so
	// any referenced components or mutations are
	// accurate
	newDm := dm.Clone()
	newDm.Remove("parts")
	for _, p := range parts {
		newDm.AddToSlice("parts", p.ToDynMap())
	}
	be.originalMap = newDm

	doc := &Document{
		BasicElement: be,
		Parts:        parts,
		ElementsByID: elementsByID,
		Context:      dc,
	}
	for _, p := range doc.Parts {
		p.SetParent(doc)
	}
	// set the pointer to itself
	be.doc = doc
	// fmt.Printf("DOCUMENT **** \n %s \n", doc.ToDynMap().ToJSON())

	return doc, nil
}

func populateElementsByID(elem Element, elements map[string]Element) error {

	elements[elem.Id()] = elem
	component, ok := elem.(Component)
	if ok {
		for _, e := range component.Children() {
			err := populateElementsByID(e, elements)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
