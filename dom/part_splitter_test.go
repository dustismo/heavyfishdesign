package dom

import (
	"math"
	"strings"
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

// diagonalSegments returns line segments whose angle is neither ~0° nor ~90°.
func diagonalSegments(p path.Path) []path.LineSegment {
	var out []path.LineSegment
	for _, seg := range p.Segments() {
		line, ok := seg.(path.LineSegment)
		if !ok {
			continue
		}
		dx := line.End().X - line.Start().X
		dy := line.End().Y - line.Start().Y
		if dx == 0 || dy == 0 {
			continue
		}
		out = append(out, line)
	}
	return out
}

func TestJoinOutlineWithSeamNoDiagonalChords(t *testing.T) {
	so := path.NewSegmentOperators()
	precision := 3

	// Open bottom-half outline: right seam → bottom → left seam.
	outline, err := path.ParsePathFromSvg(
		"M 12 4 L 12 8 L 0.25 8 L 0.25 4",
	)
	if err != nil {
		t.Fatalf("parse outline: %v", err)
	}

	// Finger-joint-like seam with tabs (simplified zig-zag along y=4).
	seam, err := path.ParsePathFromSvg(
		"M 0.25 4 L 0.25 4.25 L 0.55 4.25 L 0.55 4 L 1.15 4 L 1.15 4.25 L 12 4.25 L 12 4",
	)
	if err != nil {
		t.Fatalf("parse seam: %v", err)
	}

	joined, err := joinOutlineWithSeam(outline, seam, 0.25, 12, 4, precision, so)
	if err != nil {
		t.Fatalf("joinOutlineWithSeam: %v", err)
	}

	diags := diagonalSegments(joined)
	if len(diags) > 0 {
		t.Fatalf("unexpected diagonal chord(s) in joined path:\n%s\ndiagonals: %d",
			path.SvgString(joined, precision), len(diags))
	}
}

func TestJoinOutlineWithSeamFingerJointProfile(t *testing.T) {
	// Minimal reproduction: outline open at seam + realistic tab profile.
	so := path.NewSegmentOperators()
	precision := 3

	outline, err := path.ParsePathFromSvg(
		"M 12 0 L 12 4 L 0.25 4 L 0.25 0",
	)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	seam, err := path.ParsePathFromSvg(
		"M 0.25 0 L 0.804 0 L 0.804 0.25 L 1.104 0.25 L 1.104 0 L 11.204 0 L 12.007 0",
	)
	if err != nil {
		t.Fatalf("parse seam: %v", err)
	}

	joined, err := joinOutlineWithSeam(outline, seam, 0.25, 12.007, 0, precision, so)
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	s := path.SvgString(joined, precision)
	if strings.Contains(s, "L 0.250 0.000 L 9.") {
		t.Fatalf("found errant diagonal chord:\n%s", s)
	}
}

func TestJoinOutlineWithSeamFlambeauBottomHalf(t *testing.T) {
	// Outline fragments from a split front_flat_top bottom half (before seam join).
	so := path.NewSegmentOperators()
	precision := 3

	fragA, err := path.ParsePathFromSvg(
		"M 9.607 4.011 L 9.607 3.760 L 9.904 3.760 L 10.200 3.760 L 10.200 4.011 L 10.907 4.011 L 10.907 3.760 L 11.204 3.760 L 11.757 3.760 L 11.757 2.482 L 11.757 2.185 L 12.007 2.185 L 12.007 1.478 L 11.757 1.478 L 11.757 1.182 L 11.757 0.885 L 12.007 0.885 L 12.007 0.178 L 11.757 0.178 L 11.757 0.000",
	)
	if err != nil {
		t.Fatalf("parse fragA: %v", err)
	}
	fragB, err := path.ParsePathFromSvg(
		"M 0.250 0.000 L 0.250 0.178 L 0.000 0.178 L 0.000 0.885 L 0.250 0.885 L 0.250 1.182 L 0.250 1.478 L 0.000 1.478 L 0.000 2.185 L 0.250 2.185 L 0.250 2.482 L 0.250 3.760 L 0.803 3.760 L 1.100 3.760 L 1.100 4.011 L 1.807 4.011 L 1.807 3.760 L 2.103 3.760 L 2.400 3.760 L 2.400 4.011 L 3.107 4.011 L 3.107 3.760 L 3.403 3.760 L 3.700 3.760 L 3.700 4.011 L 4.407 4.011 L 4.407 3.760 L 4.704 3.760 L 5.000 3.760 L 5.000 4.011 L 5.707 4.011 L 5.707 3.760 L 6.003 3.760 L 6.300 3.760 L 6.300 4.011 L 7.007 4.011 L 7.007 3.760 L 7.303 3.760 L 7.600 3.760 L 7.600 4.011 L 8.307 4.011 L 8.307 3.760 L 8.604 3.760 L 8.900 3.760 L 8.900 4.011 L 9.607 4.011",
	)
	if err != nil {
		t.Fatalf("parse fragB: %v", err)
	}
	seam, err := path.ParsePathFromSvg(
		"M 0.250 4.000 L 0.000 4.000 L 0.804 4.000 L 1.104 4.000 L 1.104 4.250 L 1.804 4.250 L 1.804 4.000 L 2.104 4.000 L 2.404 4.000 L 2.404 4.250 L 3.103 4.250 L 3.103 4.000 L 3.404 4.000 L 3.704 4.000 L 3.704 4.250 L 4.404 4.250 L 4.404 4.000 L 4.704 4.000 L 5.004 4.000 L 5.004 4.250 L 5.704 4.250 L 5.704 4.000 L 6.004 4.000 L 6.303 4.000 L 6.303 4.250 L 7.004 4.250 L 7.004 4.000 L 7.303 4.000 L 7.604 4.000 L 7.604 4.250 L 8.303 4.250 L 8.303 4.000 L 8.604 4.000 L 8.904 4.000 L 8.904 4.250 L 9.604 4.250 L 9.604 4.000 L 9.904 4.000 L 10.204 4.000 L 10.204 4.250 L 10.903 4.250 L 10.903 4.000 L 11.204 4.000 L 12.007 4.000 L 11.757 4.000",
	)
	if err != nil {
		t.Fatalf("parse seam: %v", err)
	}

	combined := transforms.SimpleJoin{}.JoinPaths(fragA, fragB)
	joined, err := joinOutlineWithSeam(combined, seam, 0.25, 11.757, 4, precision, so)
	if err != nil {
		t.Fatalf("joinOutlineWithSeam: %v", err)
	}

	diags := diagonalSegments(joined)
	if len(diags) > 0 {
		t.Fatalf("unexpected diagonal chord(s):\n%s", path.SvgString(joined, precision))
	}
	s := path.SvgString(joined, precision)
	if strings.Contains(s, "L 0.250 0.000 L 9.") {
		t.Fatalf("found errant diagonal chord:\n%s", s)
	}
	// fragA and fragB are disconnected pre-join fragments; merged output may be
	// multiple subpaths until joined — assert no stray diagonals only.
}

func TestPartitionPerimeterStraightLine(t *testing.T) {
	so := path.NewSegmentOperators()
	outline, err := path.ParsePathFromSvg("M 13.507 0 L 13.507 6.003 L 0 6.004 L 0 0")
	if err != nil {
		t.Fatal(err)
	}
	tl, br, err := path.BoundingBoxTrimWhitespace(outline, so)
	if err != nil {
		t.Fatal(err)
	}
	perimeter, slots := partitionPerimeterSegments(path.TrimMove(outline.Segments()), tl, br, 3)
	if len(slots) != 0 {
		t.Fatalf("expected no slots, got %d slots and %d perimeter segs", len(slots), len(perimeter))
	}
	if len(perimeter) != 3 {
		t.Fatalf("expected 3 perimeter segs, got %d", len(perimeter))
	}
}

func TestReplaceSeamRunStraightLine(t *testing.T) {
	so := path.NewSegmentOperators()
	precision := 3
	outline, err := path.ParsePathFromSvg("M 13.507 0 L 13.507 6.003 L 0 6.004 L 0 0")
	if err != nil {
		t.Fatal(err)
	}
	segs := path.TrimMove(outline.Segments())
	left := path.NewPoint(0, 6)
	right := path.NewPoint(13.507, 6)
	seam, err := path.ParsePathFromSvg("M 0 6 L 13.507 6")
	if err != nil {
		t.Fatal(err)
	}
	seamChain := path.TrimMove(seam.Segments())

	out := replaceSeamRun(segs, left, right, 6, seamChain, precision, so)
	if out == nil {
		t.Fatal("replaceSeamRun returned nil")
	}
	s := path.SvgString(path.NewPathFromSegmentsWithoutMove(out), precision)
	if strings.Contains(s, "6.003 L 13.507") || strings.Contains(s, "6.004 L 13.507") {
		t.Fatalf("horizontal seam row not replaced:\n%s", s)
	}
}

func TestJoinOutlineWithSeamStraightLine(t *testing.T) {
	so := path.NewSegmentOperators()
	precision := 3

	outline, err := path.ParsePathFromSvg(
		"M 13.507 0 L 13.507 6.003 L 0 6.004 L 0 0 L 13.507 0",
	)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	seam, err := path.ParsePathFromSvg("M 0 6 L 13.507 6")
	if err != nil {
		t.Fatalf("parse seam: %v", err)
	}

	joined, err := joinOutlineWithSeam(outline, seam, 0, 13.507, 6, precision, so)
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	s := path.SvgString(joined, precision)
	if strings.Count(s, "M ") > 1 {
		t.Fatalf("expected single subpath, got:\n%s", s)
	}
	if !strings.Contains(s, "L 13.507 6.000 L 0.000 6.000") &&
		!strings.Contains(s, "L 0.000 6.000 L 13.507 6.000") {
		t.Fatalf("missing straight seam:\n%s", s)
	}
}

// fingerJointWall is a closed 6×3.5 outline with vertical finger notches on
// both sides and horizontal fingers on the top and bottom, matching a
// full-height box wall before it is split.
func fingerJointWall(t *testing.T) path.Path {
	t.Helper()
	p, err := path.ParsePathFromSvg(strings.Join([]string{
		"M 0 0",
		"L 0.5 0 L 0.5 0.25 L 1 0.25 L 1 0",
		"L 6 0",
		"L 6 0.4 L 5.75 0.4 L 5.75 0.9 L 6 0.9",
		"L 6 1.4 L 5.75 1.4 L 5.75 1.9 L 6 1.9",
		"L 6 3.5",
		"L 5.5 3.5 L 5.5 3.25 L 5 3.25 L 5 3.5",
		"L 0 3.5",
		"L 0 1.9 L 0.25 1.9 L 0.25 1.4 L 0 1.4",
		"L 0 0.9 L 0.25 0.9 L 0.25 0.4 L 0 0.4",
		"L 0 0",
	}, " "))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestStraightSplitThroughVerticalFingers(t *testing.T) {
	so := path.NewSegmentOperators()
	precision := 3
	wall := fingerJointWall(t)
	seamY := 1.5

	sliced, err := hSlicePath(wall, seamY, so, precision)
	if err != nil {
		t.Fatal(err)
	}

	joined, err := closeSlicedOutline(sliced, seamY, precision, so)
	if err != nil {
		t.Fatal(err)
	}
	s := path.SvgString(joined, precision)
	tl, br, err := path.BoundingBoxTrimWhitespace(joined, so)
	if err != nil {
		t.Fatal(err)
	}
	if h := br.Y - tl.Y; h < 1.4 || h > 1.6 {
		t.Fatalf("top half height = %.3f, want ~1.5\n%s", h, s)
	}
	if strings.Count(s, "M ") > 1 {
		t.Fatalf("top half is not one closed outline:\n%s", s)
	}
	if !strings.Contains(s, "L 0.500 0.000") {
		t.Fatalf("top half lost the top finger edge:\n%s", s)
	}
	if !strings.Contains(s, "5.750 1.500") || !strings.Contains(s, "0.250 1.500") {
		t.Fatalf("top half seam does not meet the cut fingers:\n%s", s)
	}
	if strings.Contains(s, "6.000 1.500") || strings.Contains(s, "0.000 1.500") {
		t.Fatalf("top half squared off the finger the cut passes through:\n%s", s)
	}
	if len(diagonalSegments(joined)) > 0 {
		t.Fatalf("top half has a diagonal seam:\n%s", s)
	}

	// Bottom half: mirror, slice, mirror back, join at the original seam Y
	// the way PartSplitter.transformPart does.
	flipped, err := transforms.MirrorTransform{
		Axis:             transforms.Horizontal,
		Handle:           path.TopLeft,
		SegmentOperators: so,
	}.PathTransform(wall)
	if err != nil {
		t.Fatal(err)
	}
	_, brWall, err := path.BoundingBoxTrimWhitespace(wall, so)
	if err != nil {
		t.Fatal(err)
	}
	bottomY := brWall.Y - seamY
	bottomSliced, err := hSlicePath(flipped, bottomY, so, precision)
	if err != nil {
		t.Fatal(err)
	}
	bottomOutline, err := transforms.MirrorTransform{
		Axis:             transforms.Horizontal,
		Handle:           path.TopLeft,
		SegmentOperators: so,
	}.PathTransform(bottomSliced)
	if err != nil {
		t.Fatal(err)
	}
	joinedBottom, err := closeSlicedOutline(bottomOutline, math.NaN(), precision, so)
	if err != nil {
		t.Fatal(err)
	}
	bs := path.SvgString(joinedBottom, precision)
	btl, bbr, err := path.BoundingBoxTrimWhitespace(joinedBottom, so)
	if err != nil {
		t.Fatal(err)
	}
	if h := bbr.Y - btl.Y; h < 1.9 || h > 2.1 {
		t.Fatalf("bottom half height = %.3f, want ~2.0\n%s", h, bs)
	}
	if strings.Count(bs, "M ") > 1 {
		t.Fatalf("bottom half is not one closed outline:\n%s", bs)
	}
	if !strings.Contains(bs, "5.500 2.000") || !strings.Contains(bs, "5.000 1.750") {
		t.Fatalf("bottom half lost the bottom finger edge:\n%s", bs)
	}
	if !strings.Contains(bs, "5.750 0.000") || !strings.Contains(bs, "0.250 0.000") {
		t.Fatalf("bottom half seam does not meet the cut fingers:\n%s", bs)
	}
	if strings.Contains(bs, "1.500") {
		t.Fatalf("bottom half seam was drawn through the middle:\n%s", bs)
	}
	if len(diagonalSegments(joinedBottom)) > 0 {
		t.Fatalf("bottom half has a diagonal seam:\n%s", bs)
	}
}

func TestCloseSlicedOutlineSnapsBleedToSeam(t *testing.T) {
	so := path.NewSegmentOperators()
	precision := 3
	rect, err := path.ParsePathFromSvg("M 0 0 L 10 0 L 10 8 L 0 8 L 0 0")
	if err != nil {
		t.Fatal(err)
	}
	sliced, err := hSlicePath(rect, 4.0035, so, precision)
	if err != nil {
		t.Fatal(err)
	}
	closed, err := closeSlicedOutline(sliced, 4, precision, so)
	if err != nil {
		t.Fatal(err)
	}
	s := path.SvgString(closed, precision)
	if strings.Count(s, "M ") > 1 {
		t.Fatalf("expected one outline:\n%s", s)
	}
	if !strings.Contains(s, "10.000 4.000") || !strings.Contains(s, "0.000 4.000") {
		t.Fatalf("seam not snapped to y=4:\n%s", s)
	}
	if strings.Contains(s, "4.003") || strings.Contains(s, "4.004") {
		t.Fatalf("bleed stub left on the seam:\n%s", s)
	}
}

func TestCloseSlicedOutlineWithSeamKeepsFingers(t *testing.T) {
	so := path.NewSegmentOperators()
	precision := 3
	// Open top half: the knife stopped on the finger faces at x=0.25 and x=5.75.
	sliced, err := path.ParsePathFromSvg(
		"M 0.25 1.5 L 0.25 0.5 L 0 0.5 L 0 0 L 6 0 L 6 0.5 L 5.75 0.5 L 5.75 1.5",
	)
	if err != nil {
		t.Fatal(err)
	}
	seam, err := path.ParsePathFromSvg("M 0.25 1.5 L 3 1.9 L 5.75 1.5")
	if err != nil {
		t.Fatal(err)
	}
	chain, err := slicedOpenChain(sliced, 1.5, precision, so)
	if err != nil {
		t.Fatal(err)
	}
	start, end := pathEndpoints(chain)
	chain = append(chain, orientSeamOnto(path.TrimMove(seam.Segments()), end, start, so)...)
	closed, err := transforms.CleanupTransform{Precision: precision}.PathTransform(
		path.NewPathFromSegmentsWithoutMove(chain))
	if err != nil {
		t.Fatal(err)
	}
	s := path.SvgString(closed, precision)
	if strings.Count(s, "M ") > 1 {
		t.Fatalf("expected one outline:\n%s", s)
	}
	if !strings.Contains(s, "3.000 1.900") {
		t.Fatalf("curve was dropped:\n%s", s)
	}
	if !strings.Contains(s, "0.000 0.000") || !strings.Contains(s, "0.250 0.500") {
		t.Fatalf("finger beside the cut was dropped:\n%s", s)
	}
	if strings.Contains(s, "0.000 1.500") || strings.Contains(s, "6.000 1.500") {
		t.Fatalf("seam squared off to the outer corners:\n%s", s)
	}
}

func TestPruneHolePathKeepsSmallClosedHole(t *testing.T) {
	// 1/8 in hole approximated by chords of ~0.088, under minHoleSegmentLength.
	hole, err := path.ParsePathFromSvg(
		"M 0 0.0625 L 0.0625 0 L 0.125 0.0625 L 0.0625 0.125 L 0 0.0625",
	)
	if err != nil {
		t.Fatal(err)
	}
	kept, ok := pruneHolePath(hole, 3)
	if !ok {
		t.Fatal("small closed hole was dropped")
	}
	if len(path.TrimMove(kept.Segments())) == 0 {
		t.Fatal("kept hole has no segments")
	}
	stub, err := path.ParsePathFromSvg("M 0 0 L 0.02 0.01")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := pruneHolePath(stub, 3); ok {
		t.Fatal("open sliver was kept")
	}
}

func TestOrientPathLeftToRight(t *testing.T) {
	so := path.NewSegmentOperators()
	p, _ := path.ParsePathFromSvg("M 10 0 L 0 0")
	out, err := orientPathLeftToRight(p, 3, so)
	if err != nil {
		t.Fatal(err)
	}
	s, e := path.GetStartAndEnd(path.TrimMove(out.Segments()))
	if s.X > e.X {
		t.Fatalf("expected left-to-right, got start=%v end=%v", s, e)
	}
}
