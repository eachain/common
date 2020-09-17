// gson设计为一个带有缓存的json解析库。
// 返回的JSON对象永远不为nil，可以一直链式调用下去。
// 设计参考标准库reflect，尽量做到简单易懂易用。
package gson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

type GsonType string

const (
	TypObject  GsonType = "object"
	TypList             = "list"
	TypString           = "string"
	TypNumber           = "number"
	TypBool             = "bool"
	TypNull             = "null"
	TypUnknown          = "unknown"
)

// 缓存解析结果，避免重复解析
type cache struct {
	obj map[string]*JSON
	oe  error
	ls  []*JSON
	le  error
	s   string
	i   int64
	b   bool
}

// JSON为可复用、带缓存的json对象
type JSON struct {
	// p为json.RawMessage，延迟解析，等用到的时候再解析
	p []byte

	// 解析结果缓存，如果有缓存，可避免二次解析
	c cache

	// 当前路径
	s string

	// 盲猜类型
	t GsonType

	// 解析错误信息
	e error
}

var ErrEmptyRaw = errors.New("gson: empty raw json message")

type WrappedError struct {
	Path     string
	Expected GsonType
	Err      error
}

func (e *WrappedError) Error() string {
	return fmt.Sprintf("gson: path '%v' expected type: %v, error: %v",
		e.Path, e.Expected, e.Err)
}

type IndexOutOfRangeError struct {
	Path  string
	Index int
	Range int
}

func (e *IndexOutOfRangeError) Error() string {
	var r string
	if e.Range <= 0 {
		r = "empty list"
	} else {
		r = fmt.Sprintf("accept range: [0~%v)", e.Range)
	}
	return fmt.Sprintf("gson: index out of range: %v[%v], %v",
		e.Path, e.Index, r)
}

type KeyNotFoundError struct {
	Path string
	Key  string
	Keys []string
}

func (e *KeyNotFoundError) Error() string {
	var keys string
	if len(e.Keys) == 0 {
		keys = "empty object"
	} else {
		keys = fmt.Sprintf("accept keys: ['%v']", strings.Join(e.Keys, "', '"))
	}
	return fmt.Sprintf("gson: object key not found: '%v.%v', %v",
		e.Path, e.Key, keys)
}

type UnmarshalTypeError struct {
	Path     string
	Expected GsonType
	Real     GsonType
}

func (e *UnmarshalTypeError) Error() string {
	return fmt.Sprintf("gson: type error: '%v', expected: %v, real: %v",
		e.Path, e.Expected, e.Real)
}

func guessType(p []byte) GsonType {
	if len(p) == 0 {
		return TypNull
	}
	t, err := json.NewDecoder(bytes.NewReader(p)).Token()
	if err != nil {
		return TypUnknown
	}
	switch x := t.(type) {
	case json.Delim:
		switch x {
		case '{', '}':
			return TypObject
		case '[', ']':
			return TypList
		}
	case string:
		return TypString
	case float64:
		return TypNumber
	case bool:
		return TypBool
	case nil:
		return TypNull
	}
	return TypUnknown
}

// IsNotExists判断err是不是以下几种情况之一：
// ErrEmptyRaw, 空json串;
// IndexOutOfRangeError, 数组下标越界;
// KeyNotFoundError, 键值未找到。
func IsNotExists(err error) bool {
	if err == nil {
		return false
	}
	if err == ErrEmptyRaw {
		return true
	}
	switch e := err.(type) {
	case *WrappedError:
		return e.Err == ErrEmptyRaw
	case *IndexOutOfRangeError:
		return true
	case *KeyNotFoundError:
		return true
	}
	return false
}

func FromString(s string) *JSON {
	return &JSON{p: []byte(s)}
}

func FromBytes(p []byte) *JSON {
	return &JSON{p: p}
}

func FromReader(r io.Reader) *JSON {
	p, e := ioutil.ReadAll(r)
	return &JSON{p: p, e: e}
}

func (j *JSON) Bytes() []byte {
	return j.p
}

func (j *JSON) Type() GsonType {
	if j.t != "" {
		return j.t
	}
	if len(j.p) == 0 {
		return TypUnknown
	}
	return guessType(j.p)
}

// 由于容易和Str()方法混淆，故删除该方法
// 如果想打印json原文，请用Bytes()方法
/*
func (j *JSON) String() string {
	return string(j.p)
}
*/

func (j *JSON) NotExists() bool {
	return IsNotExists(j.e)
}

func (j *JSON) Err() error {
	return j.e
}

