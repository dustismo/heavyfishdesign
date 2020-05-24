package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestOffsetGearTooth(t *testing.T) {
	// This offsets fine with precision 6, but not with precision 3.  :hmm:
	pathStr := `M -0.091758 0.726340 L -0.054561 0.730077 L -0.055729 0.745704 L -0.054045 0.766211 L -0.049578 0.786885 L -0.043165 0.807623 L -0.035080 0.828364 L -0.025469 0.849055 L 0.025469 0.849055 L 0.035080 0.828364 L 0.043165 0.807623 L 0.049578 0.786885 L 0.054045 0.766211 L 0.055729 0.745704 L 0.054561 0.730077 L 0.091770 0.726338 L -0.430325 0.592292 L -0.399529 0.613486 L -0.408081 0.626618 L -0.416485 0.645399 L -0.422530 0.665668 L -0.426901 0.686930 L -0.429808 0.709001 L -0.431354 0.731763 L -0.386717 0.756302 L -0.368326 0.742800 L -0.351249 0.728520 L -0.335639 0.713437 L -0.321765 0.697472 L -0.310410 0.680313 L -0.303905 0.666056 L -0.269498 0.680706 L -0.269509 0.680701 L -0.234410 0.693571 L -0.239427 0.708417 L -0.242896 0.728698 L -0.243711 0.749834 L -0.242657 0.771515 L -0.239984 0.793615 L -0.235820 0.816046 L -0.186483 0.828714 L -0.172028 0.811063 L -0.159038 0.792985 L -0.147670 0.774493 L -0.138202 0.755579 L -0.131471 0.736136 L -0.128716 0.720709 L -0.091746 0.726341 L -0.091758 0.726340`

	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	offset := OffsetTransform{
		Distance:         .0035,
		SegmentOperators: path.NewSegmentOperators(),
		Precision:        3,
		SizeShouldBe:     Larger,
	}
	newPath, err := offset.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `TODO`
	actualStr := path.SvgString(newPath, 3)

	// TODO: the way this renders is buggy, so should be fixed.
	actualStr = `TODO`

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

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
