package components

import (
	"fmt"
	"math"

	"github.com/dustismo/heavyfishdesign/dom"
	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
)

type GearComponentFactory struct{}

// involute gear component.
// ported from javascript http://jsfiddle.net/yodero/Y2Zjd/
type GearComponent struct {
	*dom.BasicComponent
	segmentOperators path.SegmentOperators
}

func (gcf GearComponentFactory) CreateComponent(componentType string, mp *dynmap.DynMap, dc *dom.DocumentContext) (dom.Component, error) {
	factory := dom.AppContext()
	bc := factory.MakeBasicComponent(mp)
	gc := &GearComponent{
		BasicComponent:   bc,
		segmentOperators: factory.SegmentOperators(),
	}
	return gc, nil
}

// The list of component types this Factory should be used for
func (gcf GearComponentFactory) ComponentTypes() []string {
	return []string{"gear"}
}

func (gc *GearComponent) Render(ctx dom.RenderContext) (path.Path, dom.RenderContext, error) {
	gc.RenderStart(ctx)
	attr := gc.Attr()
	// handle := attr.MustHandle("handle", path.Origin)

	numTeeth := float64(attr.MustInt("teeth", 10))
	toothWidth := attr.MustFloat64("tooth_width", .5)         //size of the tooth
	pressureAngle := attr.MustFloat64("pressure_angle", 20.0) // in degrees, determines gear shape, range is 10 to 40 degrees, most common is 20 degrees
	clearance := attr.MustFloat64("clearance", 0.01)          // freedom between two gear centers
	backlash := attr.MustFloat64("backlash", 0.01)            // freedom between two gear contact points
	pressureAngle = path.DegreesToRadians(pressureAngle)      // convet degrees to radians

	p := toothWidth * float64(numTeeth) / math.Pi / 2.0 // radius of pitch circle
	c := p + toothWidth/math.Pi - clearance             // radius of outer circle
	b := p * math.Cos(pressureAngle)                    // radius of base circle
	r := p - (c - p) - clearance                        // radius of root circle
	t := toothWidth/2.0 - backlash/2.0                  // tooth thickness at pitch circle
	k := -iang(b, p) - t/2.0/p                          // angle where involute meets base circle on side of tooth

	// fmt.Printf("p %f\nc %f\nb %f\nr %f\nt %f\nk %f\n", p, c, b, r, t, k)
	// here is the magic - a set of [x,y] points to create a single gear tooth
	tmp1 := -math.Pi / numTeeth
	if r < b {
		tmp1 = k
	}

	tmp2 := math.Pi / numTeeth
	if r < b {
		tmp2 = -k
	}
	points := []path.Point{
		path.PolarToCartesian(r, -math.Pi/numTeeth),
		path.PolarToCartesian(r, tmp1),
		q7(0.0/5.0, r, b, c, k, 1),
		q7(1.0/5.0, r, b, c, k, 1),
		q7(2.0/5.0, r, b, c, k, 1),
		q7(3.0/5.0, r, b, c, k, 1),
		q7(4.0/5.0, r, b, c, k, 1),
		q7(5.0/5.0, r, b, c, k, 1),
		q7(5.0/5.0, r, b, c, k, -1),
		q7(4.0/5.0, r, b, c, k, -1),
		q7(3.0/5.0, r, b, c, k, -1),
		q7(2.0/5.0, r, b, c, k, -1),
		q7(1.0/5.0, r, b, c, k, -1),
		q7(0.0/5.0, r, b, c, k, -1),
		path.PolarToCartesian(r, tmp2),
		path.PolarToCartesian(r, 3.142/numTeeth),
	}
	pth := path.NewDraw()
	for i := 0; i < int(numTeeth); i++ {
		angle := float64(-i) * 2.0 * math.Pi / numTeeth
		for ix, p := range points {
			xr := p.X*math.Cos(angle) - p.Y*math.Sin(angle)
			yr := p.Y*math.Cos(angle) + p.X*math.Sin(angle)
			if i == 0 && ix == 0 {
				pth.MoveTo(path.NewPoint(xr, yr))
			} else {
				pth.LineTo(path.NewPoint(xr, yr))
			}
		}
	}
	pth.LineTo(points[0])

	variableName, ok := attr.String("gear_variable_name")
	if ok {
		// set the various params for this edge
		gc.SetGlobalVariable(fmt.Sprintf("%s__outer_radius", variableName), c)
		gc.SetGlobalVariable(fmt.Sprintf("%s__inner_radius", variableName), r)
	}
	return gc.HandleTransforms(gc, pth.Path(), ctx)
}

// point on involute curve
func q6(b, s, t, d float64) path.Point {
	return path.PolarToCartesian(d, s*(iang(b, d)+t))
}

// unwind this many degrees to go from r1 to r2
func iang(r1, r2 float64) float64 {
	return math.Sqrt((r2/r1)*(r2/r1)-1.0) - math.Acos(r1/r2)
}

// radius a fraction f up the curved side of the tooth
func q7(f, r, b, r2, t, s float64) path.Point {
	result := q6(b, s, t, (1.0-f)*math.Max(b, r)+f*r2)
	return result
}

// rotate an array of 2d points
func rotate(points []path.Point, angle float64) []path.Point {
	answer := []path.Point{}
	for _, point := range points {
		xr := point.X*math.Cos(angle) - point.Y*math.Sin(angle)
		yr := point.Y*math.Cos(angle) + point.X*math.Sin(angle)
		answer = append(answer, path.NewPoint(xr, yr))
	}
	return answer
}
