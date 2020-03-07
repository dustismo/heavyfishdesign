package dynmap

import (
	"fmt"
	"sort"
	"time"
)

type DynMapSlice struct {
	valsPtr []*DynMap
	vals    []DynMap
	sortCol string
}

func NewDynMapSlice(vals []DynMap, sortCol string) DynMapSlice {
	return DynMapSlice{
		vals:    vals,
		sortCol: sortCol,
	}
}

func Sort(vals []DynMap, sortCol string) {
	sort.Sort(NewDynMapSlice(vals, sortCol))
}

func (dms DynMapSlice) Swap(i, j int) {
	if len(dms.valsPtr) > 0 {
		dms.valsPtr[i], dms.valsPtr[j] = dms.valsPtr[j], dms.valsPtr[i]
	} else {
		dms.vals[i], dms.vals[j] = dms.vals[j], dms.vals[i]
	}
}

func (dms DynMapSlice) Len() int {
	if len(dms.valsPtr) > 0 {
		return len(dms.valsPtr)
	}

	return len(dms.vals)
}

func (dms DynMapSlice) Less(i, j int) bool {

	if len(dms.valsPtr) > 0 {
		v1, exists1 := dms.valsPtr[i].Get(dms.sortCol)
		v2, exists2 := dms.valsPtr[j].Get(dms.sortCol)
		if !exists1 {
			return false
		}
		if !exists2 {
			return true
		}
		return Less(v1, v2)
	}

	v1, exists1 := dms.vals[i].Get(dms.sortCol)
	v2, exists2 := dms.vals[j].Get(dms.sortCol)
	if !exists1 {
		return false
	}
	if !exists2 {
		return true
	}
	return Less(v1, v2)
}

// big sort func.
func Less(val1, val2 interface{}) bool {
	v1, order1 := ToSortVal(val1)
	v2, order2 := ToSortVal(val2)
	if order1 != order2 {
		return order1 < order2
	}

	// we know the values are of the same type

	switch t := v1.(type) {
	default:
		return fmt.Sprintf("%s", v1) < fmt.Sprintf("%s", v2)
	case bool:
		return t
	case int64:
		return t < v2.(int64)
	case time.Time:
		return t.Before(v2.(time.Time))
	case string:
		return t < v2.(string)
	}
}

// cleans up to one of the types we know about, attempting to convert:
// the conversion here is not as aggressive as the above methods.  Will deference any pointers
// but strings will not be attempted to be parsed
// dynmap
// time
// int64
// bool
// string
//
// integer returned is the sort order, so bool < int < string < time < map
func ToSortVal(value interface{}) (interface{}, int) {
	switch v := value.(type) {
	case bool:
		return v, 0
	case *bool:
		return *v, 0
	case time.Time:
		return v, 3
	case *time.Time:
		return *v, 3
	case string:
		return v, 2
	case *string:
		return *v, 2
	}

	v, err := ToInt64(value)
	if err == nil {
		return v, 1
	}

	d, b := ToDynMap(value)
	if b {
		return *d, 4
	}

	// couldnt convert,
	// sending back
	return value, 10
}
