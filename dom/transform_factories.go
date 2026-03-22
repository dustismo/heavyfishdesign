package dom

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

// transform factories live here.
// when adding new factories, remember to update DefaultFactories in parser.go

func createMissingAttributeError(attribute string, transformType string, dm *dynmap.DynMap) error {
	if dm.Contains(attribute) {
		return fmt.Errorf(
			"unable to parse %s value (%s) for %s transform\n%s",
			attribute,
			dm.MustString(attribute, "unknown"),
			transformType,
			dm.ToJSON())
	}
	return fmt.Errorf(
		"%s must be specified for %s transform\n%s",
		attribute,
		transformType,
		dm.ToJSON(),
	)
}

type CleanupTransformFactory struct {
}

func (cf CleanupTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)
	return transforms.CleanupTransform{
		Precision: attr.MustInt("precision", AppContext().Precision()),
	}, nil
}

// // The list of component types this Factory should be used for
func (cf CleanupTransformFactory) TransformTypes() []string {
	return []string{"cleanup"}
}

type RotateTransformFactory struct {
}

func (tf RotateTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)
	degrees, ok := attr.Float64("degrees")
	if !ok {
		return nil, createMissingAttributeError("degrees", transformType, dm)
	}
	handle := attr.MustHandle("axis", path.TopLeft)
	return transforms.RotateTransform{
		Degrees:          degrees,
		Axis:             handle,
		SegmentOperators: AppContext().SegmentOperators(),
	}, nil
}

// // The list of component types this Factory should be used for
func (tf RotateTransformFactory) TransformTypes() []string {
	return []string{"rotate"}
}

type ReorderTransformFactory struct {
}

func (tf ReorderTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)
	return transforms.ReorderTransform{
		Precision: attr.MustInt("precision", 3),
	}, nil
}

// // The list of component types this Factory should be used for
func (cf ReorderTransformFactory) TransformTypes() []string {
	return []string{"reorder"}
}

type ReverseTransformFactory struct {
}

func (tf ReverseTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	return transforms.PathReverse{}, nil
}

// // The list of component types this Factory should be used for
func (cf ReverseTransformFactory) TransformTypes() []string {
	return []string{"reverse"}
}

type JoinTransformFactory struct {
}

func (tf JoinTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	closePath := dm.MustBool("close_path", false)
	attr := NewAttr(element, dm)
	return transforms.JoinTransform{
		Precision:        attr.MustInt("precision", 3),
		SegmentOperators: AppContext().SegmentOperators(),
		ClosePath:        closePath}, nil
}

// // The list of component types this Factory should be used for
func (cf JoinTransformFactory) TransformTypes() []string {
	return []string{"join"}
}

type OffsetTransformFactory struct {
}

func (tf OffsetTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)
	distance, ok := attr.Float64("distance")
	if !ok {
		return nil, createMissingAttributeError("distance", transformType, dm)
	}

	sizeShouldBeStr, ok := attr.String("size_should_be")
	var sizeShouldBe transforms.SizeShouldBe = transforms.Unknown
	if ok {
		switch sizeShouldBeStr {
		case "smaller":
			sizeShouldBe = transforms.Smaller
		case "larger":
			sizeShouldBe = transforms.Larger
		}
	}

	return transforms.OffsetTransform{
		Precision:        attr.MustInt("precision", 3),
		Distance:         distance,
		SegmentOperators: AppContext().SegmentOperators(),
		SizeShouldBe:     sizeShouldBe,
	}, nil
}

// // The list of component types this Factory should be used for
func (tf OffsetTransformFactory) TransformTypes() []string {
	return []string{"offset"}
}

type MoveTransformFactory struct {
}

func (tf MoveTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)

	to, ok := attr.Point("to")
	if !ok {
		return nil, createMissingAttributeError("to", transformType, dm)
	}
	handle := attr.MustHandle("handle", path.TopLeft)
	return transforms.MoveTransform{
		Point:            to,
		Handle:           handle,
		SegmentOperators: AppContext().SegmentOperators(),
	}, nil
}

// // The list of component types this Factory should be used for
func (tf MoveTransformFactory) TransformTypes() []string {
	return []string{"move"}
}

type TrimTransformFactory struct {
}

func (tf TrimTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	return transforms.TrimWhitespaceTransform{
		SegmentOperators: AppContext().SegmentOperators(),
	}, nil
}

// The list of component types this Factory should be used for
func (tf TrimTransformFactory) TransformTypes() []string {
	return []string{"trim"}
}

type SliceTransformFactory struct {
}

