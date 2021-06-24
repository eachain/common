package gson

import (
	"fmt"
	"strconv"
	"strings"
)

// Any可用于err_code, errcode兼容的情况
func (g *GSON) Any(keys ...string) *GSON {
	if len(keys) == 0 {
		return &GSON{e: KeyNotFoundErr{Path: g.Path()}}
	}
	g.objInit()
	if g.v.o == nil {
		return &GSON{e: TypOpErr{Op: "any", Typ: g.Type(), Path: g.Path()}}
	}

	ks := g.v.o.Keys()
	for _, key := range keys {
		for _, k := range ks {
			if k == key {
				return g.v.o.Index(k)
			}
		}
	}
	return &GSON{e: KeyNotFoundErr{Key: strings.Join(keys, "|"), Path: g.Path()}}
}

// Get支持类型'key.list[0][1].k'式取值。
func (g *GSON) Get(smartKey string) *GSON {
	es, err := parseSmartKey(smartKey)
	if err != nil {
		return &GSON{e: err}
	}
	for _, e := range es {
		g = e(g)
	}
	return g
}

type entry func(*GSON) *GSON

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

			entries = append(entries, func(j *GSON) *GSON {
				return j.ObjIdx(part)
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

			entries = append(entries, func(j *GSON) *GSON {
				return j.ObjIdx(part[:l])
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
			entries = append(entries, func(j *GSON) *GSON {
				return j.Index(idx)
			})

			tmp = tmp[r+1:]
		}
	}
	return entries, nil
}
