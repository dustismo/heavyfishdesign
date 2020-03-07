package dom

import (
	"testing"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

func TestDistancePoints(t *testing.T) {
	variables := dynmap.New()
	variables.Put("from", path.NewPoint(0, 0))
	variables.Put("to", path.NewPoint(0, 1))
	e := &BasicElement{
		id:          "test",
		elementType: "testType",
		originalMap: dynmap.New(),
		params:      variables,
	}

	v, _ := EvalExpression("distance(from, to)", e)
	if v != 1.0 {
		t.Errorf("Error expected 1.0 got %.3f", v)
	}
}

func mockElement(variables *dynmap.DynMap) Element {
	e := &BasicElement{
		id:          "test",
		elementType: "testType",
		originalMap: dynmap.New(),
		params:      variables,
	}
	return e
}

func TestFunctions(t *testing.T) {
	variables := dynmap.New()
	variables.Put("finger_height", 36)
	lookup := mockElement(variables)

	v, _ := EvalExpression("1.0 + sqrt(36)", lookup)
	if v != 7.0 {
		t.Errorf("Error expected 7.0 got %.3f", v)
	}

	v, _ = EvalExpression("1.0 + sqrt(finger_height)", lookup)
	if v != 7.0 {
		t.Errorf("Error expected 7.0 got %.3f", v)
	}
}

func TestExpressionRecursion(t *testing.T) {
	variables := dynmap.New()

	variables.Put("test", "test3")
	variables.Put("test2", "test")
	variables.Put("test3", "test2")
	variables.Put("finger_height", 0.6)
	lookup := mockElement(variables)

	_, err := EvalExpression("test2", lookup)
	if err == nil {
		t.Errorf("Error expected error from infinite recursion")
	}
}

func TestExpressionNested(t *testing.T) {
	variables := dynmap.New()
	variables.Put("test", "finger_height + 0.4")
	variables.Put("test2", "test")
	variables.Put("finger_height", 0.6)
	lookup := mockElement(variables)

	v, err := EvalExpression("test2", lookup)
	if v != 1.0 {
		t.Errorf("Error expected 1.0 got %.3f.. %s", v, err.Error())
	}
}

func TestExpressionEvalDotOperator(t *testing.T) {
	// TODO:

	// variables := dynmap.New()
	// variables.Put("test", 0.4)
	// variables.PutWithDot("test2.finger_height", 0.6)
	// lookup := &DynMapParamLookerUpper{
	// 	variables,
	// }

	// v, err := EvalExpression("test2.finger_height", lookup)
	// if v != 0.6 {
	// 	t.Errorf("Error expected 0.6 got %.3f.. %s", v, err.Error())
	// }
}

func TestExpressionEvalVariableParser(t *testing.T) {
	variables := dynmap.New()
	variables.Put("test", 0.4)
	variables.Put("finger_height", 0.6)
	variables.Put("test2", "test")
	lookup := mockElement(variables)

	v, _ := EvalExpression("test", lookup)
	if v != 0.4 {
		t.Errorf("Error expected 0.4 got %.3f", v)
	}

	v, err := EvalExpression("finger_height + 10", lookup)
	if v != 10.6 {
		println(err.Error())
		t.Errorf("Error expected 10.6 got %.3f", v)
	}

	v, _ = EvalExpression("test + 1", lookup)
	if v != 1.4 {
		t.Errorf("Error expected 0.4 got %.3f", v)
	}

	v, _ = EvalExpression("2 * test", lookup)
	if v != 0.8 {
		t.Errorf("Error expected 0.4 got %.3f", v)
	}

	v, _ = EvalExpression("0-1 * test", lookup)
	if v != -0.4 {
		t.Errorf("Error expected 0.4 got %.3f", v)
	}

	v, _ = EvalExpression("test2 * -1", lookup)
	if v != -0.4 {
		t.Errorf("Error expected 0.4 got %.3f", v)
	}
}
func TestEvalExpressionParser(t *testing.T) {
	element := mockElement(dynmap.New())
	v, _ := EvalExpression("6.3", element)
	if v != 6.3 {
		t.Errorf("Error expected 6.3 got %.3f", v)
	}

	v, _ = EvalExpression("6 + 76 + 10", element)
	if v != 92.0 {
		t.Errorf("Error expected 92 got %.3f", v)
	}

	// test with no spaces
	v, _ = EvalExpression("6+76+10", element)
	if v != 92.0 {
		t.Errorf("Error expected 92 got %.3f", v)
	}

	// test with no paranthesis
	v, _ = EvalExpression("1+2*3", element)
	if v != 7.0 {
		t.Errorf("Error expected 7 got %.3f", v)
	}

	// test with parenthesis
	v, _ = EvalExpression("( 1 + 2 ) * 3", element)
	if v != 9.0 {
		t.Errorf("Error expected 9 got %.3f", v)
	}

	// test with decimals
	v, _ = EvalExpression("( 1 + 2.006 ) - .13", element)
	if v != 2.876 {
		t.Errorf("Error expected 2.876 got %.3f", v)
	}

	// test negatives
	v, _ = EvalExpression("-0.4", element)
	if v != -0.4 {
		t.Errorf("Error expected -0.4 got %.3f", v)
	}
}
