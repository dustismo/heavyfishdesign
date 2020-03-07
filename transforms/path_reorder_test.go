package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestReorder(t *testing.T) {
	pathStr := `M 0.000 0.000 L 5.000 0.000 
	M 5.000 0.000 L 5.000 5.000 
	M 0.000 5.000 L 5.000 5.000 
	M 0.000 0.000 L 0.000 5.000`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	reorder := ReorderTransform{
		Precision: 3,
	}
	newPath, err := reorder.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	// M 0.000 0.000 L 5.000 0.000  ____
	// M 5.000 0.000 L 5.000 5.000      | right
	// M 0.000 5.000 L 5.000 5.000  ____  bottom
	// M 0.000 0.000 L 0.000 5.000 |

	expectedStr := `M 0.000 0.000 L 5.000 0.000 L 5.000 5.000 L 0.000 5.000 L 0.000 0.000`
	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestReorderComplex(t *testing.T) {
	pathStr := `M 0.000 -0.050 L 0.250 -0.050 M 0.300 0.000 L 0.300 0.500 M 0.250 0.450 L 1.000 0.450 M 0.950 0.500 L 0.950 0.000 M 1.000 -0.050 L 1.250 -0.050 M 1.250 -0.050 L 1.500 -0.050 M 1.550 0.000 L 1.550 0.500 M 1.500 0.450 L 2.250 0.450 M 2.200 0.500 L 2.200 0.000 M 2.250 -0.050 L 2.500 -0.050 M 2.500 -0.050 L 2.750 -0.050 M 2.800 0.000 L 2.800 0.500 M 2.750 0.450 L 3.500 0.450 M 3.450 0.500 L 3.450 0.000 M 3.500 -0.050 L 3.750 -0.050 M 3.750 -0.050 L 4.000 -0.050 M 4.050 0.000 L 4.050 0.500 M 4.000 0.450 L 4.750 0.450 M 4.700 0.500 L 4.700 0.000 M 4.750 -0.050 L 5.000 -0.050 M 5.050 0.000 L 5.050 5.000 M 5.000 5.050 L 0.000 5.050 M -0.050 5.000 L -0.050 0.000`
	p, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}
	reorder := ReorderTransform{
		Precision: 3,
	}
	newPath, err := reorder.PathTransform(p)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	expectedStr := `M 0.000 -0.050 L 0.250 -0.050 M 0.300 0.000 L 0.300 0.500 M 0.250 0.450 L 1.000 0.450 M 0.950 0.500 L 0.950 0.000 M 1.000 -0.050 L 1.250 -0.050 L 1.500 -0.050 M 1.550 0.000 L 1.550 0.500 M 1.500 0.450 L 2.250 0.450 M 2.200 0.500 L 2.200 0.000 M 2.250 -0.050 L 2.500 -0.050 L 2.750 -0.050 M 2.800 0.000 L 2.800 0.500 M 2.750 0.450 L 3.500 0.450 M 3.450 0.500 L 3.450 0.000 M 3.500 -0.050 L 3.750 -0.050 L 4.000 -0.050 M 4.050 0.000 L 4.050 0.500 M 4.000 0.450 L 4.750 0.450 M 4.700 0.500 L 4.700 0.000 M 4.750 -0.050 L 5.000 -0.050 M 5.050 0.000 L 5.050 5.000 M 5.000 5.050 L 0.000 5.050 M -0.050 5.000 L -0.050 0.000`

	actualStr := path.SvgString(newPath, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
