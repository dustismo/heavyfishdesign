package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

// func TestSimpleCurve(t *testing.T) {
// 	pathStr := `M 3.545 2.415 C 3.545 2.415 3.314 2.898 3.124 3.121 C 2.934 3.344 2.779 3.561 2.766 3.707`
// 	p, err := path.ParsePathFromSvg(pathStr)

// 	if err != nil {
// 		t.Errorf("Error %s", err)
// 	}
// 	offset := OffsetTransform{
// 		Distance:         .02,
// 		SegmentOperators: path.NewSegmentOperators(),
// 		Precision:        3,
// 		SizeShouldBe:     Larger,
// 	}
// 	newPath, err := offset.PathTransform(p)
// 	if err != nil {
// 		t.Errorf("Error %s", err)
// 	}

// 	expectedStr := ""
// 	actualStr := path.SvgString(newPath, 3)

// 	if expectedStr != actualStr {
// 		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
// 	}
// }

// func TestOffsetCabinet(t *testing.T) {
// 	pathStr := `M 0.000 10.776 L 0.000 0.776 L 2.898 0.000
// 				L 3.545 2.415
// 				C 3.545 2.415 3.314 2.898 3.124 3.121
// 				C 2.934 3.344 2.779 3.561 2.766 3.707
// 				C 2.752 3.856 2.774 4.020 2.863 4.241
// 				C 2.903 4.339 3.855 6.687 4.050 6.936
// 				C 4.254 7.196 5.034 7.436 5.244 7.564
// 				C 5.384 7.649 5.679 7.944 5.919 8.054
// 				C 6.074 8.125 8.000 8.526 8.000 8.526
// 				L 8.000 10.776 L 0.000 10.776
// 				`
// 	pathStr = `M 3.545 2.415 C 3.545 2.415 3.314 2.898 3.124 3.121 C 2.934 3.344 2.779 3.561 2.766 3.707`
// 	p, err := path.ParsePathFromSvg(pathStr)

// 	if err != nil {
// 		t.Errorf("Error %s", err)
// 	}
// 	offset := OffsetTransform{
// 		Distance:         .02,
// 		SegmentOperators: path.NewSegmentOperators(),
// 		Precision:        3,
// 		SizeShouldBe:     Larger,
// 	}
// 	newPath, err := offset.PathTransform(p)
// 	if err != nil {
// 		t.Errorf("Error %s", err)
// 	}

// 	expectedStr := ""
// 	actualStr := path.SvgString(newPath, 3)

// 	if expectedStr != actualStr {
// 		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
// 	}
// }

