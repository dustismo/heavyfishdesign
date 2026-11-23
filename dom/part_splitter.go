package dom

import (
	"fmt"
	"math"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
)

type PartSplitter struct {
	mp *dynmap.DynMap
}

func orientPathLeftToRight(p path.Path, precision int, so path.SegmentOperators) (path.Path, error) {
	segs := path.TrimMove(p.Segments())
	if len(segs) == 0 {
		return p, nil
	}
	s, e := pathEndpoints(segs)
	if s.X > e.X {
		return transforms.PathReverse{}.PathTransform(p)
	}
	return p, nil
}

const seamEndpointSlop = 0.5

func pathEndpoints(segs []path.Segment) (start, end path.Point) {
	if len(segs) == 0 {
		return
	}
	start = segs[0].Start()
	end = segs[len(segs)-1].End()
	return
}

func seamRowExtents(outline path.Path, seamY float64) (leftX, rightX float64, ok bool) {
	minX := 0.0
	maxX := 0.0
	found := false
	for _, seg := range path.TrimMove(outline.Segments()) {
		for _, pt := range []path.Point{seg.Start(), seg.End()} {
			if pt.Y < seamY-seamEndpointSlop || pt.Y > seamY+seamEndpointSlop {
				continue
			}
			if !found {
				minX, maxX = pt.X, pt.X
				found = true
			} else {
				if pt.X < minX {
					minX = pt.X
				}
				if pt.X > maxX {
					maxX = pt.X
				}
			}
		}
	}
	return minX, maxX, found && maxX-minX > seamEndpointSlop
}

func seamSide(pt, left, right path.Point) string {
	dLeft := path.Distance(pt, left)
	dRight := path.Distance(pt, right)
	if dLeft <= dRight && dLeft <= seamEndpointSlop {
		return "left"
	}
	if dRight < dLeft && dRight <= seamEndpointSlop {
		return "right"
	}
	if pt.X <= left.X+seamEndpointSlop && math.Abs(pt.Y-left.Y) <= seamEndpointSlop {
		return "left"
	}
	if pt.X >= right.X-seamEndpointSlop && math.Abs(pt.Y-right.Y) <= seamEndpointSlop {
		return "right"
	}
	return ""
}

func chainIsOnlySeamRow(segs []path.Segment, seamY, leftX, rightX float64) bool {
	if len(segs) == 0 {
		return false
	}
	for _, seg := range segs {
		if !segmentTouchesSeamRow(seg, seamY, leftX, rightX) {
			return false
		}
	}
	return true
}

func appendOutlineAndSeam(outlineSegs, seamChain []path.Segment, startSide, endSide string, seamY float64, precision int, so path.SegmentOperators) []path.Segment {
	stub := func(from, to path.Point) []path.Segment {
		return seamStubSegments(from, to, seamY, precision)
	}
	switch {
	case endSide == "left" && startSide == "right":
		out := append([]path.Segment{}, outlineSegs...)
		if len(out) > 0 && len(seamChain) > 0 {
			out = append(out, stub(out[len(out)-1].End(), seamChain[0].Start())...)
		}
		out = append(out, seamChain...)
		if len(outlineSegs) > 0 && len(seamChain) > 0 {
			out = append(out, stub(seamChain[len(seamChain)-1].End(), outlineSegs[0].Start())...)
		}
		return out
	case endSide == "right" && startSide == "left":
		seam := reverseSegmentChain(seamChain, so)
		out := append([]path.Segment{}, outlineSegs...)
		if len(out) > 0 && len(seam) > 0 {
			out = append(out, stub(out[len(out)-1].End(), seam[0].Start())...)
		}
		out = append(out, seam...)
		if len(outlineSegs) > 0 && len(seam) > 0 {
			out = append(out, stub(seam[len(seam)-1].End(), outlineSegs[0].Start())...)
		}
		return out
	default:
		return nil
	}
}

// joinOutlineWithSeam stitches an HSlice'd outline to a seam edge profile without
// SimpleJoin Move gaps (which cause ReorderTransform to pick wrong subpaths and
// JoinTransform to draw diagonal connector chords).
func joinOutlineWithSeam(outline, seam path.Path, leftX, rightX, seamY float64, precision int, so path.SegmentOperators) (path.Path, error) {
	seam, err := orientPathLeftToRight(seam, precision, so)
	if err != nil {
		return nil, err
	}
	edgeSegs := path.TrimMove(seam.Segments())
	if len(edgeSegs) == 0 {
		return outline, nil
	}
	es, ee := pathEndpoints(edgeSegs)

	left := path.NewPoint(leftX, seamY)
	right := path.NewPoint(rightX, seamY)
	if rowLeft, rowRight, ok := seamRowExtents(outline, seamY); ok {
		left = path.NewPoint(rowLeft, seamY)
		right = path.NewPoint(rowRight, seamY)
	}
	seamChain := buildSeamChain(left, right, edgeSegs, es, ee)

	tl, br, err := path.BoundingBoxTrimWhitespace(outline, so)
	if err != nil {
		return nil, err
	}
	perimeterSegs, slotSegs := partitionPerimeterSegments(path.TrimMove(outline.Segments()), tl, br, precision)

	straightSeam := seamIsStraightLine(seamChain, precision)
	ordered := joinPerimeterWithSeam(perimeterSegs, seamChain, left, right, seamY, straightSeam, precision, so)
	if ordered == nil {
		joined, err := legacyJoinOutlineWithSeam(outline, seam, precision, so)
		if err != nil {
			return nil, err
		}
		joined, err = appendSlotSegments(joined, slotSegs, precision)
		if err != nil {
			return nil, err
		}
		return scrubOrphanMoves(joined), nil
	}

	joined, err := transforms.CleanupTransform{Precision: precision}.PathTransform(
		path.NewPathFromSegmentsWithoutMove(ordered))
	if err != nil {
		return nil, err
	}
	joined, err = appendSlotSegments(joined, slotSegs, precision)
	if err != nil {
		return nil, err
	}
	return scrubOrphanMoves(joined), nil
}

// legacyJoinOutlineWithSeam is the pre-rewrite join path (SimpleJoin + JoinTransform).
// Used when joinPerimeterWithSeam cannot stitch the outline — keyed edges and simple
// trapezoids work here; finger-joint flambeau panels must succeed via joinPerimeterWithSeam.
func legacyJoinOutlineWithSeam(outline, seam path.Path, precision int, so path.SegmentOperators) (path.Path, error) {
	combined := transforms.SimpleJoin{}.JoinPaths(outline, seam)
	joined, err := transforms.JoinTransform{
		Precision:        precision,
		SegmentOperators: so,
		ClosePath:        true,
	}.PathTransform(combined)
	if err != nil {
		return joined, err
	}
	return closePathEndpoints(joined, precision), nil
}

func closePathEndpoints(p path.Path, precision int) path.Path {
	segs := path.TrimMove(p.Segments())
	if len(segs) == 0 {
		return p
	}
	start, end := pathEndpoints(segs)
	if start.EqualsPrecision(end, precision) {
		return path.NewPathFromSegmentsWithoutMove(segs)
	}
	segs = append(segs, path.LineSegment{StartPoint: end, EndPoint: start})
	return path.NewPathFromSegmentsWithoutMove(segs)
}

