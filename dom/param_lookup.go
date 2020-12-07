package dom

import (
	"fmt"
	"strings"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

// Attributes should be local to the component
// Referenced params will get looked up through the stack,
// but it is expected that the attributes are available at the component
// level.
type Attr struct {
	element Element
	mp      *dynmap.DynMap
}

var recursionMax = 12

func NewAttr(e Element, mp *dynmap.DynMap) *Attr {
	return &Attr{
		element: e,
		mp:      mp,
	}
}

func NewAttrElement(e Element) *Attr {
	return &Attr{
		element: e,
		mp:      e.ToDynMap(),
	}
}

// if this is a string, we will try to look up the string as a param
// boolean is whether this lookup is complete or if the value should be
func (b *Attr) realize(param string, v interface{}, recursion int) (interface{}, bool) {
	// test if this is a string that we should look up
	vStr, isString := v.(string)
	if isString && vStr != param {
		v2, ok := b.lookup(vStr, recursion+1)
		if ok {
			return v2, true
		} else {
			// return the first result
			return vStr, true
		}
	} else if !isString {
		// not a string, so just return it
		return v, true
	} else {
		// it is a string with the same name as the param name
		// pass, and lookup one level higher.
		return v, false
	}
}

func (b *Attr) lookup(param string, recursion int) (interface{}, bool) {
	if recursion > recursionMax {
		fmt.Printf("Lookup of param %s failed due to max recursion", param)
		return nil, false
	}
	// first look in the local map
	v, ok := b.mp.Get(param)

	if ok {
		v2, ok := b.realize(param, v, recursion)
		if ok {
			return v2, true
		}
	}
	// now look at the element parents
	v, ok = b.element.ParamLookerUpper().Lookup(param)
	if ok {
		return v, true
	}
	// didn't find it, now look at defaults
	v, ok = b.lookupDefault(b.element, param, recursion)
	if !ok {
		fmt.Printf("Unable to find param %s", param)
	}
	return v, ok
}

func (b *Attr) lookupDefault(elem Element, param string, recursion int) (interface{}, bool) {
	v, ok := elem.Defaults().Get(param)
	if ok {
		return b.realize(param, v, recursion)
	}
	com, ok := elem.(Component)
	if ok {
		return b.lookupDefault(com.Parent(), param, recursion+1)
	}
	return v, false
}

func (b *Attr) MustPoint(param string, def path.Point) path.Point {
	v, ok := b.Point(param)
	if !ok {
		return def
	}
	return v
}
func (b *Attr) MustPoint2(param string, from path.Point, def path.Point) path.Point {
	v, ok := b.Point2(param, from)
	if !ok {
		return def
	}
	return v
}

// Returns a Point either by x + y coords or by angle + length
func (b *Attr) Point2(param string, from path.Point) (path.Point, bool) {
	// try normal style point
	p, ok := b.Point(param)
	if ok {
		return p, ok
	}
	v, ok := b.lookup(param, 0)
	if !ok || !dynmap.IsDynMapConvertable(v) {
		return path.NewPoint(0, 0), false
	}

	angle, ok := b.Float64(fmt.Sprintf("%s.angle", param))
	if !ok {
		return path.NewPoint(0, 0), false
	}
	length, ok := b.Float64(fmt.Sprintf("%s.length", param))
	if !ok {
		return path.NewPoint(0, 0), false
	}
	return path.NewLineSegmentAngle(from, length, angle).End(), true
}

// attempt to coerce into a point
func (b *Attr) ToPoint(v interface{}) (path.Point, bool) {
	p, ok := v.(path.Point)
	if ok {
		return p, ok
	}
	// else try a dynmap
	mp, ok := dynmap.ToDynMap(v)
	if !ok {
		return path.NewPoint(0, 0), false
	}
	x, err := b.element.ParamLookerUpper().ToFloat64(mp.Must("x", ""))
	if err != nil {
		return path.NewPoint(0, 0), false
	}
	y, err := b.element.ParamLookerUpper().ToFloat64(mp.Must("y", ""))
	if err != nil {
		return path.NewPoint(0, 0), false
	}
	return path.NewPoint(x, y), true
}

func (b *Attr) Point(param string) (path.Point, bool) {
	if strings.HasPrefix(param, "$") {
		handle, err := path.ToPathAttr(param)
		if err != nil {
			fmt.Printf("%s is not a parsable Handle", param)
			return path.NewPoint(0, 0), false
		}
		if handle != path.Cursor {
			fmt.Printf("Error, %s is not an available handle in this context.\n", param)
			return path.NewPoint(0, 0), false
		}
		c, ok := b.element.(Component)
		if !ok {
			fmt.Printf("Error, %s is not an available handle for a non component.\n", param)
			return path.NewPoint(0, 0), false
		}
		ctx, ok := c.RenderContext()
		if !ok {
			fmt.Printf("Error, %s is not an available handle outside of a render.\n", param)
			return path.NewPoint(0, 0), false
		}
		return ctx.Cursor, true
	}

	v, ok := b.lookup(param, 0)
	if !ok {
		return path.NewPoint(0, 0), false
	}
	p, ok := b.ToPoint(v)
	if ok {
		return p, ok
	}

	if !dynmap.IsDynMapConvertable(v) {
		str, _ := b.String(param)
		// its a string, so lets try float,float
		vs := strings.Split(str, ",")
		if len(vs) == 1 && vs[0] != param {
			// try to evaluate this as a param
			return b.Point(vs[0])
		} else if len(vs) == 1 && vs[0] == param {
			// look one level up.
			// ?
		}

		if len(vs) != 2 {
			fmt.Printf("Error, %s is not a parsable point (%s) (%+v).  Point must be in the form x,y\n", str, param, v)
			return path.NewPoint(0, 0), false
		}
		x, err := b.element.ParamLookerUpper().ToFloat64(vs[0])
		if err != nil {
			fmt.Printf("Error, %s is not a parsable point: %s -- %s\n", param, vs[0], err.Error())
			return path.NewPoint(0, 0), false
		}

		y, err := b.element.ParamLookerUpper().ToFloat64(vs[1])
		if err != nil {
			fmt.Printf("Error, %s is not a parsable point: %s -- %s\n", param, vs[1], err.Error())
			return path.NewPoint(0, 0), false
		}
		return path.NewPoint(x, y), true
	}
	return path.NewPoint(0, 0), false
}

func (b *Attr) MustString(param string, def string) string {
	v, ok := b.String(param)
	if ok {
		return v
	}
	return def
}

func (b *Attr) MustHandle(param string, def path.PathAttr) path.PathAttr {
	v, ok := b.Handle(param)
	if ok {
		return v
	}
	return def
}

func (b *Attr) Handle(param string) (path.PathAttr, bool) {
	handleStr, ok := b.String(param)
	if ok {
		h, err := path.ToPathAttr(handleStr)
		if err != nil {
			fmt.Printf("%s is not a parsable Handle", handleStr)
			return path.TopLeft, false
		}
		return h, true
	}
	return path.TopLeft, false
}

func (b *Attr) SvgString(param string) (string, bool) {
	return b.String(param)
}

func (b *Attr) String(param string) (string, bool) {
	vTmp, ok := b.lookup(param, 0)
	v := dynmap.ToString(vTmp)
	if ok && !strings.ContainsAny(v, " ,$") {
		// check if this is a parameter?
		s, ok := b.element.ParamLookerUpper().String(v)
		if ok {
			return s, true
		}
	}
	return v, ok
}

func (b *Attr) Bool(param string) (bool, bool) {
	bl, ok := b.String(param)
	if !ok {
		return false, ok
	}
	v, err := dynmap.ToBool(bl)
	if err != nil {
		println(err)
		return false, false
	}
	return v, true
}

func (b *Attr) MustBool(param string, def bool) bool {
	bl, ok := b.Bool(param)
	if ok {
		return bl
	}
	return def
}

func (b *Attr) MustInt(param string, def int) int {
	i, ok := b.Int(param)
	if !ok {
		return def
	}
	return i
}

func (b *Attr) Int(param string) (int, bool) {
	f, ok := b.Float64(param)
	if !ok {
		return 0, ok
	}
	return int(f), ok
}

func (b *Attr) MustFloat64(param string, def float64) float64 {
	v, ok := b.Float64(param)
	if !ok {
		return def
	}
	return v
}
func (b *Attr) Float64(param string) (float64, bool) {
	v, ok := b.lookup(param, 0)
	if !ok {
		return 0, false
	}
	f, err := b.element.ParamLookerUpper().ToFloat64(v)
	if err != nil {
		fmt.Printf("Error looking up %s, got value %s, error: %s\n", param, v, err.Error())
		return 0, false
	}
	return f, true
}

// a lookerupper based on a single dynmap.  mostly useful for tests
type DynMapParamLookerUpper struct {
	*dynmap.DynMap
}

func (p *DynMapParamLookerUpper) Lookup(param string) (interface{}, bool) {
	return p.Get(param)
}
func (p *DynMapParamLookerUpper) ToFloat64(value interface{}) (float64, error) {
	return dynmap.ToFloat64(value)
}
func (p *DynMapParamLookerUpper) String(param string) (string, bool) {
	return p.GetString(param)
}
func (p *DynMapParamLookerUpper) Float64(param string) (float64, bool) {
	return p.GetFloat64(param)
}

type BasicParamLookerUpper struct {
	element Element
}

// does a local lookup, without traversing parents
func (p *BasicParamLookerUpper) localLookup(param string, recursion int) (interface{}, bool) {
	if recursion > maxRecursion {
		fmt.Printf("Error hit max recursion for %s\n", param)
		return nil, false
	}
	// look in the params
	v, found := p.element.Params().Get(param)
	// look in original map..
	if !found {
		v, found = p.element.ToDynMap().Get(param)
	}

	if !found {
		// check if this is a Document. if so look in the document context
		switch c := p.element.(type) {
		case *Document:
			v, found = c.Context.Params.Get(param)
		case *BasicElement:
			if c.doc != nil {
				v, found = c.doc.Context.Params.Get(param)
			}
		}
	}

	if found {
		// check if the params map contains additional meta data.
		// typically this should only apply to the Document
		if dynmap.IsDynMapConvertable(v) {
			// value, description, data_type
			v1, _ := dynmap.ToDynMap(v)
			if v1.Contains("value") {
				v, found = v1.Get("value")
			}
		}
	}

	if !found {
		if recursion > 0 {
			// if we recured, means this was a lookup on a value, so
			// the param is the previous value and we should return that
			return param, true
		}
		return v, found
	}

	// now check if v is another parameter
	vStr, ok := v.(string)
	if !ok {
		// not a string, so definitely not a param
		return v, found
	}
	return p.localLookup(vStr, recursion+1)
}
func (p *BasicParamLookerUpper) Lookup(param string) (interface{}, bool) {
	v, found := p.localLookup(param, 0)

	// now try parents if possible
	component, isComponent := p.element.(Component)

	if !found && isComponent {
		return component.Parent().ParamLookerUpper().Lookup(param)
	}
	vStr, isString := v.(string)
	if isString && isComponent {
		// try looking this up as a param
		v1, found1 := component.Parent().ParamLookerUpper().Lookup(vStr)
		if found1 {
			return v1, found1
		}
	}
	return v, found
}

func (p *BasicParamLookerUpper) MustString(param string, def string) string {
	v, ok := p.String(param)
	if !ok {
		return def
	}
	return v
}
func (p *BasicParamLookerUpper) String(param string) (string, bool) {
	o, ok := p.Lookup(param)
	if !ok {
		return "", false
	}
	return dynmap.ToString(o), true
}
func (p *BasicParamLookerUpper) MustFloat64(param string, def float64) float64 {
	v, ok := p.Float64(param)
	if !ok {
		return def
	}
	return v
}

// converts the given object
func (p *BasicParamLookerUpper) ToFloat64(obj interface{}) (float64, error) {
	switch vt := obj.(type) {
	case string:
		// it's a string, so see if the value is an expression
		variable := strings.Replace(vt, "$", "", -1)
		v, err := EvalExpression(variable, p.element)
		if err != nil {
			return 0, err
		}
		return dynmap.ToFloat64(v)
	}

	return dynmap.ToFloat64(obj)
}

func (p *BasicParamLookerUpper) Float64(param string) (float64, bool) {
	o, ok := p.Lookup(param)
	if !ok {
		return 0, false
	}
	v, err := p.ToFloat64(o)
	if err != nil {
		fmt.Printf("Error in float conversion for %s :: %s\n", param, err.Error())
		return v, false
	}
	return v, true
}

func FindElementByID(id string, element Element) (Element, error) {
	switch elem := element.(type) {
	case *Document:
		v, ok := elem.FindElementByID(id)
		if !ok {
			return v, fmt.Errorf("Unable to find element %s", id)
		}
		return v, nil
	case Component:
		return FindElementByID(id, elem.Parent())
	default:
		if element.Id() == id {
			return element, nil
		}
		return elem, fmt.Errorf("Unable to find element %s", id)
	}
}
