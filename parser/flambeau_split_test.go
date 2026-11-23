package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/path"
)

func chdirRepoRoot(t *testing.T) {
	t.Helper()
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(filepath.Join(repoRoot, "parser")) })
}

func loadFlambeauBottomHalves(t *testing.T, hfdPath string) []path.Path {
	t.Helper()
	InitContext()
	chdirRepoRoot(t)
	ps, ctx, err := LoadPlanSet(hfdPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var bottomPaths []path.Path
	err = ps.InitWithPartsFilter(ctx, func(rp *dom.RenderedPart) bool {
		if rp.Part.Id() == "bottom" {
			bottomPaths = append(bottomPaths, rp.Path)
		}
		return false
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if len(bottomPaths) != 2 {
		t.Fatalf("expected 2 bottom halves, got %d", len(bottomPaths))
	}
	return bottomPaths
}

func assertNoFingerJointSeamChord(t *testing.T, i int, s string) {
	t.Helper()
	if strings.Contains(s, "L 13.507 6.011 L 0.000 6.000") ||
		strings.Contains(s, "L 13.507 6.000 L 0.000 6.000") {
		t.Fatalf("bottom half %d: finger-joint seam skipped tabs (full-width chord):\n%s", i, s)
	}
}

func longestSubpath(d string) string {
	subpaths := strings.Split(d, " M ")
	longest := subpaths[0]
	for _, part := range subpaths[1:] {
		if len(part) > len(longest) {
			longest = part
		}
	}
	return longest
}

func TestFlambeauShelvesBackPanelSplit(t *testing.T) {
	bottomPaths := loadFlambeauBottomHalves(t, "../heavyfishdesign-studio/designs/divided_box/flambeau_shelves.hfd")

	precision := 3
	for i, p := range bottomPaths {
		s := path.SvgString(p, precision)
		assertNoFingerJointSeamChord(t, i, s)
		longest := longestSubpath(s)
		pts := extractPathPoints(longest)
		if len(pts) < 4 {
			t.Fatalf("bottom half %d: outline too short:\n%s", i, s)
		}
		if pts[0] != pts[len(pts)-1] {
			t.Fatalf("bottom half %d: outline not closed (start=%s end=%s):\n%s", i, pts[0], pts[len(pts)-1], longest)
		}
		assertNoDegenerateHoleStubs(t, i, p, precision)
	}
}

func assertNoDegenerateHoleStubs(t *testing.T, i int, p path.Path, precision int) {
	t.Helper()
	s := path.SvgString(p, precision)
	// Move-only debris from HSlice (e.g. "M 11.295 1.103" with no following L).
	for _, part := range strings.Split(s, " M ") {
		chunk := strings.TrimSpace(part)
		if chunk == "" {
			continue
		}
		if !strings.Contains(chunk, " L ") {
			t.Fatalf("bottom half %d: move-only hole stub in path:\n%s", i, chunk)
		}
	}
}

func TestFlambeauShelvesBackPanelSplitLineSeam(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	srcPath := filepath.Join(repoRoot, "..", "heavyfishdesign-studio", "designs", "divided_box", "flambeau_shelves.hfd")
	src, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(src)
	// Straight butt seam on back panel split only (keep finger joints on shelves).
	content = strings.Replace(content,
		`"id": "bottom",
            "transforms": [
                {
                    "type": "rotate",
                    "degrees": 90
                }
            ],
            "part_transformers": [
                {
                    "type": "splitter",
                    "y": "back_split_y",
                    "bleed_top": "offset",
                    "bleed_bottom": "offset",
                    "plug_edge": {
                        "id": "plug_edge",
                        "type": "finger_joint_plug",`,
		`"id": "bottom",
            "transforms": [
                {
                    "type": "rotate",
                    "degrees": 90
                }
            ],
            "part_transformers": [
                {
                    "type": "splitter",
                    "y": "back_split_y",
                    "bleed_top": "offset",
                    "bleed_bottom": "offset",
                    "plug_edge": {
                        "id": "plug_edge",
                        "type": "line",`, 1)
	content = strings.Replace(content,
		`"socket_edge": {
                        "type": "finger_joint_socket",
                        "padding_left": "finger_padding",
                        "padding_right": "finger_padding"
                    }
                }
            ],
            "components": [
                {
                    "type": "box_bottom",`,
		`"socket_edge": {
                        "type": "line",
                        "padding_left": "finger_padding",
                        "padding_right": "finger_padding"
                    }
                }
            ],
            "components": [
                {
                    "type": "box_bottom",`, 1)

	tmp := t.TempDir() + "/flambeau_line.hfd"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	bottomPaths := loadFlambeauBottomHalves(t, tmp)
	precision := 3
	for i, p := range bottomPaths {
		s := path.SvgString(p, precision)
		longest := longestSubpath(s)
		// Top half keeps the seam at the split (y=6). The mirrored bottom half
		// is trimmed so that same butt cut sits on y=0, and the original
		// finger joints remain on the far edge.
		hasStraightSeam := strings.Contains(s, "13.507 6.000 L 0.000 6.000") ||
			strings.Contains(s, "0.000 6.000 L 13.507 6.000") ||
			strings.Contains(s, "0.000 0.000 L 13.507 0.000") ||
			strings.Contains(s, "13.507 0.000 L 0.000 0.000")
		if !hasStraightSeam {
			t.Fatalf("bottom half %d: missing straight seam line in path:\n%s", i, s)
		}
		if strings.Contains(longest, "L 13.507 6.003 L 13.507 6.000") {
			t.Fatalf("bottom half %d: stray HSlice stub at seam:\n%s", i, longest)
		}
		if strings.Contains(s, "0.000 0.000 L 13.507 0.000") || strings.Contains(s, "13.507 0.000 L 0.000 0.000") {
			if !strings.Contains(s, "5.761") {
				t.Fatalf("bottom half %d: finger joints on the uncut edge were replaced:\n%s", i, longest)
			}
		}
		pts := extractPathPoints(longest)
		if len(pts) < 4 {
			t.Fatalf("bottom half %d: outline too short:\n%s", i, longest)
		}
		if pts[0] != pts[len(pts)-1] {
			t.Fatalf("bottom half %d: line-seam outline not closed (start=%s end=%s):\n%s", i, pts[0], pts[len(pts)-1], longest)
		}
	}
}

func extractPathPoints(d string) []string {
	fields := strings.Fields(strings.ReplaceAll(strings.ReplaceAll(d, "L", " "), "M", " "))
	var pairs []string
	for i := 0; i+1 < len(fields); i += 2 {
		pairs = append(pairs, fields[i]+" "+fields[i+1])
	}
	return pairs
}