func joinPerimeterWithSeam(
	perimeterSegs, seamChain []path.Segment,
	left, right path.Point,
	seamY float64,
	straightSeam bool,
	precision int,
	so path.SegmentOperators,
) []path.Segment {
	if len(perimeterSegs) == 0 {
		return nil
	}

	candidates := seamJoinCandidates(perimeterSegs, left, right, seamY, precision, so)
	for _, cand := range candidates {
		cand = dropSeamRowVerticalStubs(cand, left, right, seamY)
		if straightSeam {
			if out := replaceStraightSeamRun(cand, left, right, seamY, seamChain, precision, so); out != nil {
				return closeTopEdgeIfNeeded(out, seamY)
			}
		}
		if out := replaceSeamRun(cand, left, right, seamY, seamChain, precision, so); out != nil {
			return closeTopEdgeIfNeeded(out, seamY)
		}
	}

	for _, cand := range candidates {
		cand = dropSeamRowVerticalStubs(cand, left, right, seamY)
		if hasSeamRowHorizontal(cand, seamY, left.X, right.X) {
			continue
		}
		if out := tryJoinPerimeterWithSeam(cand, seamChain, left, right, seamY, precision, so); out != nil {
			return closeTopEdgeIfNeeded(out, seamY)
		}
	}

	merged := mergeConnectedSegments(perimeterSegs, precision, so, seamY, left, right)
	merged = dropSeamRowVerticalStubs(merged, left, right, seamY)
	if straightSeam {
		if out := replaceStraightSeamRun(merged, left, right, seamY, seamChain, precision, so); out != nil {
			return closeTopEdgeIfNeeded(out, seamY)
		}
		if out := stitchStraightSeamFallback(merged, seamChain, left, right, seamY, precision, so); len(out) > 0 {
			return closeTopEdgeIfNeeded(out, seamY)
		}
	}
	if out := replaceSeamRun(merged, left, right, seamY, seamChain, precision, so); out != nil {
		return closeTopEdgeIfNeeded(out, seamY)
	}
	if !hasSeamRowHorizontal(merged, seamY, left.X, right.X) {
		if out := tryJoinPerimeterWithSeam(merged, seamChain, left, right, seamY, precision, so); out != nil {
			return closeTopEdgeIfNeeded(out, seamY)
		}
		rotated := rotateChainToStartNear(merged, left, precision, so)
		if out := tryJoinPerimeterWithSeam(rotated, seamChain, left, right, seamY, precision, so); out != nil {
			return closeTopEdgeIfNeeded(out, seamY)
		}
	}
	return nil
}

func hasSeamRowHorizontal(segs []path.Segment, seamY, leftX, rightX float64) bool {
	for _, seg := range segs {
		if segmentTouchesSeamRow(seg, seamY, leftX, rightX) {
			return true
		}
	}
	return false
}

func seamJoinCandidates(segs []path.Segment, left, right path.Point, seamY float64, precision int, so path.SegmentOperators) [][]path.Segment {
	seen := make(map[string]bool)
	var out [][]path.Segment
	push := func(cand []path.Segment) {
		if len(cand) == 0 {
			return
		}
		key := path.SvgString(path.NewPathFromSegmentsWithoutMove(cand), precision)
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, cand)
	}

	merged := mergeConnectedSegments(segs, precision, so, seamY, left, right)
	if segmentChainIsConnected(segs, precision) {
		push(segs)
		push(rotateChainToStartNear(segs, left, precision, so))
		push(rotateChainToStartNear(segs, right, precision, so))
	} else if len(merged) > 0 {
		push(merged)
		if segmentChainIsConnected(merged, precision) {
			push(rotateChainToStartNear(merged, left, precision, so))
			push(rotateChainToStartNear(merged, right, precision, so))
		}
	}
	if rev := reverseSegmentChain(segs, so); segmentChainIsConnected(rev, precision) {
		push(rev)
		push(rotateChainToStartNear(rev, left, precision, so))
		push(rotateChainToStartNear(rev, right, precision, so))
	}
	return out
}

func tryJoinPerimeterWithSeam(
	segs, seamChain []path.Segment,
	left, right path.Point,
	seamY float64,
	precision int,
	so path.SegmentOperators,
) []path.Segment {
	if chainIsOnlySeamRow(segs, seamY, left.X, right.X) {
		return nil
	}
	oStart, oEnd := pathEndpoints(segs)
	if oStart.EqualsPrecision(oEnd, precision) {
		return joinClosedOutlineAtSeam(segs, left, right, seamChain, so)
	}
	ss, es := seamSide(oStart, left, right), seamSide(oEnd, left, right)
	if ss != "" && es != "" && ss != es {
		return appendOutlineAndSeam(segs, seamChain, ss, es, seamY, precision, so)
	}
	return nil
}

func segmentChainIncludesSeam(chain, seamChain []path.Segment, precision int) bool {
	if len(seamChain) == 0 {
		return true
	}
	seamStart := seamChain[0].Start()
	seamEnd := seamChain[len(seamChain)-1].End()
	for _, seg := range chain {
		for _, pt := range []path.Point{seg.Start(), seg.End()} {
			if pt.EqualsPrecision(seamStart, precision) || pt.EqualsPrecision(seamEnd, precision) {
				return true
			}
		}
	}
	return false
}

func stitchStraightSeamFallback(outlineSegs, seamChain []path.Segment, left, right path.Point, seamY float64, precision int, so path.SegmentOperators) []path.Segment {
	if len(outlineSegs) == 0 || len(seamChain) == 0 {
		return append(append([]path.Segment{}, outlineSegs...), seamChain...)
	}
	chain := append([]path.Segment{}, outlineSegs...)
	if line, ok := chain[len(chain)-1].(path.LineSegment); ok {
		line.EndPoint = snapCornerPointToSeamRow(line.End(), left, right, seamY)
		chain[len(chain)-1] = line
	}
	seam := seamChain
	end := chain[len(chain)-1].End()
	if path.Distance(end, right) <= path.Distance(end, left) {
		seam = reverseSegmentChain(seamChain, so)
	}
	chain = append(chain, seamStubSegments(end, seam[0].Start(), seamY, precision)...)
	chain = append(chain, seam...)
	start := outlineSegs[0].Start()
	chain = append(chain, seamStubSegments(seam[len(seam)-1].End(), start, seamY, precision)...)
	return chain
}

func joinClosedOutlineAtSeam(outlineSegs []path.Segment, left, right path.Point, seamChain []path.Segment, so path.SegmentOperators) []path.Segment {
	if len(outlineSegs) == 0 {
		return nil
	}
	rightIdx := -1
	leftIdx := -1
	rightDist := seamEndpointSlop
	leftDist := seamEndpointSlop
	for i, seg := range outlineSegs {
		for _, pt := range []path.Point{seg.Start(), seg.End()} {
			if d := path.Distance(pt, right); d <= rightDist {
				rightDist = d
				rightIdx = i
			}
			if d := path.Distance(pt, left); d <= leftDist {
				leftDist = d
				leftIdx = i
			}
		}
	}
	if rightIdx < 0 || leftIdx < 0 || rightIdx == leftIdx {
		return nil
	}
	rotated := append([]path.Segment{}, outlineSegs[rightIdx:]...)
	rotated = append(rotated, outlineSegs[:rightIdx]...)
	endIdx := (leftIdx - rightIdx + len(outlineSegs)) % len(outlineSegs)
	rotated = rotated[:endIdx+1]

	oStart, oEnd := pathEndpoints(rotated)
	if seamSide(oStart, left, right) != "right" || seamSide(oEnd, left, right) != "left" {
		rotated = append([]path.Segment{}, outlineSegs[leftIdx:]...)
		rotated = append(rotated, outlineSegs[:leftIdx]...)
		endIdx = (rightIdx - leftIdx + len(outlineSegs)) % len(outlineSegs)
		rotated = rotated[:endIdx+1]
		oStart, oEnd = pathEndpoints(rotated)
		if seamSide(oStart, left, right) != "left" || seamSide(oEnd, left, right) != "right" {
			return nil
		}
		return append(append([]path.Segment{}, rotated...), reverseSegmentChain(seamChain, so)...)
	}
	return append(append([]path.Segment{}, rotated...), seamChain...)
}

