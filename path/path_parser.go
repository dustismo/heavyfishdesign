package path

// parses a string from a path element
import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dustismo/heavyfishdesign/dynmap"
)

type SvgCommand rune

const (
	// These are the only supported Command types
	Move         SvgCommand = 'M'
	Line         SvgCommand = 'L'
	CurveCommand SvgCommand = 'C'
	ClosePath    SvgCommand = 'Z'
	// these are available to the parser, and are automatically
	// converted to the ones above
	relLine               SvgCommand = 'l'
	relMove               SvgCommand = 'm'
	relCurveCommand       SvgCommand = 'c'
	smoothCurveCommand    SvgCommand = 'S'
	relSmoothCurveCommand SvgCommand = 's'
	vLine                 SvgCommand = 'V'
	vRelLine              SvgCommand = 'v'
	hLine                 SvgCommand = 'H'
	hRelLine              SvgCommand = 'h'
	relClosePath          SvgCommand = 'z' // for parsing only, always use closePath
	qCurveCommand         SvgCommand = 'Q'
	relQCurveCommand      SvgCommand = 'q'
)

// parses a path from the SVG style string
func ParsePathFromSvg(path string) (Path, error) {

	d := NewDraw()
	// Now we normalize to make parsing easier.
	// replace comma delim with space
	// replace line breaks with space.  (TODO: just use a single regex! :))
	path = strings.Replace(path, ",", " ", -1)
	path = strings.Replace(path, "\n", " ", -1)
	path = strings.Replace(path, "\t", " ", -1)
	path = strings.Replace(path, "\r", " ", -1)

	// Make sure there is a space after the command
	re := regexp.MustCompile("([[:alpha:]])([0-9\\.]+)")
	path = string(re.ReplaceAll([]byte(path), []byte("$1 $2")))
	// make sure there is a space before the command
	re2 := regexp.MustCompile("([0-9\\.]+)([[:alpha:]])")
	path = string(re2.ReplaceAll([]byte(path), []byte("$1 $2")))
	// split and clean up any empty elements
	items := []string{}
	for _, elem := range strings.Split(path, " ") {
		if len(elem) > 0 {
			items = append(items, elem)
		}
	}

	var err error
	index := 0
	for index < len(items) {
		index, err = next(index, items, d)
		if err != nil {
			return d.Path(), fmt.Errorf("Unable to parse svg (%s) --> %s", StringElipses(path, 10), err.Error())
		}
	}

	return d.Path(), nil
}

func next(index int, items []string, d *Draw) (int, error) {
	if index >= len(items) {
		return -1, nil
	}

	op := SvgCommand([]rune(items[index])[0])
	switch op {
	case Move:
		// M 3.1 2
		vals, err := getNextFloats(items, index+1, index+2)
		if err != nil {
			return index + 3, err
		}
		d.MoveTo(NewPoint(vals[0], vals[1]))
		return index + 3, nil
	case relMove:
		// M 3.1 2
		vals, err := getNextFloats(items, index+1, index+2)
		if err != nil {
			return index + 3, err
		}
		d.RelMoveTo(NewPoint(vals[0], vals[1]))
		return index + 3, nil
	case Line:
		// M 3.1 2
		vals, err := getNextFloats(items, index+1, index+2)
		if err != nil {
			return index + 3, err
		}
		d.LineTo(NewPoint(vals[0], vals[1]))
		return index + 3, nil
	case relLine:
		// M 3.1 2
		vals, err := getNextFloats(items, index+1, index+2)
		if err != nil {
			return index + 3, err
		}
		d.RelLineTo(NewPoint(vals[0], vals[1]))
		return index + 3, nil

	case CurveCommand:
		vals, err := getNextFloats(items, index+1, index+2, index+3, index+4, index+5, index+6)
		if err != nil {
			return index + 3, err
		}
		d.CurveTo(
			NewPoint(vals[0], vals[1]), // controlstart
			NewPoint(vals[2], vals[3]), // controlend
			NewPoint(vals[4], vals[5]), // point
		)
		return index + 7, nil

	case relCurveCommand:
		vals, err := getNextFloats(items, index+1, index+2, index+3, index+4, index+5, index+6)
		if err != nil {
			return index + 3, err
		}
		d.RelCurveTo(
			NewPoint(vals[0], vals[1]), // controlstart
			NewPoint(vals[2], vals[3]), // controlend
			NewPoint(vals[4], vals[5]), // point
		)
		return index + 7, nil

	case smoothCurveCommand:
		vals, err := getNextFloats(items, index+1, index+2, index+3, index+4)
		if err != nil {
			return index + 3, err
		}
		d.SmoothCurveTo(
			NewPoint(vals[0], vals[1]), // controlend
			NewPoint(vals[2], vals[3]), // point
		)
		return index + 5, nil
	case relSmoothCurveCommand:
		vals, err := getNextFloats(items, index+1, index+2, index+3, index+4)
		if err != nil {
			return index + 3, err
		}
		d.RelSmoothCurveTo(
			NewPoint(vals[0], vals[1]), // controlend
			NewPoint(vals[2], vals[3]), // point
		)
		return index + 5, nil
	case ClosePath:
		// do nothing. maybe figure out something later?
		fmt.Printf("Warning, skipping Z close path in svg parsing\n")
		return index + 1, nil
	}
	return index, fmt.Errorf("Unable to parse, unknown command type '%s'", string(rune(op)))
}

func getNextFloats(items []string, index ...int) ([]float64, error) {
	values := []float64{}
	for _, i := range index {
		if i >= len(items) {
			return values, fmt.Errorf("could not parse, not enough numbers")
		}

		x, err := dynmap.ToFloat64(items[i])
		if err != nil {
			return values, err
		}
		values = append(values, x)
	}
	return values, nil
}
