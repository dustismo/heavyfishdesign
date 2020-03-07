package dom

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

func TestDynMapPointIndirectLookup(t *testing.T) {
	mp := dynmap.New()
	mp.Put("to", "from")

	e := &BasicElement{
		id:          "test",
		elementType: "testType",
		originalMap: dynmap.New(),
	}

	e.SetLocalVariable("from", path.NewPoint(1, 1))
	attr := NewAttr(e, mp)
	actual, ok := attr.Point("to")
	if !ok {
		t.Errorf("Unable to find key")
	}

	expected := "(1,1)"
	if !actual.Equals(path.NewPoint(1, 1)) {
		t.Errorf("Expected '%s' but got '%s'", expected, actual)
	}
}

func TestDynMapPointLookup(t *testing.T) {
	mp := dynmap.New()
	mp.Put("finger_width", 12)

	e := &BasicElement{
		id:          "test",
		elementType: "testType",
		originalMap: mp,
	}

	innerMp := dynmap.New()
	innerMp.PutWithDot("to.x", "12")
	innerMp.PutWithDot("to.y", "finger_width / 2")
	attr := NewAttr(e, innerMp)
	actual, ok := attr.Point("to")
	if !ok {
		t.Errorf("Unable to find key")
	}

	expected := "(12, 6)"
	if !actual.Equals(path.NewPoint(12, 6)) {
		t.Errorf("Expected '%s' but got '%s'", expected, actual)
	}
}

func TestSameNamePointLookup(t *testing.T) {
	mp := dynmap.New()
	mp.Put("to", "to")

	e := &BasicElement{
		id:          "test",
		elementType: "testType",
		originalMap: dynmap.New(),
	}

	e.SetLocalVariable("to", path.NewPoint(1, 1))
	attr := NewAttr(e, mp)
	actual, ok := attr.Point("to")
	if !ok {
		t.Errorf("Unable to find key")
	}

	expected := "(1,1)"
	if !actual.Equals(path.NewPoint(1, 1)) {
		t.Errorf("Expected '%s' but got '%s'", expected, actual)
	}
}

func TestLookup(t *testing.T) {
	mp := dynmap.New()
	mp.Put("index", 1)
	mp.Put("test", "index")

	e := &BasicElement{
		id:          "test",
		elementType: "testType",
		originalMap: dynmap.New(),
		params:      mp,
	}

	attr := NewAttr(e, mp)

	lu := attr.element.ParamLookerUpper()
	v, _ := lu.ToFloat64("index")
	if v != 1 {
		t.Errorf("index %.3f ", v)
	}

	v = attr.MustFloat64("test", 0)
	if v != 1 {
		t.Errorf("v index %.3f ", v)
	}

}

func TestStringLookup(t *testing.T) {
	mp := dynmap.New()
	mp.Put("index", "this is a string")
	mp.Put("test", "index")

	e := &BasicElement{
		id:          "test",
		elementType: "testType",
		originalMap: dynmap.New(),
		params:      mp,
	}
	attr := NewAttr(e, mp)
	actual, ok := attr.String("test")
	if !ok {
		t.Errorf("Unable to find test key")
	}

	expected := "this is a string"
	if expected != actual {
		t.Errorf("Expected '%s' but got '%s'", expected, actual)
	}
}

func TestSvgLookup(t *testing.T) {
	mp := dynmap.New()
	mp.Put("svg", "test")
	mp.Put("test", "index")

	originalMap := dynmap.New()
	originalMap.Put("index", "L 12 12 M 0 0")

	e := &BasicElement{
		id:          "test",
		elementType: "testType",
		originalMap: originalMap,
		params:      mp,
	}
	attr := NewAttr(e, mp)
	actual, ok := attr.SvgString("svg")
	if !ok {
		t.Errorf("Unable to find test key")
	}

	expected := "L 12 12 M 0 0"
	if expected != actual {
		t.Errorf("Expected '%s' but got '%s'", expected, actual)
	}
}
