package sampling

import (
	"reflect"
)

type mapKeys []reflect.Value

func (keys mapKeys) Range(f func(interface{})) {
	for _, key := range keys {
		f(key.Interface())
	}
}

// MapKeysIterator 将任意map的所有key生成一个Iterator。
func MapKeysIterator(m interface{}) Iterator {
	v := reflect.Indirect(reflect.ValueOf(m))
	if v.Kind() != reflect.Map {
		panic("sampling: map keys interator should be a map")
	}
	return mapKeys(v.MapKeys())
}

type mapValues []reflect.Value

func (vals mapValues) Range(f func(interface{})) {
	for _, val := range vals {
		f(val.Interface())
	}
}

// MapValuesIterator 将任意map的所有value生成一个Iterator。
func MapValuesIterator(m interface{}) Iterator {
	v := reflect.Indirect(reflect.ValueOf(m))
	if v.Kind() != reflect.Map {
		panic("sampling: map keys interator should be a map")
	}
	keys := v.MapKeys()
	vals := make([]reflect.Value, 0, len(keys))
	for _, key := range keys {
		vals = append(vals, v.MapIndex(key))
	}
	return mapValues(vals)
}

type MapPair struct {
	Key, Value interface{}
}

type mapIterator []MapPair

func (pairs mapIterator) Range(f func(interface{})) {
	for _, p := range pairs {
		f(p)
	}
}

// MapIterator 将任意map类型转为Iterator，
// Range(func(v interface{})) 中的`v`必为MapPair类型。
func MapIterator(m interface{}) Iterator {
	v := reflect.Indirect(reflect.ValueOf(m))
	if v.Kind() != reflect.Map {
		panic("sampling: map interator should be a map")
	}
	keys := v.MapKeys()
	pairs := make([]MapPair, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, MapPair{key.Interface(), v.MapIndex(key).Interface()})
	}
	return mapIterator(pairs)
}

type mapWriter struct {
	isPtr bool
	ptr   reflect.Value
	val   reflect.Value
	keys  []reflect.Value
}

func (w *mapWriter) Store(i int, v interface{}) {
	p := v.(MapPair)
	if w.isPtr && w.val.IsNil() {
		w.val = reflect.MakeMap(w.ptr.Type().Elem())
		w.ptr.Elem().Set(w.val)
	}
	key := reflect.ValueOf(p.Key)
	val := reflect.ValueOf(p.Value)
	if i < len(w.keys) {
		w.val.SetMapIndex(w.keys[i], reflect.Value{})
		w.keys[i] = key
	} else if i == len(w.keys) {
		w.keys = append(w.keys, key)
	} else {
		panic("index out of range")
	}
	w.val.SetMapIndex(key, val)
}

// MapWriter 将任意map类型转为Writer，
// Store(i int, v interface{}) 需要保证`v`是MapPair类型。
func MapWriter(m interface{}) Writer {
	w := &mapWriter{}
	val := reflect.ValueOf(m)
	if val.Kind() == reflect.Ptr {
		w.isPtr = true
		w.ptr = val
		val = val.Elem()
	}
	if val.Kind() != reflect.Map {
		panic("sampling: map writer should be a map")
	}
	w.val = val
	return w
}
