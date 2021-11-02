package gson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"unsafe"
)

type Type string

const (
	TypUnknown Type = "unknown"
	TypString  Type = "string"
	TypNumber  Type = "number"
	TypObject  Type = "object"
	TypList    Type = "list"
	TypBool    Type = "bool"
	TypNull    Type = "null"
)

type GSON struct {
	b []byte // json raw message
	p *GSON  // parents
	t Type
	v value
	u bool  // updated
	e error // if any error
}

type value struct {
	s string
	i int64
	f float64
	o *object
	l *list
	b bool
	u func() // on update

	set   bool
	isInt bool
}

func FromBytes(b []byte) *GSON {
	g := &GSON{}
	g.reset(b)
	return g
}

func FromString(s string) *GSON {
	g := &GSON{}
	g.reset([]byte(s))
	return g
}

type TypOpErr struct {
	Op   string
	Typ  Type
	Path string
}

func (to TypOpErr) Error() string {
	return fmt.Sprintf("gson: %v on '%v', path: %v", to.Op, to.Typ, to.Path)
}

type KeyNotFoundErr struct {
	Key  string
	Path string
}

func (kn KeyNotFoundErr) Error() string {
	return fmt.Sprintf("gson: key '%v' not found, path: %v", kn.Key, kn.Path)
}

func (g *GSON) Err() error {
	return g.e
}

func (g *GSON) Type() Type {
	if g.t == "" || g.t == TypUnknown {
		g.t = guess(g.b)
	}
	return g.t
}

