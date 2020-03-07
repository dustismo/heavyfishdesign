package dom

import (
	"strings"

	"github.com/dustismo/heavyfishdesign/dynmap"
)

type Units struct {
	Name string
	Abv  string
}

func (u Units) FromInch(in float64) float64 {
	switch u.Abv {
	case "mm":
		return InchToMM(in)
	default:
		return in
	}
}

func (u Units) FromMM(mm float64) float64 {
	switch u.Abv {
	case "in":
		return MMToInch(mm)
	default:
		return mm
	}
}

var MilliMeters Units = Units{
	"MilliMeters", "mm",
}
var Inches = Units{
	"Inches", "in",
}

/*
 * Returns the requested units.  and true.
 * Else returns MM, false
 */
func NewUnits(in string) (Units, bool) {
	switch strings.ToLower(in) {
	case "in", "inches":
		return Inches, true
	case "mm", "millimeters":
		return MilliMeters, true
	default:
		return MilliMeters, false
	}
}

func MustUnits(in string, defaultUnits Units) Units {
	u, ok := NewUnits(in)
	if !ok {
		return defaultUnits
	}
	return u
}

func MMToInch(mm float64) float64 {
	return mm / 25.4
}

func InchToMM(inch float64) float64 {
	return inch * 25.4
}

// Converts a list of components to the list of DynMaps
func ComponentsToDynMap(c []Component) []*dynmap.DynMap {
	mps := []*dynmap.DynMap{}
	for _, cm := range c {
		mps = append(mps, cm.ToDynMap())
	}
	return mps
}