func TestOffsetTJoint(t *testing.T) {
	// when a path has multiple elements (here it has multiple rectangels)
	// we need to offset each element individually instead of trying to
	// join them together.
	//
	pathStr := `M 1.350 0.500 M 1.350 0.700 L 1.350 1.000 L 1.150 1.000 L 1.150 0.700 L 1.350 0.700 M 1.350 1.200 M 1.350 1.200 M 1.350 1.400 L 1.350 1.700 L 1.150 1.700 L 1.150 1.400 L 1.350 1.400 M 1.350 1.900 M 1.350 1.900 M 1.350 2.100 L 1.350 2.400 L 1.150 2.400 L 1.150 2.100 L 1.350 2.100 M 1.350 2.600`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	offset := OffsetTransform{
		Distance:         .007,
		SegmentOperators: path.NewSegmentOperators(),
		Precision:        3,
		SizeShouldBe:     Smaller,
	}
	newPath, err := offset.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 1.343 0.707 L 1.343 0.993 L 1.157 0.993 L 1.157 0.707 L 1.343 0.707 M 1.343 1.407 L 1.343 1.693 L 1.157 1.693 L 1.157 1.407 L 1.343 1.407 M 1.343 2.107 L 1.343 2.393 L 1.157 2.393 L 1.157 2.107 L 1.343 2.107`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestOffsetFingerJoint(t *testing.T) {
	pathStr := `M 0.000 0.000 L 0.250 0.000 L 0.250 0.500 L 1.000 0.500 L 1.000 0.000 L 1.250 0.000 
				M 1.250 0.000 L 1.500 0.000 L 1.500 0.500 L 2.250 0.500 L 2.250 0.000 L 2.500 0.000 
				M 2.500 0.000 L 2.750 0.000 L 2.750 0.500 L 3.500 0.500 L 3.500 0.000 L 3.750 0.000 
				M 3.750 0.000 L 4.000 0.000 L 4.000 0.500 L 4.750 0.500 L 4.750 0.000 L 5.000 0.000 
				
				M 5.000 0.000 
				M 5.000 0.000 L 5.000 5.000 
				M 5.000 5.000 L 0.000 5.000 
				M 0.000 5.000 L 0.000 0.000`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	offset := OffsetTransform{
		Distance:         .007,
		SegmentOperators: path.NewSegmentOperators(),
		Precision:        3,
		SizeShouldBe:     Larger,
	}
	newPath, err := offset.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 1.250 -0.007 L 0.993 -0.007 L 0.993 0.493 L 0.257 0.493 L 0.257 -0.007 L 0.000 -0.007 M 2.500 -0.007 L 2.243 -0.007 L 2.243 0.493 L 1.507 0.493 L 1.507 -0.007 L 1.250 -0.007 M 3.750 -0.007 L 3.493 -0.007 L 3.493 0.493 L 2.757 0.493 L 2.757 -0.007 L 2.500 -0.007 M 5.000 -0.007 L 4.743 -0.007 L 4.743 0.493 L 4.007 0.493 L 4.007 -0.007 L 3.750 -0.007 M 5.007 5.000 L 5.007 0.000 M 0.000 5.007 L 5.000 5.007 M -0.007 0.000 L -0.007 5.000`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestLineOffset(t *testing.T) {
	pathStr := "M50,50 L150,50"

	// https://codepen.io/anon/pen/orPwgd
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	offset := OffsetTransform{
		Distance:         50,
		SegmentOperators: path.NewSegmentOperators(),
	}
	newPath, err := offset.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := "M 50.000 100.000 L 150.000 100.000"
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestSquareOffsetSmaller(t *testing.T) {
	pathStr := "M0,0 L5,0 L5,5 L0,5 L0,0"
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	offset := OffsetTransform{
		Precision:        3,
		Distance:         1,
		SegmentOperators: path.NewSegmentOperators(),
		SizeShouldBe:     Larger,
	}
	newPath, err := offset.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := "M -1.000 -1.000 L -1.000 6.000 L 6.000 6.000 L 6.000 -1.000 L -1.000 -1.000"
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestSquareOffsetLarger(t *testing.T) {
	pathStr := "M0,0 L5,0 L5,5 L0,5 L0,0"
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	offset := OffsetTransform{
		Precision:        3,
		Distance:         1,
		SegmentOperators: path.NewSegmentOperators(),
		SizeShouldBe:     Larger,
	}
	newPath, err := offset.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := "M -1.000 -1.000 L -1.000 6.000 L 6.000 6.000 L 6.000 -1.000 L -1.000 -1.000"
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestSquareOffset(t *testing.T) {
	pathStr := "M50,50 L150,50 L150,150 L50,150 L50,50"

	// https://codepen.io/anon/pen/orPwgd
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	offset := OffsetTransform{
		Precision:        3,
		Distance:         50,
		SegmentOperators: path.NewSegmentOperators(),
		SizeShouldBe:     Larger,
	}
	newPath, err := offset.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := "M 0.000 0.000 L 0.000 200.000 L 200.000 200.000 L 200.000 0.000 L 0.000 0.000"
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestSimpleCurveOffset(t *testing.T) {
	pathStr := `M 1.875 0 L 1.875 0.200 C 1.875 0.366 1.741 0.500 1.575 0.500 L 1.500 0.500 L 1.425 0.500 C 1.259 0.500 1.125 0.366 1.125 0.200 L 1.125 0.000`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	offset := OffsetTransform{
		Precision:        3,
		Distance:         .01,
		SegmentOperators: path.NewSegmentOperators(),
		SizeShouldBe:     Smaller,
	}
	newPath, err := offset.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 1.865 0.000 L 1.865 0.200 C 1.865 0.308 1.807 0.401 1.720 0.451 C 1.678 0.476 1.628 0.490 1.575 0.490 L 1.500 0.490 L 1.425 0.490 C 1.317 0.490 1.224 0.432 1.174 0.345 C 1.149 0.303 1.135 0.253 1.135 0.200 L 1.135 0.000`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestOffsetEmptyPath(t *testing.T) {
	pathStr := `M 1.350 0.500`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	offset := OffsetTransform{
		Distance:         .007,
		SegmentOperators: path.NewSegmentOperators(),
		Precision:        3,
		SizeShouldBe:     Smaller,
	}
	newPath, err := offset.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := "M 1.350 0.500"
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
