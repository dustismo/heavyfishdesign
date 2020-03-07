package dom

import (
	"github.com/dustismo/heavyfishdesign/dynmap"
)

type PartSplitterTransformerFactory struct{}

func (pf PartSplitterTransformerFactory) CreateTransformer(transformType string, dm *dynmap.DynMap, part *Part) (PartTransformer, error) {
	return &PartSplitter{mp: dm}, nil
}

// // The list of component types this Factory should be used for
func (cf PartSplitterTransformerFactory) TransformerTypes() []string {
	return []string{"splitter"}
}
