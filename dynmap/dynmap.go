package dynmap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
)

//Dont make this a map type, since we want the option of
//extending this and adding members.
type DynMap struct {
	//Map        map[string]interface{} `bson:",inline"`
	OrderedMap *OrderedMap
}

type DynMaper interface {
	ToDynMap() *DynMap
}

func CreateFromOrderedMap(mp *OrderedMap) *DynMap {
	return &DynMap{
		mp,
	}
}

func CreateFromMap(mp map[string]interface{}) *DynMap {
	return CreateFromOrderedMap(NewOrderedMapFromMap(mp))
}

// Creates a new dynmap
func New() *DynMap {
	return &DynMap{
		NewOrderedMap(),
	}
}

func ParseJSON(json string) (*DynMap, error) {
	// strip any comments
	re := regexp.MustCompile("(?s)//.*?\n|/\\*.*?\\*/")
	newBytes := re.ReplaceAll([]byte(json), nil)
	mp := New()
	return mp, mp.UnmarshalJSON(newBytes)
}

// attempts to convert the passed in object to
// a dynmap then JSON, ignoring any errors
func PrettyJSON(obj interface{}) string {
	d, err := Convert(obj)
	if err != nil {
		return err.Error()
	}
	return d.ToJSON()
}

// creates a dynmap from the passed in object, if possible
func Convert(obj interface{}) (*DynMap, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	dm := New()
	err = dm.UnmarshalJSON(b)
	return dm, err
}

// Attempts to fill the given struct with the values contained
// in this dynmap.
func (this *DynMap) ConvertTo(val interface{}) error {
	// do the easy thing, and just use json as an intermediary
	bytes, err := this.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, val)
}

// flattens this dynmap into a one level map
// where all keys use the dot operator, and values
// are not maps (primitives or objects)
// arrays and slices are indexed by integer
// for instance
// key1: ["v1", "v2"] => key1.0:v1, key1.1:v2
// Note that the slice syntax is not supported yet by PutWithDot so
// flattened keys cannot be automatically used to recreate the nested map (todo)
//
func (this *DynMap) Flatten() map[string]interface{} {
	mp := make(map[string]interface{})
	for _, k := range this.OrderedMap.Keys() {
		dm, ok := this.GetDynMap(k)
		if ok {
			for k2, v2 := range dm.Flatten() {
				mp[fmt.Sprintf("%s.%s", k, k2)] = v2
			}
		} else {
			// try as a slice
			v, _ := this.OrderedMap.Get(k)
			slice, ok := v.([]interface{})
			if ok {
				// need to flatten the members of the slice
				for i, sliceItem := range slice {
					dmItem, ok := ToDynMap(sliceItem)
					if ok {
						for k2, v2 := range dmItem.Flatten() {
							mp[fmt.Sprintf("%s.%d.%s", k, i, k2)] = v2
						}
					} else {
						mp[fmt.Sprintf("%s.%d", k, i)] = sliceItem
					}
				}
			} else {
				mp[k] = v
			}
		}
	}
	return mp
}

// Takes any dot operator keys and makes nested maps
// returns a new DynMap instance
func (dm *DynMap) UnFlatten() (*DynMap, error) {
	newMp := New()
	for _, key := range dm.OrderedMap.Keys() {
		value, _ := dm.OrderedMap.Get(key)
		err := newMp.PutWithDot(key, value)
		if err != nil {
			return dm, err
		}
	}
	return newMp, nil
}

// Recursively converts this to a regular go map.
// (will convert any sub maps)
func (this *DynMap) ToMap() map[string]interface{} {
	mp := make(map[string]interface{})
	for _, k := range this.OrderedMap.Keys() {
		v, _ := this.OrderedMap.Get(k)
		submp, ok := ToDynMap(v)
		if ok {
			v = submp.ToMap()
		}
		mp[k] = v
	}
	return mp
}

// Returns self. Here so that we satisfy the DynMaper interface
func (this *DynMap) ToDynMap() *DynMap {
	return this
}

// Recursive clone
func (dm *DynMap) Clone() *DynMap {
	return New().Merge(dm)
}

// recursively merges the requested dynmap into the current dynmap
// returns self in order to support chaining.
func (this *DynMap) Merge(mp *DynMap) *DynMap {
	for _, key := range mp.OrderedMap.Keys() {
		value, _ := mp.OrderedMap.Get(key)
		m, ok := mp.GetDynMap(key)
		if ok {
			m2, ok := this.GetDynMap(key)
			if ok {
				m2.Merge(m)
			} else {
				this.Put(key, m.Clone())
			}
		} else {
			this.Put(key, value)
		}
	}
	return this
}