func buildSeamChain(left, right path.Point, edgeSegs []path.Segment, es, ee path.Point) []path.Segment {
	chain := []path.Segment{}
	if !left.Equals(es) {
		chain = append(chain, path.LineSegment{StartPoint: left, EndPoint: es})
	}
	chain = append(chain, edgeSegs...)
	if !ee.Equals(right) {
		chain = append(chain, path.LineSegment{StartPoint: ee, EndPoint: right})
	}
	return chain
}

func reverseSegmentChain(segs []path.Segment, so path.SegmentOperators) []path.Segment {
	out := make([]path.Segment, 0, len(segs))
	for i := len(segs) - 1; i >= 0; i-- {
		rev, err := so.Reverse(segs[i])
		if err != nil {
			rev = segs[i]
		}
		out = append(out, rev)
	}
	return out
}

func mergeConnectedSegments(segs []path.Segment, precision int, so path.SegmentOperators, seamY float64, left, right path.Point) []path.Segment {
	segs = path.TrimMove(segs)
	if len(segs) == 0 {
		return nil
	}

	tryAttach := func(chain []path.Segment, s path.Segment) ([]path.Segment, path.Segment, bool) {
		if len(chain) == 0 {
			return []path.Segment{s}, s, true
		}
		start, end := pathEndpoints(chain)
		if end.EqualsPrecision(s.Start(), precision) {
			return append(chain, s), s, true
		}
		if end.EqualsPrecision(s.End(), precision) {
			rev, err := so.Reverse(s)
			if err != nil {
				rev = s
			}
			return append(chain, rev), rev, true
		}
		if start.EqualsPrecision(s.End(), precision) {
			rev, err := so.Reverse(s)
			if err != nil {
				rev = s
			}
			return append([]path.Segment{rev}, chain...), rev, true
		}
		if start.EqualsPrecision(s.Start(), precision) {
			return append([]path.Segment{s}, chain...), s, true
		}
		return chain, s, false
	}

	buildFromStart := func(startIdx int) []path.Segment {
		used := make([]bool, len(segs))
		chain := []path.Segment{segs[startIdx]}
		used[startIdx] = true
		for {
			attached := false
			for i, s := range segs {
				if used[i] {
					continue
				}
				var ok bool
				chain, _, ok = tryAttach(chain, s)
				if ok {
					used[i] = true
					attached = true
					break
				}
			}
			if !attached {
				break
			}
		}
		return chain
	}

	scoreChain := func(chain []path.Segment) int {
		if chainIsOnlySeamRow(chain, seamY, left.X, right.X) {
			return -1
		}
		score := len(chain)
		if len(chain) != len(segs) {
			return score
		}
		s0, e0 := pathEndpoints(chain)
		ss, es := seamSide(s0, left, right), seamSide(e0, left, right)
		if (ss == "left" && es == "right") || (ss == "right" && es == "left") {
			score += 10000
		} else if ss != "" && es != "" {
			score += 5000
		} else if math.Abs(s0.Y-seamY) <= seamEndpointSlop && math.Abs(e0.Y-seamY) <= seamEndpointSlop {
			score += 3000
		} else if s0.Y <= seamEndpointSlop && e0.Y <= seamEndpointSlop && e0.X > s0.X+1 {
			score += 2000
		}
		return score
	}

	best := buildFromStart(0)
	bestScore := scoreChain(best)
	for i := 1; i < len(segs); i++ {
		chain := buildFromStart(i)
		if sc := scoreChain(chain); sc > bestScore {
			best = chain
			bestScore = sc
		}
	}
	return best
}

func isInteriorHorizontalSlot(seg path.Segment, tl, br path.Point) bool {
	s, e := seg.Start(), seg.End()
	dy := math.Abs(s.Y - e.Y)
	dx := math.Abs(s.X - e.X)
	if dy > 0.05 || dx < 2.0 {
		return false
	}
	if math.Abs(s.Y-tl.Y) <= seamEndpointSlop && math.Abs(e.Y-tl.Y) <= seamEndpointSlop {
		return false
	}
	if math.Abs(s.Y-br.Y) <= seamEndpointSlop && math.Abs(e.Y-br.Y) <= seamEndpointSlop {
		return false
	}
	width := br.X - tl.X
	return dx >= width*0.5-seamEndpointSlop
}

func isInteriorVerticalSlot(seg path.Segment, tl, br path.Point) bool {
	s, e := seg.Start(), seg.End()
	dx := math.Abs(s.X - e.X)
	dy := math.Abs(s.Y - e.Y)
	if dx > 0.05 || dy < 2.0 {
		return false
	}
	if math.Abs(s.X-tl.X) <= seamEndpointSlop && math.Abs(e.X-tl.X) <= seamEndpointSlop {
		return false
	}
	if math.Abs(s.X-br.X) <= seamEndpointSlop && math.Abs(e.X-br.X) <= seamEndpointSlop {
		return false
	}
	height := br.Y - tl.Y
	return dy >= height*0.5-seamEndpointSlop
}

func partitionPerimeterSegments(segs []path.Segment, tl, br path.Point, precision int) (perimeter, slots []path.Segment) {
	for _, s := range segs {
		if isInteriorHorizontalSlot(s, tl, br) || isInteriorVerticalSlot(s, tl, br) {
			slots = append(slots, s)
		} else {
			perimeter = append(perimeter, s)
		}
	}
	return perimeter, slots
}

func ensureOppositeSeamEndpoints(segs []path.Segment, left, right path.Point, precision int, so path.SegmentOperators) []path.Segment {
	if len(segs) == 0 {
		return segs
	}
	if !segmentChainIsConnected(segs, precision) {
		return segs
	}
	s0, e0 := pathEndpoints(segs)
	ss, es := seamSide(s0, left, right), seamSide(e0, left, right)
	if ss != "" && es != "" && ss != es {
		return segs
	}
	if ss != "" && es != "" && ss == es {
		return rotateChainToStartNear(segs, left, precision, so)
	}
	return segs
}

func segmentChainIsConnected(segs []path.Segment, precision int) bool {
	if len(segs) <= 1 {
		return true
	}
	for i := 1; i < len(segs); i++ {
		prevEnd := segs[i-1].End()
		if !prevEnd.EqualsPrecision(segs[i].Start(), precision) &&
			!prevEnd.EqualsPrecision(segs[i].End(), precision) {
			return false
		}
	}
	return true
}

func isSeamRowVerticalStub(seg path.Segment, left, right path.Point, seamY float64) bool {
	s, e := seg.Start(), seg.End()
	if math.Abs(s.X-e.X) > 0.05 {
		return false
	}
	onCorner := math.Abs(s.X-left.X) <= seamEndpointSlop || math.Abs(s.X-right.X) <= seamEndpointSlop
	if !onCorner {
		return false
	}
	midY := (s.Y + e.Y) / 2
	if math.Abs(midY-seamY) > seamEndpointSlop {
		return false
	}
	return math.Abs(s.Y-e.Y) < 0.05
}

func dropSeamRowVerticalStubs(segs []path.Segment, left, right path.Point, seamY float64) []path.Segment {
	if len(segs) == 0 {
		return segs
	}
	out := make([]path.Segment, 0, len(segs))
	for _, seg := range segs {
		if isSeamRowVerticalStub(seg, left, right, seamY) {
			continue
		}
		out = append(out, seg)
	}
	return out
}

