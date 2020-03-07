package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/transforms"
	"github.com/dustismo/heavyfishdesign/util"

	"github.com/dustismo/heavyfishdesign/path"
)

type SVGParser struct {
}

// parses an SVG file for any Path Elements and joining them together
// note this will ignore any transforms.
func (s SVGParser) ParseSVG(xml string, logger *util.HfdLog) (path.Path, error) {
	element, err := Parse(strings.NewReader(xml), true)
	p, err := ElementToPath(element, util.NewLog())
	if err != nil {
		return p, err
	}
	return p, nil
}

// Most of the parsing code below was originally from:
// https://github.com/JoshVarga/svgparser

// Element is a representation of an SVG element.
type Element struct {
	Name       string
	Attributes *dynmap.DynMap // map[string]string
	Children   []*Element
	Content    string
}

// convert this element to a path
func ElementToPath(elem *Element, log *util.HfdLog) (path.Path, error) {
	// see https://developer.mozilla.org/en-US/docs/Web/SVG/Tutorial/Basic_Shapes
	switch elem.Name {
	case "path":
		// <path d="M1927.2,1663.2C1927.2Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
		return path.ParsePathFromSvg(elem.Attributes.MustString("d", ""))
	case "circle":
		// <circle id="screw_hole" cx="88.8" cy="88.8" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
		draw := path.NewDraw()
		cx, ok := elem.Attributes.GetFloat64("cx")
		if !ok {
			return nil, fmt.Errorf("Circle must have 'cx' property")
		}

		cy, ok := elem.Attributes.GetFloat64("cy")
		if !ok {
			return nil, fmt.Errorf("Circle must have 'cy' property")
		}
		r, ok := elem.Attributes.GetFloat64("r")
		if !ok {
			return nil, fmt.Errorf("Circle must have 'r' property")
		}
		draw.MoveTo(path.NewPoint(
			cx-r,
			cy-r,
		))
		draw.Circle(r)
		return draw.Path(), nil
	case "rect":
		// <rect id="base_outer" x="60" y="60" width="1896" height="1896" style="fill:none;stroke:black;stroke-width:1.33px;"/>
		draw := path.NewDraw()
		x, ok := elem.Attributes.GetFloat64("x")
		if !ok {
			return nil, fmt.Errorf("Rect must have 'x' property")
		}

		y, ok := elem.Attributes.GetFloat64("y")
		if !ok {
			return nil, fmt.Errorf("Rect must have 'y' property")
		}
		w, ok := elem.Attributes.GetFloat64("width")
		if !ok {
			return nil, fmt.Errorf("Rect must have 'width' property")
		}
		h, ok := elem.Attributes.GetFloat64("height")
		if !ok {
			return nil, fmt.Errorf("Rect must have 'height' property")
		}
		draw.MoveTo(path.NewPoint(
			x,
			y,
		))
		draw.Rect(w, h)
		return draw.Path(), nil

	case "elipse":
		// <ellipse cx="75" cy="75" rx="20" ry="5" stroke="red" fill="transparent" stroke-width="5"/>
		return nil, fmt.Errorf("error, 'elipse' is not supported in svg parsing. yet...")
	case "polyline":
		// <polyline points="60 110 65 120 70 115 75 130 80 125 85 140 90 135 95 150 100 145" stroke="orange" fill="transparent" stroke-width="5"/>
		return nil, fmt.Errorf("error, 'polyline' is not supported in svg parsing. yet...")

	case "line":
		// <line x1="10" x2="50" y1="110" y2="150"/>
		draw := path.NewDraw()
		x1, ok := elem.Attributes.GetFloat64("x1")
		if !ok {
			return nil, fmt.Errorf("Line must have 'x1' property")
		}

		y1, ok := elem.Attributes.GetFloat64("y1")
		if !ok {
			return nil, fmt.Errorf("Line must have 'y1' property")
		}
		x2, ok := elem.Attributes.GetFloat64("x2")
		if !ok {
			return nil, fmt.Errorf("Line must have 'x2' property")
		}

		y2, ok := elem.Attributes.GetFloat64("y2")
		if !ok {
			return nil, fmt.Errorf("Line must have 'y2' property")
		}
		draw.MoveTo(path.NewPoint(
			x1,
			y1,
		))
		draw.LineTo(path.NewPoint(x2, y2))
		return draw.Path(), nil
	case "polygon":
		// <polygon points="50, 160 55, 180 70, 180 60, 190 65, 205 50, 195 35, 205 40, 190 30, 180 45, 180"/>
		return nil, fmt.Errorf("error, 'polygon' is not supported in svg parsing. yet...")

	default:
		// process the children
		paths := []path.Path{}
		for _, c := range elem.Children {
			p, err := ElementToPath(c, log)
			if err != nil {
				return p, err
			}
			paths = append(paths, p)
		}
		// join the paths..
		return transforms.SimpleJoin{}.JoinPaths(paths...), nil
	}
}