//encodes this map into a url encoded string.
//maps are encoded in the rails style (key[key2][key2]=value)
// TODO: we should sort the keynames so ordering is consistent and then this
// can be used a cache key
func (this *DynMap) MarshalUrl() (string, error) {
	vals := &url.Values{}
	for _, key := range this.OrderedMap.Keys() {
		value, _ := this.OrderedMap.Get(key)
		err := this.urlEncode(vals, key, value)
		if err != nil {
			return "", err
		}
	}

	str := vals.Encode()
	return str, nil
}

// Unmarshals a url encoded string.
// will also parse rails style maps in the form key[key1][key2]=val
func (this *DynMap) UnmarshalUrl(urlstring string) error {
	//TODO: split on ?
	values, err := url.ParseQuery(urlstring)
	if err != nil {
		return err
	}

	return this.UnmarshalUrlValues(values)
}

// Unmarshals url.Values into the map.
// Will correctly handle rails style maps in the form key[key1][key2]=val
func (this *DynMap) UnmarshalUrlValues(values url.Values) error {
	for k := range values {
		var v = values[k]
		key := strings.Replace(k, "[", ".", -1)
		key = strings.Replace(key, "]", "", -1)

		if len(v) == 1 {
			this.PutWithDot(key, v[0])
		} else {
			this.PutWithDot(key, v)
		}
	}
	return nil
}

//adds the requested value to the Values
func (this *DynMap) urlEncode(vals *url.Values, key string, value interface{}) error {

	if IsDynMapConvertable(value) {
		mp, ok := ToDynMap(value)
		if !ok {
			return fmt.Errorf("Unable to convert %+v", mp)
		}
		for _, k := range mp.OrderedMap.Keys() {
			v, _ := mp.OrderedMap.Get(key)
			//encode in rails style key[key2]=value
			this.urlEncode(vals, fmt.Sprintf("%s[%s]", key, k), v)
		}
		return nil
	}
	r := reflect.ValueOf(value)
	//now test if it is an array
	if r.Kind() == reflect.Array || r.Kind() == reflect.Slice {
		for i := 0; i < r.Len(); i++ {
			this.urlEncode(vals, key, r.Index(i).Interface())
		}
	}

	vals.Add(key, ToString(value))
	return nil
}

func (this DynMap) MarshalJSON() ([]byte, error) {
	return this.OrderedMap.MarshalJSON()
}

// converts to indented json, throws away any errors
// this is useful for logging purposes.  MarshalJSON
// should be used for most uses.
func (this DynMap) ToJSON() string {
	bytes, _ := this.MarshalJSON()
	return string(bytes)
}

func (this *DynMap) UnmarshalJSON(bytes []byte) error {
	return this.OrderedMap.UnmarshalJSON(bytes)
}

func (this *DynMap) Length() int {
	return this.OrderedMap.Length()
}

func (this *DynMap) IsEmpty() bool {
	return this.Length() == 0
}

func (this *DynMap) Keys() []string {
	return this.OrderedMap.Keys()
}

// Gets the value at the specified key as an int64.  returns -1,false if value not available or is not convertable
func (this DynMap) GetInt64(key string) (int64, bool) {
	tmp, ok := this.Get(key)
	if !ok {
		return -1, ok
	}
	val, err := ToInt64(tmp)
	if err == nil {
		return val, true
	}
	return -1, false
}

func (this DynMap) MustInt64(key string, def int64) int64 {
	v, ok := this.GetInt64(key)
	if ok {
		return v
	}
	return def
}

func (this DynMap) MustInt(key string, def int) int {
	v, ok := this.GetInt(key)
	if ok {
		return v
	}
	return def
}

func (this DynMap) GetInt(key string) (int, bool) {
	v, ok := this.GetInt64(key)
	if !ok {
		return -1, ok
	}
	return int(v), true
}

func (this DynMap) GetFloat64(key string) (float64, bool) {
	tmp, ok := this.Get(key)
	if !ok {
		return -1, ok
	}
	val, err := ToFloat64(tmp)
	if err == nil {
		return val, true
	}
	return -1, false
}

func (this DynMap) MustFloat64(key string, def float64) float64 {
	v, ok := this.GetFloat64(key)
	if ok {
		return v
	}
	return def
}

func (this DynMap) Contains(key string) bool {
	_, ok := this.Get(key)
	return ok
}

func (this DynMap) ContainsAll(keys ...string) bool {
	for _, k := range keys {
		if !this.Contains(k) {
			return false
		}
	}
	return true
}

func (this DynMap) ContainsString(key string) bool {
	_, ok := this.GetString(key)
	return ok
}

func (this DynMap) ContainsDynMap(key string) bool {
	_, ok := this.GetDynMap(key)
	return ok
}

//
// Gets a string representation of the value at key
//
func (this DynMap) GetString(key string) (string, bool) {
	tmp, ok := this.Get(key)
	if !ok {
		return "", ok
	}
	str := ToString(tmp)
	if len(str) == 0 {
		return str, false
	}
	return str, true
}

