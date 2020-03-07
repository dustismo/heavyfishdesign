package dynmap

import (
	"testing"
)

//go test -v github.com/dustismo/open-books/dynmap
func TestDynMapSort(t *testing.T) {
	arr := make([]DynMap, 5, 5)
	arr[0] = *New()
	arr[0].PutWithDot("test", "z")
	arr[1] = *New()
	arr[1].PutWithDot("test", "x")
	arr[2] = *New()
	arr[2].PutWithDot("test", "w")
	arr[3] = *New()
	arr[3].PutWithDot("test", "n")
	arr[4] = *New()
	arr[4].PutWithDot("test", "m")

	Sort(arr, "test")
	if arr[4].MustString("test", "") != "z" {
		t.Errorf("array sort failed.  Expected z, got %s", arr[4].MustString("test", ""))
	}
}