// NewElement creates element from decoder token.
func NewElement(token xml.StartElement) *Element {
	element := &Element{}
	attributes := dynmap.New()
	for _, attr := range token.Attr {
		attributes.Put(attr.Name.Local, attr.Value)
	}
	element.Name = token.Name.Local
	element.Attributes = attributes
	return element
}

// Compare compares two elements.
func (e *Element) Compare(o *Element) bool {
	if e.Name != o.Name || e.Content != o.Content ||
		e.Attributes.Length() != o.Attributes.Length() ||
		len(e.Children) != len(o.Children) {
		return false
	}

	for k, v := range e.Attributes.Map {
		v1, _ := o.Attributes.Get(k)
		if v != v1 {
			return false
		}
	}

	for i, child := range e.Children {
		if !child.Compare(o.Children[i]) {
			return false
		}
	}
	return true
}

// DecodeFirst creates the first element from the decoder.
func DecodeFirst(decoder *xml.Decoder) (*Element, error) {
	for {
		token, err := decoder.Token()
		if token == nil && err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		switch element := token.(type) {
		case xml.StartElement:
			return NewElement(element), nil
		}
	}
	return &Element{}, nil
}

// Decode decodes the child elements of element.
func (e *Element) Decode(decoder *xml.Decoder) error {
	for {
		token, err := decoder.Token()
		if token == nil && err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch element := token.(type) {
		case xml.StartElement:
			nextElement := NewElement(element)
			err := nextElement.Decode(decoder)
			if err != nil {
				return err
			}

			e.Children = append(e.Children, nextElement)

		case xml.CharData:
			data := strings.TrimSpace(string(element))
			if data != "" {
				e.Content = string(element)
			}

		case xml.EndElement:
			if element.Name.Local == e.Name {
				return nil
			}
		}
	}
	return nil
}

// Parse creates an Element instance from an SVG input.
func Parse(source io.Reader, validate bool) (*Element, error) {
	raw, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	// decoder.CharsetReader = charset.NewReaderLabel
	element, err := DecodeFirst(decoder)
	if err != nil {
		return nil, err
	}
	if err := element.Decode(decoder); err != nil && err != io.EOF {
		return nil, err
	}
	return element, nil
}

// FindID finds the first child with the specified ID.
func (e *Element) FindID(id string) *Element {
	for _, child := range e.Children {
		if childID, ok := child.Attributes.GetString("id"); ok && childID == id {
			return child
		}
		if element := child.FindID(id); element != nil {
			return element
		}
	}
	return nil
}

// FindAll finds all children with the given name.
func (e *Element) FindAll(name string) []*Element {
	var elements []*Element
	for _, child := range e.Children {
		if child.Name == name {
			elements = append(elements, child)
		}
		elements = append(elements, child.FindAll(name)...)
	}
	return elements
}