// gets a string. if string is not available in the map, then the default
//is returned
func (this DynMap) MustString(key string, def string) string {
	tmp, ok := this.GetString(key)
	if !ok {
		return def
	}
	return tmp
}

func (this DynMap) GetTime(key string) (time.Time, bool) {
	tmp, ok := this.Get(key)
	if !ok {
		return time.Now(), false
	}
	t, err := ToTime(tmp)
	if err != nil {
		return time.Now(), false
	}
	return t, true
}

func (this DynMap) MustTime(key string, def time.Time) time.Time {
	tmp, ok := this.GetTime(key)
	if !ok {
		return def
	}
	return tmp
}

func (this DynMap) GetBool(key string) (bool, bool) {
	tmp, ok := this.Get(key)
	if !ok {
		return false, ok
	}
	b, err := ToBool(tmp)
	if err != nil {
		return false, false
	}
	return b, true
}

func (this DynMap) MustBool(key string, def bool) bool {
	tmp, ok := this.GetBool(key)
	if !ok {
		return def
	}
	return tmp
}

//Gets a dynmap from the requested.
// This will update the value in the map if the
// value was not already a dynmap.
func (this DynMap) GetDynMap(key string) (*DynMap, bool) {
	tmp, ok := this.Get(key)
	if !ok {
		return nil, ok
	}
	mp, ok := ToDynMap(tmp)
	return mp, ok
}

func (this DynMap) MustDynMap(key string, def *DynMap) *DynMap {
	tmp, ok := this.GetDynMap(key)
	if !ok {
		return def
	}
	return tmp
}

// gets a slice of dynmaps
func (this DynMap) GetDynMapSlice(key string) ([]*DynMap, bool) {
	lst, ok := this.Get(key)
	if !ok {
		return nil, false
	}
	switch v := lst.(type) {
	case []*DynMap:
		return v, true
	case []interface{}:
		retlist := make([]*DynMap, 0)
		for _, tmp := range v {
			in, ok := ToDynMap(tmp)
			if !ok {
				return nil, false
			}
			retlist = append(retlist, in)
		}
		return retlist, true
	}
	return nil, false
}

//Returns a slice of ints
func (this DynMap) GetIntSlice(key string) ([]int, bool) {
	lst, ok := this.Get(key)
	if !ok {
		return nil, false
	}
	switch v := lst.(type) {
	case []int:
		return v, true
	case []interface{}:
		retlist := make([]int, 0)
		for _, tmp := range v {
			in, err := ToInt(tmp)
			if err != nil {
				return nil, false
			}
			retlist = append(retlist, in)
		}
		return retlist, true
	}
	return nil, false
}

//gets a slice of ints.  if the value is a string it will
//split by the requested delimiter
func (this DynMap) GetIntSliceSplit(key, delim string) ([]int, bool) {
	lst, ok := this.Get(key)
	if !ok {
		return nil, false
	}
	switch v := lst.(type) {
	case string:
		retlist := make([]int, 0)
		for _, tmp := range strings.Split(v, delim) {
			in, err := ToInt(tmp)
			if err != nil {
				return nil, false
			}
			retlist = append(retlist, in)
		}
		return retlist, true
	}
	ret, ok := this.GetIntSlice(key)
	return ret, ok
}

func (this DynMap) MustStringSlice(key string, def []string) []string {
	lst, ok := this.GetStringSlice(key)
	if !ok {
		return def
	}
	return lst
}

func (this DynMap) MustDynMapSlice(key string, def []*DynMap) []*DynMap {
	lst, ok := this.GetDynMapSlice(key)
	if !ok {
		return def
	}
	return lst
}

//Returns a slice of strings
func (this DynMap) GetStringSlice(key string) ([]string, bool) {
	lst, ok := this.Get(key)
	if !ok {
		return nil, false
	}
	switch v := lst.(type) {
	case []string:
		return v, true
	case []interface{}:
		retlist := make([]string, 0)
		for _, tmp := range v {
			in := ToString(tmp)
			retlist = append(retlist, in)
		}
		return retlist, true
	}
	return nil, false
}

//gets a slice of strings.  if the value is a string it will
//split by the requested delimiter
func (this DynMap) GetStringSliceSplit(key, delim string) ([]string, bool) {
	lst, ok := this.Get(key)
	if !ok {
		return nil, false
	}
	switch v := lst.(type) {
	case string:
		return strings.Split(v, delim), true
	}
	ret, ok := this.GetStringSlice(key)
	return ret, ok
}

// Adds the item to a slice
func (this DynMap) AddToSlice(key string, mp ...interface{}) error {
	this.PutIfAbsent(key, make([]interface{}, 0))
	lst, _ := this.Get(key)
	switch v := lst.(type) {
	case []interface{}:
		v = append(v, mp...)
		this.Put(key, v)
	}
	return nil
}

