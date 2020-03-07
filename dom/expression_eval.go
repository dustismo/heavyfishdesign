package dom

import (
	"fmt"
	"math"

	"github.com/dustismo/govaluate"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

// wrapper to implement the govaluate Parameters interface
type pllExpression struct {
	p         ParamLookerUpper
	attr      *Attr
	recursion int
}

var maxRecursion = 20

func (pll *pllExpression) get(name string) (interface{}, error) {

	v, ok := pll.p.Lookup(name)
	if !ok {
		return v, fmt.Errorf("No param with name %s found", name)
	}

	// Attempt to coerse to the appropriate type
	// type can be either boolean or float64
	// I don't believe there is any reason this should ever be a string??
	// maybe a Point though?

	// bool
	b, err := dynmap.ToBool(v)
	if err == nil {
		return b, nil
	}

	// float64
	f, err := dynmap.ToFloat64(v)
	if err == nil {
		return f, nil
	}

	// string?
	switch s := v.(type) {
	case string:
		// if its a string, try to get it
		pll.recursion = pll.recursion + 1
		return pll.evalExp(s)
	default:
		return s, nil
	}
}

func (pll *pllExpression) Get(name string) (interface{}, error) {
	return pll.get(name)
}

func (pll *pllExpression) evalExp(expression string) (interface{}, error) {
	if pll.recursion > maxRecursion {
		return nil, fmt.Errorf("Error evaluating %s, recursion is too deep", expression)
	}
	params := pll.p
	var functions = make(map[string]govaluate.ExpressionFunction)

	// flattens array arguments
	// so arg1, arg2, arg3[arg4,arg5]
	// will flatten to
	// so arg1, arg2, arg4, arg5
	flatten := func(args ...interface{}) []interface{} {
		vals := []interface{}{}
		for _, arg := range args {
			a, ok := dynmap.ToArray(arg)
			if ok {
				vals = append(vals, a...)
			} else {
				vals = append(vals, arg)
			}
		}
		return vals
	}

	fl := func(args ...interface{}) ([]float64, error) {
		vals := []float64{}
		for _, arg := range args {
			v, err := params.ToFloat64(arg)
			if err != nil {
				return vals, err
			}
			vals = append(vals, v)
		}
		return vals, nil
	}
	// retrives a pair of points from the given args
	pointPair := func(args ...interface{}) (path.Point, path.Point, error) {
		args = flatten(args...)
		if len(args) == 2 {
			//try to load points?
			p1, ok := pll.attr.ToPoint(args[0])
			if !ok {
				return path.NewPoint(0, 0), path.NewPoint(0, 0), fmt.Errorf("Error evaluating, %s must be a Point", args[0])
			}
			p2, ok := pll.attr.ToPoint(args[1])
			if !ok {
				return path.NewPoint(0, 0), path.NewPoint(0, 0), fmt.Errorf("Error evaluating, %s must be a Point", args[1])
			}
			return p1, p2, nil
		}
		vals, err := fl(args...)
		if err != nil {
			return path.NewPoint(0, 0), path.NewPoint(0, 0), err
		}
		if len(vals) != 4 {
			return path.NewPoint(0, 0), path.NewPoint(0, 0), fmt.Errorf("Error requires 4 inputs")
		}
		return path.NewPoint(vals[0], vals[1]), path.NewPoint(vals[2], vals[3]), nil
	}
	//
	// add the functions...
	//
	functions["sqrt"] = func(args ...interface{}) (interface{}, error) {
		v, err := params.ToFloat64(args[0])
		if err != nil {
			return v, err
		}
		return math.Sqrt(v), nil
	}

	functions["distance"] = func(args ...interface{}) (interface{}, error) {
		p1, p2, err := pointPair(args...)
		if err != nil {
			return nil, fmt.Errorf("Error in distance function: %s", err.Error())
		}
		return path.Distance(p1, p2), nil
	}

	functions["angle"] = func(args ...interface{}) (interface{}, error) {
		p1, p2, err := pointPair(args...)
		if err != nil {
			return nil, fmt.Errorf("Error in angle function: %s", err.Error())
		}
		return path.LineSegment{p1, p2}.Angle(), nil
	}

	functions["mmToInch"] = func(args ...interface{}) (interface{}, error) {
		vals, err := fl(args...)
		if err != nil {
			return nil, err
		}
		if len(vals) != 1 {
			return nil, fmt.Errorf("Error 'mmToInch' requires 1 input")
		}
		return MMToInch(vals[0]), nil
	}

	functions["inchToMM"] = func(args ...interface{}) (interface{}, error) {
		vals, err := fl(args...)
		if err != nil {
			return nil, err
		}
		if len(vals) != 1 {
			return nil, fmt.Errorf("Error 'inchToMM' requires 1 input")
		}
		return InchToMM(vals[0]), nil
	}

	exp, err := govaluate.NewEvaluableExpressionWithFunctions(expression, functions)
	if err != nil {
		return nil, err
	}

	v, err := exp.Eval(pll)
	return v, err
}

func EvalExpression(expression string, element Element) (interface{}, error) {
	pms := &pllExpression{
		p:         element.ParamLookerUpper(),
		attr:      NewAttrElement(element),
		recursion: 0,
	}
	return pms.evalExp(expression)
}
