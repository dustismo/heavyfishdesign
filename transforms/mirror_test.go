package transforms

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/path"
)

func TestMirrorTransform(t *testing.T) {
	pathStr := "M425.474,539.286L345.318,539.286C330.065,539.286 317.682,551.67 317.682,566.923L317.682,622.195C317.682,637.448 330.065,649.832 345.318,649.832C383.05,617.532 417.826,603.513 450.386,603.194C482.947,603.513 517.723,617.532 555.455,649.832C570.708,649.832 583.091,637.448 583.091,622.195L583.091,566.923C583.091,551.67 570.708,539.286 555.455,539.286L475.298,539.286"
	originalPath, err := path.ParsePathFromSvg(pathStr)

	if err != nil {
		t.Errorf("Error %s", err)
	}

	move := MirrorTransform{
		Axis:             Horizontal,
		Handle:           path.MiddleMiddle,
		SegmentOperators: path.NewSegmentOperators(),
	}

	p, err := move.PathTransform(originalPath)
	if err != nil {
		t.Errorf("Error %s", err)
	}
	expectedStr := "M 425.474 705.105 L 345.318 705.105 C 330.065 705.105 317.682 692.721 317.682 677.468 L 317.682 622.196 C 317.682 606.943 330.065 594.559 345.318 594.559 C 383.050 626.859 417.826 640.878 450.386 641.197 C 482.947 640.878 517.723 626.859 555.455 594.559 C 570.708 594.559 583.091 606.943 583.091 622.196 L 583.091 677.468 C 583.091 692.721 570.708 705.105 555.455 705.105 L 475.298 705.105"
	actualStr := path.SvgString(p, 3)

	if expectedStr != actualStr {
		t.Errorf("Expected: %s\nActual: %s", expectedStr, actualStr)
	}
}