// Adds the item to a slice
func (this DynMap) AddToSliceWithDot(key string, mp ...interface{}) error {
	this.PutIfAbsentWithDot(key, make([]interface{}, 0))
	lst, _ := this.Get(key)
	switch v := lst.(type) {
	case []interface{}:
		v = append(v, mp...)
		this.PutWithDot(key, v)
	}
	return nil
}

// puts all the values from the passed in map into this dynmap
// the passed in map must be convertable to a DynMap via ToDynMap.
// returns false if the passed value is not convertable to dynmap
func (this *DynMap) PutAll(mp interface{}) bool {
	dm, ok := ToDynMap(mp)
	if !ok {
		return false
	}
	for _, k := range dm.OrderedMap.Keys() {
		v, _ := dm.OrderedMap.Get(k)
		this.Put(k, v)
	}
	return true
}

//
// Puts the value into the map if and only if no value exists at the
// specified key.
// This does not honor the dot operator on insert.
func (this *DynMap) PutIfAbsent(key string, value interface{}) (interface{}, bool) {
	v, ok := this.Get(key)
	if ok {
		return v, false
	}
	this.Put(key, value)
	return value, true
}

//
// Same as PutIfAbsent but honors the dot operator
//
func (this *DynMap) PutIfAbsentWithDot(key string, value interface{}) (interface{}, bool) {
	v, ok := this.Get(key)
	if ok {
		return v, false
	}
	this.PutWithDot(key, value)
	return value, true
}

//
// Put's a value into the map
//
func (this *DynMap) Put(key string, value interface{}) {
	this.OrderedMap.Set(key, this.beforePut(value))
}

func (this *DynMap) beforePut(value interface{}) interface{} {
	// allow some tranformation here
	d, ok := value.(DynMaper)
	if ok {
		return d.ToDynMap()
	} else {
		return value
	}
}

//
// puts the value into the map, honoring the dot operator.
// so PutWithDot("map1.map2.value", 100)
// would result in:
// {
//   map1 : { map2 : { value: 100 }}
//
// }
func (this *DynMap) PutWithDot(key string, value interface{}) error {
	splitStr := strings.Split(key, ".")
	if len(splitStr) == 1 {
		this.Put(key, value)
		return nil
	}
	mapKeys := splitStr[:(len(splitStr) - 1)]
	var mp = this.OrderedMap
	for _, k := range mapKeys {
		tmp, o := mp.Get(k)
		if !o {
			//create a new map and insert
			newmap := NewOrderedMap()
			mp.Set(k, newmap)
			mp = newmap
		} else {
			mp, o = ToOrderedMap(tmp)
			if !o {
				//error
				return errors.New("Error, value at key was not a map")
			}
		}
	}
	mp.Set(splitStr[len(splitStr)-1], this.beforePut(value))
	return nil
}

func (this *DynMap) Exists(key string) bool {
	_, ok := this.Get(key)
	return ok
}

func (dm *DynMap) RemoveAll(key ...string) {
	for _, k := range key {
		dm.Remove(k)
	}
}

//Remove a mapping
func (this *DynMap) Remove(key string) (interface{}, bool) {
	val, ok := this.OrderedMap.Get(key)
	if ok {
		this.OrderedMap.Delete(key)
		return val, true
	}
	// dot op..
	splitStr := strings.Split(key, ".")
	if len(splitStr) == 1 {
		return val, false
	}
	var mp = this.OrderedMap
	for index, k := range splitStr {
		tmp, o := mp.Get(k)
		if !o {
			return val, ok
		}

		if index == (len(splitStr) - 1) {
			mp.Delete(k)
			return tmp, o
		} else {
			mp, o = ToOrderedMap(tmp)
			if !o {
				return val, ok
			}
		}
	}
	return val, false

}

func (this *DynMap) Must(key string, def interface{}) interface{} {
	val, ok := this.Get(key)
	if ok {
		return val
	}
	return def
}

//
// Get's the value.  will honor the dot operator if needed.
// key = 'map.map2'
// will first attempt to matche the literal key 'map.map2'
// if no value is present it will look for a sub map at key 'map'
//
func (this *DynMap) Get(key string) (interface{}, bool) {
	val, ok := this.OrderedMap.Get(key)
	if ok {
		return val, true
	}
	//look for dot operator.
	splitStr := strings.Split(key, ".")
	if len(splitStr) == 1 {
		return val, false
	}

	var mp = this.OrderedMap
	for index, k := range splitStr {
		tmp, o := mp.Get(k)
		if !o {
			return val, ok
		}

		if index == (len(splitStr) - 1) {
			return tmp, o
		} else {
			mp, o = ToOrderedMap(tmp)
			if !o {
				return val, ok
			}
		}
	}
	return val, ok
}