func seamIsStraightLine(seamChain []path.Segment, precision int) bool {
	if len(seamChain) != 1 {
		return false
	}
	line, ok := seamChain[0].(path.LineSegment)
	if !ok {
		return false
	}
	return math.Abs(line.Start().Y-line.End().Y) <= seamEndpointSlop
}

func appendSlotSegments(outline path.Path, slots []path.Segment, precision int) (path.Path, error) {
	if len(slots) == 0 {
		return outline, nil
	}
	slotPath, err := transforms.CleanupTransform{Precision: precision}.PathTransform(
		path.NewPathFromSegmentsWithoutMove(slots))
	if err != nil {
		return outline, err
	}
	return transforms.SimpleJoin{}.JoinPaths(outline, slotPath), nil
}

func rotateChainToStartNear(segs []path.Segment, pt path.Point, precision int, so path.SegmentOperators) []path.Segment {
	if len(segs) == 0 {
		return segs
	}
	bestI := 0
	atStart := true
	bestDist := math.MaxFloat64
	for i, seg := range segs {
		for _, useStart := range []bool{true, false} {
			p := seg.Start()
			if !useStart {
				p = seg.End()
			}
			if d := path.Distance(p, pt); d < bestDist {
				bestDist = d
				bestI = i
				atStart = useStart
			}
		}
	}
	rotated := append([]path.Segment{}, segs[bestI:]...)
	rotated = append(rotated, segs[:bestI]...)
	if !atStart {
		rev, err := so.Reverse(rotated[0])
		if err != nil {
			rev = rotated[0]
		}
		rotated[0] = rev
	}
	return rotated
}

func segmentTouchesSeamRow(seg path.Segment, seamY, leftX, rightX float64) bool {
	s, e := seg.Start(), seg.End()
	dy := math.Abs(s.Y - e.Y)
	dx := math.Abs(s.X - e.X)
	if dy <= 0.05 && dx > 2.0 {
		minX := math.Min(s.X, e.X)
		maxX := math.Max(s.X, e.X)
		atSeamY := math.Abs((s.Y+e.Y)/2-seamY) <= seamEndpointSlop
		spansSeam := minX <= leftX+seamEndpointSlop && maxX >= rightX-seamEndpointSlop
		if !(atSeamY && spansSeam) {
			return false
		}
	}
	if dy > 0.05 && dy > dx {
		return false
	}
	midY := (s.Y + e.Y) / 2
	if midY < seamY-seamEndpointSlop || midY > seamY+seamEndpointSlop {
		return false
	}
	minX := math.Min(s.X, e.X)
	maxX := math.Max(s.X, e.X)
	return maxX >= leftX-seamEndpointSlop && minX <= rightX+seamEndpointSlop
}

func seamRunScore(segs []path.Segment, first, last int, seamY float64) float64 {
	var horizontalLen float64
	for i := first; i <= last; i++ {
		s, e := segs[i].Start(), segs[i].End()
		if math.Abs(s.Y-seamY) <= seamEndpointSlop && math.Abs(e.Y-seamY) <= seamEndpointSlop {
			horizontalLen += math.Abs(e.X - s.X)
		}
	}
	return horizontalLen
}

func spanIsSeamReplaceable(segs []path.Segment, first, last int, seamY, leftX, rightX float64) bool {
	for i := first; i <= last; i++ {
		if segmentTouchesSeamRow(segs[i], seamY, leftX, rightX) {
			continue
		}
		s, e := segs[i].Start(), segs[i].End()
		midY := (s.Y + e.Y) / 2
		if math.Abs(midY-seamY) > seamEndpointSlop+0.05 {
			return false
		}
	}
	return true
}

func replaceStraightSeamRun(outlineSegs []path.Segment, left, right path.Point, seamY float64, seamChain []path.Segment, precision int, so path.SegmentOperators) []path.Segment {
	var runs [][2]int
	for i := 0; i < len(outlineSegs); {
		if !segmentTouchesSeamRow(outlineSegs[i], seamY, left.X, right.X) {
			i++
			continue
		}
		start := i
		for i < len(outlineSegs) && segmentTouchesSeamRow(outlineSegs[i], seamY, left.X, right.X) {
			i++
		}
		runs = append(runs, [2]int{start, i - 1})
	}
	if len(runs) == 0 {
		return nil
	}
	first := runs[0][0]
	last := runs[len(runs)-1][1]
	prefix := outlineSegs[:first]
	suffix := outlineSegs[last+1:]
	prefix = snapPrefixSuffixToSeamRow(prefix, left, right, seamY, precision)
	suffix = snapSuffixStartToSeamRow(suffix, left, right, seamY, precision)

	seam := seamChain
	if len(prefix) > 0 {
		pe := prefix[len(prefix)-1].End()
		if path.Distance(pe, right) < path.Distance(pe, left) {
			seam = reverseSegmentChain(seamChain, so)
		}
	} else if len(suffix) > 0 {
		ss := suffix[0].Start()
		if path.Distance(ss, right) < path.Distance(ss, left) {
			seam = reverseSegmentChain(seamChain, so)
		}
	}

	out := append([]path.Segment{}, prefix...)
	if len(prefix) > 0 && len(seam) > 0 {
		out = append(out, seamStubSegments(prefix[len(prefix)-1].End(), seam[0].Start(), seamY, precision)...)
	}
	out = append(out, seam...)
	if len(suffix) > 0 && len(seam) > 0 {
		out = append(out, seamStubSegments(seam[len(seam)-1].End(), suffix[0].Start(), seamY, precision)...)
	}
	out = append(out, suffix...)
	return out
}

func replaceSeamRun(outlineSegs []path.Segment, left, right path.Point, seamY float64, seamChain []path.Segment, precision int, so path.SegmentOperators) []path.Segment {
	var runs [][2]int
	for i := 0; i < len(outlineSegs); {
		if !segmentTouchesSeamRow(outlineSegs[i], seamY, left.X, right.X) {
			i++
			continue
		}
		start := i
		for i < len(outlineSegs) && segmentTouchesSeamRow(outlineSegs[i], seamY, left.X, right.X) {
			i++
		}
		runs = append(runs, [2]int{start, i - 1})
	}
	if len(runs) == 0 {
		return nil
	}

	first, last := -1, -1
	for _, run := range runs {
		if seamRunScore(outlineSegs, run[0], run[1], seamY) < 0.01 {
			continue
		}
		if first < 0 || run[0] < first {
			first = run[0]
		}
		if last < 0 || run[1] > last {
			last = run[1]
		}
	}
	if first < 0 {
		return nil
	}
	if !spanIsSeamReplaceable(outlineSegs, first, last, seamY, left.X, right.X) {
		best := runs[0]
		bestScore := seamRunScore(outlineSegs, best[0], best[1], seamY)
		for _, run := range runs[1:] {
			if score := seamRunScore(outlineSegs, run[0], run[1], seamY); score > bestScore {
				best = run
				bestScore = score
			}
		}
		first, last = best[0], best[1]
	}

	prefix := outlineSegs[:first]
	suffix := outlineSegs[last+1:]
	prefix = snapPrefixSuffixToSeamRow(prefix, left, right, seamY, precision)
	suffix = snapSuffixStartToSeamRow(suffix, left, right, seamY, precision)

	seam := seamChain
	if len(prefix) > 0 {
		pe := prefix[len(prefix)-1].End()
		if path.Distance(pe, right) < path.Distance(pe, left) {
			seam = reverseSegmentChain(seamChain, so)
		}
	} else if len(suffix) > 0 {
		ss := suffix[0].Start()
		if path.Distance(ss, right) < path.Distance(ss, left) {
			seam = reverseSegmentChain(seamChain, so)
		}
	}

	out := append([]path.Segment{}, prefix...)
	if len(prefix) > 0 && len(seam) > 0 {
		out = append(out, seamStubSegments(prefix[len(prefix)-1].End(), seam[0].Start(), seamY, precision)...)
	}
	out = append(out, seam...)
	if len(suffix) > 0 && len(seam) > 0 {
		out = append(out, seamStubSegments(seam[len(seam)-1].End(), suffix[0].Start(), seamY, precision)...)
	}
	out = append(out, suffix...)
	return out
}

