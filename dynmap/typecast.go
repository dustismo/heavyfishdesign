package dynmap

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//parse time
func ToTime(value interface{}) (tm time.Time, err error) {
	switch v := value.(type) {
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			return t, nil
		}
		return time.Now(), err
	case time.Time:
		return v, nil
	case *time.Time:
		return *v, nil
	}
	return time.Now(), fmt.Errorf("Unable to parse (%s) into a time", value)
}

func ToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case *bool:
		return *v, nil
	case string:
		tmp := strings.ToLower(v)
		if tmp == "true" || tmp == "t" || tmp == "yes" || tmp == "y" || tmp == "on" {
			return true, nil
		}

		if tmp == "false" || tmp == "f" || tmp == "no" || tmp == "n" || tmp == "off" {
			return false, nil
		}
	}
	return false, fmt.Errorf("Unable to convert to bool (%s)", value)
}

func MustBool(value interface{}, def bool) bool {
	i, err := ToBool(value)
	if err != nil {
		return def
	}
	return i
}

func ToInt(value interface{}) (int, error) {
	i, err := ToInt64(value)
	return int(i), err
}

func MustInt(value interface{}, def int) int {
	i, err := ToInt(value)
	if err != nil {
		return def
	}
	return i
}

func ToFloat64(value interface{}) (f float64, err error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		i, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return i, err
	}
	v, err := ToInt64(value)
	if err == nil {
		return float64(v), err
	}
	return -1, fmt.Errorf("Could not convert to float from %s", value)
}

func ToInt64(value interface{}) (i int64, err error) {
	switch v := value.(type) {
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(v), 0, 64)
		return i, err
	case *int:
		return int64(*v), nil
	case int:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		log.Printf("uhh unable to convert to int64 %+v \n", v)

	}
	return -1, fmt.Errorf("Could not convert")
}

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case *string:
		return *v
	default:
		return fmt.Sprint(value)
	}
}

func MustString(value interface{}, def string) string {
	if value == nil {
		return def
	}
	i := ToString(value)
	if len(i) == 0 {
		return def
	}
	return i
}

//Returns true if the given value is
// a map, dynmap, DynMaper or pointer of one of those types
func IsDynMapConvertable(value interface{}) bool {
	switch value.(type) {
	case DynMaper:
		return true
	case map[string]interface{}:
		return true
	case *map[string]interface{}:
		return true
	case map[string]string:
		return true
	case *map[string]string:
		return true
	case DynMap:
		return true
	case *DynMap:
		return true
	}
	return false
}

func ToDynMap(value interface{}) (*DynMap, bool) {
	switch v := value.(type) {
	case DynMaper:
		return v.ToDynMap(), true
	case map[string]interface{}:
		dynmap := New()
		dynmap.Map = v
		return dynmap, true
	case *map[string]interface{}:
		dynmap := New()
		dynmap.Map = *v
		return dynmap, true
	case map[string]string:
		dynmap := New()
		for k, v := range v {
			dynmap.Put(k, v)
		}
		return dynmap, true
	case *map[string]string:
		dynmap := New()
		for k, v := range *v {
			dynmap.Put(k, v)
		}
		return dynmap, true
	case DynMap:
		return &v, true
	case *DynMap:
		return v, true
	}
	return nil, false
}

//
// attempts to convert the given value to a map.
// returns
func ToMap(value interface{}) (map[string]interface{}, bool) {
	switch v := value.(type) {
	case map[string]interface{}:
		return v, true
	case *map[string]interface{}:
		return *v, true
	default:
		dynmap, ok := ToDynMap(value)
		if ok {
			return dynmap.Map, true
		}
	}
	return nil, false
}

// convert the object into an array if possible
func ToArray(value interface{}) ([]interface{}, bool) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice:
		retVals := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			retVals[i] = v.Index(i).Interface()
		}
		return retVals, true
	case reflect.Array:
		retVals := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			retVals[i] = v.Index(i).Interface()
		}
		return retVals, true
	default:
		return nil, false
	}

}