func guess(p []byte) Type {
	if len(p) == 0 {
		return TypUnknown
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

func (g *GSON) Str() string {
	if g.Type() == TypString {
		if !g.v.set {
			e := json.Unmarshal(g.b, &g.v.s)
			if e == nil {
				g.v.set = true
			}
		}
		return g.v.s
	}

	if g.u {
		g.MarshalJSON()
	}

	return *(*string)(unsafe.Pointer(&g.b))
}

var isInt = regexp.MustCompile("^0$|^[1-9][0-9]*$|^-[1-9][0-9]*$")

func (g *GSON) IsInt() bool {
	switch g.Type() {
	default:
		return false
	case TypNumber, TypString:
	}

	if g.v.set {
		return g.v.isInt
	}

	s := g.Str()

	if !isInt.MatchString(s) {
		f, e := strconv.ParseFloat(s, 64)
		if e == nil {
			g.v.f = f
			g.v.set = true
		}
		return false
	}

	i, e := strconv.ParseInt(s, 10, 64)
	if e != nil {
		return false
	}
	g.v.i = i
	g.v.isInt = true
	g.v.set = true
	return true
}

func (g *GSON) Int() int64 {
	if g.IsInt() {
		return g.v.i
	}
	return int64(g.v.f)
}

func (g *GSON) Float() float64 {
	if g.IsInt() {
		return float64(g.v.i)
	}
	return g.v.f
}

// object keys
func (g *GSON) objInit() {
	if g.v.o == nil {
		var b []byte
		if g.Type() == TypObject {
			b = g.b
		} else if g.Type() == TypString {
			b = []byte(g.Str())
		} else {
			return
		}

		o := object{p: g}
		e := json.Unmarshal(b, &o)
		if e == nil {
			g.v.o = &o
		}
	}
}

func (g *GSON) Keys() []string {
	g.objInit()
	if g.v.o == nil {
		return nil
	}
	return g.v.o.Keys()
}

// get object value of the key
func (g *GSON) ObjIdx(key string) *GSON {
	if g.e != nil {
		return g
	}

	g.objInit()
	if g.v.o == nil {
		if g.Type() != TypUnknown || len(g.b) > 0 {
			return &GSON{e: TypOpErr{Op: "objidx", Typ: g.Type(), Path: g.Path()}}
		}

		g.v.o = &object{p: g}
		g.t = TypObject
	}
	return g.v.o.Index(key)
}

func (g *GSON) listInit() {
	if g.v.l == nil {
		var b []byte
		if g.Type() == TypList {
			b = g.b
		} else if g.Type() == TypString {
			b = []byte(g.Str())
		} else {
			return
		}

		l := list{p: g}
		e := json.Unmarshal(b, &l)
		if e == nil {
			g.v.l = &l
		}
	}
}

// list length
func (g *GSON) Len() int {
	g.listInit()
	if g.v.l == nil {
		return 0
	}
	return g.v.l.Len()
}

// list index
func (g *GSON) Index(i int) *GSON {
	if g.e != nil {
		return g
	}

	g.listInit()
	if g.v.l == nil {
		if g.Type() != TypUnknown || len(g.b) > 0 {
			return &GSON{e: TypOpErr{Op: "index", Typ: g.Type(), Path: g.Path()}}
		}
		g.v.l = &list{p: g}
		g.t = TypList
	}
	return g.v.l.Index(i)
}

// bool value
func (g *GSON) Bool() bool {
	if g.Type() == TypBool {
		if g.v.set {
			return g.v.b
		}
		b, e := strconv.ParseBool(string(g.b))
		if e == nil {
			g.v.b = b
			g.v.set = true
		}
		return g.v.b
	}

	if g.Type() == TypNumber {
		if g.IsInt() {
			return g.Int() != 0
		} else {
			return g.Float() != 0
		}
	}

	if g.Type() == TypString {
		s := g.Str()
		b, _ := strconv.ParseBool(s)
		return b
	}
	return false
}

// returns is null or not
func (g *GSON) IsNull() bool {
	return g.Type() == TypNull
}

func (g *GSON) update(set bool) {
	if g == nil {
		return
	}
	if set && g.v.u != nil {
		g.v.u()
		g.v.u = nil
	}
	g.u = true
	g.p.update(set)
}

func (g *GSON) reset(b []byte) {
	if len(b) > 0 {
		g.b = make([]byte, len(b))
		copy(g.b, b)
	} else {
		g.b = nil
	}
	g.t = ""
	if g.v.u != nil {
		g.v.u()
	}
	if g.v.o != nil {
		g.v.o.orphan()
	}
	if g.v.l != nil {
		g.v.l.orphan()
	}
	g.v = value{}
	g.e = nil
	g.p.update(true)
}

func (g *GSON) Set(v interface{}) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	g.reset(b)
	return nil
}

func (g *GSON) Remove() {
	if g.p == nil {
		return
	}

	if g.p.v.o != nil {
		g.p.v.o.Remove(g)
	}
	if g.p.v.l != nil {
		g.p.v.l.Remove(g)
	}
	g.p.update(false)
	g.p = nil
}

// usage: g.Index(100).Exists()
func (g *GSON) Exists() bool {
	return len(g.b) == 0
}

func (g *GSON) MarshalJSON() ([]byte, error) {
	if !g.u {
		b := make([]byte, len(g.b))
		copy(b, g.b)
		return b, nil
	}
	g.u = false

	var e error

	switch g.Type() {
	default:
		// case "", TypUnknown, TypBool, TypNull, TypNumber:
		return g.b, nil

	case TypObject:
		g.objInit()
		if g.v.o == nil {
			return []byte{'{', '}'}, nil
		}
		g.b, e = g.v.o.MarshalJSON()
		return g.b, e

	case TypList:
		g.listInit()
		if g.v.l == nil {
			return []byte{'[', ']'}, nil
		}
		g.b, e = g.v.l.MarshalJSON()
		return g.b, e

	case TypString:
		if g.v.o != nil {
			b, e := g.v.o.MarshalJSON()
			if e != nil {
				return nil, e
			}
			g.b, e = json.Marshal(string(b))
			return g.b, e
		}
		if g.v.l != nil {
			b, e := g.v.l.MarshalJSON()
			if e != nil {
				return nil, e
			}
			g.b, e = json.Marshal(string(b))
			return g.b, e
		}
		return g.b, nil
	}
}

func (g *GSON) UnmarshalJSON(b []byte) error {
	g.reset(b)
	return nil
}

func (g *GSON) routeOf(c *GSON) string {
	if g == nil {
		return ""
	}
	switch g.Type() {
	case TypObject:
		g.objInit()
		if g.v.o != nil {
			return g.v.o.routeOf(c)
		}
	case TypList:
		g.listInit()
		if g.v.l != nil {
			return g.v.l.routeOf(c)
		}
	case TypString:
		if g.v.o != nil {
			return g.v.o.routeOf(c)
		}
		if g.v.l != nil {
			return g.v.l.routeOf(c)
		}
	}
	return ""
}

func (g *GSON) Path() string {
	return g.p.routeOf(g)
}
