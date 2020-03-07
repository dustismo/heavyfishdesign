package dom

import (
	"strings"

	"github.com/dustismo/heavyfishdesign/dynmap"
)

type DocumentContext struct {
	customComponents map[string]Component
	Params           *dynmap.DynMap
}

// Creates a custom component if possible
// returns nil, false, nil if there is no custom component available with that
// type
func (dc *DocumentContext) CreateCustomComponent(componentType string, dm *dynmap.DynMap) (Component, bool, error) {
	c, ok := dc.customComponents[componentType]
	if !ok {
		return nil, ok, nil
	}

	replaced := dm.Clone()
	replaced.RemoveAll("type") // we need the type from the referenced component

	// grab the replaced transforms and append to the end of the transforms
	// from the referenced component
	transforms := replaced.MustDynMapSlice("transforms", []*dynmap.DynMap{})
	transforms = append(c.Transforms(), transforms...)
	// get the map of the custom component, then merge in our local changes.
	newComponent := c.ToDynMap().Clone()
	newComponent.RemoveAll("custom_component")
	// change to a group component if this is a part,
	// this is because parts will automatically strip whitespace and move to 0,0
	if newComponent.MustString("type", "part") == "part" {
		newComponent.Put("type", "group")
		// need to remove the default part index
		newComponent.Remove("params.index")
	}

	newComponent.Merge(replaced)
	// set top level defaults
	newComponent.Put("transforms", transforms)

	c, err := AppContext().MakeComponent(newComponent, dc)
	return c, true, err
}

// returns the document context of the owning document. or nil
func FindDocumentContext(c Element) *DocumentContext {
	switch v := c.(type) {
	case *Document:
		return v.Context
	case *Part:
		return v.Document().Context
	case Component:
		return FindDocumentContext(v.Parent())
	case *BasicElement:
		if v.doc != nil {
			return v.doc.Context
		}
	}
	return nil

}

// top level document, has multiple parts
type Document struct {
	*BasicElement
	ElementsByID map[string]Element
	Parts        []*Part
	Context      *DocumentContext
}

// all the elements that this document contains.
func (d *Document) AllElements() []Element {
	elems := make([]Element, 0, len(d.ElementsByID))
	for _, v := range d.ElementsByID {
		elems = append(elems, v)
	}
	return elems
}

func (d *Document) FindElementByID(id string) (Element, bool) {
	id = strings.Trim(id, "@")
	e, ok := d.ElementsByID[id]
	return e, ok
}

func (d *Document) FindPartByID(id string) (*Part, bool) {
	id = strings.Trim(id, "@")
	for _, p := range d.Parts {
		if p.Id() == id {
			return p, true
		}
	}
	return nil, false
}
