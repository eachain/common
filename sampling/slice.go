package sampling

import (
	"reflect"
)

type sliceIterator struct {
	reflect.Value
}

func (ls sliceIterator) Range(f func(interface{})) {
	n := ls.Len()
	for i := 0; i < n; i++ {
		f(ls.Index(i).Interface())
	}
}

// SliceIterator 将任意slice转化为Iterator，兼容array类型。
func SliceIterator(slice interface{}) Iterator {
	ls := reflect.Indirect(reflect.ValueOf(slice))
	switch ls.Kind() {
	case reflect.Array, reflect.Slice:
	default:
		panic("sampling: slice iterator only accept 'slice' and 'array' type")
	}
	return sliceIterator{ls}
}

type sliceWriter struct {
	reflect.Value
}

func (ls *sliceWriter) Store(i int, v interface{}) {
	n := ls.Len()
	if i < n {
		ls.Index(i).Set(reflect.ValueOf(v))
	} else if i == n {
		// slice = append(slice, value)
		ls.Set(reflect.Append(ls.Value, reflect.ValueOf(v)))
	} else {
		panic("index out of range")
	}
}

// SliceWriter 将任意slice类型转为Writer，
// 注意：slice必须是指针。
func SliceWriter(slice interface{}) Writer {
	ptr := reflect.ValueOf(slice)
	if ptr.Kind() != reflect.Ptr {
		panic("sampling: slice writer should be a pointer")
	}
	ls := ptr.Elem()
	switch ls.Kind() {
	case reflect.Slice:
	default:
		panic("sampling: slice writer only accept 'slice' type")
	}
	return &sliceWriter{ls}
}
