package path

import (
	"fmt"
	"strings"

	"github.com/dustismo/heavyfishdesign/dynmap"
)

// defines the special points on a box
type PathAttr string

const (
	// positions based on the bounding box of the drawn shape
	// Note this will take whitespace into account
	TopLeft      PathAttr = "$TOP_LEFT"
	TopRight     PathAttr = "$TOP_RIGHT"
	TopMiddle    PathAttr = "$TOP_MIDDLE"
	BottomLeft   PathAttr = "$BOTTOM_LEFT"
	BottomRight  PathAttr = "$BOTTOM_RIGHT"
	BottomMiddle PathAttr = "$BOTTOM_MIDDLE"
	MiddleLeft   PathAttr = "$MIDDLE_LEFT"
	MiddleRight  PathAttr = "$MIDDLE_RIGHT"
	MiddleMiddle PathAttr = "$MIDDLE_MIDDLE"

	// the first and last visible pixel
	StartPoint PathAttr = "$START_POINT"
	EndPoint   PathAttr = "$END_POINT"

	// current cursor location
	Cursor PathAttr = "$CURSOR"

	// the first point in the path, this
	// differs from StartPoint because this could be
	// a Move.
	StartPosition PathAttr = "$START_POSITION"
	EndPosition   PathAttr = "$END_POSITION"

	Width  PathAttr = "$WIDTH"
	Height PathAttr = "$HEIGHT"

	Origin PathAttr = "$ORIGIN"
)

var enumAttr = map[string]PathAttr{
	string(TopLeft):       TopLeft,
	string(TopRight):      TopRight,
	string(TopMiddle):     TopMiddle,
	string(BottomLeft):    BottomLeft,
	string(BottomRight):   BottomRight,
	string(BottomMiddle):  BottomMiddle,
	string(MiddleLeft):    MiddleLeft,
	string(MiddleRight):   MiddleRight,
	string(MiddleMiddle):  MiddleMiddle,
	string(StartPoint):    StartPoint,
	string(EndPoint):      EndPoint,
	string(StartPosition): StartPosition,
	string(EndPosition):   EndPosition,
	string(Width):         Width,
	string(Height):        Height,
	string(Origin):        Origin,
	string(Cursor):        Cursor,
}

func ToPathAttrFromPoint(point Point, precision int) PathAttr {
	return PathAttr(fmt.Sprintf(precisionStr("%.3f,%.3f", precision), point.X, point.Y))
}

func ToPathAttr(str string) (PathAttr, error) {
	if strings.Contains(str, ",") {
		_, err := parsePoint(str)
		return PathAttr(str), err
	}

	v, ok := enumAttr[str]
	if !ok {
		return TopLeft, fmt.Errorf("Error unknown path attr %s", str)
	}
	return v, nil
}

func PointPathAttribute(pos PathAttr, p Path, so SegmentOperators) (Point, error) {
	v, err := PathAttribute(string(pos), p, so)
	if err != nil {
		return NewPoint(0, 0), err
	}

	switch pt := v.(type) {
	case Point:
		return pt, err
	default:
		return NewPoint(0, 0), fmt.Errorf("Error, type of %s is not a Point. it is %T", string(pos), v)
	}
}

func parsePoint(attr string) (Point, error) {
	commaPoint := strings.Split(attr, ",")
	if len(commaPoint) == 2 {
		// this is a X,Y handle
		x, err := dynmap.ToFloat64(commaPoint[0])
		if err != nil {
			return NewPoint(0, 0), err
		}
		y, err := dynmap.ToFloat64(commaPoint[1])
		if err != nil {
			return NewPoint(0, 0), err
		}
		return NewPoint(x, y), nil
	}
	return NewPoint(0, 0), fmt.Errorf("%s is not parsable to a Point", attr)
}

func PathAttribute(attr string, p Path, so SegmentOperators) (interface{}, error) {

	point, err := parsePoint(attr)
	if err == nil {
		return point, err
	}

	dots := strings.Split(attr, ".")

	tl, br, err := BoundingBoxTrimWhitespace(p, so)
	if err != nil {
		return nil, err
	}

	middleX := tl.X + ((br.X - tl.X) / 2)
	middleY := tl.Y + ((br.Y - tl.Y) / 2)

	pnt := NewPoint(0, 0)
	switch dots[0] {
	case string(Origin):
		return NewPoint(0, 0), nil
	case string(Width):
		return br.X - tl.X, nil
	case string(Height):
		return br.Y - tl.Y, nil
	case string(StartPosition):
		s := p.Segments()
		if len(s) == 0 {
			return p, fmt.Errorf("No available StartPosition on an empty path!")
		}
		return s[0].Start(), nil
	case string(EndPosition):
		s := p.Segments()
		if len(s) == 0 {
			return p, fmt.Errorf("no available endpos!")
		}
		return s[len(s)-1].End(), nil

	case string(StartPoint):
		s := TrimMove(p.Segments())
		if len(s) == 0 {
			return p, fmt.Errorf("no available startpoint!")
		}
		return s[0].Start(), nil
	case string(EndPoint):
		s := TrimMove(p.Segments())
		if len(s) == 0 {
			return p, fmt.Errorf("no available endpoint!")
		}
		return s[len(s)-1].End(), nil
	case string(TopLeft):
		pnt = tl
	case string(TopRight):
		pnt = NewPoint(br.X, tl.Y)
	case string(TopMiddle):
		pnt = NewPoint(middleX, tl.Y)
	case string(BottomRight):
		pnt = br
	case string(BottomLeft):
		pnt = NewPoint(tl.X, br.Y)
	case string(BottomMiddle):
		pnt = NewPoint(middleX, br.Y)
	case string(MiddleMiddle):
		pnt = NewPoint(middleX, middleY)
	case string(MiddleLeft):
		pnt = NewPoint(tl.X, middleY)
	case string(MiddleRight):
		pnt = NewPoint(br.X, middleY)
	default:
		return tl, fmt.Errorf("Error unknown field %s", dots[0])
	}

	// now look at the point, and see if we need to select the X or Y
	if len(dots) > 1 {
		switch dots[1] {
		case "X":
			return pnt.X, nil
		case "Y":
			return pnt.Y, nil
		default:
			return pnt, fmt.Errorf("Error unknown subfield %s", dots[1])
		}
	} else {
		return pnt, nil
	}
}
