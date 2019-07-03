package utils

import (
	"reflect"
)

/*
TypeConverter 是一个类型转换工具.
它常用于将结构体中的某一字段类型转换成其它类型，
以便于操作，比如：

	type T struct {
		Dur time.Duration `json:"dur"`
	}

	type Duration time.Duration

	func (dur *Duration) UnmarshalJSON(p []byte) error {
		d, err := time.ParseDuration(string(p[1 : len(p)-1]))
		if err != nil {
			return err
		}
		*dur = Duration(d)
		return nil
	}

	func (dur Duration) MarshalJSON() ([]byte, error) {
		return []byte("\"" + time.Duration(dur).String() + "\""), nil
	}

	b, _ := json.Marshal(T{2 * time.Second})
	fmt.Println(string(b)) // Output: {"dur":2000000000}

	tc := make(TypeConverter)
	tc.Add(time.Duration(0), Duration(0))
	p, _ := json.Marshal(tc.ReplaceStructFieldType(T{2 * time.Second}))
	fmt.Println(string(p)) // Output: {"dur":"2s"}
*/
type TypeConverter map[reflect.Type]reflect.Type

func (tc TypeConverter) Add(old, new interface{}) {
	src := reflect.TypeOf(old)
	dst := reflect.TypeOf(new)
	if src.ConvertibleTo(dst) {
		tc[src] = dst
	}
}

func (tc TypeConverter) Convert(x interface{}) interface{} {
	v := reflect.ValueOf(x)
	t, ok := tc[v.Type()]
	if !ok {
		return x
	}
	return v.Convert(t).Interface()
}

func (tc TypeConverter) Cached(x interface{}) bool {
	_, ok := tc[reflect.TypeOf(x)]
	return ok
}

func (tc TypeConverter) ReplaceStructFieldType(x interface{}) interface{} {
	t := reflect.TypeOf(x)
	if typ, ok := tc[t]; ok {
		return reflect.ValueOf(x).Convert(typ).Interface()
	}

	isPtr := false
	if t.Kind() == reflect.Ptr {
		isPtr = true
		t = t.Elem()
		if typ, ok := tc[t]; ok {
			return reflect.ValueOf(x).Convert(reflect.PtrTo(typ)).Interface()
		}
	} else {
		if typ, ok := tc[reflect.PtrTo(t)]; ok {
			return reflect.ValueOf(x).Convert(typ.Elem()).Interface()
		}
	}

	if t.Kind() != reflect.Struct {
		return x
	}

	modified := false
	n := t.NumField()
	fields := make([]reflect.StructField, n)
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		if typ, ok := tc[ft.Type]; ok {
			ft.Type = typ
			modified = true
		}
		fields[i] = ft
	}
	if !modified {
		return x
	}

	typ := reflect.StructOf(fields)
	if isPtr {
		typ = reflect.PtrTo(typ)
	}
	return reflect.ValueOf(x).Convert(typ).Interface()
}
