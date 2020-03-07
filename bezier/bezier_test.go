package bezier

import (
	"fmt"
	"testing"
)

func TestDeCasteljau(t *testing.T) {

	point := DeCasteljau(CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	},
		.7,
	)

	if fmt.Sprintf("%.3f,%.3f", point.X, point.Y) != "160.035,110.490" {
		t.Errorf("DeCasteljau failed. Please learn maths. Expected (%s) Got (%.3f, %.3f)", "160.035,110.490", point.X, point.Y)
	}
}

func curveString(c CubicCurve) string {
	return fmt.Sprintf("start: %.3f,%.3f startControl: %.3f,%.3f end: %.3f,%.3f endControl: %.3f,%.3f",
		c.Start.X, c.Start.Y,
		c.StartControl.X, c.StartControl.Y,
		c.End.X, c.End.Y,
		c.EndControl.X, c.EndControl.Y)
}
func TestSplitCurve(t *testing.T) {
	left, right := SplitCurve(
		CubicCurve{
			Start:        NewPoint(120, 160),
			StartControl: NewPoint(35, 200),
			End:          NewPoint(220, 40),
			EndControl:   NewPoint(220, 260),
		}, .75)

	expectedLeft := "start: 120.000,160.000 startControl: 56.250,190.000 end: 192.422,157.188 endControl: 144.375,231.250"
	expectedRight := "start: 192.422,157.188 startControl: 208.438,132.500 end: 220.000,40.000 endControl: 220.000,95.000"

	if curveString(left) != expectedLeft {
		t.Errorf("Left section was %s but expected %s", curveString(left), expectedLeft)
	}

	if curveString(right) != expectedRight {
		t.Errorf("Right section was %s but expected %s", curveString(right), expectedRight)
	}

	// should be ~200, 160 start point
}

func TestDerive(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}

	// in js:
	// var points =
	// var curve = new Bezier(90.000,110.000,  25.000,40.000,  230.000,40.000,  150.000,240.000);
	// console.log(curve.dpoints)
	dpoints := derive(curve)

	expectedStr := "[[(-195.000,-210.000) (615.000,0.000) (-240.000,600.000)] [(1620.000,420.000) (-1710.000,1200.000)] [(-3330.000,780.000)]]"
	actualStr := fmt.Sprintf("%+v", dpoints)
	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestDerivative(t *testing.T) {
	// in js:
	// var curve = new Bezier(100,25 , 10,90 , 110,100 , 150,195);
	// curve.derivative(.5)
	// == {x : 112.5, y: 135}

	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}

	point := Derivative(curve, .5)
	if point.X != 198.75 {
		t.Errorf("Expected X: %.3f Actual X: %.3f", 198.75, point.X)
	}

	if point.Y != 97.5 {
		t.Errorf("Expected Y: %.3f Actual Y: %.3f", 97.5, point.Y)
	}
}

func TestFindPoint(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}
	point := FindPoint(curve, .43)
	expected := "111.793,68.865"
	actual := fmt.Sprintf("%.3f,%.3f", point.X, point.Y)
	if expected != actual {
		t.Errorf("Expected: %s  Actual: %s", expected, actual)
	}
}

func TestExtrema(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}
	x, y := extrema(curve)
	expectedStr := "X: [0.14072358446134256 0.8322493885116303 0.4864864864864865] Y: [0.3717045820153255]"
	actualStr := fmt.Sprintf("X: %+v Y: %+v", x, y)
	if expectedStr != actualStr {
		t.Errorf("Expected %s\nActual %s", expectedStr, actualStr)
	}
}

func TestMinMax(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}

	xMin, xMax, yMin, yMax := getMinMax(curve)
	expectedStr := "(77.053,84.969), (168.820,155.620), (100.928,67.633), (150.000,240.000)"
	actualStr := fmt.Sprintf("%+v, %+v, %+v, %+v", xMin, xMax, yMin, yMax)
	if expectedStr != actualStr {
		t.Errorf("Expected %s\nActual %s", expectedStr, actualStr)
	}
}

