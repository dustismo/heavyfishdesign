package dom

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

type Element interface {
	// The id of this element.
	// note: this should be guarenteed to be set
	Id() string
	ElementType() string

	ToDynMap() *dynmap.DynMap

	// @Deprecated
	Params() *dynmap.DynMap

	SetLocalVariable(key string, val interface{})
	SetGlobalVariable(key string, val interface{})

	// Looks up values in the params map, if no value is found
	// this should recursively search parent elements until
	// it is found
	ParamLookerUpper() ParamLookerUpper
}

type Component interface {
	Element
	// Render the path.  This is typically called by the
	// owning element, which handles the transforms
	// Note that Render should not be ever considered threadsafe
	// and should never be called concurrently
	Render(ctx RenderContext) (path.Path, RenderContext, error)

	// If rendering is currently in progress, this should return the
	// current context
	RenderContext() (RenderContext, bool)

	Transforms() []*dynmap.DynMap
	Parent() Element
	SetParent(element Element)
	// measure this component
	// returns w, h
	// Note: this typical calls render, so care should be taken when calling this
	Measure() (float64, float64, error)
	Children() []Element
}

type ParamLookerUpper interface {
	Lookup(param string) (interface{}, bool)
	// attempt to convert the given object to a float.
	// if this is a number than that will be returned
	// if it is a string, the string will be evaluated
	ToFloat64(value interface{}) (float64, error)
	MustString(param string, def string) string
	String(param string) (string, bool)
	MustFloat64(param string, def float64) float64
	Float64(param string) (float64, bool)
}

type BasicElement struct {
	id          string
	elementType string
	originalMap *dynmap.DynMap // The map that this element is parsed from
	params      *dynmap.DynMap
	doc         *Document // this will only be set for the actual document element (pointer to itself)
}

func (b *BasicElement) Id() string {
	if len(b.id) == 0 {
		b.id = nextId(b.ToDynMap())
	}
	return b.id
}
func (b *BasicElement) ElementType() string {
	return b.elementType
}
func (b *BasicElement) ToDynMap() *dynmap.DynMap {
	return b.originalMap
}
func (b *BasicElement) Params() *dynmap.DynMap {
	if b.params == nil || b.params.Length() == 0 {
		b.params = b.ToDynMap().MustDynMap("params", dynmap.New())
	}
	return b.params
}

func (b *BasicElement) ParamLookerUpper() ParamLookerUpper {
	return &BasicParamLookerUpper{element: b}
}

// sets a parameter available to this component and all of its
// children
func (b *BasicElement) SetLocalVariable(key string, val interface{}) {
	b.originalMap.PutWithDot(key, val)
}

// Sets a global variable available to all components rendered after this
func (b *BasicElement) SetGlobalVariable(key string, val interface{}) {
	docCtx := FindDocumentContext(b)
	docCtx.Params.PutWithDot(key, val)
}

// returns an attribute finder based on the
// the original map of this element
func (b *BasicElement) Attr() *Attr {
	return b.DmAttr(b.ToDynMap())
}

// returns an attribute finder based on the passed in dynmap
func (b *BasicElement) DmAttr(mp *dynmap.DynMap) *Attr {
	return &Attr{
		element: b,
		mp:      mp,
	}
}

// provides most of the component functionality
// Note we copy pasta the BAsicElement here because embedding
// does not work for the param lookup pieces
type BasicComponent struct {
	id          string
	elementType string
	originalMap *dynmap.DynMap // The map that this element is parsed from
	params      *dynmap.DynMap
	parent      Element
	children    []Element
	ctx         RenderContext
	rendering   bool
}

func (b *BasicComponent) Id() string {
	if len(b.id) == 0 {
		b.id = nextId(b.ToDynMap())
	}
	return b.id
}
func (b *BasicComponent) ElementType() string {
	return b.elementType
}
func (b *BasicComponent) ToDynMap() *dynmap.DynMap {
	return b.originalMap
}
func (b *BasicComponent) Params() *dynmap.DynMap {
	if b.params == nil || b.params.Length() == 0 {
		b.params = b.ToDynMap().MustDynMap("params", dynmap.New())
	}
	return b.params
}

// sets a parameter available to this component and all of its
// children
func (b *BasicComponent) SetLocalVariable(key string, val interface{}) {
	b.originalMap.PutWithDot(key, val)
}

// Sets a global variable available to all components rendered after this
func (b *BasicComponent) SetGlobalVariable(key string, val interface{}) {
	docCtx := FindDocumentContext(b)
	docCtx.Params.PutWithDot(key, val)
}

func (b *BasicComponent) ParamLookerUpper() ParamLookerUpper {
	return &BasicParamLookerUpper{element: b}
}

// returns an attribute finder based on the
// the original map of this element
func (b *BasicComponent) Attr() *Attr {
	return b.DmAttr(b.ToDynMap())
}

// returns an attribute finder based on the passed in dynmap
func (b *BasicComponent) DmAttr(mp *dynmap.DynMap) *Attr {
	return &Attr{
		element: b,
		mp:      mp,
	}
}

// Render the path.  This is typically called by the
// owning element, which handles the transforms
func (b *BasicComponent) Render(ctx RenderContext) (path.Path, RenderContext, error) {
	return nil, ctx, fmt.Errorf("Render not implemented")
}

func (b *BasicComponent) RenderStart(ctx RenderContext) {
	b.ctx = ctx
	b.rendering = true
}

func (b *BasicComponent) SetChildren(c []Element) {
	b.children = c
}

// sets the children componenets
func (b *BasicComponent) SetComponents(c []Component) {
	b.children = CtoE(c)
}

func (b *BasicComponent) RenderContext() (RenderContext, bool) {
	return b.ctx, b.rendering
}

// internal method to handle the transforms
func (b *BasicComponent) HandleTransforms(self Component, pth path.Path, ctx RenderContext) (path.Path, RenderContext, error) {
	// handle the transforms..
	p := pth
	context := ctx.Clone()

	transforms, err := AppContext().MakeTransforms(self.Transforms(), self)
	if err != nil {
		return nil, context, err
	}
	for _, t := range transforms {
		p1, err := t.PathTransform(p)
		if err != nil {
			return p, ctx, err
		}
		p = p1
	}
	b.rendering = false
	context.Cursor = path.PathCursor(p)
	return p, context, nil
}

func (b *BasicComponent) Transforms() []*dynmap.DynMap {
	return b.ToDynMap().MustDynMapSlice("transforms", []*dynmap.DynMap{})
}

func (b *BasicComponent) Parent() Element {
	return b.parent
}
func (b *BasicComponent) SetParent(p Element) {
	b.parent = p
}

func (b *BasicComponent) Children() []Element {
	return b.children
}

// measure this component
// returns w, h
// Note: this typical calls render, so care should be taken when calling this
func (b *BasicComponent) Measure() (float64, float64, error) {
	ctx := RenderContext{
		Origin: path.NewPoint(0, 0),
		Cursor: path.NewPoint(0, 0),
	}
	p, _, err := b.Render(ctx)
	if err != nil {
		return 0, 0, err
	}

	tl, br, err := path.BoundingBoxTrimWhitespace(p, AppContext().SegmentOperators())
	if err != nil {
		return 0, 0, err
	}
	return br.X - tl.X, br.Y - tl.Y, nil
}

// Array of components to an array of elements
func CtoE(components []Component) []Element {
	e := []Element{}
	for _, c := range components {
		e = append(e, c)
	}
	return e
}
