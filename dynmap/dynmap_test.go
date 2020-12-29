package dynmap

import (
	"log"
	"strings"
	"testing"
)

func TestJsonMarshal(t *testing.T) {
	mp := New()
	mp.PutWithDot("this.that.test", 80)
	mp.PutWithDot("this.eight", 8)

	inner := New()
	inner.Put("ex", "val")
	mp.PutWithDot("this.nine", *inner)

	bytes, _ := mp.MarshalJSON()
	log.Printf("Got JSON %s", bytes)
	un := New()
	un.UnmarshalJSON(bytes)

	unbytes, _ := mp.MarshalJSON()
	if string(unbytes) != string(bytes) {
		t.Errorf("JSon marshal failed (%s) != (%s)", unbytes, bytes)
	}

	if un.MustString("this.nine.ex", "") != "val" {
		t.Errorf("JSon marshal failed (%s)", bytes)
	}
}

func TestPutAll(t *testing.T) {
	mp := make(map[string]string)
	mp["key1"] = "val1"
	mp["key2"] = "key2"
	dm := New()
	dm.PutAll(mp)
	if dm.MustString("key1", "") != "val1" {
		t.Errorf("Error in string map putall")
	}
}

func TestMerge(t *testing.T) {
	mp := New()
	mp.PutWithDot("this.that.test", 80)
	mp.PutWithDot("this.eight", 8)
	mp.PutWithDot("this.that.also", 18)

	mp2 := New()
	mp2.PutWithDot("this.that.test2", 23)
	mp2.PutWithDot("this.second", 2)

	mp.Merge(mp2)

	if mp.MustInt("this.that.test2", 0) != 23 {
		t.Errorf("Error on Recursive Merge: %s", mp.ToJSON())
	}

	if mp.MustInt("this.that.test", 0) != 80 {
		t.Errorf("Error on Recursive Merge: %s", mp.ToJSON())
	}

	if mp.MustInt("this.second", 0) != 2 {
		t.Errorf("Error on Recursive Merge: %s", mp.ToJSON())
	}
}

func TestFlatten(t *testing.T) {
	mp := New()
	mp.PutWithDot("this.that.test", 80)
	mp.PutWithDot("this.eight", 8)
	mp.PutWithDot("this.that.also", 18)

	flattened := mp.Flatten()
	if flattened["this.that.test"] != 80 {
		t.Errorf("Error during flatten")
	}
	if flattened["this.eight"] != 8 {
		t.Errorf("Error during flatten")
	}

	mp = New()

	inner := New()
	inner.PutWithDot("key1.key2", "blah")

	mp.AddToSliceWithDot("something.array", inner)
	mp.AddToSliceWithDot("something.array", 2)

	flattened = mp.Flatten()
	if flattened["something.array.1"] != 2 {
		t.Errorf("Error during array flatten")
	}

	if flattened["something.array.0.key1.key2"] != "blah" {
		t.Error("Error during array dynmap flatten")
	}
}

func TestRemove(t *testing.T) {
	mp := New()
	mp.PutWithDot("this.that.test", 80)
	mp.PutWithDot("this.eight", 8)
	mp.PutWithDot("this.that.also", 18)

	mp.Remove("this.that.test")
	if mp.MustInt("this.that.test", 0) != 0 {
		t.Errorf("Error on dot operator remove: %s", mp.ToJSON())
	}

	if mp.MustInt("this.that.also", 0) != 18 {
		t.Errorf("Dot operator remove also removed what it shouldnt")
	}
}

func TestToString(t *testing.T) {
	mp := New()
	mp.Put("testempty", "")
	mp.Put("testnumber", 10)

	str, b := mp.GetString("testempty")
	if b {
		t.Errorf("Empty string should return false")
	}

	str, b = mp.GetString("testnonexist")
	if b {
		t.Errorf("Non existent mapping should return false")
	}

	str, b = mp.GetString("testnumber")
	if str != "10" {
		t.Errorf("Incorrect conversion from number to string")
	}
}

// func TestURLEncode(t *testing.T) {
// 	mp := New()
// 	mp.PutWithDot("this.that.test", 80)
// 	mp.PutWithDot("this.eight", 8)
// 	url, err := mp.MarshalUrl()
// 	if err != nil {
// 		t.Errorf("Error in url %s", err)
// 	}

// 	log.Printf("Got URL : %s", url)

// 	un := New()
// 	un.UnmarshalUrl(url)

// 	if un.MustInt("this.that.test", 0) != 80 {
// 		t.Errorf("Unmarshal URL failure ")
// 	}
// }

func TestAddToSlice(t *testing.T) {
	mp := New()
	mp.AddToSliceWithDot("key.key2", "as")

	mp.AddToSlice("key2", "as2")
	mp.AddToSlice("key2", "as3")
	mp.AddToSlice("key2", "as4", "as5")
	val, _ := mp.GetStringSlice("key.key2")
	if len(val) != 1 {
		t.Errorf("Error in add to AddToSliceWithDot")
	}

	val, _ = mp.GetStringSlice("key2")
	if len(val) != 4 {
		t.Errorf("Error in add to AddToSlice")
	}
	// add dynmap
	mpInner := New()
	mpInner.Put("Example", "1")
	mp.AddToSliceWithDot("key.key3", mpInner)
	mp.AddToSlice("key3", mpInner)

	vD, _ := mp.GetDynMapSlice("key.key3")
	if len(vD) != 1 {
		t.Errorf("Error in add to AddToSlice DynMap")
	}

	bytes, _ := mp.MarshalJSON()
	log.Println(string(bytes))

}

type TestStruct struct {
	Key   string
	Value DynMap
}

func TestStructSerialize(t *testing.T) {
	mp := New()

	subMp := New()
	subMp.Put("subkey1", "1")
	subMp.Put("subkey2", "2")

	strct := &TestStruct{
		Key:   "key1",
		Value: *subMp,
	}

	log.Println(strct)
	mp.Put("submp", strct)
	mp.Put("submp2", "adsfasdf")

	bytes, _ := mp.MarshalJSON()

	// make sure we arent serializing DynMap.Map
	if strings.Contains(string(bytes), "map") {
		t.Errorf("Map serialization error")
	}
}
