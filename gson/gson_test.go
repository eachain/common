package gson

import (
	"testing"
)

func testInt(t *testing.T, s string, v int64) {
	g := FromString(s)
	if g.Type() != TypNumber {
		t.Fatal("type is:", g.Type())
	}
	if !g.IsInt() {
		t.Fatal("number should be int")
	}
	if g.Int() != v {
		t.Fatal("number is:", g.Int())
	}
}

func testIntStr(t *testing.T, s string, v int64) {
	g := FromString(s)
	if g.Type() != TypString {
		t.Fatal("type is:", g.Type())
	}
	if !g.IsInt() {
		t.Fatal("number should be int")
	}
	if g.Int() != v {
		t.Fatal("number is:", g.Int())
	}
}

func testFloat(t *testing.T, s string, v float64) {
	g := FromString(s)
	if g.Type() != TypNumber {
		t.Fatal("type is:", g.Type())
	}
	if g.IsInt() {
		t.Fatal("number should be float")
	}
	if g.Float() != v {
		t.Fatal("number is:", g.Float())
	}
}

func testFloatStr(t *testing.T, s string, v float64) {
	g := FromString(s)
	if g.Type() != TypString {
		t.Fatal("type is:", g.Type())
	}
	if g.IsInt() {
		t.Fatal("number should be float")
	}
	if g.Float() != v {
		t.Fatal("number is:", g.Float())
	}
}

func TestNumber(t *testing.T) {
	testInt(t, `123`, 123)
	testInt(t, `-123`, -123)
	testIntStr(t, `"123"`, 123)
	testIntStr(t, `"-123"`, -123)

	testFloat(t, `1.23`, 1.23)
	testFloat(t, `-1.23`, -1.23)
	testFloatStr(t, `"1.23"`, 1.23)
	testFloatStr(t, `"-1.23"`, -1.23)
}

func stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestObject(t *testing.T) {
	g := FromString(`{}`)
	if len(g.Keys()) > 0 {
		t.Fatal("object keys:", g.Keys())
	}

	g = FromString(`{"a":1, "x":"7", "b":2, "y":"8", "c": 3, "z":"9"}`)
	if !stringsEqual(g.Keys(), []string{"a", "x", "b", "y", "c", "z"}) {
		t.Fatal("object keys:", g.Keys())
	}

	g = FromString(`{"a": 123}`)
	if i := g.ObjIdx("a").Int(); i != 123 {
		t.Fatal("object value:", i)
	}
	g = FromString(`{"a": {"b": {"c": 123}}}`)
	if i := g.ObjIdx("a").ObjIdx("b").ObjIdx("c").Int(); i != 123 {
		t.Fatal("object value:", i)
	}
}

func TestList(t *testing.T) {
	g := FromString(`[]`)
	if g.Len() != 0 {
		t.Fatal("list len:", g.Len())
	}

	g = FromString("[123]")
	if i := g.Index(0).Int(); i != 123 {
		t.Fatal("list 0:", i)
	}
	g = FromString(`[[[["123"]]]]`)
	if i := g.Index(0).Index(0).Index(0).Index(0).Int(); i != 123 {
		t.Fatal("list 0:", i)
	}
	/*
		g = FromString(`[]`)
		g.Insert(0, 123)
		g.Insert(9, 789)
		g.Insert(1, 456)
		if i := g.Index(0).Int(); i != 123 {
			t.Fatal("list 0:", i)
		}
		if i := g.Index(1).Int(); i != 456 {
			t.Fatal("list 0:", i)
		}
		if i := g.Index(2).Int(); i != 789 {
			t.Fatal("list 0:", i)
		}
	*/
}

func TestAny(t *testing.T) {
	g := FromString(`{"a": 123, "b": "456", "c": "789"}`)
	if i := g.Any("a", "b").Int(); i != 123 {
		t.Fatal("object any:", i)
	}
	if i := g.Any("c", "b").Int(); i != 789 {
		t.Fatal("object any:", i)
	}
}

func TestGet(t *testing.T) {
	g := FromString(`{"a": [{"b": {"c": 123}}]}`)
	if i := g.Get("a[0].b.c").Int(); i != 123 {
		t.Fatal("object get:", i)
	}
	g = FromString(`[{"a": [[{"b": {"c": ["123"]}}]]}]`)
	if i := g.Get("[0].a[0][0].b.c[0]").Int(); i != 123 {
		t.Fatal("object get:", i)
	}

	g = FromString(`{"": {"": {"": {"": 123}}}}`)
	if i := g.Get("...").Int(); i != 123 {
		t.Fatal("object get:", i)
	}

	g = FromString(`{"": {"": {"": {"a": 123}}}}`)
	if i := g.Get("...a").Int(); i != 123 {
		t.Fatal("object get:", i)
	}
}

func TestIsNull(t *testing.T) {
	g := FromString("null")
	if !g.IsNull() {
		t.Fatal("json is not null")
	}
}

func TestInsert(t *testing.T) {
	g := &GSON{}
	g.Index(5).Set(123)
	if i := g.Index(0).Int(); i != 123 {
		t.Fatal("insert value:", i)
	}
	g.Index(0).Set(456)
	if i := g.Index(0).Int(); i != 456 {
		t.Fatal("insert value:", i)
	}
}

func TestSetKey(t *testing.T) {
	g := &GSON{}
	g.ObjIdx("a").Set("123")
	if i := g.ObjIdx("a").Int(); i != 123 {
		t.Fatal("object set value:", i)
	}
}

func TestRemove(t *testing.T) {
	g := &GSON{}
	g.Get("a.b.c").Set(123)
	if i := g.Get("a.b.c").Int(); i != 123 {
		t.Fatal("object set:", i)
	}
	g.Get("a.b.c").Remove()
	if i := g.Get("a.b.c").Int(); i == 123 {
		t.Fatal("object remove:", i)
	}
}

func TestPath(t *testing.T) {
	g := &GSON{}
	if path := g.Get("a.b.c").Path(); path != "" {
		t.Fatal("not exists path:", path)
	}
	g.Get("a.b.c").Set("123")
	if path := g.Get("a.b.c").Path(); path != "a.b.c" {
		t.Fatal("path:", path)
	}
}
