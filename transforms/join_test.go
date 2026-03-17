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

	// ReorderTransform returns the full path; traversal may start at a different vertex (same closed outline).
	expectedStr := `M 1.550 -0.050 L 1.550 0.450 L 2.200 0.450 L 2.200 -0.050 L 2.500 -0.050 L 2.800 -0.050 L 2.800 0.450 L 3.450 0.450 L 3.450 -0.050 L 3.750 -0.050 L 4.050 -0.050 L 4.050 0.450 L 4.700 0.450 L 4.700 -0.050 L 5.050 -0.050 L 5.050 5.050 L -0.050 5.050 L -0.050 -0.050 L 0.300 -0.050 L 0.300 0.450 L 0.950 0.450 L 0.950 -0.050 L 1.250 -0.050 L 1.550 -0.050`
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

// TestJoinKeyedEdgeNonSymmetricKey reproduces the bug where a keyed edge with a
// non-symmetric key (e.g. chessboard) can "close one side completely" due to
// JoinTransform (or ReorderTransform) connecting segments in the wrong order.
// Key path from user: non-symmetric polygon that when HSlice'd and joined must
// produce a single open edge outline (left straight + key pocket + right straight).
// Passes when key is oriented left-to-right after the first join (keyed_edge step 8b)
// and ReorderTransform returns the full reordered path (not just the last subpath).
func TestJoinKeyedEdgeNonSymmetricKey(t *testing.T) {
	// Non-symmetric key SVG path (e.g. pawn silhouette) — closed polygon.
	keySVG := "M 350 183.3333 L 440.4444 172.5 L 455.4966 287.0521 L 350 316.6667 L 305.5556 250 L 350 183.3333"
	keyPath, err := path.ParsePathFromSvg(keySVG)
	if err != nil {
		t.Fatalf("parse key_svg: %v", err)
	}

	so := path.NewSegmentOperators()
	precision := 3
	edgeLen := 200.0
	keyWidth := 100.0

	// Replicate keyed_edge pipeline: trim, scale, HSlice at mid, mirror, trim, join (key only).
	keyPath, err = TrimWhitespaceTransform{SegmentOperators: so}.PathTransform(keyPath)
	if err != nil {
		t.Fatalf("trim: %v", err)
	}
	tl, br, err := path.BoundingBoxTrimWhitespace(keyPath, so)
	if err != nil {
		t.Fatalf("bbox: %v", err)
	}
	naturalWidth := br.X - tl.X
	naturalHeight := br.Y - tl.Y
	if naturalWidth <= 0 || naturalHeight <= 0 {
		t.Fatalf("key has zero bbox: w=%.4f h=%.4f", naturalWidth, naturalHeight)
	}
	keyHeight := keyWidth * (naturalHeight / naturalWidth)

	keyPath, err = ScaleTransform{
		Width:            keyWidth,
		Height:           keyHeight,
		SegmentOperators: so,
	}.PathTransform(keyPath)
	if err != nil {
		t.Fatalf("scale: %v", err)
	}

	keyPath, err = HSliceTransform{
		Y:                keyHeight / 2,
		SegmentOperators: so,
		Precision:        precision,
	}.PathTransform(keyPath)
	if err != nil {
		t.Fatalf("hslice: %v", err)
	}

	keyPath, err = MirrorTransform{
		Axis:             Horizontal,
		Handle:           path.TopLeft,
		SegmentOperators: so,
	}.PathTransform(keyPath)
	if err != nil {
		t.Fatalf("mirror: %v", err)
	}
	keyPath, err = TrimWhitespaceTransform{SegmentOperators: so}.PathTransform(keyPath)
	if err != nil {
		t.Fatalf("trim after mirror: %v", err)
	}

	// First JoinTransform: reconnect disjoint segments of the key profile (from HSlice).
	keyPath, err = JoinTransform{
		Precision:        precision,
		SegmentOperators: so,
	}.PathTransform(keyPath)
	if err != nil {
		t.Fatalf("join key profile: %v", err)
	}

	// Orient key left-to-right (same as keyed_edge step 8b) so assembly order is correct.
	keySegs := path.TrimMove(keyPath.Segments())
	if len(keySegs) > 0 {
		ks, ke := path.GetStartAndEnd(keySegs)
		if ks.X > ke.X {
			keyPath, err = PathReverse{}.PathTransform(keyPath)
			if err != nil {
				t.Fatalf("reverse key: %v", err)
			}
		}
	}

	// Build full edge: left straight + key + right straight (same as keyed_edge step 11).
	offset := (edgeLen - keyWidth) / 2
	if offset < 0 {
		offset = 0
	}
	leftDraw := path.NewDraw()
	leftDraw.MoveTo(path.NewPoint(0, 0))
	leftDraw.LineTo(path.NewPoint(offset, 0))

	rightDraw := path.NewDraw()
	rightDraw.MoveTo(path.NewPoint(offset+keyWidth, 0))
	rightDraw.LineTo(path.NewPoint(edgeLen, 0))

	var allSegs []path.Segment
	allSegs = append(allSegs, leftDraw.Path().Segments()...)
	allSegs = append(allSegs, keyPath.Segments()...)
	allSegs = append(allSegs, rightDraw.Path().Segments()...)

	fullPath, err := JoinTransform{
		Precision:        precision,
		SegmentOperators: so,
	}.PathTransform(path.NewPathFromSegments(allSegs))
	if err != nil {
		t.Fatalf("join full edge: %v", err)
	}

	// Assert: result must be a single contiguous outline (no extra Move = no "closed" side).
	subpaths := path.SplitPathOnMove(fullPath)
	if len(subpaths) != 1 {
		t.Errorf("keyed edge should produce a single contiguous path; got %d subpaths (bug: one side closed). SVG:\n%s",
			len(subpaths), path.SvgString(fullPath, precision))
	}

	// Edge outline must span the full seam 0..edgeLen (left straight + key + right straight).
	// Bug: JoinTransform can connect segments in the wrong order so one side is "closed" or
	// dropped, and the path no longer reaches x=0 (or x=edgeLen).
	segs := path.TrimMove(fullPath.Segments())
	if len(segs) == 0 {
		t.Errorf("path has no segments after TrimMove")
		return
	}
	start, end := path.GetStartAndEnd(segs)
	minX := start.X
	if end.X < minX {
		minX = end.X
	}
	maxX := start.X
	if end.X > maxX {
		maxX = end.X
	}
	if minX > 1 {
		t.Errorf("keyed edge bug: path does not reach left side (x=0); minX=%.2f — one side closed or dropped. start=%v end=%v",
			minX, start, end)
	}
	if maxX < edgeLen-1 {
		t.Errorf("keyed edge bug: path does not reach right side (x=%.0f); maxX=%.2f", edgeLen, maxX)
	}
}
