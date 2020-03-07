package dom

import (
	"fmt"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

// transform factories live here.
// when adding new factories, remember to update DefaultFactories in parser.go

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
		return nil, fmt.Errorf("Error, degrees must be specified for rotate transform")
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
		fmt.Printf("%+v \n", element)
		return nil, fmt.Errorf("Error, distance must be specified for offset transform")
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
		return nil, fmt.Errorf("Error, 'to' must be specified for move transform")
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
		return nil, fmt.Errorf("Error, 'y' must be specified for slice transform")
	}
	return transforms.HSliceTransform{
		Y:                y,
		SegmentOperators: AppContext().SegmentOperators(),
		Precision:        AppContext().Precision(),
	}, nil
}

type ScaleTransformFactory struct {
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

	width := attr.MustFloat64("width", 0)
	height := attr.MustFloat64("height", 0)

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
