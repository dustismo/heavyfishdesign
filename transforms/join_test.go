package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestJoinLineWithCurve2(t *testing.T) {
	pathStr := `M 0.000 0.011 C 0.004 0.011 0.626 -0.092 0.905 0.333 C 1.179 0.751 1.084 0.900 1.252 1.126 C 1.437 1.375 1.789 1.230 1.522 0.706 C 1.256 0.183 1.822 0.074 1.822 0.074 C 1.822 0.074 1.766 0.895 2.303 0.839 C 2.562 0.812 3.364 0.812 3.559 1.021 C 3.559 1.021 3.025 1.002 3.101 1.741 C 3.131 2.031 3.505 2.412 3.902 2.240 C 4.393 2.028 4.707 1.990 5.027 2.245 M 4.500 2.011 L 4.500 3.011 M 4.500 2.761 L 4.000 2.761 L 3.850 2.761 L 3.850 3.011 L 3.450 3.011 L 3.450 2.761 L 3.300 2.761 L 3.150 2.761 L 3.150 3.011 L 2.750 3.011 L 2.750 2.761 L 2.600 2.761 L 2.450 2.761 L 2.450 3.011 L 2.050 3.011 L 2.050 2.761 L 1.900 2.761 L 1.750 2.761 L 1.750 3.011 L 1.350 3.011 L 1.350 2.761 L 1.200 2.761 L 1.050 2.761 L 1.050 3.011 L 0.650 3.011 L 0.650 2.761 L 0.500 2.761 L 0.000 2.761 M 0.250 3.011 L 0.250 2.561 L 0.250 2.411 L 0.000 2.411 L 0.000 2.011 L 0.250 2.011 L 0.250 1.861 L 0.250 1.711 L 0.000 1.711 L 0.000 1.311 L 0.250 1.311 L 0.250 1.161 L 0.250 1.011 L 0.000 1.011 L 0.000 0.611 L 0.250 0.611 L 0.250 0.461 L 0.250 0.011 M 0.750 2.261 C 0.750 2.261 0.750 2.261 0.750 2.261 C 0.750 2.261 0.750 2.261 0.750 2.261 C 0.750 2.261 0.750 2.261 0.750 2.261 C 0.750 2.261 0.750 2.261 0.750 2.261 M 3.750 2.261 C 3.750 2.261 3.750 2.261 3.750 2.261 C 3.750 2.261 3.750 2.261 3.750 2.261 C 3.750 2.261 3.750 2.261 3.750 2.261 C 3.750 2.261 3.750 2.261 3.750 2.261`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
		ClosePath:        true,
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 4.500 2.069 C 4.327 2.078 4.135 2.140 3.902 2.240 C 3.505 2.412 3.131 2.031 3.101 1.741 C 3.025 1.002 3.559 1.021 3.559 1.021 C 3.364 0.812 2.562 0.812 2.303 0.839 C 1.766 0.895 1.822 0.074 1.822 0.074 C 1.822 0.074 1.256 0.183 1.522 0.706 C 1.789 1.230 1.437 1.375 1.252 1.126 C 1.084 0.900 1.179 0.751 0.905 0.333 C 0.740 0.082 0.456 0.015 0.250 0.003 L 0.250 0.461 L 0.250 0.611 L 0.000 0.611 L 0.000 1.011 L 0.250 1.011 L 0.250 1.161 L 0.250 1.311 L 0.000 1.311 L 0.000 1.711 L 0.250 1.711 L 0.250 1.861 L 0.250 2.011 L 0.000 2.011 L 0.000 2.411 L 0.250 2.411 L 0.250 2.561 L 0.250 2.761 L 0.500 2.761 L 0.650 2.761 L 0.650 3.011 L 1.050 3.011 L 1.050 2.761 L 1.200 2.761 L 1.350 2.761 L 1.350 3.011 L 1.750 3.011 L 1.750 2.761 L 1.900 2.761 L 2.050 2.761 L 2.050 3.011 L 2.450 3.011 L 2.450 2.761 L 2.600 2.761 L 2.750 2.761 L 2.750 3.011 L 3.150 3.011 L 3.150 2.761 L 3.300 2.761 L 3.450 2.761 L 3.450 3.011 L 3.850 3.011 L 3.850 2.761 L 4.000 2.761 L 4.500 2.761 L 4.500 2.069`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestJoinLineWithCurve(t *testing.T) {
	pathStr := `M 0.000 0.066 C 2.673 0.994 2.854 0.953 3.055 1.084 M 2.800 1.066 L 2.800 1.366`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
		ClosePath:        false,
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 0.000 0.066 C 1.846 0.707 2.504 0.886 2.800 0.978 L 2.800 1.366`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestSquare2(t *testing.T) {
	pathStr := `M 0.000 0.000 L 5.000 0.000 
				M 5.000 0.000 L 5.000 5.000 
				M 0.000 5.000 L 5.000 5.000 
				M 0.000 0.000 L 0.000 5.000`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 0.000 0.000 L 5.000 0.000 L 5.000 5.000 L 0.000 5.000 L 0.000 0.000`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestJoinCrossed(t *testing.T) {
	pathStr := `M 5 3 L 2 3 M 3 1 L 3 4`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 5.000 3.000 L 3.000 3.000 L 3.000 1.000`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestJoinAfterOffset(t *testing.T) {
	pathStr := `M 0.000 -0.050 L 0.250 -0.050 M 0.300 0.000 L 0.300 0.500 M 0.250 0.450 L 1.000 0.450 M 0.950 0.500 L 0.950 0.000 M 1.000 -0.050 L 1.250 -0.050 M 1.250 -0.050 L 1.500 -0.050 M 1.550 0.000 L 1.550 0.500 M 1.500 0.450 L 2.250 0.450 M 2.200 0.500 L 2.200 0.000 M 2.250 -0.050 L 2.500 -0.050 M 2.500 -0.050 L 2.750 -0.050 M 2.800 0.000 L 2.800 0.500 M 2.750 0.450 L 3.500 0.450 M 3.450 0.500 L 3.450 0.000 M 3.500 -0.050 L 3.750 -0.050 M 3.750 -0.050 L 4.000 -0.050 M 4.050 0.000 L 4.050 0.500 M 4.000 0.450 L 4.750 0.450 M 4.700 0.500 L 4.700 0.000 M 4.750 -0.050 L 5.000 -0.050 M 5.050 0.000 L 5.050 5.000 M 5.000 5.050 L 0.000 5.050 M -0.050 5.000 L -0.050 0.000`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
		ClosePath:        true,
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 0.950 -0.050 L 0.950 0.450 L 0.300 0.450 L 0.300 -0.050 L -0.050 -0.050 L -0.050 5.050 L 5.050 5.050 L 5.050 -0.050 L 4.700 -0.050 L 4.700 0.450 L 4.050 0.450 L 4.050 -0.050 L 3.750 -0.050 L 3.450 -0.050 L 3.450 0.450 L 2.800 0.450 L 2.800 -0.050 L 2.500 -0.050 L 2.200 -0.050 L 2.200 0.450 L 1.550 0.450 L 1.550 -0.050 L 1.250 -0.050 L 0.950 -0.050`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestJoinSquare2(t *testing.T) {
	pathStr := `M.5,0 L3.5, 0 
				M0,0, L0,4`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 3.500 0.000 L 0.000 0.000 L 0.000 4.000`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestJoinFingerJoint3(t *testing.T) {

	// This is the shape:
	// ____
	// |    |
	// |    |
	//
	// it should transform to:
	// ______
	// |    |
	// |    |

	pathStr := `M 4.180 0.000  L 1.170 0.000 L 0.000 0.000 L 0.000 5.000 
				M 0.000 5.000 M 5.000 0.000 L 5.000 5.000`

	// this double move is the problem:
	// M 0.000 5.000 M 5.000 0.000
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 0.000 5.000 L 0.000 0.000 L 1.170 0.000 L 5.000 0.000 L 5.000 5.000`
	actualStr := path.SvgString(newPath, 3)
	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestJoinFingerJoint2(t *testing.T) {
	pathStr := `M 0.820 0.000 L 1.170 0.000 L 1.170 -0.450 L 2.150 -0.450 L 2.150 0.000 L 2.500 0.000 
				L 2.850 0.000 L 2.850 -0.450 L 3.830 -0.450 L 3.830 0.000 L 4.180 0.000 
				M 0.000 0.000 L 0.000 5.000
				M 5.000 0.000 L 5.000 5.000`

	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := `M 0.000 5.000 L 0.000 0.000 L 1.170 0.000 L 1.170 -0.450 L 2.150 -0.450 L 2.150 0.000 L 2.500 0.000 L 2.850 0.000 L 2.850 -0.450 L 3.830 -0.450 L 3.830 0.000 L 5.000 0.000 L 5.000 5.000`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestJoinFingerJoint(t *testing.T) {
	pathStr := `M 0.820 -0.450 M 0.820 0.000 L 1.170 0.000 L 4.180 0.000 M 4.180 -0.450 M 4.180 -0.450 M 0.000 0.000 L 0.000 5.000`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	expectedStr := "M 4.180 0.000 L 1.170 0.000 L 0.000 0.000 L 0.000 5.000"
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestJoinRightAngle(t *testing.T) {
	pathStr := "M5,5 L50,5 M75,30 L75, 80"

	// https://codepen.io/anon/pen/orPwgd
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := "M 5.000 5.000 L 75.000 5.000 L 75.000 80.000"
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestJoinSquare(t *testing.T) {
	pathStr := "M10,0 L50,0 M60,10 L60,50 M50,60 L10,60 M0,50 L0,10"

	// https://codepen.io/anon/pen/orPwgd
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	join := JoinTransform{
		Precision:        3,
		SegmentOperators: path.NewSegmentOperators(),
		ClosePath:        true,
	}
	newPath, err := join.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}

	expectedStr := "M 0.000 0.000 L 60.000 0.000 L 60.000 60.000 L 0.000 60.000 L 0.000 0.000"
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