func snapSuffixStartToSeamRow(segs []path.Segment, left, right path.Point, seamY float64, precision int) []path.Segment {
	if len(segs) == 0 {
		return segs
	}
	first := segs[0]
	start := first.Start()
	snapped := snapCornerPointToSeamRow(start, left, right, seamY)
	if start.EqualsPrecision(snapped, precision) {
		return segs
	}
	line, ok := first.(path.LineSegment)
	if !ok {
		return segs
	}
	line.StartPoint = snapped
	out := append([]path.Segment{line}, segs[1:]...)
	return out
}

func snapPrefixSuffixToSeamRow(segs []path.Segment, left, right path.Point, seamY float64, precision int) []path.Segment {
	if len(segs) == 0 {
		return segs
	}
	last := segs[len(segs)-1]
	end := last.End()
	snapped := snapCornerPointToSeamRow(end, left, right, seamY)
	if end.EqualsPrecision(snapped, precision) {
		return segs
	}
	line, ok := last.(path.LineSegment)
	if !ok {
		return segs
	}
	line.EndPoint = snapped
	out := append([]path.Segment{}, segs[:len(segs)-1]...)
	out = append(out, line)
	return out
}

func snapCornerPointToSeamRow(pt, left, right path.Point, seamY float64) path.Point {
	onCorner := math.Abs(pt.X-left.X) <= seamEndpointSlop || math.Abs(pt.X-right.X) <= seamEndpointSlop
	if onCorner && math.Abs(pt.Y-seamY) <= 0.05 {
		return path.NewPoint(pt.X, seamY)
	}
	return pt
}

func seamStubSegments(from, to path.Point, seamY float64, precision int) []path.Segment {
	if from.EqualsPrecision(to, precision) {
		return nil
	}
	if math.Abs(from.X-to.X) <= 0.05 &&
		math.Abs(from.Y-seamY) <= 0.05 &&
		math.Abs(to.Y-seamY) <= 0.05 {
		return nil
	}
	if math.Abs(from.Y-seamY) <= seamEndpointSlop && math.Abs(to.Y-seamY) <= seamEndpointSlop &&
		math.Abs(from.X-to.X) > 1.0 {
		return nil
	}
	return []path.Segment{path.LineSegment{StartPoint: from, EndPoint: to}}
}

func closingSegments(outlineSegs []path.Segment, from, to path.Point, seamY, leftX, rightX float64, precision int, so path.SegmentOperators) []path.Segment {
	remaining := append([]path.Segment{}, outlineSegs...)
	chain := []path.Segment{}
	current := from
	for len(remaining) > 0 && !current.EqualsPrecision(to, precision) {
		attached := false
		for i, s := range remaining {
			if segmentTouchesSeamRow(s, seamY, leftX, rightX) {
				continue
			}
			if current.EqualsPrecision(s.Start(), precision) {
				chain = append(chain, s)
				current = s.End()
				remaining = append(remaining[:i], remaining[i+1:]...)
				attached = true
				break
			}
			if current.EqualsPrecision(s.End(), precision) {
				rev, err := so.Reverse(s)
				if err != nil {
					rev = s
				}
				chain = append(chain, rev)
				current = rev.End()
				remaining = append(remaining[:i], remaining[i+1:]...)
				attached = true
				break
			}
		}
		if !attached {
			break
		}
	}
	return chain
}

func closeTopEdgeIfNeeded(segs []path.Segment, seamY float64) []path.Segment {
	if len(segs) == 0 {
		return segs
	}
	start, end := pathEndpoints(segs)
	if start.EqualsPrecision(end, 3) {
		return segs
	}
	if start.Y >= seamY-seamEndpointSlop || end.Y >= seamY-seamEndpointSlop {
		return segs
	}
	if start.Y > end.Y+seamEndpointSlop || end.Y > start.Y+seamEndpointSlop {
		return segs
	}
	return append(segs, path.LineSegment{StartPoint: end, EndPoint: start})
}

// splitOutlineFromHoles separates the outer perimeter from interior cutouts.
// joinOutlineWithSeam must only walk the perimeter — shelf/tab holes share the
// split row and otherwise pollute mergeConnectedSegments.
func splitOutlineFromHoles(p path.Path, so path.SegmentOperators) (outline path.Path, holes []path.Path) {
	subpaths := path.SplitPathOnMove(p)
	if len(subpaths) <= 1 {
		return p, nil
	}
	bestI := 0
	bestArea := 0.0
	for i, sp := range subpaths {
		tl, br, err := path.BoundingBoxTrimWhitespace(sp, so)
		if err != nil {
			continue
		}
		area := (br.X - tl.X) * (br.Y - tl.Y)
		if area > bestArea {
			bestArea = area
			bestI = i
		}
	}
	for i, sp := range subpaths {
		if i == bestI {
			outline = sp
		} else {
			holes = append(holes, sp)
		}
	}
	return outline, holes
}

// mirrorFrame is the horizontal-flip frame MirrorTransform uses: the path's
// top-left, and the Y extent after that corner is moved to the origin.
func mirrorFrame(p path.Path, so path.SegmentOperators) (origin path.Point, height float64, err error) {
	origin, err = path.PointPathAttribute(path.TopLeft, p, so)
	if err != nil {
		return origin, 0, err
	}
	moved, err := transforms.MoveTransform{
		Point:            path.NewPoint(0, 0),
		Handle:           path.TopLeft,
		SegmentOperators: so,
	}.PathTransform(p)
	if err != nil {
		return origin, 0, err
	}
	_, br, err := path.BoundingBoxWithWhitespace(moved, so)
	if err != nil {
		return origin, 0, err
	}
	return origin, br.Y, nil
}

// mirrorHorizontalInFrame flips p inside a frame taken from another path.
// MirrorTransform flips each path inside its own bounding box, which leaves a
// small hole unmoved. Holes have to flip inside the outline's frame so a
// split keeps the cutouts that belong on that half.
func mirrorHorizontalInFrame(p path.Path, origin path.Point, height float64, so path.SegmentOperators) (path.Path, error) {
	shifted, err := transforms.ShiftTransform{
		DeltaX:           -origin.X,
		DeltaY:           -origin.Y,
		SegmentOperators: so,
	}.PathTransform(p)
	if err != nil {
		return nil, err
	}
	segs := make([]path.Segment, 0, len(shifted.Segments()))
	for _, seg := range shifted.Segments() {
		s, err := so.TransformPoints(seg, func(pt path.Point) path.Point {
			np := pt.Clone()
			np.Y = height - pt.Y
			return np
		})
		if err != nil {
			return nil, err
		}
		segs = append(segs, s)
	}
	return transforms.ShiftTransform{
		DeltaX:           origin.X,
		DeltaY:           origin.Y,
		SegmentOperators: so,
	}.PathTransform(path.NewPathFromSegments(segs))
}

func hSlicePath(p path.Path, y float64, so path.SegmentOperators, precision int) (path.Path, error) {
	return transforms.HSliceTransform{
		Y:                y,
		SegmentOperators: so,
		Precision:        precision,
	}.PathTransform(p)
}

const straightCutSnapSlop = 0.05