func (j *JSON) setErr(e error, typ GsonType) {
	if e == nil {
		j.e = nil
		return
	}
	_, ok := e.(*json.UnmarshalTypeError)
	if !ok {
		j.e = &WrappedError{
			Path:     j.s,
			Expected: typ,
			Err:      e,
		}
		return
	}
	if j.t == "" {
		j.t = guessType(j.p)
	}
	j.e = &UnmarshalTypeError{
		Path:     j.s,
		Expected: typ,
		Real:     j.t,
	}
}

// Str方法将当前json值以字符串格式返回。
func (j *JSON) Str() string {
	if j.t == TypString {
		return j.c.s
	}
	if len(j.p) == 0 {
		if j.e == nil {
			j.setErr(ErrEmptyRaw, TypString)
		}
		return ""
	}
	if j.p[0] == '"' {
		j.setErr(json.Unmarshal(j.p, &j.c.s), TypString)
		if j.e == nil {
			j.t = TypString
		}
		return j.c.s
	}
	return string(j.p)
}

// Int将当前json值以int64格式返回。
func (j *JSON) Int() int64 {
	if j.t == TypNumber {
		return j.c.i
	}
	if len(j.p) == 0 {
		if j.e == nil {
			j.setErr(ErrEmptyRaw, TypNumber)
		}
		return 0
	}

	var n json.Number
	j.setErr(json.Unmarshal(j.p, &n), TypNumber)
	if j.e != nil {
		return 0
	}
	v, e := n.Int64()
	if e != nil {
		j.setErr(e, TypNumber)
		return 0
	}
	j.c.i = v
	j.t = TypNumber
	return v
}

// TryBool将当前json值以bool格式返回。
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
func (j *JSON) Bool() bool {
	if j.t == TypBool {
		return j.c.b
	}
	if len(j.p) == 0 {
		if j.e != nil {
			j.setErr(ErrEmptyRaw, TypBool)
		}
		return false
	}

	var s string
	if j.p[0] == '"' {
		j.setErr(json.Unmarshal(j.p, &s), TypBool)
		if j.e != nil {
			return false
		}
	} else {
		s = string(j.p)
	}
	b, e := strconv.ParseBool(s)
	if e != nil {
		j.setErr(e, TypBool)
		return false
	}
	j.c.b = b
	j.t = TypBool
	return b
}

// TryStr将当前json值以字符串格式返回。
// 如果出错，将返回error信息。
func (j *JSON) TryStr() (string, error) {
	s := j.Str()
	return s, j.Err()
}

// TryInt将当前json值以int64格式返回。
// 如果出错，将返回error信息。
func (j *JSON) TryInt() (int64, error) {
	v := j.Int()
	return v, j.Err()
}

// TryBool将当前json值以bool格式返回。
// 如果出错，将返回error信息。
func (j *JSON) TryBool() (bool, error) {
	b := j.Bool()
	return b, j.Err()
}

// Len将返回当前json list对象长度。
func (j *JSON) Len() int {
	j.Index(0)
	return len(j.c.ls)
}

func (j *JSON) retListErr(e error) *JSON {
	path := j.s
	if path == "" {
		path = "list"
	}
	r := &JSON{p: j.p, s: path}
	r.setErr(e, TypList)
	r.p = nil // reset to nil
	return r
}

// Index方法从当前json list对象查找第i个对象，
// 如果下标越界或类型错误，将返回带有错误信息的JSON。
func (j *JSON) Index(i int) *JSON {
	if j.t == TypList {
		goto IndexEnd
	}

	if len(j.p) == 0 {
		if j.e != nil {
			return j
		}
		return j.retListErr(ErrEmptyRaw)
	}

	if j.c.le != nil {
		return j.retListErr(j.c.le)
	}

	if j.c.ls == nil {
		var ls []json.RawMessage
		e := json.Unmarshal(j.p, &ls)
		if e != nil {
			j.c.le = e
			return j.retListErr(e)
		}
		if j.s == "" {
			j.s = "list"
		}
		j.c.ls = make([]*JSON, len(ls))
		for k := 0; k < len(ls); k++ {
			path := fmt.Sprintf("%v[%v]", j.s, k)
			j.c.ls[k] = &JSON{s: path, p: ls[k]}
		}
		j.t = TypList
	}

IndexEnd:
	if 0 <= i && i < len(j.c.ls) {
		return j.c.ls[i]
	}
	return &JSON{
		s: fmt.Sprintf("%v[%v]", j.s, i),
		e: &IndexOutOfRangeError{
			Path:  j.s,
			Index: i,
			Range: len(j.c.ls),
		}}
}

func (j *JSON) retObjErr(e error) *JSON {
	path := j.s
	if path == "" {
		path = "object"
	}
	r := &JSON{p: j.p, s: path}
	r.setErr(e, TypObject)
	r.p = nil // reset to nil
	return r
}

