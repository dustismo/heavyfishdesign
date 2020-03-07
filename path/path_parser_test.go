package path

import "testing"

func TestPathParser(t *testing.T) {
	path := "m 6.1222251,228.86139 c 3.2732061,-2.13129 5.8707069,-4.49412 10.6737049,-3.83158 C 4.802998,0.66254 7.471412,20.84919 7.896028,31.75029"
	p, err := ParsePathFromSvg(path)

	if err != nil {
		t.Errorf("Error parsing path %s", err)
	}

	expectedStr := "M 6.122 228.861 C 9.395 226.730 11.993 224.367 16.796 225.030 C 4.803 0.663 7.471 20.849 7.896 31.750"
	actualStr := SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestPathParserWhitespace(t *testing.T) {
	path := "  m 6,228 L 10  20"
	p, err := ParsePathFromSvg(path)

	if err != nil {
		t.Errorf("Error parsing path %s", err)
	}

	expectedStr := "M 6.000 228.000 L 10.000 20.000"
	actualStr := SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}

func TestPathParserDecimals(t *testing.T) {
	path := "L3.5,0"
	p, err := ParsePathFromSvg(path)

	if err != nil {
		t.Errorf("Error parsing path %s", err)
	}

	expectedStr := "L 3.500 0.000"
	actualStr := SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
