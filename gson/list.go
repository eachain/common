package gson

import (
	"encoding/json"
	"fmt"
)

type L []interface{}

type list struct {
	els []*GSON
	p   *GSON
}

func (l *list) UnmarshalJSON(b []byte) error {
	var raws []json.RawMessage
	err := json.Unmarshal(b, &raws)
	if err != nil {
		return err
	}
	l.els = make([]*GSON, len(raws))
	for i, raw := range raws {
		l.els[i] = &GSON{b: raw, p: l.p}
	}
	return nil
}

func (l *list) MarshalJSON() ([]byte, error) {
	var err error
	raws := make([]json.RawMessage, len(l.els))
	for i, g := range l.els {
		raws[i], err = g.MarshalJSON()
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(raws)
}

func (l *list) Len() int {
	return len(l.els)
}

func (l *list) Index(i int) *GSON {
	if 0 <= i && i < len(l.els) {
		return l.els[i]
	}
	g := &GSON{}
	g.v.u = func() { l.Insert(i, g) }
	return g
}

func (l *list) Insert(i int, g *GSON) {
	g.p = l.p
	gs := make([]*GSON, len(l.els)+1)
	if i <= 0 {
		gs[0] = g
		copy(gs[1:], l.els)
	} else if i >= len(l.els) {
		copy(gs, l.els)
		gs[len(l.els)] = g
	} else {
		copy(gs[:i], l.els)
		gs[i] = g
		copy(gs[i+1:], l.els[i:])
	}
	l.els = gs
}

func (l *list) Remove(c *GSON) {
	for i, g := range l.els {
		if g == c {
			l.els = append(l.els[:i], l.els[i+1:]...)
			break
		}
	}
}

func (l *list) routeOf(c *GSON) string {
	for i, g := range l.els {
		if g == c {
			return fmt.Sprintf("%v[%v]", l.p.p.routeOf(l.p), i)
		}
	}
	return ""
}

func (l *list) orphan() {
	for _, g := range l.els {
		g.p = nil
	}
}