func TestSplitCurveSection(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}
	result := SplitCurveSection(curve, .2, .3)

	expectedStr := "[(78.960,77.440) (81.040,73.760) (84.710,71.040) (89.415,69.410)]"
	actualStr := fmt.Sprintf("%+v", cubicCurveToArray(result))
	if expectedStr != actualStr {
		t.Errorf("Expected %s\nActual %s", expectedStr, actualStr)
	}
}

func TestBoundingBox(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}

	topLeft, bottomRight := BoundingBox(curve)

	actualStr := fmt.Sprintf("topLeft %+v bottemRight %+v", topLeft, bottomRight)
	expectedStr := "topLeft (77.053,67.633) bottemRight (168.820,240.000)"
	if expectedStr != actualStr {
		t.Errorf("Expected %s\nActual %s", expectedStr, actualStr)
	}
}

func TestScale(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}
	result, _ := Scale(curve, .3)

	// 	points: Array(4)
	// 0: {x: 90.21983804748788, y: 109.79586467018981}
	// 1: {x: 25.04073818265443, y: 39.60298789267687}
	// 2: {x: 229.56204468703783, y: 40.28711356133528}
	// 3: {x: 149.72145699273443, y: 239.88858279709376}

	points := cubicCurveToArray(result)
	// test a random point
	if points[1].X != 25.0407381826544295 {
		t.Errorf("expected %f, got %f", 25.0407381826544295, points[1].X)
	}
}

func TestReduce(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}
	result := reduce(curve)

	// check a random point
	// points: Array[3]
	// 0: {x: 100.92793889156523, y: 67.63285925823757}
	// 1: {x: 107.70458101879049, y: 67.63285925823757}
	// 2: {x: 115.32051549113373, y: 69.19173274438701}
	// 3: {x: 122.93644996347697, y: 72.50607071644325}

	if result[3].curve.EndControl.X != 115.32051549113373 {
		t.Errorf("expected %f, got %f", 115.32051549113373, result[3].curve.EndControl.X)
	}
}

func TestOffsetPoint(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}

	op := OffsetPoint(curve, .1, .3)
	actualStr := fmt.Sprintf("Offset point: %.9f, %.9f", op.X, op.Y)
	expectedStr := "Offset point: 78.332144822, 91.143121631"
	if expectedStr != actualStr {
		t.Errorf("Expected %s\nActual %s", expectedStr, actualStr)
	}
}

func TestOffset2(t *testing.T) {
	curve := CubicCurve{
		NewPoint(3.545, 2.415),
		NewPoint(3.545, 2.415),
		NewPoint(3.124, 3.121),
		NewPoint(3.314, 2.898),
	}

	curves := Offset(curve, .02)

	if len(curves) == 0 {
		t.Errorf("Offset curve is empty")
	}
	//random point check
	actualStr := cubicCurveToArray(curves[1])[1].String()
	expectedStr := "(3.513,2.435)"
	if actualStr != expectedStr {
		t.Errorf("Expected %s but got %s", expectedStr, actualStr)
	}
}

func TestOffset(t *testing.T) {
	curve := CubicCurve{
		NewPoint(90, 110),
		NewPoint(25, 40),
		NewPoint(150, 240),
		NewPoint(230, 40),
	}

	curves := Offset(curve, .3)

	// for i, c := range curves {
	// 	// 0: {x: 168.5201215565434, y: 155.620122575963}
	// 	// 1: {x: 168.52012155654333, y: 178.49789941112564}
	// 	// 2: {x: 163.12378575909608, y: 206.3827608811896}
	// 	// 3: {x: 149.72145699273443, y: 239.88858279709376}
	// 	// [(168.5201215565433870,155.6201225759629949)
	// 	// (168.5201215565433301,178.4978994111256441)
	// 	// (163.1237857590960800,206.3827608811895971)
	// 	// (149.7214569927344314,239.8885827970937612)]
	// 	t.Errorf("result[%d] %+v\n", i, cubicCurveToArray(c))
	// }
	// random point check
	if cubicCurveToArray(curves[6])[3].X != 149.7214569927344314 {
		t.Errorf("Expected %f but got %f", 149.7214569927344314, cubicCurveToArray(curves[6])[3].X)
	}
}
