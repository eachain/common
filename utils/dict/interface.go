package dict

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// Map 类似于 sync.Map, 函数`LoadOrStore`参数 value 改为new,
// 用于省略不必要的对象创建.
type Map interface {
	Load(key interface{}) (value interface{}, ok bool)
	LoadOrStore(key interface{}, new func() interface{}) (actual interface{}, loaded bool)
	Store(key, value interface{})
	Delete(key interface{})
	Range(f func(key, value interface{}) bool)
}

// Followings are for json (un)marshal.

type iterator interface {
	Range(f func(key, value interface{}) bool)
}

func marshalJSON(it iterator) ([]byte, error) {
	var buf bytes.Buffer
	var enc = json.NewEncoder(&buf)
	var err error
	var i int

	buf.WriteByte('{')
	it.Range(func(k, v interface{}) bool {
		if i > 0 {
			buf.WriteByte(',')
		}
		err = enc.Encode(k)
		if err != nil {
			return false
		}
		buf.WriteByte(':')
		err = enc.Encode(v)
		if err != nil {
			return false
		}
		i++
		return true
	})
	buf.WriteByte('}')

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type writer interface {
	Store(key, value interface{})
}

func unmarshalJSON(p []byte, w writer) error {
	dec := json.NewDecoder(bytes.NewReader(p))
	token, err := dec.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		value := "array"
		if !ok {
			switch token.(type) {
			case bool:
				value = "bool"
			case float64:
				value = "float64"
			case json.Number:
				value = "json.Number"
			case string:
				value = "string"
			case nil:
				value = "null"
			}
		}
		return &json.UnmarshalTypeError{
			Type:  reflect.TypeOf(w),
			Value: value,
		}
	}
	err = decodeObject(dec, w)
	if err != nil {
		return err
	}
	dec.Token() // ignore '}'
	return nil
}

func decodeValue(dec *json.Decoder) (interface{}, error) {
	token, err := dec.Token()
	if err != nil {
		return token, err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return token, nil
	}

	switch delimiter {
	case '[':
		slice := make([]interface{}, 0)
		for dec.More() {
			val, err := decodeValue(dec)
			if err != nil {
				return nil, err
			}
			slice = append(slice, val)
		}
		dec.Token() // ignore ']'
		return slice, nil
	case '{':
		obj := NewLinkedMap() // default
		err = decodeObject(dec, obj)
		if err != nil {
			return nil, err
		}
		dec.Token() // ignore '}'
		return obj, nil
	}
	return nil, fmt.Errorf("unexpected delimiter: %s", delimiter)
}

func decodeObject(dec *json.Decoder, w writer) error {
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return err
		}
		val, err := decodeValue(dec)
		if err != nil {
			return err
		}
		w.Store(key, val)
	}
	return nil
}