// The list of component types this Factory should be used for
func (tf SliceTransformFactory) TransformTypes() []string {
	return []string{"slice"}
}
func (tf SliceTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)
	y, ok := attr.Float64("y")
	if !ok {
		return nil, createMissingAttributeError("y", transformType, dm)
	}
	return transforms.HSliceTransform{
		Y:                y,
		SegmentOperators: AppContext().SegmentOperators(),
		Precision:        AppContext().Precision(),
	}, nil
}

type ScaleTransformFactory struct {
}

// float64FromTransformMap reads and evaluates a numeric field only if it appears on this transform's
// JSON object. It does not fall back to parent/part params via Attr.lookup.
//
// Without this, a scale like {"type":"scale","width":"…"} would still resolve "height" from the
// owning part's param "height" (panel height), producing non-uniform X/Y scale and stretching circles.
func float64FromTransformMap(e Element, dm *dynmap.DynMap, key string, def float64) float64 {
	v, ok := dm.Get(key)
	if !ok {
		return def
	}
	f, err := e.ParamLookerUpper().ToFloat64(v)
	if err != nil {
		return def
	}
	return f
}

// // The list of component types this Factory should be used for
func (tf ScaleTransformFactory) TransformTypes() []string {
	return []string{"scale"}
}
func (tf ScaleTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)

	scaleX := attr.MustFloat64("scale_x", 1)
	scaleY := attr.MustFloat64("scale_y", 1)

	startPoint := attr.MustPoint("start_point", path.NewPoint(0, 0))
	endPoint := attr.MustPoint("end_point", path.NewPoint(0, 0))

	width := float64FromTransformMap(element, dm, "width", 0)
	height := float64FromTransformMap(element, dm, "height", 0)

	return transforms.ScaleTransform{
		ScaleX:           scaleX,
		ScaleY:           scaleY,
		StartPoint:       startPoint,
		EndPoint:         endPoint,
		Width:            width,
		Height:           height,
		SegmentOperators: AppContext().SegmentOperators(),
	}, nil
}

type MirrorTransformFactory struct {
	Axis transforms.Axis
}

func (tf MirrorTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)
	axis := attr.MustString("axis", "horizontal")
	a := transforms.Horizontal
	if axis == "vertical" {
		a = transforms.Vertical
	}
	return transforms.MirrorTransform{
		Axis:             a,
		SegmentOperators: AppContext().SegmentOperators(),
	}, nil
}

// // The list of component types this Factory should be used for
func (tf MirrorTransformFactory) TransformTypes() []string {
	return []string{"mirror"}
}

type RotateScaleTransformFactory struct {
}

// // The list of component types this Factory should be used for
func (tf RotateScaleTransformFactory) TransformTypes() []string {
	return []string{"rotate_scale"}
}
func (tf RotateScaleTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)

	startPoint := attr.MustPoint("start_point", path.NewPoint(0, 0))
	endPoint := attr.MustPoint("end_point", path.NewPoint(0, 0))
	return transforms.RotateScaleTransform{
		StartPoint:       startPoint,
		EndPoint:         endPoint,
		SegmentOperators: AppContext().SegmentOperators(),
	}, nil
}

type MatrixTransformFactory struct {
	Values []float64
}

func (tf MatrixTransformFactory) CreateTransform(transformType string, dm *dynmap.DynMap, element Element) (path.PathTransform, error) {
	attr := NewAttr(element, dm)

	a, ok := attr.Float64("a")
	if !ok {
		return nil, createMissingAttributeError("a", transformType, dm)
	}
	b, ok := attr.Float64("b")
	if !ok {
		return nil, createMissingAttributeError("b", transformType, dm)
	}
	c, ok := attr.Float64("c")
	if !ok {
		return nil, createMissingAttributeError("c", transformType, dm)
	}
	d, ok := attr.Float64("d")
	if !ok {
		return nil, createMissingAttributeError("d", transformType, dm)
	}

	e, ok := attr.Float64("e")
	if !ok {
		return nil, createMissingAttributeError("e", transformType, dm)
	}
	f, ok := attr.Float64("f")
	if !ok {
		return nil, createMissingAttributeError("f", transformType, dm)
	}

	return transforms.MatrixTransform{
		A:                a,
		B:                b,
		C:                c,
		D:                d,
		E:                e,
		F:                f,
		SegmentOperators: AppContext().SegmentOperators(),
	}, nil
}

// // The list of component types this Factory should be used for
func (tf MatrixTransformFactory) TransformTypes() []string {
	return []string{"matrix"}
}