// Keys返回object内所有键。
func (j *JSON) Keys() []string {
	x := j.Key("__not_exists_key__")
	e, ok := x.e.(*KeyNotFoundError)
	if ok {
		return e.Keys
	}

	var ks []string
	for k := range j.c.obj {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// Key查找并返回键对应的值。
// 如果未找到或类型错误，将返回带有错误信息的JSON。
func (j *JSON) Key(k string) *JSON {
	if j.t == TypObject {
		goto KeyEnd
	}

	if len(j.p) == 0 {
		if j.e != nil {
			return j
		}
		return j.retObjErr(ErrEmptyRaw)
	}

	if j.c.oe != nil {
		return j.retObjErr(j.c.oe)
	}

	if j.c.obj == nil {
		var m map[string]json.RawMessage
		e := json.Unmarshal(j.p, &m)
		if e != nil {
			j.c.oe = e
			return j.retObjErr(e)
		}

		if j.s == "" {
			j.s = "object"
		}
		j.c.obj = make(map[string]*JSON)
		for key, val := range m {
			j.c.obj[key] = &JSON{s: j.s + "." + key, p: val}
		}
		j.t = TypObject
	}

KeyEnd:
	v := j.c.obj[k]
	if v == nil {
		var ks []string
		for key := range j.c.obj {
			ks = append(ks, key)
		}
		sort.Strings(ks)
		v = &JSON{s: j.s + "." + k, e: &KeyNotFoundError{
			Path: j.s,
			Key:  k,
			Keys: ks,
		}}
	}
	return v
}

// Any可用于err_code, errcode兼容的情况
func (j *JSON) Any(keys ...string) *JSON {
	if len(keys) == 0 {
		return &JSON{s: j.s, e: fmt.Errorf("gson: no keys specified")}
	}
	j.Keys()
	if j.t != TypObject {
		r := &JSON{p: j.p, s: j.s}
		r.setErr(&json.UnmarshalTypeError{}, TypObject)
		r.p = nil
		return r
	}
	for _, k := range keys {
		v := j.c.obj[k]
		if v != nil {
			return v
		}
	}
	return &JSON{s: j.s, e: fmt.Errorf("gson: no any keys found")}
}

// Get支持类型'key.list[0][1].k'式取值。
func (j *JSON) Get(smartKey string) *JSON {
	es, err := parseSmartKey(smartKey)
	if err != nil {
		return &JSON{s: j.s, e: err}
	}
	for _, e := range es {
		j = e(j)
	}
	return j
}

func (j *JSON) UnmarshalJSON(p []byte) error {
	j.p = p
	j.e = nil
	j.c = cache{}
	return nil
}

// Value stores the result in the value pointed to by v.
func (j *JSON) Value(v interface{}) error {
	if j.e != nil {
		return j.e
	}
	if len(j.p) == 0 {
		return ErrEmptyRaw
	}

	return json.Unmarshal(j.p, v)
}

type entry func(*JSON) *JSON

func invalidFormat(i int, s string) error {
	return fmt.Errorf("gson: smart key %vth part invalid format: '%v'", i, s)
}

func parseSmartKey(keys string) ([]entry, error) {
	var entries []entry
	parts := strings.Split(keys, ".")
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		l := strings.IndexByte(part, '[')
		r := strings.IndexByte(part, ']')
		if l < 0 {
			if r > 0 {
				return nil, invalidFormat(i, part)
			}

			entries = append(entries, func(j *JSON) *JSON {
				return j.Key(part)
			})
			continue
		}
		if l == 0 {
			if i > 0 { // "[0]"
				return nil, invalidFormat(i, part)
			}
			if r < 0 {
				return nil, invalidFormat(i, part)
			}
		}
		if l > 0 {
			if (r < 0) || (r > 0 && r < l) {
				return nil, invalidFormat(i, part)
			}

			entries = append(entries, func(j *JSON) *JSON {
				return j.Key(part[:l])
			})
		}

		tmp := part[l:]
		for tmp != "" {
			if tmp[0] != '[' {
				return nil, invalidFormat(i, part)
			}
			r = strings.IndexByte(tmp, ']')
			if r < 0 {
				return nil, invalidFormat(i, part)
			}
			idx, err := strconv.Atoi(tmp[1:r])
			if err != nil {
				return nil, fmt.Errorf("gson: smart key %vth invalid index: '%v'",
					i, part)
			}
			entries = append(entries, func(j *JSON) *JSON {
				return j.Index(idx)
			})

			tmp = tmp[r+1:]
		}
	}
	return entries, nil
}