// slicedOpenChain joins the open chains left by a horizontal knife cut.
// snapY pulls the cut ends back onto the nominal seam when the slice bled
// past it by a kerf offset; NaN leaves the cut where the knife made it.
func slicedOpenChain(p path.Path, snapY float64, precision int, so path.SegmentOperators) ([]path.Segment, error) {
	var chains [][]path.Segment
	for _, sp := range path.SplitPathOnMove(p) {
		segs := path.TrimMove(sp.Segments())
		if len(segs) == 0 {
			continue
		}
		chains = append(chains, segs)
	}
	if len(chains) == 0 {
		return nil, nil
	}
	chains = mergeOpenChains(chains, precision, so)
	chain := append([]path.Segment{}, chains[longestChainIndex(chains)]...)
	snapOpenChainEnds(chain, snapY, straightCutSnapSlop)
	return chain, nil
}

// closeSlicedOutline closes the knife cut with one straight segment between
// the two cut ends.
func closeSlicedOutline(p path.Path, snapY float64, precision int, so path.SegmentOperators) (path.Path, error) {
	chain, err := slicedOpenChain(p, snapY, precision, so)
	if err != nil || len(chain) == 0 {
		return p, err
	}
	start, end := pathEndpoints(chain)
	if !start.EqualsPrecision(end, precision) {
		chain = append(chain, path.LineSegment{StartPoint: end, EndPoint: start})
	}
	return transforms.CleanupTransform{Precision: precision}.PathTransform(
		path.NewPathFromSegmentsWithoutMove(chain))
}

// closeWithRenderedSeam closes the knife cut with edge, drawn left to right
// between the two cut ends so a curve keeps the same side on both halves.
func closeWithRenderedSeam(sliced path.Path, edge Component, snapY float64, ctx RenderContext, precision int, so path.SegmentOperators) (path.Path, error) {
	chain, err := slicedOpenChain(sliced, snapY, precision, so)
	if err != nil || len(chain) == 0 {
		return sliced, err
	}
	start, end := pathEndpoints(chain)
	if start.EqualsPrecision(end, precision) {
		return transforms.CleanupTransform{Precision: precision}.PathTransform(
			path.NewPathFromSegmentsWithoutMove(chain))
	}
	seam, err := renderEdgeLeftToRight(edge, start, end, ctx)
	if err != nil {
		return nil, err
	}
	chain = append(chain, orientSeamOnto(path.TrimMove(seam.Segments()), end, start, so)...)
	start, end = pathEndpoints(chain)
	if !start.EqualsPrecision(end, precision) {
		chain = append(chain, path.LineSegment{StartPoint: end, EndPoint: start})
	}
	return transforms.CleanupTransform{Precision: precision}.PathTransform(
		path.NewPathFromSegmentsWithoutMove(chain))
}

// renderEdgeLeftToRight draws edge along the horizontal chord from the left
// cut end to the right one. Left to right keeps the edge's +y downward.
func renderEdgeLeftToRight(edge Component, a, b path.Point, ctx RenderContext) (path.Path, error) {
	left, right := a, b
	if left.X > right.X {
		left, right = right, left
	}
	edge.SetLocalVariable("from", left.ToDynMap())
	edge.SetLocalVariable("to", right.ToDynMap())
	p, _, err := edge.Render(ctx)
	return p, err
}

// orientSeamOnto turns seam so it runs from `from` to `to`, and pins the
// ends so a float gap does not become a broken move.
func orientSeamOnto(segs []path.Segment, from, to path.Point, so path.SegmentOperators) []path.Segment {
	if len(segs) == 0 {
		return []path.Segment{path.LineSegment{StartPoint: from, EndPoint: to}}
	}
	ss, se := pathEndpoints(segs)
	if path.Distance(ss, to)+path.Distance(se, from) < path.Distance(ss, from)+path.Distance(se, to) {
		segs = reverseSegmentChain(segs, so)
	}
	segs[0] = setSegmentStart(segs[0], from)
	segs[len(segs)-1] = setSegmentEnd(segs[len(segs)-1], to)
	return segs
}

func setSegmentStart(seg path.Segment, p path.Point) path.Segment {
	switch s := seg.(type) {
	case path.LineSegment:
		s.StartPoint = p
		return s
	case path.CurveSegment:
		s.StartPoint = p
		return s
	default:
		return seg
	}
}

func setSegmentEnd(seg path.Segment, p path.Point) path.Segment {
	switch s := seg.(type) {
	case path.LineSegment:
		s.EndPoint = p
		return s
	case path.CurveSegment:
		s.EndPoint = p
		return s
	default:
		return seg
	}
}

func longestChainIndex(chains [][]path.Segment) int {
	best := 0
	bestLen := -1.0
	for i, chain := range chains {
		n := 0.0
		for _, seg := range chain {
			n += segmentLength(seg)
		}
		if n > bestLen {
			bestLen = n
			best = i
		}
	}
	return best
}

func mergeOpenChains(chains [][]path.Segment, precision int, so path.SegmentOperators) [][]path.Segment {
	for {
		merged := false
		for i := 0; i < len(chains) && !merged; i++ {
			for j := i + 1; j < len(chains); j++ {
				joined, ok := joinTwoChains(chains[i], chains[j], precision, so)
				if !ok {
					continue
				}
				chains[i] = joined
				chains = append(chains[:j], chains[j+1:]...)
				merged = true
				break
			}
		}
		if !merged {
			return chains
		}
	}
}

func joinTwoChains(a, b []path.Segment, precision int, so path.SegmentOperators) ([]path.Segment, bool) {
	if len(a) == 0 || len(b) == 0 {
		return nil, false
	}
	as, ae := pathEndpoints(a)
	bs, be := pathEndpoints(b)
	// A closed chain meets itself; don't stitch another chain onto that point.
	if as.EqualsPrecision(ae, precision) || bs.EqualsPrecision(be, precision) {
		return nil, false
	}
	switch {
	case ae.EqualsPrecision(bs, precision):
		return append(append([]path.Segment{}, a...), b...), true
	case ae.EqualsPrecision(be, precision):
		return append(append([]path.Segment{}, a...), reverseSegmentChain(b, so)...), true
	case as.EqualsPrecision(be, precision):
		return append(append([]path.Segment{}, b...), a...), true
	case as.EqualsPrecision(bs, precision):
		return append(reverseSegmentChain(b, so), a...), true
	default:
		return nil, false
	}
}

func snapOpenChainEnds(segs []path.Segment, snapY, tol float64) {
	if len(segs) == 0 || math.IsNaN(snapY) {
		return
	}
	snapLineEnd(segs, true, snapY, tol)
	snapLineEnd(segs, false, snapY, tol)
}

func snapLineEnd(segs []path.Segment, atStart bool, y, tol float64) {
	idx := 0
	if !atStart {
		idx = len(segs) - 1
	}
	line, ok := segs[idx].(path.LineSegment)
	if !ok {
		return
	}
	if atStart {
		if math.Abs(line.StartPoint.Y-y) > tol {
			return
		}
		line.StartPoint.Y = y
		if math.Abs(line.EndPoint.Y-y) <= tol {
			line.EndPoint.Y = y
		}
	} else {
		if math.Abs(line.EndPoint.Y-y) > tol {
			return
		}
		line.EndPoint.Y = y
		if math.Abs(line.StartPoint.Y-y) <= tol {
			line.StartPoint.Y = y
		}
	}
	segs[idx] = line
}

const minHoleSegmentLength = 0.1

func segmentLength(seg path.Segment) float64 {
	return path.Distance(seg.Start(), seg.End())
}

