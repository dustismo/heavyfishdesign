package dom

import (
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

type GroupComponentFactory struct{}

type GroupComponent struct {
	*BasicComponent
	components       []Component
	segmentOperators path.SegmentOperators
}

func (ccf GroupComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *DocumentContext) (Component, error) {
	factory := AppContext()
	dm := mp.Clone()
	components, err := factory.MakeComponents(dm.MustDynMapSlice("components", []*dynmap.DynMap{}), dc)
	if err != nil {
		return nil, err
	}
	dm.Put("components", ComponentsToDynMap(components))
	bc := factory.MakeBasicComponent(dm)
	bc.SetComponents(components)
	gc := &GroupComponent{
		BasicComponent:   bc,
		components:       components,
		segmentOperators: factory.SegmentOperators(),
	}
	for _, c := range components {
		c.SetParent(gc)
	}
	return gc, nil
}

// The list of component types this Factory should be used for
func (ccf GroupComponentFactory) ComponentTypes() []string {
	return []string{"group"}
}

func (cc *GroupComponent) Children() []Element {
	return CtoE(cc.components)
}

func (cc *GroupComponent) Render(ctx RenderContext) (path.Path, RenderContext, error) {
	cc.RenderStart(ctx)

	// render all the children then merge into one path.
	paths := []path.Path{}

	context := ctx.Clone()
	for _, e := range cc.components {
		p, c, err := e.Render(context)
		if err != nil {
			return p, c, err
		}
		paths = append(paths, p)
		context = c
		// find the new cursor location
		context.Cursor = path.PathCursor(p)
	}

	p := transforms.SimpleJoin{}.JoinPaths(paths...)
	return cc.HandleTransforms(cc, p, context)
}
