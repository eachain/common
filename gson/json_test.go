package gson

import (
	"testing"
)

func TestKey(t *testing.T) {
	s := `{"a":{"b":{"c":{"d": "123"}}}}`
	js := FromString(s)
	d := js.Key("a").Key("b").Key("c").Key("d")
	if d.Int() != 123 {
		t.Fatal(d.Err())
	}
}

func TestIndex(t *testing.T) {
	s := `[[[0],[1],[2]], [[10],[11],[12]]]`
	js := FromString(s)
	for i := 0; i < js.Len(); i++ {
		for j := 0; j < js.Index(i).Len(); j++ {
			x := js.Index(i).Index(j).Index(0)
			if x.Int() != int64(i*10+j) {
				t.Fatal(i, j, x.Err())
			}
		}
	}
}

func TestKeyIndex(t *testing.T) {
	s := `{"a":[{"b":"0"},{"b":1},{"b":"2"},{"b":3}]}`
	js := FromString(s)
	for i := 0; i < js.Key("a").Len(); i++ {
		if js.Key("a").Index(i).Key("b").Int() != int64(i) {
			t.Fatal(i)
		}
	}
}

func TestGet1(t *testing.T) {
	s := `{"a":[[[],[{},{},{"b":123}]]]}`
	js := FromString(s)
	if js.Get("a[0][1][2].b").Int() != 123 {
		t.Fatal(js.Get("a[0][1][2].b").Err())
	}
}

func TestGet2(t *testing.T) {
	s := `[[[],[{},{},{"a":[[[{"b":[123]}]]]}]]]`
	js := FromString(s)
	k := "[0][1][2].a[0][0][0].b[0]"
	if js.Get(k).Int() != 123 {
		t.Fatal(js.Get(k).Err())
	}
}

func TestAny(t *testing.T) {
	s := `{"errcode": 123}`
	js := FromString(s).Any("errcode", "err_code")
	if js.Int() != 123 {
		t.Fatal(js.Err())
	}

	s = `{"err_code": "123"}`
	js = FromString(s).Any("errcode", "err_code")
	if js.Int() != 123 {
		t.Fatal(js.Err())
	}
}
