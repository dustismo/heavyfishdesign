package parser

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/path"
)

func TestChessBoardAutoSplitPathsTrimmedToOrigin(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	srcPath := filepath.Join(repoRoot, "..", "heavyfishdesign-studio", "designs", "chess", "chess_board_light.hfd")
	src, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	content := strings.Replace(string(src),
		`"material_width":     24,`,
		`"material_width":     20,`, 1)
	tmp := t.TempDir() + "/chess_split.hfd"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	InitContext()
	chdirRepoRoot(t)
	ps, ctx, err := LoadPlanSet(tmp)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	var frames []path.Path
	var heights []float64
	err = ps.InitWithPartsFilter(ctx, func(rp *dom.RenderedPart) bool {
		if rp.Part.Id() == "outer_frame_miter" {
			frames = append(frames, rp.Path)
			heights = append(heights, rp.Height)
		}
		return false
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if len(frames) != 8 {
		t.Fatalf("expected 8 split frame parts, got %d", len(frames))
	}

	so := path.NewSegmentOperators()
	for i, p := range frames {
		tl, br, err := path.BoundingBoxTrimWhitespace(p, so)
		if err != nil {
			t.Fatalf("frame %d bbox: %v", i, err)
		}
		if tl.X < -1e-3 || tl.Y < -1e-3 {
			t.Fatalf("frame %d not trimmed to origin: min=(%.3f, %.3f)", i, tl.X, tl.Y)
		}
		if math.Abs(br.Y-tl.Y-heights[i]) > 1e-2 {
			t.Fatalf("frame %d height mismatch: bbox=%.3f part=%.3f", i, br.Y-tl.Y, heights[i])
		}
		if br.Y > heights[i]+1e-2 {
			t.Fatalf("frame %d path extends past height: maxY=%.3f height=%.3f", i, br.Y, heights[i])
		}
		s := path.SvgString(p, 3)
		pts := extractPathPoints(longestSubpath(s))
		if len(pts) < 4 {
			t.Fatalf("frame %d outline too short: %s", i, s)
		}
		if pts[0] != pts[len(pts)-1] {
			t.Fatalf("frame %d outline not closed: %s", i, longestSubpath(s))
		}
		if heights[i] < 1.0 {
			t.Fatalf("frame %d suspiciously thin (%.3f in): %s", i, heights[i], s)
		}
	}
}
