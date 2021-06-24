package gson

import (
	"bytes"
	"encoding/json"
	"errors"
)

type M map[string]interface{}

type object struct {
	ks []string
	mp map[string]*GSON
	p  *GSON
}

func (o *object) UnmarshalJSON(p []byte) error {
	dec := json.NewDecoder(bytes.NewReader(p))
	t, e := dec.Token()
	if e != nil {
		return e
	}
	if d, ok := t.(json.Delim); !ok || d != '{' {
		return errors.New("gson: object expect '{'")
	}
	for dec.More() {
		t, e := dec.Token()
		if e != nil {
			return e
		}
		k, ok := t.(string)
		if !ok {
			return errors.New("gson: object expect key type of string")
		}
		var b json.RawMessage
		e = dec.Decode(&b)
		if e != nil {
			return e
		}
		o.ks = append(o.ks, k)
		if o.mp == nil {
			o.mp = make(map[string]*GSON)
		}
		o.mp[k] = &GSON{b: b, p: o.p}
	}
	t, e = dec.Token() // ignore '}'
	if e != nil {
		return e
	}
	if d, ok := t.(json.Delim); !ok || d != '}' {
		return errors.New("gson: object expect '}'")
	}
	return nil
}

func (o *object) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	buf.WriteByte('{')
	for i, k := range o.ks {
		if i > 0 {
			buf.WriteByte(',')
		}
		err := enc.Encode(k)
		if err != nil {
			return nil, err
		}
		if buf.Bytes()[buf.Len()-1] == '\n' {
			buf.Truncate(buf.Len() - 1)
		}
		buf.WriteByte(':')
		err = enc.Encode(o.mp[k])
		if err != nil {
			return nil, err
		}
		if buf.Bytes()[buf.Len()-1] == '\n' {
			buf.Truncate(buf.Len() - 1)
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o *object) Keys() []string {
	ks := make([]string, len(o.ks))
	copy(ks, o.ks)
	return ks
}

func (o *object) Index(k string) *GSON {
	g := o.mp[k]
	if g == nil {
		g = &GSON{}
		g.v.u = func() {
			if o.mp == nil {
				o.mp = make(map[string]*GSON)
			}
			if _, exists := o.mp[k]; !exists {
				o.ks = append(o.ks, k)
			}
			g.p = o.p
			o.mp[k] = g
		}
	}
	return g
}

func (o *object) Remove(c *GSON) {
	var key string
	var found bool
	for k, g := range o.mp {
		if g == c {
			key = k
			found = true
			break
		}
	}

	if !found {
		return
	}

	delete(o.mp, key)
	for i := 0; i < len(o.ks); i++ {
		if o.ks[i] == key {
			o.ks = append(o.ks[:i], o.ks[i+1:]...)
			break
		}
	}
}

func (o *object) routeOf(c *GSON) string {
	for k, g := range o.mp {
		if g == c {
			r := o.p.p.routeOf(o.p)
			if r == "" {
				return k
			}
			return r + "." + k
		}
	}
	return ""
}

func (o *object) orphan() {
	for _, g := range o.mp {
		g.p = nil
	}
}