func subpathHasDrawnSegment(segs []path.Segment, minLen float64) bool {
	for _, s := range segs {
		if segmentLength(s) >= minLen {
			return true
		}
	}
	return false
}

// subpathWorthKeeping keeps a subpath that has a long segment, or a closed
// shape whose shorter side is at least minLen. Screw holes are four short
// cubics (a 1/8 in circle's chord is ~0.088, under minHoleSegmentLength) but
// the hole itself is larger than a sliced finger-slot sliver.
func subpathWorthKeeping(segs []path.Segment, minLen float64) bool {
	if subpathHasDrawnSegment(segs, minLen) {
		return true
	}
	return closedSubpathMinSpan(segs) >= minLen
}

func closedSubpathMinSpan(segs []path.Segment) float64 {
	if len(segs) < 2 {
		return 0
	}
	if !segs[0].Start().EqualsPrecision(segs[len(segs)-1].End(), 3) {
		return 0
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, s := range segs {
		for _, pt := range []path.Point{s.Start(), s.End()} {
			if pt.X < minX {
				minX = pt.X
			}
			if pt.Y < minY {
				minY = pt.Y
			}
			if pt.X > maxX {
				maxX = pt.X
			}
			if pt.Y > maxY {
				maxY = pt.Y
			}
		}
	}
	w, h := maxX-minX, maxY-minY
	if w < h {
		return w
	}
	return h
}

// pruneHolePath drops move-only stubs left when HSlice clips finger-joint hole paths.
func pruneHolePath(p path.Path, precision int) (path.Path, bool) {
	subpaths := path.SplitPathOnMove(p)
	kept := make([]path.Path, 0, len(subpaths))
	for _, sp := range subpaths {
		segs := path.TrimMove(sp.Segments())
		if subpathWorthKeeping(segs, minHoleSegmentLength) {
			kept = append(kept, sp)
		}
	}
	if len(kept) == 0 {
		return path.NewPath(), false
	}
	return transforms.SimpleJoin{}.JoinPaths(kept...), true
}

// dropDegenerateSubpaths removes zero-length subpaths after SimpleJoin merges outline + holes.
func dropDegenerateSubpaths(p path.Path, precision int) path.Path {
	p = scrubOrphanMoves(p)
	subpaths := path.SplitPathOnMove(p)
	if len(subpaths) <= 1 {
		return p
	}
	kept := make([]path.Path, 0, len(subpaths))
	for _, sp := range subpaths {
		segs := path.TrimMove(sp.Segments())
		if subpathWorthKeeping(segs, minHoleSegmentLength) {
			kept = append(kept, sp)
		}
	}
	if len(kept) == 0 {
		return p
	}
	joined := transforms.SimpleJoin{}.JoinPaths(kept...)
	return scrubOrphanMoves(joined)
}

// scrubOrphanMoves removes Move commands not followed by a drawable segment before the
// next Move. HSlice leaves these as SVG debris (e.g. "M 11.295 1.103") that still
// sit in the same subpath as real geometry so SplitPathOnMove alone cannot drop them.
func scrubOrphanMoves(p path.Path) path.Path {
	segs := p.Segments()
	if len(segs) == 0 {
		return p
	}
	out := make([]path.Segment, 0, len(segs))
	for i := 0; i < len(segs); i++ {
		seg := segs[i]
		if path.IsMove(seg) {
			var following []path.Segment
			for j := i + 1; j < len(segs); j++ {
				if path.IsMove(segs[j]) {
					break
				}
				following = append(following, segs[j])
			}
			if !subpathWorthKeeping(following, minHoleSegmentLength) {
				continue
			}
		}
		out = append(out, seg)
	}
	if len(out) == 0 {
		return path.NewPath()
	}
	cleaned, err := transforms.CleanupTransform{Precision: 3}.PathTransform(
		path.NewPathFromSegmentsWithoutMove(out))
	if err != nil {
		return path.NewPathFromSegmentsWithoutMove(out)
	}
	return cleaned
}

func joinSplitPaths(outline path.Path, holes []path.Path, precision int) path.Path {
	paths := []path.Path{outline}
	for _, hp := range holes {
		if pruned, ok := pruneHolePath(hp, precision); ok {
			paths = append(paths, pruned)
		}
	}
	joined := transforms.SimpleJoin{}.JoinPaths(paths...)
	return dropDegenerateSubpaths(joined, precision)
}

func (ps *PartSplitter) transformPart(part *RenderedPart, y float64, ctx RenderContext) ([]*RenderedPart, error) {
	so := AppContext().SegmentOperators()
	precision := AppContext().Precision()
	attr := part.Part.DmAttr(ps.mp)

	plugEdge, err := AppContext().MakeComponent(
		ps.mp.MustDynMap("plug_edge", dynmap.New()),
		FindDocumentContext(part.Part))
	if err != nil {
		return nil, err
	}
	socketEdge, err := AppContext().MakeComponent(
		ps.mp.MustDynMap("socket_edge", dynmap.New()),
		FindDocumentContext(part.Part))
	if err != nil {
		return nil, err
	}
	plugEdge.SetParent(part.Part)
	socketEdge.SetParent(part.Part)

	originalPath, err := transforms.TrimWhitespaceTransform{
		SegmentOperators: so,
	}.PathTransform(part.Path)
	if err != nil {
		return nil, err
	}

	outlinePath, holePaths := splitOutlineFromHoles(originalPath, so)

	tl, br, err := path.BoundingBoxTrimWhitespace(outlinePath, so)
	if err != nil {
		return nil, err
	}

	yPos := y + attr.MustFloat64("y_offset", 0)

	if yPos >= br.Y {
		return []*RenderedPart{
			&RenderedPart{
				Part:   part.Part,
				Path:   originalPath,
				Width:  br.X - tl.X,
				Height: br.Y - tl.Y,
			},
		}, nil
	}

	plugEdge.SetLocalVariable("from", path.NewPoint(br.X, yPos).ToDynMap())
	plugEdge.SetLocalVariable("to", path.NewPoint(tl.X, yPos).ToDynMap())

	plugEdgePath, _, err := plugEdge.Render(ctx)
	if err != nil {
		return nil, err
	}

	_, brPl, err := path.BoundingBoxTrimWhitespace(plugEdgePath, so)
	if err != nil {
		return nil, err
	}
	bleedTop := attr.MustFloat64("bleed_top", brPl.Y-yPos)

	topOutline, err := hSlicePath(outlinePath, yPos+bleedTop, so, precision)
	if err != nil {
		return nil, err
	}

	// A straight butt seam closes the knife cut between the two points where
	// the outline actually crosses the split. Replacing a "seam run" squares
	// off vertical fingers the cut passes through, and on the mirrored bottom
	// half it can land in the middle of the piece and delete the rest.
	straightPlug := seamIsStraightLine(path.TrimMove(plugEdgePath.Segments()), precision)
	if straightPlug {
		topOutline, err = closeSlicedOutline(topOutline, yPos, precision, so)
	} else if plugEdge.ElementType() == "keyed_edge" {
		topOutline, err = joinOutlineWithSeam(topOutline, plugEdgePath, tl.X, br.X, yPos, precision, so)
	} else {
		// Draw the profile left to right between the knife's cut ends. The
		// first render runs right to left, and that 180° turn flips the
		// curve to the other side of the seam.
		topOutline, err = closeWithRenderedSeam(topOutline, plugEdge, yPos, ctx, precision, so)
	}
	if err != nil {
		return nil, err
	}

	topHoles := make([]path.Path, 0, len(holePaths))
	for _, hp := range holePaths {
		sliced, err := hSlicePath(hp, yPos+bleedTop, so, precision)
		if err != nil {
			return nil, err
		}
		topHoles = append(topHoles, sliced)
	}
	topPath := joinSplitPaths(topOutline, topHoles, precision)

	bottomY := br.Y - yPos
	flippedOutline, err := transforms.MirrorTransform{
		Axis:             transforms.Horizontal,
		Handle:           path.TopLeft,
		SegmentOperators: so,
	}.PathTransform(outlinePath)
	if err != nil {
		return nil, err
	}

	socketEdge.SetLocalVariable("from", path.NewPoint(br.X, bottomY).ToDynMap())
	socketEdge.SetLocalVariable("to", path.NewPoint(tl.X, bottomY).ToDynMap())
	socketEdgePath, _, err := socketEdge.Render(ctx)
	if err != nil {
		return nil, err
	}
	straightSocket := seamIsStraightLine(path.TrimMove(socketEdgePath.Segments()), precision)

	_, brSo, err := path.BoundingBoxTrimWhitespace(socketEdgePath, so)
	if err != nil {
		return nil, err
	}
	bleedBottom := attr.MustFloat64("bleed_bottom", brSo.Y-bottomY)

	bottomOutline, err := hSlicePath(flippedOutline, bottomY+bleedBottom, so, precision)
	if err != nil {
		return nil, err
	}
	// Capture the flip frames before the outline is mirrored back. Holes must
	// use these, not their own bounding boxes.
	outlineOrigin, outlineH, err := mirrorFrame(outlinePath, so)
	if err != nil {
		return nil, err
	}
	bottomOrigin, bottomH, err := mirrorFrame(bottomOutline, so)
	if err != nil {
		return nil, err
	}

	var bottomPath path.Path
	if plugEdge.ElementType() == "keyed_edge" {
		// Keyed miter seams (chess frame): join in mirrored coords, then mirror back.
		// Re-rendering the socket in normal coords after mirroring the outline alone
		// collapses trapezoids to seam slivers.
		bottomPath, err = legacyJoinOutlineWithSeam(bottomOutline, socketEdgePath, precision, so)
		if err != nil {
			return nil, err
		}
		bottomPath, err = transforms.MirrorTransform{
			Axis:             transforms.Horizontal,
			Handle:           path.TopLeft,
			SegmentOperators: so,
		}.PathTransform(bottomPath)
		if err != nil {
			return nil, err
		}
	} else {
		bottomOutline, err = transforms.MirrorTransform{
			Axis:             transforms.Horizontal,
			Handle:           path.TopLeft,
			SegmentOperators: so,
		}.PathTransform(bottomOutline)
		if err != nil {
			return nil, err
		}

		if straightSocket {
			// Mirror-back puts the knife cut on the piece's top edge (y≈0),
			// not at the original split Y. Close that cut; snapping to yPos
			// would draw the seam through the middle of the lower half.
			bottomPath, err = closeSlicedOutline(bottomOutline, math.NaN(), precision, so)
		} else if socketEdge.ElementType() == "keyed_edge" {
			socketEdge.SetLocalVariable("from", path.NewPoint(tl.X, yPos).ToDynMap())
			socketEdge.SetLocalVariable("to", path.NewPoint(br.X, yPos).ToDynMap())
			socketEdgePath, _, err = socketEdge.Render(ctx)
			if err != nil {
				return nil, err
			}
			bottomPath, err = joinOutlineWithSeam(bottomOutline, socketEdgePath, tl.X, br.X, yPos, precision, so)
		} else {
			bottomPath, err = closeWithRenderedSeam(bottomOutline, socketEdge, math.NaN(), ctx, precision, so)
		}
		if err != nil {
			return nil, err
		}
	}

	bottomHoles := make([]path.Path, 0, len(holePaths))
	for _, hp := range holePaths {
		flippedHole, err := mirrorHorizontalInFrame(hp, outlineOrigin, outlineH, so)
		if err != nil {
			return nil, err
		}
		sliced, err := hSlicePath(flippedHole, bottomY+bleedBottom, so, precision)
		if err != nil {
			return nil, err
		}
		sliced, err = mirrorHorizontalInFrame(sliced, bottomOrigin, bottomH, so)
		if err != nil {
			return nil, err
		}
		bottomHoles = append(bottomHoles, sliced)
	}
	bottomPath = joinSplitPaths(bottomPath, bottomHoles, precision)

	topPath, topMinX, topMinY, twidth, theight, err := trimSplitPartPath(topPath, so)
	if err != nil {
		return nil, err
	}
	bottomPath, botMinX, botMinY, bwidth, bheight, err := trimSplitPartPath(bottomPath, so)
	if err != nil {
		return nil, err
	}

	return []*RenderedPart{
		&RenderedPart{
			Part:   part.Part,
			Path:   topPath,
			Width:  twidth,
			Height: theight,
			MinX:   topMinX,
			MinY:   topMinY,
		},
		&RenderedPart{
			Part:   part.Part,
			Path:   bottomPath,
			Width:  bwidth,
			Height: bheight,
			MinX:   botMinX,
			MinY:   botMinY,
		},
	}, nil
}

func trimSplitPartPath(p path.Path, so path.SegmentOperators) (path.Path, float64, float64, float64, float64, error) {
	trimmed, err := transforms.TrimWhitespaceTransform{
		SegmentOperators: so,
	}.PathTransform(p)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	tl, br, err := path.BoundingBoxTrimWhitespace(trimmed, so)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	return trimmed, tl.X, tl.Y, br.X - tl.X, br.Y - tl.Y, nil
}

func (ps *PartSplitter) TransformPart(part *RenderedPart, ctx RenderContext) ([]*RenderedPart, error) {
	attr := part.Part.DmAttr(ps.mp)
	so := AppContext().SegmentOperators()

	autoSplit := attr.MustBool("auto_split", false)
	yPos := 0.0
	if autoSplit {
		_, hasY := attr.Float64("y")
		if hasY {
			return nil, fmt.Errorf("param 'y' cannot be used with 'autosplit'")
		}
		tl, br, err := path.BoundingBoxTrimWhitespace(part.Path, so)
		if err != nil {
			return nil, err
		}
		w := br.X - tl.X
		h := br.Y - tl.Y

		mH := attr.MustFloat64("max_height", attr.MustFloat64("material_height", 0))
		mW := attr.MustFloat64("max_width", attr.MustFloat64("material_width", 0))
		if mH == 0 || mW == 0 {
			return nil, fmt.Errorf("for autosplit, material_height and material_width must be specified")
		}
		if (h >= mH && h >= mW) || (h >= mH && w >= mH) {
			yPos = h / 2
		} else if w >= mW && w >= mH {
			pth, err := transforms.RotateTransform{
				Degrees:          90,
				SegmentOperators: so,
			}.PathTransform(part.Path)
			if err != nil {
				return nil, err
			}
			newPart := &RenderedPart{
				Part:   part.Part,
				Path:   pth,
				Width:  h,
				Height: w,
			}
			return ps.TransformPart(newPart, ctx)
		} else {
			return []*RenderedPart{part}, nil
		}

	} else {
		y, ok := attr.Float64("y")
		if !ok {
			return nil, fmt.Errorf("Error in split transformer.  'y' must be specified")
		}
		yPos = y
	}

	if yPos == 0 {
		return []*RenderedPart{part}, nil
	}

	parts, err := ps.transformPart(part, yPos, ctx)
	if err != nil || len(parts) <= 1 || !autoSplit {
		return parts, err
	}
	retParts := []*RenderedPart{}
	for _, p := range parts {
		pts, err := ps.TransformPart(p, ctx)
		if err != nil {
			return parts, err
		}
		retParts = append(retParts, pts...)
	}
	return retParts, nil
}
