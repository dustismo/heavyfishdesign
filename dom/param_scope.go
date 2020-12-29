package dom

import (
	"fmt"
	"strings"

	"github.com/dustismo/heavyfishdesign/dynmap"
)

// defines the special points on a box
type ParamType string

const (
	// Number represents a floating point number param
	// Value will either be a formula or a constant
	Number ParamType = "NUMBER"
	// Point represents an X, Y Coord
	// Value can be of the form:
	// "x, y" or { x, y }
	Point ParamType = "POINT"
	// Handle will be $TOP_LEFT or the like
	Handle ParamType = "HANDLE"
	// Unknown params
	Unknown ParamType = "UNKNOWN"
)

func ToParamType(str string) ParamType {

	switch strings.ToUpper(str) {
	case string(Number):
		return Number
	case string(Point):
		return Point
	case string(Handle):
		return Handle
	default:
		return Unknown
	}
}

// Stateful parameter scope
// this should collect params during render and should
// represent all parameters values by the end of the render

type Param struct {
	Key           string
	ElementID     string
	RealizedValue interface{}
	Type          ParamType
	Value         string
	Description   string
	// If param gets overwritten because its Id is
	// non-unique
	// generally non-unique params should not be available
	// to other components
	Unique   bool
	lookedUp bool
	owner    Element
}

func (p *Param) ToHFD() *dynmap.DynMap {
	mp := dynmap.New()
	mp.Put("value", p.Value)
	if len(p.Description) > 0 {
		mp.Put("description", p.Description)
	}
	if len(string(p.Type)) > 0 {
		mp.Put("type", string(p.Type))
	}
	return mp
}

func ParseParams(paramsMap *dynmap.DynMap) ([]*Param, error) {
	retval := []*Param{}
	for _, k := range paramsMap.Keys() {
		v, _ := paramsMap.Get(k)
		mp := dynmap.New()
		if dynmap.IsDynMapConvertable(v) {
			mp, _ = dynmap.ToDynMap(v)
		} else {
			// only has value
			mp.Put("value", v)
		}
		mp.PutIfAbsent("key", k)
		p := &Param{
			Key:         k,
			Type:        ToParamType(mp.MustString("type", "")),
			Value:       mp.MustString("value", ""),
			Description: mp.MustString("description", ""),
		}
		retval = append(retval, p)
	}
	return retval, nil
}

type ParamTracker struct {
	params map[string]*Param
}

func (pt *ParamTracker) toKey(key string, elementID string) string {
	return fmt.Sprintf("%s__%s", key, elementID)
}

func (pt *ParamTracker) Set(p *Param) {
	pt.params[pt.toKey(p.Key, p.ElementID)] = p
}

func (pt *ParamTracker) SetFromElement(e Element) {
	// id := e.Id()
	// params := e.Params()
	// for k, v := range params.Map {

	// }
}
