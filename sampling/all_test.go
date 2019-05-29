package sampling

import (
	"testing"
)

func TestSlice(t *testing.T) {
	src := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	var dst []int
	Sampling(SliceWriter(&dst), SliceIterator(src), 3)
	if len(dst) != 3 {
		t.Fatalf("dst size is not 3, but %v", len(dst))
	}
	t.Log(dst)
}

var srcMap = map[int]string{1: "a", 2: "b", 3: "c", 4: "d", 5: "e", 6: "f", 7: "g"}

func TestMapKeys(t *testing.T) {
	var dst []int
	Sampling(SliceWriter(&dst), MapKeysIterator(srcMap), 8)
	t.Log(dst)
}

func TestMapVals(t *testing.T) {
	var dst []string
	Sampling(SliceWriter(&dst), MapValuesIterator(srcMap), 8)
	t.Log(dst)
}

func TestMap1(t *testing.T) {
	var dst = make(map[int]string)
	Sampling(MapWriter(dst), MapIterator(srcMap), 3)
	t.Log(dst)
}

func TestMap2(t *testing.T) {
	var dst map[int]string
	Sampling(MapWriter(&dst), MapIterator(srcMap), 3)
	t.Log(dst)
}

func TestMap3(t *testing.T) {
	var dst map[int]interface{}
	Sampling(MapWriter(&dst), MapIterator(srcMap), 3)
	t.Log(dst)
}
